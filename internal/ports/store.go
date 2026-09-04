// Package ports defines the boundaries the application layer needs from the
// outside world. Implementations live in internal/adapters; nothing in ports
// imports adapters.
package ports

import (
	"context"
	"time"

	"github.com/osugodbless/kweeks/internal/domain"
)

// Clock is the server clock. Everything time-critical (answer cutoff, pacing
// advance) reads from here so tests can freeze time.
type Clock interface {
	Now() time.Time
}

// Store is the persistence boundary for the whole game. A single interface
// keeps the demo runnable against an in-memory adapter and shippable against
// Postgres without touching callers.
type Store interface {
	// Quizzes
	CreateQuiz(ctx context.Context, q *domain.Quiz) error
	GetQuiz(ctx context.Context, id string) (*domain.Quiz, error)
	UpdateQuiz(ctx context.Context, q *domain.Quiz) error
	// A quiz is started by attaching it to a room; list is for the instructor
	// dashboard.
	ListQuizzes(ctx context.Context, instructorID string) ([]domain.Quiz, error)

	// Rooms
	CreateRoom(ctx context.Context, r *domain.Room) error
	GetRoom(ctx context.Context, id string) (*domain.Room, error)
	GetRoomByCode(ctx context.Context, code string) (*domain.Room, error)
	SaveRoom(ctx context.Context, r *domain.Room) error
	// ListLiveRooms returns rooms in the live state (used by the pacing
	// scheduler to advance AUTO rooms).
	ListLiveRooms(ctx context.Context) ([]domain.Room, error)
	// ListRoomsByHost lists rooms a host has opened, newest first. Used by the
	// dashboard + history.
	ListRoomsByHost(ctx context.Context, hostID string) ([]domain.Room, error)
	// LatestLiveRoom returns the most recently started room for a quiz, if one
	// is in lobby/live state (dashboard "continue" affordance).
	LatestLiveRoom(ctx context.Context, quizID string) (*domain.Room, error)

	// Participants (one row per email per room; rejoin merges)
	JoinParticipant(ctx context.Context, p *domain.Participant) (*domain.Participant, error)
	GetParticipant(ctx context.Context, roomID, email string) (*domain.Participant, error)
	ListParticipants(ctx context.Context, roomID string) ([]domain.Participant, error)

	// Answers
	// RecordAnswer stores an answer exactly once. Returning
	// domain.ErrAlreadyAnswered when a duplicate is detected is the caller's
	// job; the store only persists. Answer existence is checked by
	// HasAnswered.
	RecordAnswer(ctx context.Context, a *domain.Answer) error
	HasAnswered(ctx context.Context, roomID, participantID, questionID string) (bool, error)
	ListAnswers(ctx context.Context, roomID string) ([]domain.Answer, error)

	// Claims (exactly-once per email+quiz, enforced by the store)
	CreateClaim(ctx context.Context, c *domain.Claim) error
	GetClaimByCode(ctx context.Context, quizID, code string) (*domain.Claim, error)
	GetClaimByEmail(ctx context.Context, quizID, email string) (*domain.Claim, error)
	ListClaims(ctx context.Context, quizID string) ([]domain.Claim, error)
	ListClaimsByQuizIDs(ctx context.Context, quizIDs []string) ([]domain.Claim, error)
	UpdateClaimState(ctx context.Context, id string, to domain.ClaimState) error

	// Instructors (multi-user auth)
	CreateInstructor(ctx context.Context, i *domain.Instructor) error
	GetInstructorByEmail(ctx context.Context, email string) (*domain.Instructor, error)
	GetInstructor(ctx context.Context, id string) (*domain.Instructor, error)

	// Sessions (bearer tokens)
	CreateSession(ctx context.Context, s *domain.Session) error
	GetSession(ctx context.Context, token string) (*domain.Session, error)

	// Wallets (per-instructor NGN ledger)
	CreateWallet(ctx context.Context, w *domain.Wallet) error
	GetWalletByInstructor(ctx context.Context, instructorID string) (*domain.Wallet, error)
	// ApplyWalletTx atomically applies a signed transaction to the balance and
	// appends the ledger row.
	ApplyWalletTx(ctx context.Context, walletID string, tx *domain.WalletTransaction) error
	ListWalletTransactions(ctx context.Context, walletID string) ([]domain.WalletTransaction, error)
	// SetWalletBmoni records the external BMONI ids provisioned for a wallet.
	SetWalletBmoni(ctx context.Context, walletID string, external *domain.WalletExternal) error

	// Admin / bootstrap
	Close() error
}
