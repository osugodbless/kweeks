package app

import (
	"context"
	"testing"
	"time"

	"github.com/osugodbless/kweeks/internal/adapters/clock"
	"github.com/osugodbless/kweeks/internal/adapters/store/memory"
	"github.com/osugodbless/kweeks/internal/domain"
	"github.com/osugodbless/kweeks/internal/ports"
)

// personaRecordingMoney provisions a distinct BMONI user id derived from the
// persona phone, so two instructors provisioning with the same phone would get
// the SAME external id (the shared-identity bug) and different phones get
// different ids.
type personaRecordingMoney struct {
	personas []domain.BmoniPersona
}

func (f *personaRecordingMoney) Provision(ctx context.Context, p domain.BmoniPersona) (*domain.WalletExternal, error) {
	f.personas = append(f.personas, p)
	return &domain.WalletExternal{UserID: "usr-" + p.Phone, WalletID: "wal-" + p.Phone, Address: "0x" + p.Phone}, nil
}

func (f *personaRecordingMoney) ListNigerianBanks(ctx context.Context, userID string) ([]domain.NigerianBank, error) {
	return nil, nil
}
func (f *personaRecordingMoney) VerifyNigerianAccount(ctx context.Context, userID, accountNumber, bankCode string) (string, error) {
	return "", nil
}
func (f *personaRecordingMoney) RegisterNigerianWithdrawalAccount(ctx context.Context, userID string, acct domain.NigerianAccount) (string, error) {
	return "", nil
}
func (f *personaRecordingMoney) PayWinnerToNigerianBank(ctx context.Context, from *domain.WalletExternal, bankAccountID string, amount domain.Amount) (string, error) {
	return "", nil
}
func (f *personaRecordingMoney) DepositAccount(ctx context.Context, userID, walletID string) (string, string, error) {
	return "", "", nil
}

var _ ports.Money = (*personaRecordingMoney)(nil)

func newInstructor(t *testing.T, st *memory.Store, clk *clock.Static, id, email, phone string) {
	t.Helper()
	if err := st.CreateInstructor(context.Background(), &domain.Instructor{
		ID: id, Name: "Host", Email: email, Phone: phone, PasswordHash: "h", Avatar: "AP", CreatedAt: clk.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.CreateWallet(context.Background(), &domain.Wallet{ID: "wal-" + id, InstructorID: id, CreatedAt: clk.Now()}); err != nil {
		t.Fatal(err)
	}
}

// Two instructors must provision DISTINCT BMONI users, never the shared persona
// phone. This is the regression guard for the reported account-sharing bug.
func TestProvisionAssignsDistinctBmoniIdentity(t *testing.T) {
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	newInstructor(t, st, clk, "ins-a", "a@kweeks.ng", "+2348011111111")
	newInstructor(t, st, clk, "ins-b", "b@kweeks.ng", "+2348022222222")

	money := &personaRecordingMoney{}
	w := NewWallet(st, clk, money).WithProvisioning(domain.BmoniPersona{
		FirstName: "Bunch", LastName: "Dillon", Email: "shared@example.com",
		Phone: "+2348000000000", BVN: "95888168924", DOB: "1990-01-15",
		Address: "15 Admiralty Way", City: "Lagos", State: "Lagos",
	}, true)

	wA, err := w.Provision(context.Background(), "ins-a")
	if err != nil {
		t.Fatalf("provision A: %v", err)
	}
	wB, err := w.Provision(context.Background(), "ins-b")
	if err != nil {
		t.Fatalf("provision B: %v", err)
	}

	if wA.BmoniUserID == "" || wA.BmoniUserID == wB.BmoniUserID {
		t.Fatalf("instructors share a BMONI user: A=%q B=%q", wA.BmoniUserID, wB.BmoniUserID)
	}
	if wA.BmoniWalletID == wB.BmoniWalletID {
		t.Fatalf("instructors share a BMONI wallet: %q", wA.BmoniWalletID)
	}
	// The instructor's own phone (not the shared persona phone) is used.
	for i, p := range money.personas {
		if p.Phone == "+2348000000000" {
			t.Fatalf("persona[%d] reused the shared persona phone", i)
		}
	}
	if len(money.personas) != 2 {
		t.Fatalf("expected 2 provisions, got %d", len(money.personas))
	}
}

// An instructor with no phone must still get a deterministic unique phone, so
// re-provisioning the same instructor reuses the same BMONI user while two
// phone-less instructors never collide.
func TestUniquePhoneForIsDeterministicAndDistinct(t *testing.T) {
	a1 := uniquePhoneFor("ins-a")
	a2 := uniquePhoneFor("ins-a")
	b := uniquePhoneFor("ins-b")
	if a1 == "" || a1 != a2 {
		t.Fatalf("uniquePhoneFor not deterministic: %q vs %q", a1, a2)
	}
	if a1 == b {
		t.Fatalf("uniquePhoneFor collided: %q", a1)
	}
	for _, p := range []string{a1, b} {
		if len(p) != len("+234")+10 || p[:4] != "+234" {
			t.Fatalf("phone %q is not E.164 Nigerian format", p)
		}
	}
}

// Same-instructor re-provisioning (dashboard retry) is idempotent and reuses
// the same external identity.
func TestProvisionIdempotentPerInstructor(t *testing.T) {
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	newInstructor(t, st, clk, "ins-x", "x@kweeks.ng", "+2348033333333")

	money := &personaRecordingMoney{}
	w := NewWallet(st, clk, money).WithProvisioning(domain.BmoniPersona{
		FirstName: "Bunch", LastName: "Dillon", Email: "s@example.com",
		Phone: "+2348000000000", BVN: "95888168924", DOB: "1990-01-15",
		Address: "15 Admiralty Way", City: "Lagos", State: "Lagos",
	}, true)

	w1, err := w.Provision(context.Background(), "ins-x")
	if err != nil {
		t.Fatal(err)
	}
	w2, err := w.Provision(context.Background(), "ins-x")
	if err != nil {
		t.Fatal(err)
	}
	if w1.BmoniUserID != w2.BmoniUserID || w1.BmoniWalletID != w2.BmoniWalletID {
		t.Fatalf("re-provision changed identity: %+v vs %+v", w1, w2)
	}
	if len(money.personas) != 1 {
		t.Fatalf("re-provision provisioned again; want idempotent, got %d calls", len(money.personas))
	}
}
