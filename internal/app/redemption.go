package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/osugodbless/kweeks/internal/domain"
	"github.com/osugodbless/kweeks/internal/ports"
)

// Redemption orchestrates the winner-driven claim flow. A winner taps redeem
// on their current screen; the claim is created exactly once, an email goes
// out as a recovery artifact, and the money move settles in the background.
type Redemption struct {
	store ports.Store
	clock ports.Clock
	money ports.Money
	mail  ports.Mail

	// winnerUserID is the BMONI user that closes the payout loop live (the
	// second sandbox persona). When set, Settle pays that user from the
	// platform wallet; empty means real payouts are disabled.
	winnerUserID string
}

func NewRedemption(store ports.Store, clock ports.Clock, money ports.Money, mail ports.Mail) *Redemption {
	return &Redemption{store: store, clock: clock, money: money, mail: mail}
}

// WithWinnerUser sets the BMONI recipient that live payouts settle to.
func (r *Redemption) WithWinnerUser(winnerUserID string) *Redemption {
	r.winnerUserID = winnerUserID
	return r
}

func (r *Redemption) nowTime() time.Time {
	if r.clock != nil {
		return r.clock.Now()
	}
	return time.Now()
}

// CreateClaim is the exactly-once claim write. The winner's email must match a
// podium winner for the quiz. The claim code is generated here and returned to
// the caller; it is delivered to the winner's session only, never broadcast.
func (r *Redemption) CreateClaim(ctx context.Context, roomID, email string, amount domain.Amount) (*domain.Claim, error) {
	room, err := r.store.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room.State != domain.RoomPodium {
		return nil, domain.ErrRoomWrongState
	}
	quiz, err := r.store.GetQuiz(ctx, room.QuizID)
	if err != nil {
		return nil, err
	}
	// Winner check: the email must resolve to a podium winner. The standings
	// carry participant ids only (never emails, so a broadcast cannot leak
	// them), so we resolve the participant first and match on id.
	participant, err := r.store.GetParticipant(ctx, roomID, email)
	if err != nil {
		if errors.Is(err, domain.ErrParticipantNotFound) {
			return nil, domain.ErrNotWinner
		}
		return nil, err
	}
	standings, err := r.Standings(ctx, roomID)
	if err != nil {
		return nil, err
	}
	winners := domain.SelectWinners(standings, quiz.WinnerCount)
	if !containsWinnerID(winners, participant.ID) {
		return nil, domain.ErrNotWinner
	}

	// Exactly-once: a second claim for the same winner returns the existing
	// one rather than erroring, so a double-tap or replay is harmless.
	if existing, err := r.store.GetClaimByEmail(ctx, quiz.ID, email); err == nil && existing != nil {
		return existing, nil
	}

	code, err := newClaimCode()
	if err != nil {
		return nil, err
	}
	now := r.nowTime()
	claim := &domain.Claim{
		ID:        newID(),
		QuizID:    quiz.ID,
		RoomID:    roomID,
		Email:     email,
		Amount:    amount,
		ClaimCode: code,
		State:     domain.ClaimCreated,
		CreatedAt: now,
	}
	if err := r.store.CreateClaim(ctx, claim); err != nil {
		if errors.Is(err, domain.ErrClaimExists) {
			if existing, err2 := r.store.GetClaimByEmail(ctx, quiz.ID, email); err2 == nil && existing != nil {
				return existing, nil
			}
		}
		return nil, err
	}
	return claim, nil
}

// Standings is a thin passthrough to the game service's standings logic; the
// redemption flow needs the same ordering to confirm winners.
func (r *Redemption) Standings(ctx context.Context, roomID string) ([]domain.Standing, error) {
	participants, err := r.store.ListParticipants(ctx, roomID)
	if err != nil {
		return nil, err
	}
	answers, err := r.store.ListAnswers(ctx, roomID)
	if err != nil {
		return nil, err
	}
	byPID := map[string]*domain.Standing{}
	for _, p := range participants {
		byPID[p.ID] = &domain.Standing{ParticipantID: p.ID, Nickname: p.Nickname, Avatar: p.Avatar, JoinedAt: p.JoinedAt}
	}
	for _, a := range answers {
		s := byPID[a.ParticipantID]
		if s == nil || !a.Correct {
			continue
		}
		s.CorrectCount++
		s.TotalLatency += time.Duration(domain.LatencyMs(a.ReceivedAt, a.QuestionStartedAt)) * time.Millisecond
	}
	out := make([]domain.Standing, 0, len(byPID))
	for _, s := range byPID {
		out = append(out, *s)
	}
	return domain.SortStandings(out), nil
}

func containsWinnerID(winners []domain.Standing, participantID string) bool {
	for _, w := range winners {
		if w.ParticipantID == participantID {
			return true
		}
	}
	return false
}

// SendRedemptionEmail dispatches the recovery artifact. Failures are logged
// by the adapter; they never block the claim.
func (r *Redemption) SendRedemptionEmail(ctx context.Context, c *domain.Claim) {
	if r.mail == nil {
		return
	}
	_ = r.mail.SendRedemptionEmail(ctx, c.Email, c.ClaimCode, c.Amount.NairaString())
}

// Settle triggers the background money move for a claim that has reached the
// onboarded state. Returns the settlement reference.
//
// The money move goes from the platform wallet to the configured demo winner
// user (the second sandbox persona) — the master design's pre-provisioned
// persona that closes the payout loop live. Claims for non-persona emails are
// still created + emailed; onboarding them is the no-app/invite flow.
func (r *Redemption) Settle(ctx context.Context, c *domain.Claim) (string, error) {
	if !domain.CanTransition(c.State, domain.ClaimPaid) {
		return "", domain.ErrInvalidTransition
	}
	if r.money == nil {
		return "", errors.New("money port not configured")
	}
	recipient := r.winnerUserID
	if recipient == "" {
		recipient = c.Email // no persona configured: attempt the email (no-app invite resolves it)
	}
	ref, err := r.money.PayWinner(ctx, recipient, c.Amount)
	if err != nil {
		_ = r.store.UpdateClaimState(ctx, c.ID, domain.ClaimFailed)
		return "", err
	}
	if err := r.store.UpdateClaimState(ctx, c.ID, domain.ClaimPaid); err != nil {
		return "", err
	}
	return ref, nil
}

func newClaimCode() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
