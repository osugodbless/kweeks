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
// user identity phone, so two instructors with the same phone would get the
// SAME external id (the shared-identity bug) and different phones get
// different ids.
type personaRecordingMoney struct {
	users []domain.UserIdentity
}

func (f *personaRecordingMoney) CreateUser(ctx context.Context, id domain.UserIdentity) (string, error) {
	f.users = append(f.users, id)
	return "usr-" + id.Phone, nil
}
func (f *personaRecordingMoney) SubmitKYC(ctx context.Context, userID string, k domain.KYCProfile) error {
	return nil
}
func (f *personaRecordingMoney) UploadKycDocument(ctx context.Context, userID, kind string, data []byte, filename string) error {
	return nil
}
func (f *personaRecordingMoney) CreateWallet(ctx context.Context, userID string) (string, string, error) {
	return "wal-" + userID, "0x" + userID, nil
}
func (f *personaRecordingMoney) ActivateRail(ctx context.Context, userID, walletAddr, bvn string) error {
	return nil
}
func (f *personaRecordingMoney) DepositAccount(ctx context.Context, userID, walletID string) (string, string, error) {
	return "0123456789", "Providus", nil
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

var _ ports.Money = (*personaRecordingMoney)(nil)

func newInstructor(t *testing.T, st *memory.Store, clk *clock.Static, id, email, phone string) {
	t.Helper()
	if err := st.CreateInstructor(context.Background(), &domain.Instructor{
		ID: id, Name: "Adeola Peters", Email: email, Phone: phone, PasswordHash: "h", Avatar: "AP", CreatedAt: clk.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.CreateWallet(context.Background(), &domain.Wallet{ID: "wal-" + id, InstructorID: id, CreatedAt: clk.Now()}); err != nil {
		t.Fatal(err)
	}
}

// Two instructors must provision DISTINCT BMONI users — never a shared identity.
func TestProvisionAssignsDistinctBmoniIdentity(t *testing.T) {
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	newInstructor(t, st, clk, "ins-a", "a@kweeks.ng", "+2348011111111")
	newInstructor(t, st, clk, "ins-b", "b@kweeks.ng", "+2348022222222")

	money := &personaRecordingMoney{}
	w := NewWallet(st, clk, money)

	wA, err := w.CreateBmoniUser(context.Background(), "ins-a")
	if err != nil {
		t.Fatalf("create A: %v", err)
	}
	wB, err := w.CreateBmoniUser(context.Background(), "ins-b")
	if err != nil {
		t.Fatalf("create B: %v", err)
	}
	if wA.BmoniUserID == "" || wA.BmoniUserID == wB.BmoniUserID {
		t.Fatalf("instructors share a BMONI user: A=%q B=%q", wA.BmoniUserID, wB.BmoniUserID)
	}
	// The instructor's own phone (not a shared persona phone) is used.
	for _, id := range money.users {
		if id.Phone == "" {
			t.Fatalf("identity used an empty phone: %+v", id)
		}
	}
	if len(money.users) != 2 {
		t.Fatalf("expected 2 create-user calls, got %d", len(money.users))
	}
}

// The full strict-flow wizard: create user → KYC → wallet → rail → ready.
func TestWalletSetupWizardFullFlow(t *testing.T) {
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	newInstructor(t, st, clk, "ins-1", "host@kweeks.ng", "+2348033333333")

	money := &personaRecordingMoney{}
	w := NewWallet(st, clk, money)

	st0, err := w.SetupStatus(context.Background(), "ins-1")
	if err != nil {
		t.Fatal(err)
	}
	if st0.Stage != domain.SetupUnprovisioned {
		t.Fatalf("initial stage = %s", st0.Stage)
	}

	if _, err := w.CreateBmoniUser(context.Background(), "ins-1"); err != nil {
		t.Fatal(err)
	}
	st1, _ := w.SetupStatus(context.Background(), "ins-1")
	if st1.Stage != domain.SetupKYC {
		t.Fatalf("after user: stage = %s, want kyc", st1.Stage)
	}

	if _, err := w.SubmitKYC(context.Background(), "ins-1", domain.KYCProfile{
		FirstName: "Adeola", LastName: "Peters", DateOfBirth: "1990-01-15", Gender: "male",
		BVN: "22222222222", Street: "15 Admiralty Way", City: "Lagos", State: "Lagos", PostalCode: "101241",
	}); err != nil {
		t.Fatal(err)
	}
	st2, _ := w.SetupStatus(context.Background(), "ins-1")
	if st2.Stage != domain.SetupWallet {
		t.Fatalf("after kyc: stage = %s, want wallet", st2.Stage)
	}

	if _, err := w.CreateWallet(context.Background(), "ins-1"); err != nil {
		t.Fatal(err)
	}
	st3, _ := w.SetupStatus(context.Background(), "ins-1")
	if st3.Stage != domain.SetupRail {
		t.Fatalf("after wallet: stage = %s, want rail", st3.Stage)
	}

	if _, err := w.ActivateRail(context.Background(), "ins-1", "22222222222"); err != nil {
		t.Fatal(err)
	}
	st4, err := w.SetupStatus(context.Background(), "ins-1")
	if err != nil {
		t.Fatal(err)
	}
	if st4.Stage != domain.SetupReady {
		t.Fatalf("after rail: stage = %s, want ready", st4.Stage)
	}
	if st4.DepositAccountNumber != "0123456789" {
		t.Fatalf("ready deposit account missing: %+v", st4)
	}
}

// SetupStatus must report a "ready" wallet as ready without re-provisioning.
func TestWalletSetupStatusReadyIdempotent(t *testing.T) {
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	newInstructor(t, st, clk, "ins-x", "x@kweeks.ng", "+2348044444444")
	money := &personaRecordingMoney{}
	w := NewWallet(st, clk, money)
	for _, step := range []func() error{
		func() error { _, err := w.CreateBmoniUser(context.Background(), "ins-x"); return err },
		func() error {
			_, err := w.SubmitKYC(context.Background(), "ins-x", domain.KYCProfile{
				FirstName: "Adeola", LastName: "Peters", DateOfBirth: "1990-01-15", Gender: "male",
				BVN: "22222222222", Street: "15 Admiralty Way", City: "Lagos", State: "Lagos", PostalCode: "101241",
			})
			return err
		},
		func() error { _, err := w.CreateWallet(context.Background(), "ins-x"); return err },
		func() error { _, err := w.ActivateRail(context.Background(), "ins-x", "22222222222"); return err },
	} {
		if err := step(); err != nil {
			t.Fatal(err)
		}
	}
	st2, _ := w.SetupStatus(context.Background(), "ins-x")
	if st2.Stage != domain.SetupReady {
		t.Fatalf("expected ready, got %s", st2.Stage)
	}
	// Re-running each step must be idempotent.
	if _, err := w.ActivateRail(context.Background(), "ins-x", "22222222222"); err != nil {
		t.Fatalf("re-activate rail errored: %v", err)
	}
}

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

func TestIdentityForUsesExplicitNameParts(t *testing.T) {
	got := identityFor(domain.Instructor{
		FirstName: "Chiamaka", LastName: "Okafor-Osei",
		Name: "Chiamaka Okafor-Osei", Email: "c@kweeks.ng", Phone: "+2348012345678",
	})
	if got.FirstName != "Chiamaka" || got.LastName != "Okafor-Osei" {
		t.Fatalf("explicit name parts not used verbatim: %+v", got)
	}
}

func TestIdentityForFallsBackToSplitForLegacyAccounts(t *testing.T) {
	got := identityFor(domain.Instructor{
		Name: "Adeola Peters Okafor", Email: "a@kweeks.ng", Phone: "+2348011111111",
	})
	if got.FirstName != "Adeola" || got.LastName != "Peters Okafor" {
		t.Fatalf("legacy split wrong: first=%q last=%q", got.FirstName, got.LastName)
	}
}
