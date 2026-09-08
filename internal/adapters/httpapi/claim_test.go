package httpapi

import (
	"context"
	"encoding/json"
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

type httpFakeMoney struct {
	provisioned *domain.WalletExternal
	banks       []domain.NigerianBank
	holderName  string
	bankAcctID  string
	payoutRef   string
}

func (f *httpFakeMoney) Provision(ctx context.Context, p domain.BmoniPersona) (*domain.WalletExternal, error) {
	return f.provisioned, nil
}
func (f *httpFakeMoney) ListNigerianBanks(ctx context.Context, userID string) ([]domain.NigerianBank, error) {
	return f.banks, nil
}
func (f *httpFakeMoney) VerifyNigerianAccount(ctx context.Context, userID, accountNumber, bankCode string) (string, error) {
	return f.holderName, nil
}
func (f *httpFakeMoney) RegisterNigerianWithdrawalAccount(ctx context.Context, userID string, acct domain.NigerianAccount) (string, error) {
	return f.bankAcctID, nil
}
func (f *httpFakeMoney) PayWinnerToNigerianBank(ctx context.Context, from *domain.WalletExternal, bankAccountID string, amount domain.Amount) (string, error) {
	return f.payoutRef, nil
}
func (f *httpFakeMoney) DepositAccount(ctx context.Context, userID, walletID string) (string, string, error) {
	return "0123456789", "Providus", nil
}

var _ ports.Money = (*httpFakeMoney)(nil)

func buildClaimServer(t *testing.T, money ports.Money) (*Server, *memory.Store) {
	t.Helper()
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	game := app.NewGame(st, clk, bus.NewInMemory())
	join := app.NewJoin(st, clk)
	red := app.NewRedemption(st, clk, money, nil)
	return New(game, join, red), st
}

func seedClaimQuiz(t *testing.T, api *Server, st *memory.Store, clk *clock.Static) (roomID, winnerEmail, claimCode string) {
	t.Helper()
	ctx := context.Background()
	// The quiz from seedQuizRoom is owned by "instructor-demo", so the
	// provisioned wallet that pays the winner must belong to that same host.
	ins := &domain.Instructor{ID: "instructor-demo", Name: "Adeola", Email: "host@kweeks.ng", Avatar: "AP", CreatedAt: clk.Now()}
	if err := st.CreateInstructor(ctx, ins); err != nil {
		t.Fatal(err)
	}
	if err := st.CreateWallet(ctx, &domain.Wallet{
		ID: "wal-http", InstructorID: "instructor-demo", Balance: 100000,
		BmoniUserID: "usr_host", BmoniWalletID: "wal_host", BmoniWalletAddr: "0xhost", CreatedAt: clk.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	roomID, _ = seedQuizRoom(t, api, st, clk, true)

	// Create the claim via the redeem endpoint.
	rr := doJSON(t, api, "POST", "/api/rooms/"+roomID+"/redeem", map[string]string{"email": "alice@x.com"})
	if rr.Code != http.StatusOK {
		t.Fatalf("redeem: %d %s", rr.Code, rr.Body.String())
	}
	var claim struct {
		ClaimCode string `json:"claimCode"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &claim)
	return roomID, "alice@x.com", claim.ClaimCode
}

func TestResolveClaimEndpoint(t *testing.T) {
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	game := app.NewGame(st, clk, bus.NewInMemory())
	join := app.NewJoin(st, clk)
	money := &httpFakeMoney{
		provisioned: &domain.WalletExternal{UserID: "usr_host", WalletID: "wal_host"},
		banks:       []domain.NigerianBank{{Code: "058", Name: "GTB"}},
		holderName:  "Alice Winner", bankAcctID: "ba_1", payoutRef: "prop_1",
	}
	api := New(game, join, app.NewRedemption(st, clk, money, nil))
	_, _, claimCode := seedClaimQuiz(t, api, st, clk)

	rr := doJSON(t, api, "POST", "/api/claims/resolve", map[string]string{"claimCode": claimCode, "email": "alice@x.com"})
	if rr.Code != http.StatusOK {
		t.Fatalf("resolve: %d %s", rr.Code, rr.Body.String())
	}
	var out struct {
		Claim struct {
			State string `json:"state"`
		} `json:"claim"`
		Banks []struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"banks"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	if out.Claim.State != "created" {
		t.Fatalf("claim state = %q", out.Claim.State)
	}
	if len(out.Banks) != 1 || out.Banks[0].Code != "058" {
		t.Fatalf("banks mismatch: %+v", out.Banks)
	}

	// Wrong email -> 403.
	rr2 := doJSON(t, api, "POST", "/api/claims/resolve", map[string]string{"claimCode": claimCode, "email": "mallory@x.com"})
	if rr2.Code != http.StatusForbidden {
		t.Fatalf("wrong email resolve got %d, want 403", rr2.Code)
	}
}

func TestSubmitPayoutEndpoint(t *testing.T) {
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	game := app.NewGame(st, clk, bus.NewInMemory())
	join := app.NewJoin(st, clk)
	money := &httpFakeMoney{
		provisioned: &domain.WalletExternal{UserID: "usr_host", WalletID: "wal_host"},
		banks:       []domain.NigerianBank{{Code: "058", Name: "GTB"}},
		holderName:  "Alice Winner", bankAcctID: "ba_1", payoutRef: "prop_1",
	}
	api := New(game, join, app.NewRedemption(st, clk, money, nil))
	_, email, claimCode := seedClaimQuiz(t, api, st, clk)

	rr := doJSON(t, api, "POST", "/api/claims/payout", map[string]any{
		"claimCode": claimCode, "email": email,
		"accountNumber": "0123456789", "bankCode": "058", "bankName": "GTB",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("payout: %d %s", rr.Code, rr.Body.String())
	}
	var out struct {
		Claim struct {
			State     string `json:"state"`
			PayoutRef string `json:"payoutRef"`
		} `json:"claim"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	if out.Claim.State != "paid" || out.Claim.PayoutRef != "prop_1" {
		t.Fatalf("claim mismatch: %+v", out.Claim)
	}
}
