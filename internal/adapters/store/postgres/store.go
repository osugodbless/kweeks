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

func (s *Store) UpdateQuiz(ctx context.Context, q *domain.Quiz) error {
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
	tag, err := s.pool.Exec(ctx, `
		update quizzes set title=$2, pool_kobo=$3, winner_count=$4, pacing=$5,
			default_duration_ms=$6, questions=$7
		where id=$1`,
		q.ID, q.Title, int64(q.Pool), q.WinnerCount, string(q.Pacing),
		q.DefaultDuration.Milliseconds(), b)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrQuizNotFound
	}
	return nil
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
	// Treat an empty code as NULL so the partial unique index still allows
	// multiple rows without a code (test fixtures) while real rooms carry one.
	code := (*string)(nil)
	if r.Code != "" {
		code = &r.Code
	}
	_, err := s.pool.Exec(ctx, `
		insert into rooms (id, quiz_id, state, host_id, current_question_idx, question_started_at, started_at, code)
		values ($1,$2,$3,$4,$5,$6,$7,$8)`,
		r.ID, r.QuizID, string(r.State), r.HostID, r.CurrentQuestionIdx, r.QuestionStartedAt, r.StartedAt, code)
	if isUniqueViolation(err) {
		return domain.ErrRoomCodeTaken
	}
	if isFKViolation(err) {
		return domain.ErrQuizNotFound
	}
	return err
}

func (s *Store) GetRoom(ctx context.Context, id string) (*domain.Room, error) {
	return s.scanRoom(s.pool.QueryRow(ctx, `
		select id, quiz_id, state, host_id, current_question_idx, question_started_at, started_at, code
		from rooms where id = $1`, id))
}

func (s *Store) GetRoomByCode(ctx context.Context, code string) (*domain.Room, error) {
	return s.scanRoom(s.pool.QueryRow(ctx, `
		select id, quiz_id, state, host_id, current_question_idx, question_started_at, started_at, code
		from rooms where code = $1`, code))
}

func (s *Store) scanRoom(row pgx.Row) (*domain.Room, error) {
	var r domain.Room
	var state string
	var startedAt *time.Time
	var code *string
	err := row.Scan(&r.ID, &r.QuizID, &state, &r.HostID, &r.CurrentQuestionIdx, &r.QuestionStartedAt, &startedAt, &code)
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
	if code != nil {
		r.Code = *code
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
		select id, quiz_id, state, host_id, current_question_idx, question_started_at, started_at, code
		from rooms where state='live'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Room
	for rows.Next() {
		var r domain.Room
		var startedAt *time.Time
		var code *string
		if err := rows.Scan(&r.ID, &r.QuizID, &r.State, &r.HostID, &r.CurrentQuestionIdx, &r.QuestionStartedAt, &startedAt, &code); err != nil {
			return nil, err
		}
		if startedAt != nil {
			r.StartedAt = *startedAt
		}
		if code != nil {
			r.Code = *code
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) ListRoomsByHost(ctx context.Context, hostID string) ([]domain.Room, error) {
	rows, err := s.pool.Query(ctx, `
		select id, quiz_id, state, host_id, current_question_idx, question_started_at, started_at, code
		from rooms where host_id=$1 order by created_at desc`, hostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Room
	for rows.Next() {
		var r domain.Room
		var startedAt *time.Time
		var code *string
		if err := rows.Scan(&r.ID, &r.QuizID, &r.State, &r.HostID, &r.CurrentQuestionIdx, &r.QuestionStartedAt, &startedAt, &code); err != nil {
			return nil, err
		}
		if startedAt != nil {
			r.StartedAt = *startedAt
		}
		if code != nil {
			r.Code = *code
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) LatestLiveRoom(ctx context.Context, quizID string) (*domain.Room, error) {
	row := s.pool.QueryRow(ctx, `
		select id, quiz_id, state, host_id, current_question_idx, question_started_at, started_at, code
		from rooms
		where quiz_id=$1 and state in ('lobby','live')
		order by started_at desc nulls last, created_at desc
		limit 1`, quizID)
	return s.scanRoom(row)
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

func (s *Store) ListClaimsByQuizIDs(ctx context.Context, quizIDs []string) ([]domain.Claim, error) {
	if len(quizIDs) == 0 {
		return []domain.Claim{}, nil
	}
	rows, err := s.pool.Query(ctx, `
		select id, quiz_id, room_id, email, amount_kobo, claim_code, state, created_at, paid_at
		from claims where quiz_id = any($1) order by created_at`, quizIDs)
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

// --- Instructors ---

func (s *Store) CreateInstructor(ctx context.Context, i *domain.Instructor) error {
	_, err := s.pool.Exec(ctx, `
		insert into instructors (id, name, email, password_hash, avatar, created_at)
		values ($1,$2,$3,$4,$5,$6)`,
		i.ID, i.Name, i.Email, i.PasswordHash, i.Avatar, i.CreatedAt)
	if isUniqueViolation(err) {
		return domain.ErrEmailTaken
	}
	return err
}

func scanInstructor(row pgx.Row) (*domain.Instructor, error) {
	var i domain.Instructor
	if err := row.Scan(&i.ID, &i.Name, &i.Email, &i.PasswordHash, &i.Avatar, &i.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInstructorNotFound
		}
		return nil, err
	}
	return &i, nil
}

func (s *Store) GetInstructorByEmail(ctx context.Context, email string) (*domain.Instructor, error) {
	return scanInstructor(s.pool.QueryRow(ctx, `
		select id, name, email, password_hash, avatar, created_at
		from instructors where email=$1`, email))
}

func (s *Store) GetInstructor(ctx context.Context, id string) (*domain.Instructor, error) {
	return scanInstructor(s.pool.QueryRow(ctx, `
		select id, name, email, password_hash, avatar, created_at
		from instructors where id=$1`, id))
}

// --- Sessions ---

func (s *Store) CreateSession(ctx context.Context, sess *domain.Session) error {
	_, err := s.pool.Exec(ctx, `
		insert into sessions (token, instructor_id, created_at, expires_at)
		values ($1,$2,$3,$4)`,
		sess.Token, sess.InstructorID, sess.CreatedAt, sess.ExpiresAt)
	return err
}

func (s *Store) GetSession(ctx context.Context, token string) (*domain.Session, error) {
	var sess domain.Session
	err := s.pool.QueryRow(ctx, `
		select token, instructor_id, created_at, expires_at
		from sessions where token=$1`, token).
		Scan(&sess.Token, &sess.InstructorID, &sess.CreatedAt, &sess.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

// --- Wallets ---

func (s *Store) CreateWallet(ctx context.Context, w *domain.Wallet) error {
	_, err := s.pool.Exec(ctx, `
		insert into wallets (id, instructor_id, balance_kobo, created_at)
		values ($1,$2,$3,$4)`,
		w.ID, w.InstructorID, int64(w.Balance), w.CreatedAt)
	if isUniqueViolation(err) {
		return domain.ErrWalletExists
	}
	return err
}

func (s *Store) GetWalletByInstructor(ctx context.Context, instructorID string) (*domain.Wallet, error) {
	var w domain.Wallet
	var bal int64
	err := s.pool.QueryRow(ctx, `
		select id, instructor_id, balance_kobo, created_at
		from wallets where instructor_id=$1`, instructorID).
		Scan(&w.ID, &w.InstructorID, &bal, &w.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrWalletNotFound
	}
	if err != nil {
		return nil, err
	}
	w.Balance = domain.Amount(bal)
	return &w, nil
}

func (s *Store) ApplyWalletTx(ctx context.Context, walletID string, tx *domain.WalletTransaction) error {
	tag, err := s.pool.Exec(ctx, `
		update wallets set balance_kobo = balance_kobo + $2 where id=$1 and balance_kobo + $2 >= 0`,
		walletID, int64(tx.Amount))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInsufficientBalance
	}
	_, err = s.pool.Exec(ctx, `
		insert into wallet_transactions (id, wallet_id, kind, amount_kobo, note, created_at)
		values ($1,$2,$3,$4,$5,$6)`,
		tx.ID, walletID, string(tx.Kind), int64(tx.Amount), tx.Note, tx.CreatedAt)
	return err
}

func (s *Store) ListWalletTransactions(ctx context.Context, walletID string) ([]domain.WalletTransaction, error) {
	rows, err := s.pool.Query(ctx, `
		select id, wallet_id, kind, amount_kobo, note, created_at
		from wallet_transactions where wallet_id=$1 order by created_at desc`, walletID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.WalletTransaction
	for rows.Next() {
		var t domain.WalletTransaction
		var kind string
		var amt int64
		if err := rows.Scan(&t.ID, &t.WalletID, &kind, &amt, &t.Note, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.Kind = domain.WalletTxKind(kind)
		t.Amount = domain.Amount(amt)
		out = append(out, t)
	}
	return out, rows.Err()
}
