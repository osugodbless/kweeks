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
		case strings.HasSuffix(r.URL.Path, "/proposals/approve"):
			_, _ = w.Write([]byte(`{}`))
		case strings.HasSuffix(r.URL.Path, "/sign-payload"):
			_, _ = w.Write([]byte(`{"data":{"hashToSign":"` + knownDigest + `"}}`))
		case strings.HasSuffix(r.URL.Path, "/proposals/sign"):
			_, _ = w.Write([]byte(`{"data":{"proposal":{"id":"prop_1","status":"COMPLETED"}}}`))
		case strings.Contains(r.URL.Path, "/proposals") && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"proposal":{"id":"prop_1","groupWalletId":"wal_1","status":"PENDING_APPROVALS"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"no mock for this path"}`))
		}
	}))
	c := New(srv.URL, "pk_test", ownerKeyForTest, "usr_platform", "wal_platform")
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

func TestCreditAndPayFlow(t *testing.T) {
	srv, c := newMockRail(t, nil)
	defer srv.Close()

	ext := &domain.WalletExternal{UserID: "usr_demo_1", WalletID: "wal_1", Address: "0xRecipientAddress"}
	ref, err := c.CreditTo(context.Background(), ext.UserID, ext.WalletID, "100.00")
	if err != nil || ref == "" {
		t.Fatalf("credit: %v ref=%q", err, ref)
	}
	ref2, err := c.PayTo(context.Background(), ext.UserID, ext.WalletID, "0xRecipientAddress", "25.00")
	if err != nil || ref2 == "" {
		t.Fatalf("pay: %v ref=%q", err, ref2)
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

func TestPayWinnerLiveShape(t *testing.T) {
	var proposalBody map[string]any
	srv, c := newMockRail(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/proposals") && r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&proposalBody)
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/proposals") && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"proposal":{"id":"prop_live","groupWalletId":"wal_platform","status":"PENDING_APPROVALS"}}`))
		case strings.HasSuffix(r.URL.Path, "/approve"):
			_, _ = w.Write([]byte(`{}`))
		case strings.HasSuffix(r.URL.Path, "/sign-payload"):
			_, _ = w.Write([]byte(`{"data":{"hashToSign":"` + knownDigest + `"}}`))
		case strings.HasSuffix(r.URL.Path, "/sign"):
			_, _ = w.Write([]byte(`{"data":{"proposal":{"id":"prop_live","status":"COMPLETED"}}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()

	ref, err := c.PayWinner(context.Background(), "winner-user-id", 2500)
	if err != nil {
		t.Fatalf("paywinner: %v", err)
	}
	if ref != "prop_live" {
		t.Fatalf("ref = %q", ref)
	}
	inner, _ := proposalBody["proposal"].(map[string]any)
	if inner == nil || inner["toUserId"] != "winner-user-id" {
		t.Fatalf("proposal recipient mismatch: %v", proposalBody)
	}
	if inner["currency"] != "CNGN" {
		t.Fatalf("proposal currency = %v", inner["currency"])
	}
}
