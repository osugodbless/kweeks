// Package app contains the application services that orchestrate domain rules
// against ports. It has no knowledge of HTTP, websockets, Postgres, or BMONI.
package app

import (
	"context"
	"errors"
	"time"

	"github.com/osugodbless/kweeks/internal/domain"
	"github.com/osugodbless/kweeks/internal/ports"
)

// Game is the room lifecycle service: lobby, start, pacing advance, answer
// acceptance, standings, and podium.
type Game struct {
	store ports.Store
	clock ports.Clock
	bus   ports.Bus
	now   func() time.Time
}

// NewGame wires the game service. clock and bus may be nil for unit tests that
// only exercise store logic; the service is used by callers who supply both.
func NewGame(store ports.Store, clock ports.Clock, bus ports.Bus) *Game {
	return &Game{store: store, clock: clock, bus: bus}
}

// nowTime returns the current time from the clock when present.
func (g *Game) nowTime() time.Time {
	if g.clock != nil {
		return g.clock.Now()
	}
	return time.Now()
}

// CreateQuiz validates and persists a fully-authored quiz.
func (g *Game) CreateQuiz(ctx context.Context, q *domain.Quiz) error {
	if err := domain.ValidateQuiz(q); err != nil {
		return err
	}
	return g.store.CreateQuiz(ctx, q)
}

// UpdateQuiz validates and persists edits to an existing quiz (builder save).
func (g *Game) UpdateQuiz(ctx context.Context, q *domain.Quiz) error {
	if err := domain.ValidateQuiz(q); err != nil {
		return err
	}
	existing, err := g.store.GetQuiz(ctx, q.ID)
	if err != nil {
		return err
	}
	// Only the owning instructor may edit; keep original ownership + createdAt.
	q.InstructorID = existing.InstructorID
	q.CreatedAt = existing.CreatedAt
	return g.store.UpdateQuiz(ctx, q)
}

// OpenRoom creates a room in the lobby for an existing quiz. A short human
// code is generated (and kept unique against existing rooms) for players to
// join by typing it.
func (g *Game) OpenRoom(ctx context.Context, r *domain.Room) error {
	if r == nil || r.ID == "" || r.QuizID == "" {
		return domain.ErrRoomNotFound
	}
	quiz, err := g.store.GetQuiz(ctx, r.QuizID)
	if err != nil {
		return err
	}
	if quiz == nil {
		return domain.ErrQuizNotFound
	}
	r.Pacing = quiz.Pacing
	r.State = domain.RoomLobby
	r.CurrentQuestionIdx = -1
	if r.Code == "" {
		code, err := g.uniqueRoomCode(ctx)
		if err != nil {
			return err
		}
		r.Code = code
	}
	return g.store.CreateRoom(ctx, r)
}

func (g *Game) uniqueRoomCode(ctx context.Context) (string, error) {
	for attempt := 0; attempt < 8; attempt++ {
		code, err := domain.GenerateRoomCode()
		if err != nil {
			return "", err
		}
		if _, err := g.store.GetRoomByCode(ctx, code); err != nil {
			if errors.Is(err, domain.ErrRoomNotFound) {
				return code, nil
			}
			return "", err
		}
	}
	return "", errors.New("could not allocate a unique room code")
}

// Start begins the first question. Pacing is irrelevant to starting; manual
// and auto rooms both begin on this call.
func (g *Game) Start(ctx context.Context, roomID string) error {
	room, err := g.store.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room.State != domain.RoomLobby {
		return domain.ErrRoomWrongState
	}
	room.State = domain.RoomLive
	room.CurrentQuestionIdx = 0
	room.QuestionStartedAt = g.nowTime()
	room.StartedAt = room.QuestionStartedAt
	if err := g.store.SaveRoom(ctx, room); err != nil {
		return err
	}
	return g.push(ctx, room.ID, "question", room)
}

// Next advances to the next question. In MANUAL rooms this is the instructor's
// next tap. In AUTO rooms the scheduler calls it when the question window
// elapses. Returns ErrRoomWrongState if already past the last question.
func (g *Game) Next(ctx context.Context, roomID string) error {
	room, err := g.store.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room.State != domain.RoomLive {
		return domain.ErrRoomWrongState
	}
	quiz, err := g.store.GetQuiz(ctx, room.QuizID)
	if err != nil {
		return err
	}
	next := room.CurrentQuestionIdx + 1
	if next >= len(quiz.Questions) {
		// No questions remain: the podium is reachable after the last answer
		// window. We leave the room live; FinalizePodium is called by the
		// scheduler/instructor.
		return domain.ErrRoomWrongState
	}
	room.CurrentQuestionIdx = next
	room.QuestionStartedAt = g.nowTime()
	if err := g.store.SaveRoom(ctx, room); err != nil {
		return err
	}
	return g.push(ctx, room.ID, "question", room)
}

// SubmitAnswer accepts a player answer. The server clock gates the cutoff:
// an answer arriving after started_at+duration is rejected. Answers are
// accepted exactly once per participant per question.
func (g *Game) SubmitAnswer(ctx context.Context, a *domain.Answer) (*domain.AnswerReceipt, error) {
	room, err := g.store.GetRoom(ctx, a.RoomID)
	if err != nil {
		return nil, err
	}
	if room.State != domain.RoomLive {
		return nil, domain.ErrRoomNotLive
	}
	quiz, err := g.store.GetQuiz(ctx, room.QuizID)
	if err != nil {
		return nil, err
	}
	q := quiz.CurrentQuestion(room.CurrentQuestionIdx)
	if q.ID != a.QuestionID {
		return nil, domain.ErrAnswerUnknownQuestion
	}
	if !domain.AnswerGate(g.nowTime(), room.QuestionStartedAt, quiz.QuestionDuration(q)) {
		return nil, domain.ErrAnswerLate
	}
	if a.OptionIndex < 0 || a.OptionIndex >= len(q.Options) {
		return nil, domain.ErrInvalidOptionIndex
	}
	answered, err := g.store.HasAnswered(ctx, a.RoomID, a.ParticipantID, a.QuestionID)
	if err != nil {
		return nil, err
	}
	if answered {
		return nil, domain.ErrAlreadyAnswered
	}
	a.QuestionStartedAt = room.QuestionStartedAt
	a.ReceivedAt = g.nowTime()
	a.Correct = a.OptionIndex == q.CorrectIndex
	if err := g.store.RecordAnswer(ctx, a); err != nil {
		return nil, err
	}
	receipt := &domain.AnswerReceipt{
		ID:        a.ID,
		Correct:   a.Correct,
		Score:     domain.ScoreFor(a.Correct),
		LatencyMs: domain.LatencyMs(a.ReceivedAt, a.QuestionStartedAt),
	}
	return receipt, nil
}

// Standings computes the live leaderboard for a room. Correct answers only;
// latency is summed across correct answers for the tie-break.
func (g *Game) Standings(ctx context.Context, roomID string) ([]domain.Standing, error) {
	participants, err := g.store.ListParticipants(ctx, roomID)
	if err != nil {
		return nil, err
	}
	answers, err := g.store.ListAnswers(ctx, roomID)
	if err != nil {
		return nil, err
	}
	byPID := map[string]*domain.Standing{}
	for _, p := range participants {
		byPID[p.ID] = &domain.Standing{
			ParticipantID: p.ID,
			Nickname:      p.Nickname,
			Avatar:        p.Avatar,
			JoinedAt:      p.JoinedAt,
		}
	}
	for _, a := range answers {
		s := byPID[a.ParticipantID]
		if s == nil || !a.Correct {
			continue
		}
		s.CorrectCount++
		s.TotalLatency += time.Duration(domain.LatencyMs(a.ReceivedAt, a.QuestionStartedAt)) * time.Millisecond
	}
	standings := make([]domain.Standing, 0, len(byPID))
	for _, s := range byPID {
		standings = append(standings, *s)
	}
	return domain.SortStandings(standings), nil
}

// FinalizePodium closes the room and declares the winners once the last
// question window has ended. Standings are read from the store; the top N are
// the winners.
func (g *Game) FinalizePodium(ctx context.Context, roomID string) ([]domain.Standing, error) {
	room, err := g.store.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room.State == domain.RoomPodium || room.State == domain.RoomEnded {
		return nil, domain.ErrRoomWrongState
	}
	quiz, err := g.store.GetQuiz(ctx, room.QuizID)
	if err != nil {
		return nil, err
	}
	standings, err := g.Standings(ctx, roomID)
	if err != nil {
		return nil, err
	}
	winners := domain.SelectWinners(standings, quiz.WinnerCount)
	if len(winners) == 0 {
		return nil, domain.ErrNoWinners
	}
	room.State = domain.RoomPodium
	if err := g.store.SaveRoom(ctx, room); err != nil {
		return nil, err
	}
	if err := g.push(ctx, room.ID, "podium", winners); err != nil {
		return nil, err
	}
	return winners, nil
}

func (g *Game) push(ctx context.Context, roomID, typ string, payload any) error {
	if g.bus == nil {
		return nil
	}
	return g.bus.Publish(ctx, ports.Event{RoomID: roomID, Type: typ, Payload: payload})
}

// EndRoom transitions a live room to the terminal ended state. It is used when
// a room finishes its questions but no eligible winner answered correctly, so
// the room still resolves rather than lingering live forever. Idempotent for
// rooms already ended.
func (g *Game) EndRoom(ctx context.Context, roomID string) error {
	room, err := g.store.GetRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room.State == domain.RoomEnded {
		return nil
	}
	if room.State != domain.RoomLive {
		return domain.ErrRoomWrongState
	}
	room.State = domain.RoomEnded
	if err := g.store.SaveRoom(ctx, room); err != nil {
		return err
	}
	return g.push(ctx, room.ID, "ended", room)
}

// ErrDuplicateEmail wraps the store error for join semantics.
var ErrDuplicateEmail = errors.New("duplicate email join")

// GetQuiz returns a quiz by id.
func (g *Game) GetQuiz(ctx context.Context, id string) (*domain.Quiz, error) {
	return g.store.GetQuiz(ctx, id)
}

// GetRoom returns a room by id.
func (g *Game) GetRoom(ctx context.Context, id string) (*domain.Room, error) {
	return g.store.GetRoom(ctx, id)
}

// ListQuizzes lists quizzes authored by an instructor.
func (g *Game) ListQuizzes(ctx context.Context, instructorID string) ([]domain.Quiz, error) {
	return g.store.ListQuizzes(ctx, instructorID)
}

// RoomParticipants lists participants of a room (public fields only).
func (g *Game) RoomParticipants(ctx context.Context, roomID string) ([]domain.Participant, error) {
	return g.store.ListParticipants(ctx, roomID)
}

// GetRoomByCode resolves a room by its short join code.
func (g *Game) GetRoomByCode(ctx context.Context, code string) (*domain.Room, error) {
	return g.store.GetRoomByCode(ctx, code)
}

// LatestLiveRoom returns the active (lobby/live) room for a quiz, if any.
func (g *Game) LatestLiveRoom(ctx context.Context, quizID string) (*domain.Room, error) {
	return g.store.LatestLiveRoom(ctx, quizID)
}

// Winners recomputes the podium winners for a room from the live standings.
// Deterministic and safe to call after FinalizePodium.
func (g *Game) Winners(ctx context.Context, roomID string) ([]domain.Standing, error) {
	room, err := g.store.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	quiz, err := g.store.GetQuiz(ctx, room.QuizID)
	if err != nil {
		return nil, err
	}
	standings, err := g.Standings(ctx, roomID)
	if err != nil {
		return nil, err
	}
	return domain.SelectWinners(standings, quiz.WinnerCount), nil
}
