package bmoni

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/osugodbless/kweeks/internal/domain"
)

const (
	ownerKeyForTest = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	knownDigest     = "8f5156823a5c2cdc7bedc12253e49e4946c6fff0273034eb485750035d21ad31"
)

func testPersona() domain.BmoniPersona {
	return domain.BmoniPersona{
		FirstName: "Samson", LastName: "Jabo", Email: "samson@example.com",
		Phone: "+2348000000001", BVN: "22222222222", DOB: "1990-01-15",
		Address: "15 Admiralty Way", City: "Lagos", State: "Lagos",
	}
}

func newMockRail(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler != nil {
			handler(w, r)
			return
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/v1/users") && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"user":{"bmoniUserId":"usr_demo_1"}}`))
		case strings.HasSuffix(r.URL.Path, "/kyc"):
			_, _ = w.Write([]byte(`{}`))
		case strings.HasSuffix(r.URL.Path, "/owner-proof-challenges"):
			_, _ = w.Write([]byte(`{"challengeId":"ch_1","message":"please prove you own this key"}`))
		case strings.HasSuffix(r.URL.Path, "/create-managed"):
			_, _ = w.Write([]byte(`{"id":"wal_1","currency":"NGN","walletAddress":"0xRecipientAddress","isActive":true}`))
		case strings.HasSuffix(r.URL.Path, "/onboarding/start-nigeria"):
			_, _ = w.Write([]byte(`{}`))
		case strings.HasSuffix(r.URL.Path, "/smart-wallets/account/wallets"):
			_, _ = w.Write([]byte(`[]`))
		case strings.HasSuffix(r.URL.Path, "/onboarding/status"):
			_, _ = w.Write([]byte(`{"status":"active"}`))
		case strings.Contains(r.URL.Path, "/kyc/documents/"):
			_, _ = w.Write([]byte(`{}`))
		case strings.HasSuffix(r.URL.Path, "/bank-accounts/nigerian-banks"):
			_, _ = w.Write([]byte(`{"banks":[{"code":"058","name":"Guaranty Trust Bank"},{"code":"044","name":"Access Bank"}]}`))
		case strings.HasSuffix(r.URL.Path, "/bank-accounts/verify-nigerian-account"):
			_, _ = w.Write([]byte(`{"accountHolderName":"Jane Doe"}`))
		case strings.HasSuffix(r.URL.Path, "/bank-accounts/withdrawal-accounts/nigeria"):
			_, _ = w.Write([]byte(`{"id":"ba_1"}`))
		case strings.HasSuffix(r.URL.Path, "/onramp/vba/nigeria"):
			_, _ = w.Write([]byte(`{}`))
		case strings.HasSuffix(r.URL.Path, "/bank-accounts/deposit-accounts/NGN"):
			_, _ = w.Write([]byte(`{"accounts":[{"id":"vba_1","accountNumber":"0123456789","bankName":"Providus Bank","currency":"NGN","targetCurrency":"NGN"}]}`))
		case strings.HasSuffix(r.URL.Path, "/offramp/nigeria"):
			_, _ = w.Write([]byte(`{"data":{"proposalId":"prop_offramp","status":"PENDING_APPROVALS"}}`))
		case strings.HasSuffix(r.URL.Path, "/proposals/approve"):
			_, _ = w.Write([]byte(`{}`))
		case strings.HasSuffix(r.URL.Path, "/sign-payload"):
			_, _ = w.Write([]byte(`{"data":{"hashToSign":"` + knownDigest + `"}}`))
		case strings.HasSuffix(r.URL.Path, "/proposals/sign"):
			_, _ = w.Write([]byte(`{"data":{"proposal":{"id":"prop_offramp","status":"COMPLETED"}}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"no mock for this path"}`))
		}
	}))
	c := New(srv.URL, "pk_test", ownerKeyForTest)
	return srv, c
}

func TestProvisionHappyPath(t *testing.T) {
	srv, c := newMockRail(t, nil)
	defer srv.Close()

	ext, err := c.Provision(context.Background(), testPersona())
	if err != nil {
		t.Fatalf("provision: %v", err)
	}
	if ext.UserID != "usr_demo_1" || ext.WalletID != "wal_1" || ext.Address != "0xRecipientAddress" {
		t.Fatalf("provision result mismatch: %+v", ext)
	}
}

func TestCreateUserConflictIsSurfaced(t *testing.T) {
	srv, c := newMockRail(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"statusCode":409,"message":"User already exists with this phoneNumber","error":"Conflict"}`))
	})
	defer srv.Close()
	if _, err := c.CreateUser(context.Background(), testPersona()); err == nil {
		t.Fatalf("expected conflict error")
	}
}

func TestProvisionWalletSignsOwnerProofEIP191(t *testing.T) {
	var gotProof string
	srv, c := newMockRail(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/create-managed") {
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			gotProof = body["ownerProofSignature"]
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/smart-wallets/account/wallets"):
			_, _ = w.Write([]byte(`[]`))
		case strings.HasSuffix(r.URL.Path, "/owner-proof-challenges"):
			_, _ = w.Write([]byte(`{"challengeId":"ch_1","message":"prove it"}`))
		default:
			_, _ = w.Write([]byte(`{"id":"wal_1","currency":"NGN","walletAddress":"0xRecipientAddress"}`))
		}
	})
	defer srv.Close()

	if _, _, err := c.ProvisionWallet(context.Background(), "usr_demo_1"); err != nil {
		t.Fatalf("provision wallet: %v", err)
	}
	if !strings.HasPrefix(gotProof, "0x") || len(gotProof) != 132 {
		t.Fatalf("ownerProofSignature malformed: %q", gotProof)
	}
}

func TestListNigerianBanks(t *testing.T) {
	srv, c := newMockRail(t, nil)
	defer srv.Close()

	banks, err := c.ListNigerianBanks(context.Background(), "usr_demo_1")
	if err != nil {
		t.Fatalf("banks: %v", err)
	}
	if len(banks) != 2 || banks[0].Code != "058" || banks[0].Name != "Guaranty Trust Bank" {
		t.Fatalf("banks mismatch: %+v", banks)
	}
}

func TestVerifyAndRegisterNigerianAccount(t *testing.T) {
	srv, c := newMockRail(t, nil)
	defer srv.Close()

	holder, err := c.VerifyNigerianAccount(context.Background(), "usr_demo_1", "0123456789", "058")
	if err != nil || holder != "Jane Doe" {
		t.Fatalf("verify: %v holder=%q", err, holder)
	}

	id, err := c.RegisterNigerianWithdrawalAccount(context.Background(), "usr_demo_1", domain.NigerianAccount{
		AccountNumber: "0123456789", BankCode: "058", BankName: "Guaranty Trust Bank", AccountHolderName: "Jane Doe",
	})
	if err != nil || id != "ba_1" {
		t.Fatalf("register: %v id=%q", err, id)
	}
}

func TestPayWinnerToNigerianBankOfframp(t *testing.T) {
	var offrampBody map[string]string
	srv, c := newMockRail(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/offramp/nigeria") {
			_ = json.NewDecoder(r.Body).Decode(&offrampBody)
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/offramp/nigeria"):
			_, _ = w.Write([]byte(`{"data":{"proposalId":"prop_offramp","status":"PENDING_APPROVALS"}}`))
		case strings.HasSuffix(r.URL.Path, "/approve"):
			_, _ = w.Write([]byte(`{}`))
		case strings.HasSuffix(r.URL.Path, "/sign-payload"):
			_, _ = w.Write([]byte(`{"data":{"hashToSign":"` + knownDigest + `"}}`))
		case strings.HasSuffix(r.URL.Path, "/sign"):
			_, _ = w.Write([]byte(`{"data":{"proposal":{"id":"prop_offramp","status":"COMPLETED"}}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()

	ext := &domain.WalletExternal{UserID: "usr_demo_1", WalletID: "wal_1", Address: "0xRecipientAddress"}
	ref, err := c.PayWinnerToNigerianBank(context.Background(), ext, "ba_1", 2500)
	if err != nil {
		t.Fatalf("offramp: %v", err)
	}
	if ref != "prop_offramp" {
		t.Fatalf("ref = %q", ref)
	}
	if offrampBody["bankAccountId"] != "ba_1" || offrampBody["fromAmount"] != "25.00" {
		t.Fatalf("offramp body mismatch: %+v", offrampBody)
	}
}

func TestPayWinnerToNigerianBankRequiresProvisionedWallet(t *testing.T) {
	srv, c := newMockRail(t, nil)
	defer srv.Close()
	if _, err := c.PayWinnerToNigerianBank(context.Background(), &domain.WalletExternal{}, "ba_1", 2500); err == nil {
		t.Fatalf("expected provisioning error for empty wallet")
	}
}

func TestDepositAccount(t *testing.T) {
	srv, c := newMockRail(t, nil)
	defer srv.Close()

	number, bank, err := c.DepositAccount(context.Background(), "usr_demo_1", "wal_1")
	if err != nil {
		t.Fatalf("deposit account: %v", err)
	}
	if number != "0123456789" || bank != "Providus Bank" {
		t.Fatalf("deposit mismatch: %q %q", number, bank)
	}
}

func TestUploadDocumentMultipart(t *testing.T) {
	srv, c := newMockRail(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Fatalf("expected multipart upload, got %q", r.Header.Get("Content-Type"))
		}
		_, _ = w.Write([]byte(`{}`))
	})
	defer srv.Close()

	dir := t.TempDir()
	path := filepath.Join(dir, "id.jpg")
	if err := os.WriteFile(path, []byte("fakejpeg"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := c.UploadDocument(context.Background(), "usr_demo_1", "identification", path); err != nil {
		t.Fatalf("upload: %v", err)
	}
}
