// Package postgres implements ports.Store against plain Postgres via pgx.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/osugodbless/kweeks/internal/domain"
)

// Store is the pgx-backed ports.Store.
type Store struct {
	pool *pgxpool.Pool
}

// Open connects a pgx pool from a DSN and returns a Store.
func Open(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Store{pool: pool}, nil
}

// OpenMigrated opens the store and applies pending migrations first, so the
// app can start against an empty database without a separate migration step.
func OpenMigrated(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	if err := Migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate postgres: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() error {
	s.pool.Close()
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isFKViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

// --- Quizzes ---

type questionRow struct {
	ID           string   `json:"id"`
	Prompt       string   `json:"prompt"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correct_index"`
	DurationMs   int64    `json:"duration_ms"`
}

func (s *Store) CreateQuiz(ctx context.Context, q *domain.Quiz) error {
	rows := make([]questionRow, 0, len(q.Questions))
	for _, question := range q.Questions {
		rows = append(rows, questionRow{
			ID: question.ID, Prompt: question.Prompt, Options: question.Options,
			CorrectIndex: question.CorrectIndex, DurationMs: question.Duration.Milliseconds(),
		})
	}
	b, err := json.Marshal(rows)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		insert into quizzes (id, instructor_id, title, pool_kobo, winner_count, pacing, default_duration_ms, questions)
		values ($1,$2,$3,$4,$5,$6,$7,$8)`,
		q.ID, q.InstructorID, q.Title, int64(q.Pool), q.WinnerCount, string(q.Pacing),
		q.DefaultDuration.Milliseconds(), b)
	return err
}

func scanQuiz(row pgx.Row) (*domain.Quiz, error) {
	var q domain.Quiz
	var pool int64
	var defaultMs int64
	var pacing string
	var questionsJSON []byte
	if err := row.Scan(&q.ID, &q.InstructorID, &q.Title, &pool, &q.WinnerCount, &pacing, &defaultMs, &questionsJSON, &q.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrQuizNotFound
		}
		return nil, err
	}
	q.Pool = domain.Amount(pool)
	q.Pacing = domain.PacingMode(pacing)
	q.DefaultDuration = time.Duration(defaultMs) * time.Millisecond
	var rows []questionRow
	if err := json.Unmarshal(questionsJSON, &rows); err != nil {
		return nil, err
	}
	for _, r := range rows {
		q.Questions = append(q.Questions, domain.Question{
			ID: r.ID, Prompt: r.Prompt, Options: r.Options, CorrectIndex: r.CorrectIndex,
			Duration: time.Duration(r.DurationMs) * time.Millisecond,
		})
	}
	return &q, nil
}

func (s *Store) GetQuiz(ctx context.Context, id string) (*domain.Quiz, error) {
	return scanQuiz(s.pool.QueryRow(ctx, `
		select id, instructor_id, title, pool_kobo, winner_count, pacing, default_duration_ms, questions, created_at
		from quizzes where id = $1`, id))
}

func (s *Store) ListQuizzes(ctx context.Context, instructorID string) ([]domain.Quiz, error) {
	rows, err := s.pool.Query(ctx, `
		select id, instructor_id, title, pool_kobo, winner_count, pacing, default_duration_ms, questions, created_at
		from quizzes where instructor_id = $1 order by created_at desc`, instructorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Quiz
	for rows.Next() {
		q, err := scanQuiz(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *q)
	}
	return out, rows.Err()
}

// --- Rooms ---

func (s *Store) CreateRoom(ctx context.Context, r *domain.Room) error {
	_, err := s.pool.Exec(ctx, `
		insert into rooms (id, quiz_id, state, host_id, current_question_idx, question_started_at, started_at)
		values ($1,$2,$3,$4,$5,$6,$7)`,
		r.ID, r.QuizID, string(r.State), r.HostID, r.CurrentQuestionIdx, r.QuestionStartedAt, r.StartedAt)
	if isFKViolation(err) {
		return domain.ErrQuizNotFound
	}
	return err
}

func (s *Store) GetRoom(ctx context.Context, id string) (*domain.Room, error) {
	var r domain.Room
	var state string
	var startedAt *time.Time
	err := s.pool.QueryRow(ctx, `
		select id, quiz_id, state, host_id, current_question_idx, question_started_at, started_at
		from rooms where id = $1`, id).
		Scan(&r.ID, &r.QuizID, &state, &r.HostID, &r.CurrentQuestionIdx, &r.QuestionStartedAt, &startedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrRoomNotFound
	}
	if err != nil {
		return nil, err
	}
	r.State = domain.RoomState(state)
	if startedAt != nil {
		r.StartedAt = *startedAt
	}
	return &r, nil
}

func (s *Store) SaveRoom(ctx context.Context, r *domain.Room) error {
	tag, err := s.pool.Exec(ctx, `
		update rooms set state=$2, current_question_idx=$3, question_started_at=$4, started_at=$5
		where id=$1`, r.ID, string(r.State), r.CurrentQuestionIdx, r.QuestionStartedAt, r.StartedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrRoomNotFound
	}
	return nil
}

func (s *Store) ListLiveRooms(ctx context.Context) ([]domain.Room, error) {
	rows, err := s.pool.Query(ctx, `
		select id, quiz_id, state, host_id, current_question_idx, question_started_at, started_at
		from rooms where state='live'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Room
	for rows.Next() {
		var r domain.Room
		var startedAt *time.Time
		if err := rows.Scan(&r.ID, &r.QuizID, &r.State, &r.HostID, &r.CurrentQuestionIdx, &r.QuestionStartedAt, &startedAt); err != nil {
			return nil, err
		}
		if startedAt != nil {
			r.StartedAt = *startedAt
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// --- Participants ---

func (s *Store) JoinParticipant(ctx context.Context, p *domain.Participant) (*domain.Participant, error) {
	_, err := s.pool.Exec(ctx, `
		insert into participants (id, room_id, email, nickname, avatar, joined_at)
		values ($1,$2,$3,$4,$5,$6)`,
		p.ID, p.RoomID, p.Email, p.Nickname, p.Avatar, p.JoinedAt)
	if isUniqueViolation(err) {
		return nil, domain.ErrDuplicateParticipant
	}
	if isFKViolation(err) {
		return nil, domain.ErrRoomNotFound
	}
	return p, err
}

func scanParticipant(row pgx.Row) (*domain.Participant, error) {
	var p domain.Participant
	if err := row.Scan(&p.ID, &p.RoomID, &p.Email, &p.Nickname, &p.Avatar, &p.JoinedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrParticipantNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (s *Store) GetParticipant(ctx context.Context, roomID, email string) (*domain.Participant, error) {
	return scanParticipant(s.pool.QueryRow(ctx, `
		select id, room_id, email, nickname, avatar, joined_at
		from participants where room_id=$1 and email=$2`, roomID, email))
}

func (s *Store) ListParticipants(ctx context.Context, roomID string) ([]domain.Participant, error) {
	rows, err := s.pool.Query(ctx, `
		select id, room_id, email, nickname, avatar, joined_at
		from participants where room_id=$1 order by joined_at`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Participant
	for rows.Next() {
		p, err := scanParticipant(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// --- Answers ---

func (s *Store) RecordAnswer(ctx context.Context, a *domain.Answer) error {
	_, err := s.pool.Exec(ctx, `
		insert into answers (id, room_id, participant_id, question_id, option_index, correct, question_started_at, received_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.RoomID, a.ParticipantID, a.QuestionID, a.OptionIndex, a.Correct,
		a.QuestionStartedAt, a.ReceivedAt)
	if isUniqueViolation(err) {
		return domain.ErrAlreadyAnswered
	}
	if isFKViolation(err) {
		return domain.ErrParticipantNotFound
	}
	return err
}

func (s *Store) HasAnswered(ctx context.Context, roomID, participantID, questionID string) (bool, error) {
	var one int
	err := s.pool.QueryRow(ctx, `
		select 1 from answers where participant_id=$1 and question_id=$2 limit 1`,
		participantID, questionID).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (s *Store) ListAnswers(ctx context.Context, roomID string) ([]domain.Answer, error) {
	rows, err := s.pool.Query(ctx, `
		select id, room_id, participant_id, question_id, option_index, correct, question_started_at, received_at
		from answers where room_id=$1 order by received_at`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Answer
	for rows.Next() {
		var a domain.Answer
		if err := rows.Scan(&a.ID, &a.RoomID, &a.ParticipantID, &a.QuestionID, &a.OptionIndex, &a.Correct, &a.QuestionStartedAt, &a.ReceivedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// --- Claims ---

func scanClaim(row pgx.Row) (*domain.Claim, error) {
	var c domain.Claim
	var amount int64
	var state string
	var paidAt *time.Time
	if err := row.Scan(&c.ID, &c.QuizID, &c.RoomID, &c.Email, &amount, &c.ClaimCode, &state, &c.CreatedAt, &paidAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrClaimNotFound
		}
		return nil, err
	}
	c.Amount = domain.Amount(amount)
	c.State = domain.ClaimState(state)
	if paidAt != nil {
		c.PaidAt = paidAt
	}
	return &c, nil
}

func (s *Store) CreateClaim(ctx context.Context, c *domain.Claim) error {
	_, err := s.pool.Exec(ctx, `
		insert into claims (id, quiz_id, room_id, email, amount_kobo, claim_code, state, created_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8)`,
		c.ID, c.QuizID, c.RoomID, c.Email, int64(c.Amount), c.ClaimCode, string(c.State), c.CreatedAt)
	if isUniqueViolation(err) {
		return domain.ErrClaimExists
	}
	return err
}

func (s *Store) GetClaimByCode(ctx context.Context, quizID, code string) (*domain.Claim, error) {
	return scanClaim(s.pool.QueryRow(ctx, `
		select id, quiz_id, room_id, email, amount_kobo, claim_code, state, created_at, paid_at
		from claims where quiz_id=$1 and claim_code=$2`, quizID, code))
}

func (s *Store) GetClaimByEmail(ctx context.Context, quizID, email string) (*domain.Claim, error) {
	return scanClaim(s.pool.QueryRow(ctx, `
		select id, quiz_id, room_id, email, amount_kobo, claim_code, state, created_at, paid_at
		from claims where quiz_id=$1 and email=$2`, quizID, email))
}

func (s *Store) ListClaims(ctx context.Context, quizID string) ([]domain.Claim, error) {
	rows, err := s.pool.Query(ctx, `
		select id, quiz_id, room_id, email, amount_kobo, claim_code, state, created_at, paid_at
		from claims where quiz_id=$1 order by created_at`, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Claim
	for rows.Next() {
		c, err := scanClaim(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func (s *Store) UpdateClaimState(ctx context.Context, id string, to domain.ClaimState) error {
	tag, err := s.pool.Exec(ctx, `
		update claims set state=$2, paid_at=case when $2='paid' then now() else paid_at end
		where id=$1`, id, string(to))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrClaimNotFound
	}
	return nil
}
