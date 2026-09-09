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
// out immediately with the claim code + a claim URL (the recovery artifact so
// the code can never be lost), and the money move settles to the winner's
// Nigerian bank account from the host's wallet when the winner submits their
// bank details.
type Redemption struct {
	store ports.Store
	clock ports.Clock
	money ports.Money
	mail  ports.Mail

	// publicURL is the externally-reachable base URL (KWEEKS_PUBLIC_URL) used
	// to build the /claim link in the redemption email.
	publicURL string
}

func NewRedemption(store ports.Store, clock ports.Clock, money ports.Money, mail ports.Mail) *Redemption {
	return &Redemption{store: store, clock: clock, money: money, mail: mail}
}

// WithPublicURL sets the external base URL used for claim links in email.
func (r *Redemption) WithPublicURL(publicURL string) *Redemption {
	r.publicURL = publicURL
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
// The claim amount is the winner's SPECIFIC share of the pool — the rank-based
// SplitPodium share for that winner — never the whole pool.
func (r *Redemption) CreateClaim(ctx context.Context, roomID, email string) (*domain.Claim, error) {
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
	rank := winnerRank(winners, participant.ID)
	if rank < 0 {
		return nil, domain.ErrNotWinner
	}
	// The winner's share mirrors the podium display: split the pool across the
	// configured winner count and take the share at this winner's rank.
	amount := domain.SplitPodium(quiz.Pool, quiz.WinnerCount)[rank]

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

// winnerRank returns the 0-based rank of a podium winner within the declared
// winners list, or -1 when the participant is not a winner.
func winnerRank(winners []domain.Standing, participantID string) int {
	for i, w := range winners {
		if w.ParticipantID == participantID {
			return i
		}
	}
	return -1
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

// ClaimURL builds the public claim link pre-filled with the code + email.
func (r *Redemption) ClaimURL(claimCode, email string) string {
	base := r.publicURL
	if base == "" {
		base = "/"
	}
	return base + "/claim?code=" + claimCode + "&email=" + email
}

// SendRedemptionEmail dispatches the recovery artifact immediately at redeem
// time. Failures are logged by the adapter; they never block the claim.
func (r *Redemption) SendRedemptionEmail(ctx context.Context, c *domain.Claim) {
	if r.mail == nil {
		return
	}
	_ = r.mail.SendRedemptionEmail(ctx, c.Email, c.ClaimCode, c.Amount.NairaString(), r.ClaimURL(c.ClaimCode, c.Email))
}

// ResolveClaim validates a claim code + email pair and returns the claim. Used
// by the public claim page to pre-fill the amount before bank details.
func (r *Redemption) ResolveClaim(ctx context.Context, claimCode, email string) (*domain.Claim, error) {
	claim, err := r.store.GetClaimByCodeOnly(ctx, claimCode)
	if err != nil {
		return nil, domain.ErrBadClaimCode
	}
	if claim.Email != email {
		return nil, domain.ErrBadClaimCode
	}
	return claim, nil
}

// HostExternal resolves the host wallet's BMONI identity for a claim (the
// wallet that pays the winner). Errors if the host is not provisioned.
func (r *Redemption) HostExternal(ctx context.Context, claim *domain.Claim) (*domain.WalletExternal, error) {
	quiz, err := r.store.GetQuiz(ctx, claim.QuizID)
	if err != nil {
		return nil, err
	}
	wallet, err := r.store.GetWalletByInstructor(ctx, quiz.InstructorID)
	if err != nil {
		return nil, err
	}
	if wallet.BmoniUserID == "" {
		return nil, errors.New("host wallet is not provisioned on the money rail")
	}
	return &domain.WalletExternal{
		UserID: wallet.BmoniUserID, WalletID: wallet.BmoniWalletID, Address: wallet.BmoniWalletAddr,
	}, nil
}

// ListBanks returns the supported Nigerian banks for the claim form, scoped to
// the host wallet.
func (r *Redemption) ListBanks(ctx context.Context, claim *domain.Claim) ([]domain.NigerianBank, error) {
	if r.money == nil {
		return nil, errors.New("money rail not configured")
	}
	ext, err := r.HostExternal(ctx, claim)
	if err != nil {
		return nil, err
	}
	return r.money.ListNigerianBanks(ctx, ext.UserID)
}

// SubmitBankPayout takes the winner's Nigerian bank details, verifies +
// registers the account on the host's rail, and offramps the prize from the
// host wallet to the winner's bank account. The claim advances
// created -> bank_submitted -> paying -> paid.
func (r *Redemption) SubmitBankPayout(ctx context.Context, claimCode, email string, acct domain.NigerianAccount) (*domain.Claim, error) {
	claim, err := r.ResolveClaim(ctx, claimCode, email)
	if err != nil {
		return nil, err
	}
	if r.money == nil {
		return nil, errors.New("money rail not configured")
	}
	if !domain.CanTransition(claim.State, domain.ClaimBankSubmitted) {
		return nil, domain.ErrInvalidTransition
	}
	if len(acct.AccountNumber) != 10 || acct.BankCode == "" || acct.BankName == "" {
		return nil, errors.New("enter a valid 10-digit Nigerian account number and bank")
	}

	ext, err := r.HostExternal(ctx, claim)
	if err != nil {
		return nil, err
	}

	// 1. Verify the account -> exact holder name the registration requires.
	holder, err := r.money.VerifyNigerianAccount(ctx, ext.UserID, acct.AccountNumber, acct.BankCode)
	if err != nil {
		return nil, err
	}
	acct.AccountHolderName = holder

	// 2. Register the withdrawal account (get-or-create).
	bankAccountID, err := r.money.RegisterNigerianWithdrawalAccount(ctx, ext.UserID, acct)
	if err != nil {
		return nil, err
	}

	// Persist the bank details + bank_submitted state before money moves, so a
	// failure below leaves the claim resumable, not lost.
	claim.BankAccountID = bankAccountID
	claim.BankAccountNumber = acct.AccountNumber
	claim.BankName = acct.BankName
	claim.AccountHolderName = holder
	if err := r.store.UpdateClaimBank(ctx, claim); err != nil {
		return nil, err
	}
	if err := r.store.UpdateClaimState(ctx, claim.ID, domain.ClaimBankSubmitted); err != nil {
		return nil, err
	}
	claim.State = domain.ClaimBankSubmitted

	// 3. Offramp the prize from the host wallet to the winner's bank account.
	if err := r.store.UpdateClaimState(ctx, claim.ID, domain.ClaimPaying); err != nil {
		return nil, err
	}
	claim.State = domain.ClaimPaying
	ref, err := r.money.PayWinnerToNigerianBank(ctx, ext, bankAccountID, claim.Amount)
	if err != nil {
		_ = r.store.UpdateClaimState(ctx, claim.ID, domain.ClaimFailed)
		claim.State = domain.ClaimFailed
		return claim, err
	}
	claim.PayoutRef = ref
	if err := r.store.UpdateClaimBank(ctx, claim); err != nil {
		return nil, err
	}
	if err := r.store.UpdateClaimState(ctx, claim.ID, domain.ClaimPaid); err != nil {
		return nil, err
	}
	claim.State = domain.ClaimPaid
	return claim, nil
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
