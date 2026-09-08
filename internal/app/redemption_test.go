package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/osugodbless/kweeks/internal/adapters/clock"
	"github.com/osugodbless/kweeks/internal/adapters/store/memory"
	"github.com/osugodbless/kweeks/internal/domain"
	"github.com/osugodbless/kweeks/internal/ports"
)

// fakeMoney implements ports.Money for the redemption payout tests.
type fakeMoney struct {
	bankAccountID string
	holderName    string
	payoutRef     string
	provisioned   *domain.WalletExternal
	payErr        error
	lastAmount    domain.Amount
}

func (f *fakeMoney) Provision(ctx context.Context, p domain.BmoniPersona) (*domain.WalletExternal, error) {
	return &domain.WalletExternal{UserID: "usr_host", WalletID: "wal_host", Address: "0xhost"}, nil
}

func (f *fakeMoney) ListNigerianBanks(ctx context.Context, userID string) ([]domain.NigerianBank, error) {
	return []domain.NigerianBank{{Code: "058", Name: "GTB"}, {Code: "044", Name: "Access"}}, nil
}

func (f *fakeMoney) VerifyNigerianAccount(ctx context.Context, userID, accountNumber, bankCode string) (string, error) {
	return f.holderName, nil
}

func (f *fakeMoney) RegisterNigerianWithdrawalAccount(ctx context.Context, userID string, acct domain.NigerianAccount) (string, error) {
	return f.bankAccountID, nil
}

func (f *fakeMoney) PayWinnerToNigerianBank(ctx context.Context, from *domain.WalletExternal, bankAccountID string, amount domain.Amount) (string, error) {
	f.lastAmount = amount
	if f.payErr != nil {
		return "", f.payErr
	}
	return f.payoutRef, nil
}

func (f *fakeMoney) DepositAccount(ctx context.Context, userID, walletID string) (string, string, error) {
	return "0123456789", "Providus", nil
}

var _ ports.Money = (*fakeMoney)(nil)

// seedHostAndWinner creates an instructor + provisioned wallet + a podium room
// with one winning participant, and returns the room id + winner email.
func seedHostAndWinner(t *testing.T, st *memory.Store, clk *clock.Static, money ports.Money) (roomID, email string) {
	t.Helper()
	ctx := context.Background()

	ins := &domain.Instructor{ID: "ins-1", Name: "Adeola", Email: "host@kweeks.ng", Avatar: "AP", CreatedAt: clk.Now()}
	if err := st.CreateInstructor(ctx, ins); err != nil {
		t.Fatal(err)
	}
	if err := st.CreateWallet(ctx, &domain.Wallet{
		ID: "wallet-1", InstructorID: "ins-1", Balance: 100000,
		BmoniUserID: "usr_host", BmoniWalletID: "wal_host", BmoniWalletAddr: "0xhost", CreatedAt: clk.Now(),
	}); err != nil {
		t.Fatal(err)
	}

	quiz := &domain.Quiz{
		ID: "quiz-1", InstructorID: "ins-1", Title: "Demo", Pool: 25000, WinnerCount: 1,
		Pacing: domain.PacingManual, DefaultDuration: 60 * time.Second,
		Questions: []domain.Question{{ID: "q1", Prompt: "Q?", Options: []string{"a", "b"}, CorrectIndex: 1}},
	}
	if err := st.CreateQuiz(ctx, quiz); err != nil {
		t.Fatal(err)
	}
	room := &domain.Room{ID: "room-1", QuizID: quiz.ID, HostID: "ins-1", State: domain.RoomPodium}
	if err := st.CreateRoom(ctx, room); err != nil {
		t.Fatal(err)
	}
	p, err := st.JoinParticipant(ctx, &domain.Participant{ID: "p-1", RoomID: "room-1", Email: "winner@x.com", Nickname: "w", Avatar: "🦁", JoinedAt: clk.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.RecordAnswer(ctx, &domain.Answer{
		ID: "a-1", RoomID: "room-1", ParticipantID: p.ID, QuestionID: "q1",
		OptionIndex: 1, Correct: true, QuestionStartedAt: clk.Now(), ReceivedAt: clk.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	// mark room podium for CreateClaim.
	room.State = domain.RoomPodium
	if err := st.SaveRoom(ctx, room); err != nil {
		t.Fatal(err)
	}
	_ = money
	return "room-1", "winner@x.com"
}

func TestCreateClaimEmailsImmediately(t *testing.T) {
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	roomID, email := seedHostAndWinner(t, st, clk, nil)

	sent := false
	mail := &fakeMail{send: func(to, code, amount, url string) { sent = true }}
	red := NewRedemption(st, clk, nil, mail).WithPublicURL("https://kweeks.ng")

	claim, err := red.CreateClaim(context.Background(), roomID, email, 25000)
	if err != nil {
		t.Fatalf("create claim: %v", err)
	}
	if claim.ClaimCode == "" || claim.State != domain.ClaimCreated {
		t.Fatalf("unexpected claim: %+v", claim)
	}
	red.SendRedemptionEmail(context.Background(), claim)
	if !sent {
		t.Fatalf("expected redemption email at redeem time")
	}
	if !strings.Contains(red.ClaimURL(claim.ClaimCode, email), "/claim?code="+claim.ClaimCode) {
		t.Fatalf("claim URL missing code: %s", red.ClaimURL(claim.ClaimCode, email))
	}
}

type fakeMail struct {
	send func(to, code, amount, url string)
}

func (f *fakeMail) SendRedemptionEmail(ctx context.Context, to, code, amount, url string) error {
	if f.send != nil {
		f.send(to, code, amount, url)
	}
	return nil
}

var _ ports.Mail = (*fakeMail)(nil)

func TestSubmitBankPayoutHappyPath(t *testing.T) {
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	roomID, email := seedHostAndWinner(t, st, clk, nil)

	money := &fakeMoney{bankAccountID: "ba_1", holderName: "Winner Name", payoutRef: "prop_1"}
	red := NewRedemption(st, clk, money, nil)

	claim, err := red.CreateClaim(context.Background(), roomID, email, 25000)
	if err != nil {
		t.Fatalf("create claim: %v", err)
	}

	paid, err := red.SubmitBankPayout(context.Background(), claim.ClaimCode, email, domain.NigerianAccount{
		AccountNumber: "0123456789", BankCode: "058", BankName: "Guaranty Trust Bank",
	})
	if err != nil {
		t.Fatalf("submit payout: %v", err)
	}
	if paid.State != domain.ClaimPaid {
		t.Fatalf("claim state = %s, want paid", paid.State)
	}
	if paid.BankAccountID != "ba_1" || paid.PayoutRef != "prop_1" || paid.AccountHolderName != "Winner Name" {
		t.Fatalf("claim payout details missing: %+v", paid)
	}
	if money.lastAmount != 25000 {
		t.Fatalf("offramp amount = %d, want 25000", money.lastAmount)
	}
}

func TestSubmitBankPayoutWrongCodeOrEmail(t *testing.T) {
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	roomID, email := seedHostAndWinner(t, st, clk, nil)

	money := &fakeMoney{bankAccountID: "ba_1", holderName: "Winner Name", payoutRef: "prop_1"}
	red := NewRedemption(st, clk, money, nil)
	claim, err := red.CreateClaim(context.Background(), roomID, email, 25000)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := red.SubmitBankPayout(context.Background(), claim.ClaimCode, "attacker@x.com", domain.NigerianAccount{AccountNumber: "0123456789", BankCode: "058", BankName: "GTB"}); err != domain.ErrBadClaimCode {
		t.Fatalf("wrong email: got %v, want ErrBadClaimCode", err)
	}
	if _, err := red.SubmitBankPayout(context.Background(), "deadbeef", email, domain.NigerianAccount{AccountNumber: "0123456789", BankCode: "058", BankName: "GTB"}); err != domain.ErrBadClaimCode {
		t.Fatalf("wrong code: got %v, want ErrBadClaimCode", err)
	}
}

func TestSubmitBankPayoutMarksFailedOnOfframpError(t *testing.T) {
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	roomID, email := seedHostAndWinner(t, st, clk, nil)

	money := &fakeMoney{bankAccountID: "ba_1", holderName: "Winner Name", payErr: domain.ErrInsufficientBalance}
	red := NewRedemption(st, clk, money, nil)
	claim, err := red.CreateClaim(context.Background(), roomID, email, 25000)
	if err != nil {
		t.Fatal(err)
	}

	paid, err := red.SubmitBankPayout(context.Background(), claim.ClaimCode, email, domain.NigerianAccount{AccountNumber: "0123456789", BankCode: "058", BankName: "GTB"})
	if err == nil {
		t.Fatalf("expected offramp error")
	}
	if paid == nil || paid.State != domain.ClaimFailed {
		t.Fatalf("claim should be failed, got %+v", paid)
	}
}

func TestResolveClaimAndListBanks(t *testing.T) {
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	roomID, email := seedHostAndWinner(t, st, clk, nil)

	money := &fakeMoney{bankAccountID: "ba_1", holderName: "Winner Name", payoutRef: "prop_1"}
	red := NewRedemption(st, clk, money, nil)
	claim, err := red.CreateClaim(context.Background(), roomID, email, 25000)
	if err != nil {
		t.Fatal(err)
	}

	got, err := red.ResolveClaim(context.Background(), claim.ClaimCode, email)
	if err != nil || got.ID != claim.ID {
		t.Fatalf("resolve: %v %+v", err, got)
	}
	banks, err := red.ListBanks(context.Background(), got)
	if err != nil || len(banks) != 2 {
		t.Fatalf("banks: %v %+v", err, banks)
	}
}
