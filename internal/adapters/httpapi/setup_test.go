package httpapi

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/osugodbless/kweeks/internal/adapters/bus"
	"github.com/osugodbless/kweeks/internal/adapters/clock"
	"github.com/osugodbless/kweeks/internal/adapters/store/memory"
	"github.com/osugodbless/kweeks/internal/app"
	"github.com/osugodbless/kweeks/internal/domain"
	"github.com/osugodbless/kweeks/internal/ports"
)

// wizardMoney drives the strict flow: create user → KYC → wallet → rail.
type wizardMoney struct {
	createdUser bool
	kyc         bool
	wallet      bool
	rail        bool
	kycSeen     domain.KYCProfile
	bvnSeen     string
	idSeen      domain.UserIdentity
}

func (f *wizardMoney) CreateUser(ctx context.Context, id domain.UserIdentity) (string, error) {
	f.createdUser = true
	f.idSeen = id
	return "usr_wiz", nil
}
func (f *wizardMoney) LookupBVN(ctx context.Context, userID, bvn string) (*domain.BVNRecord, error) {
	return &domain.BVNRecord{BVN: bvn, FirstName: "Bunch", LastName: "Dillon", DateOfBirth: "1990-01-15", Gender: "male", PhoneNumber: "+2348000000000"}, nil
}
func (f *wizardMoney) SubmitKYC(ctx context.Context, userID string, k domain.KYCProfile) error {
	f.kyc = true
	f.kycSeen = k
	return nil
}
func (f *wizardMoney) UploadKycDocument(ctx context.Context, userID, kind string, data []byte, filename string) error {
	return nil
}
func (f *wizardMoney) CreateWallet(ctx context.Context, userID string) (string, string, error) {
	f.wallet = true
	return "wal_wiz", "0xWiz", nil
}
func (f *wizardMoney) ActivateRail(ctx context.Context, userID, walletAddr, bvn string) error {
	f.rail = true
	f.bvnSeen = bvn
	return nil
}
func (f *wizardMoney) DepositAccount(ctx context.Context, userID, walletID string) (string, string, error) {
	return "9999999999", "Providus", nil
}
func (f *wizardMoney) ListNigerianBanks(ctx context.Context, userID string) ([]domain.NigerianBank, error) {
	return nil, nil
}
func (f *wizardMoney) VerifyNigerianAccount(ctx context.Context, userID, accountNumber, bankCode string) (string, error) {
	return "", nil
}
func (f *wizardMoney) RegisterNigerianWithdrawalAccount(ctx context.Context, userID string, acct domain.NigerianAccount) (string, error) {
	return "", nil
}
func (f *wizardMoney) PayWinnerToNigerianBank(ctx context.Context, from *domain.WalletExternal, bankAccountID string, amount domain.Amount) (string, error) {
	return "", nil
}

var _ ports.Money = (*wizardMoney)(nil)

func buildWizardServer(t *testing.T, money ports.Money) (*Server, *memory.Store) {
	t.Helper()
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	game := app.NewGame(st, clk, bus.NewInMemory())
	join := app.NewJoin(st, clk)
	red := app.NewRedemption(st, clk, money, nil)
	auth := app.NewAuth(st, clk)
	wallet := app.NewWallet(st, clk, money)
	auth.WithWalletProvisioning(func(ctx context.Context, instructorID string) error {
		_, err := wallet.CreateBmoniUser(ctx, instructorID)
		return err
	})
	api := New(game, join, red).WithServices(auth, wallet)
	return api, st
}

func wizardSignup(t *testing.T, api *Server) string {
	t.Helper()
	rr := authDo(api, "POST", "/api/auth/signup", map[string]string{
		"firstName": "Bunch", "lastName": "Dillon", "email": "wiz@kweeks.ng", "phone": "+2348000000000", "password": "secret1",
	}, "")
	if rr.Code != 200 {
		t.Fatalf("signup: %d %s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	decodeBody(t, rr, &body)
	return body["token"].(string)
}

func TestWalletSetupWizardEndpoints(t *testing.T) {
	money := &wizardMoney{}
	api, _ := buildWizardServer(t, money)
	token := wizardSignup(t, api)

	// signup auto-creates the BMONI user.
	if !money.createdUser {
		t.Fatalf("signup did not create the BMONI user")
	}

	// Lifecycle stage 2: stage is "wallet" after user creation (wallet precedes KYC).
	rr := authDo(api, "GET", "/api/wallet/setup", nil, token)
	if rr.Code != 200 {
		t.Fatalf("setup: %d %s", rr.Code, rr.Body.String())
	}
	var st map[string]any
	decodeBody(t, rr, &st)
	if st["stage"] != "wallet" {
		t.Fatalf("stage after signup = %v, want wallet (wallet precedes KYC)", st["stage"])
	}

	// create wallet (strict stage 2)
	if rr = authDo(api, "POST", "/api/wallet/create", nil, token); rr.Code != 200 {
		t.Fatalf("create wallet: %d %s", rr.Code, rr.Body.String())
	}

	// stage → kyc
	rr = authDo(api, "GET", "/api/wallet/setup", nil, token)
	decodeBody(t, rr, &st)
	if st["stage"] != "kyc" {
		t.Fatalf("stage after wallet = %v, want kyc", st["stage"])
	}

	// BVN lookup resolves + pre-fills (stage 3 helper).
	rr = authDo(api, "POST", "/api/wallet/kyc/lookup", map[string]string{"bvn": "95888168924"}, token)
	if rr.Code != 200 {
		t.Fatalf("bvn lookup: %d %s", rr.Code, rr.Body.String())
	}
	var rec map[string]any
	decodeBody(t, rr, &rec)
	if rec["firstName"] != "Bunch" || rec["lastName"] != "Dillon" {
		t.Fatalf("bvn lookup autofill wrong: %v", rec)
	}

	// Submit KYC with user-supplied values (host confirms the autofill).
	rr = authDo(api, "POST", "/api/wallet/kyc", map[string]any{
		"firstName": "Bunch", "lastName": "Dillon", "dateOfBirth": "1990-01-15",
		"gender": "male", "bvn": "95888168924", "street": "15 Admiralty Way",
		"city": "Lagos", "state": "Lagos", "postalCode": "101241",
	}, token)
	if rr.Code != 200 {
		t.Fatalf("kyc: %d %s", rr.Code, rr.Body.String())
	}
	if !money.kyc || money.kycSeen.BVN != "95888168924" {
		t.Fatalf("kyc not forwarded correctly: %+v", money.kycSeen)
	}

	// stage → rail
	rr = authDo(api, "GET", "/api/wallet/setup", nil, token)
	decodeBody(t, rr, &st)
	if st["stage"] != "rail" {
		t.Fatalf("stage after kyc = %v, want rail", st["stage"])
	}

	// activate rail with the host's BVN (stage 4)
	if rr = authDo(api, "POST", "/api/wallet/activate-rail", map[string]string{"bvn": "95888168924"}, token); rr.Code != 200 {
		t.Fatalf("activate rail: %d %s", rr.Code, rr.Body.String())
	}
	if money.bvnSeen != "95888168924" {
		t.Fatalf("activate rail bvn = %q", money.bvnSeen)
	}

	// stage → ready, with the NGN deposit account
	rr = authDo(api, "GET", "/api/wallet/setup", nil, token)
	if rr.Code != 200 {
		t.Fatalf("setup: %d", rr.Code)
	}
	decodeBody(t, rr, &st)
	if st["stage"] != "ready" {
		t.Fatalf("final stage = %v, want ready", st["stage"])
	}
	da := st["depositAccount"].(map[string]any)
	if da["accountNumber"] != "9999999999" {
		t.Fatalf("deposit account missing: %v", st["depositAccount"])
	}
}

func TestWalletSetupRequiresAuth(t *testing.T) {
	money := &wizardMoney{}
	api, _ := buildWizardServer(t, money)
	if rr := authDo(api, "GET", "/api/wallet/setup", nil, ""); rr.Code != http.StatusUnauthorized {
		t.Fatalf("setup without auth = %d, want 401", rr.Code)
	}
}
