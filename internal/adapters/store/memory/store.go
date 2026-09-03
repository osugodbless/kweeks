// Package memory provides an in-memory Store implementation. It is the test
// double AND the demo fallback: no Postgres required to run the show.
package memory

import (
	"context"
	"sync"

	"github.com/osugodbless/kweeks/internal/domain"
)

// Store is a thread-safe in-memory implementation of ports.Store.
type Store struct {
	mu           sync.RWMutex
	quizzes      map[string]*domain.Quiz
	rooms        map[string]*domain.Room
	participants map[string][]*domain.Participant // key roomID
	answers      map[string][]*domain.Answer      // key roomID
	claims       map[string][]*domain.Claim       // key quizID
}

func New() *Store {
	return &Store{
		quizzes:      map[string]*domain.Quiz{},
		rooms:        map[string]*domain.Room{},
		participants: map[string][]*domain.Participant{},
		answers:      map[string][]*domain.Answer{},
		claims:       map[string][]*domain.Claim{},
	}
}

func (s *Store) CreateQuiz(ctx context.Context, q *domain.Quiz) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.quizzes[q.ID] = cloneQuiz(q)
	return nil
}

func (s *Store) GetQuiz(ctx context.Context, id string) (*domain.Quiz, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	q, ok := s.quizzes[id]
	if !ok {
		return nil, domain.ErrQuizNotFound
	}
	return cloneQuiz(q), nil
}

func (s *Store) ListQuizzes(ctx context.Context, instructorID string) ([]domain.Quiz, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Quiz
	for _, q := range s.quizzes {
		if q.InstructorID == instructorID {
			out = append(out, *cloneQuiz(q))
		}
	}
	return out, nil
}

func (s *Store) CreateRoom(ctx context.Context, r *domain.Room) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *r
	s.rooms[r.ID] = &cp
	return nil
}

func (s *Store) GetRoom(ctx context.Context, id string) (*domain.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rooms[id]
	if !ok {
		return nil, domain.ErrRoomNotFound
	}
	cp := *r
	return &cp, nil
}

func (s *Store) SaveRoom(ctx context.Context, r *domain.Room) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rooms[r.ID]; !ok {
		return domain.ErrRoomNotFound
	}
	cp := *r
	s.rooms[r.ID] = &cp
	return nil
}

func (s *Store) ListLiveRooms(ctx context.Context) ([]domain.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Room
	for _, r := range s.rooms {
		if r.State == domain.RoomLive {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (s *Store) JoinParticipant(ctx context.Context, p *domain.Participant) (*domain.Participant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.participants[p.RoomID] {
		if existing.Email == p.Email {
			cp := *existing
			return &cp, domain.ErrDuplicateParticipant
		}
	}
	cp := *p
	s.participants[p.RoomID] = append(s.participants[p.RoomID], &cp)
	return &cp, nil
}

func (s *Store) GetParticipant(ctx context.Context, roomID, email string) (*domain.Participant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.participants[roomID] {
		if p.Email == email {
			cp := *p
			return &cp, nil
		}
	}
	return nil, domain.ErrParticipantNotFound
}

func (s *Store) ListParticipants(ctx context.Context, roomID string) ([]domain.Participant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ps := s.participants[roomID]
	out := make([]domain.Participant, 0, len(ps))
	for _, p := range ps {
		out = append(out, *p)
	}
	return out, nil
}

func (s *Store) RecordAnswer(ctx context.Context, a *domain.Answer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.answers[a.RoomID] = append(s.answers[a.RoomID], cloneAnswer(a))
	return nil
}

func (s *Store) HasAnswered(ctx context.Context, roomID, participantID, questionID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.answers[roomID] {
		if a.ParticipantID == participantID && a.QuestionID == questionID {
			return true, nil
		}
	}
	return false, nil
}

func (s *Store) ListAnswers(ctx context.Context, roomID string) ([]domain.Answer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	as := s.answers[roomID]
	out := make([]domain.Answer, 0, len(as))
	for _, a := range as {
		out = append(out, *cloneAnswer(a))
	}
	return out, nil
}

func (s *Store) CreateClaim(ctx context.Context, c *domain.Claim) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.claims[c.QuizID] {
		if existing.Email == c.Email {
			return domain.ErrClaimExists
		}
	}
	cp := *c
	s.claims[c.QuizID] = append(s.claims[c.QuizID], &cp)
	return nil
}

func (s *Store) GetClaimByCode(ctx context.Context, quizID, code string) (*domain.Claim, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.claims[quizID] {
		if c.ClaimCode == code {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrClaimNotFound
}

func (s *Store) GetClaimByEmail(ctx context.Context, quizID, email string) (*domain.Claim, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.claims[quizID] {
		if c.Email == email {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrClaimNotFound
}

func (s *Store) ListClaims(ctx context.Context, quizID string) ([]domain.Claim, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cs := s.claims[quizID]
	out := make([]domain.Claim, 0, len(cs))
	for _, c := range cs {
		out = append(out, *c)
	}
	return out, nil
}

func (s *Store) UpdateClaimState(ctx context.Context, id string, to domain.ClaimState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, cs := range s.claims {
		for _, c := range cs {
			if c.ID == id {
				if !domain.CanTransition(c.State, to) {
					return domain.ErrInvalidTransition
				}
				c.State = to
				return nil
			}
		}
	}
	return domain.ErrClaimNotFound
}

func (s *Store) Close() error { return nil }

func cloneQuiz(q *domain.Quiz) *domain.Quiz {
	if q == nil {
		return nil
	}
	cp := *q
	cp.Questions = append([]domain.Question(nil), q.Questions...)
	return &cp
}

func cloneAnswer(a *domain.Answer) *domain.Answer {
	if a == nil {
		return nil
	}
	cp := *a
	return &cp
}
