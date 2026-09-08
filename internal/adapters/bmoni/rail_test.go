package bmoni

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/osugodbless/kweeks/internal/domain"
)

const (
	ownerKeyForTest = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	knownDigest     = "8f5156823a5c2cdc7bedc12253e49e4946c6fff0273034eb485750035d21ad31"
)

func testIdentity() domain.UserIdentity {
	return domain.UserIdentity{FirstName: "Samson", LastName: "Jabo", Email: "samson@example.com", Phone: "+2348000000001"}
}

func testKYC() domain.KYCProfile {
	return domain.KYCProfile{
		FirstName: "Samson", LastName: "Jabo", Phone: "+2348000000001",
		DateOfBirth: "1990-01-15", Gender: "male", BVN: "22222222222",
		Street: "15 Admiralty Way", City: "Lagos", State: "Lagos", PostalCode: "101241",
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
		case strings.HasSuffix(r.URL.Path, "/kyc") && r.Method == http.MethodPatch:
			_, _ = w.Write([]byte(`{}`))
		case strings.Contains(r.URL.Path, "/kyc/bvn-lookup/"):
			_, _ = w.Write([]byte(`{"bvn":"95888168924","firstName":"Bunch","lastName":"Dillon","dateOfBirth":"1990-01-15","gender":"male","phoneNumber":"+2348000000000","nin":"63184876213"}`))
		case strings.HasSuffix(r.URL.Path, "/owner-proof-challenges"):
			_, _ = w.Write([]byte(`{"challengeId":"ch_1","message":"please prove you own this key"}`))
		case strings.HasSuffix(r.URL.Path, "/smart-wallets/account/wallets"):
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":["No embedded smart wallet group found for this user. Call POST .../owner-proof-challenges first."],"error":"Bad Request","statusCode":400}`))
		case strings.HasSuffix(r.URL.Path, "/create-managed"):
			_, _ = w.Write([]byte(`{"id":"wal_1","currency":"NGN","walletAddress":"0xRecipientAddress","isActive":true}`))
		case strings.HasSuffix(r.URL.Path, "/onboarding/start-nigeria"):
			_, _ = w.Write([]byte(`{}`))
		case strings.Contains(r.URL.Path, "/kyc/documents/"):
			_, _ = w.Write([]byte(`{}`))
		case strings.HasSuffix(r.URL.Path, "/bank-accounts/nigerian-banks"):
			_, _ = w.Write([]byte(`{"banks":[{"code":"058","name":"Guaranty Trust Bank"},{"code":"044","name":"Access Bank"}]}`))
		case strings.HasSuffix(r.URL.Path, "/bank-accounts/verify-nigerian-account"):
			_, _ = w.Write([]byte(`{"accountHolderName":"Jane Doe"}`))
		case strings.HasSuffix(r.URL.Path, "/bank-accounts/withdrawal-accounts/nigeria"):
			_, _ = w.Write([]byte(`{"id":"ba_1"}`))
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

func TestCreateUser(t *testing.T) {
	srv, c := newMockRail(t, nil)
	defer srv.Close()
	userID, err := c.CreateUser(context.Background(), testIdentity())
	if err != nil || userID != "usr_demo_1" {
		t.Fatalf("create user: %v id=%q", err, userID)
	}
}

func TestCreateUserConflictRecoversExisting(t *testing.T) {
	var gotBody map[string]string
	srv, c := newMockRail(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/v1/users") && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"message":"User already exists with this phone","error":"Conflict","statusCode":409}`))
		case strings.HasSuffix(r.URL.Path, "/v1/users") && r.Method == http.MethodGet:
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			_, _ = w.Write([]byte(`{"users":[{"bmoniUserId":"usr_recovered","phoneNumber":"+2348000000001","email":"samson@example.com"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()
	userID, err := c.CreateUser(context.Background(), testIdentity())
	if err != nil {
		t.Fatalf("recover on conflict: %v", err)
	}
	if userID != "usr_recovered" {
		t.Fatalf("recovered user = %q", userID)
	}
}

func TestSubmitKYC(t *testing.T) {
	var got map[string]any
	srv, c := newMockRail(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/kyc") && r.Method == http.MethodPatch {
			_ = json.NewDecoder(r.Body).Decode(&got)
		}
		_, _ = w.Write([]byte(`{}`))
	})
	defer srv.Close()
	if err := c.SubmitKYC(context.Background(), "usr_demo_1", testKYC()); err != nil {
		t.Fatalf("submit kyc: %v", err)
	}
	ident, _ := got["identificationNumbers"].([]any)
	if len(ident) != 1 {
		t.Fatalf("missing bvn identification: %v", got)
	}
}

func TestSubmitKYCRejectsBadBVN(t *testing.T) {
	srv, c := newMockRail(t, nil)
	defer srv.Close()
	k := testKYC()
	k.BVN = "123"
	if err := c.SubmitKYC(context.Background(), "usr_demo_1", k); err == nil {
		t.Fatalf("expected bvn length error")
	}
}

func TestLookupBVNResolvesHolder(t *testing.T) {
	srv, c := newMockRail(t, nil)
	defer srv.Close()
	rec, err := c.LookupBVN(context.Background(), "usr_demo_1", "95888168924")
	if err != nil {
		t.Fatalf("lookup bvn: %v", err)
	}
	if rec.FirstName != "Bunch" || rec.LastName != "Dillon" || rec.DateOfBirth != "1990-01-15" {
		t.Fatalf("bvn holder mismatch: %+v", rec)
	}
	if rec.NIN != "63184876213" {
		t.Fatalf("nin not surfaced: %+v", rec)
	}
	if _, err := c.LookupBVN(context.Background(), "usr_demo_1", "123"); err == nil {
		t.Fatalf("expected bvn length error")
	}
}

func TestCreateWalletSignsOwnerProofEIP191(t *testing.T) {
	var gotProof string
	srv, c := newMockRail(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/create-managed") {
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			gotProof = body["ownerProofSignature"]
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/smart-wallets/account/wallets"):
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":["No embedded smart wallet group found for this user. Call POST .../owner-proof-challenges first."],"error":"Bad Request","statusCode":400}`))
		case strings.HasSuffix(r.URL.Path, "/owner-proof-challenges"):
			_, _ = w.Write([]byte(`{"challengeId":"ch_1","message":"prove it"}`))
		default:
			_, _ = w.Write([]byte(`{"id":"wal_1","currency":"NGN","walletAddress":"0xRecipientAddress"}`))
		}
	})
	defer srv.Close()

	walletID, addr, err := c.CreateWallet(context.Background(), "usr_demo_1")
	if err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	if walletID != "wal_1" || addr != "0xRecipientAddress" {
		t.Fatalf("wallet mismatch: %q %q", walletID, addr)
	}
	if !strings.HasPrefix(gotProof, "0x") || len(gotProof) != 132 {
		t.Fatalf("ownerProofSignature malformed: %q", gotProof)
	}
}

func TestActivateRail(t *testing.T) {
	srv, c := newMockRail(t, nil)
	defer srv.Close()
	if err := c.ActivateRail(context.Background(), "usr_demo_1", "0xRecipientAddress", "22222222222"); err != nil {
		t.Fatalf("activate rail: %v", err)
	}
	if err := c.ActivateRail(context.Background(), "usr_demo_1", "0xRecipientAddress", "bad"); err == nil {
		t.Fatalf("expected bvn length error")
	}
}

func TestUploadKycDocumentMultipart(t *testing.T) {
	var bodyTxt string
	srv, c := newMockRail(t, func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "multipart/form-data") {
			t.Fatalf("expected multipart upload, got %q", ct)
		}
		if err := r.ParseMultipartForm(2 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		// Image must ride under the `files` field (not `file`), with type metadata.
		if len(r.MultipartForm.File["files"]) == 0 {
			t.Fatalf("image not under the `files` field: %v", r.MultipartForm.File)
		}
		if got := r.FormValue("type"); got != "passport" {
			t.Fatalf("type = %q", got)
		}
		if got := r.FormValue("documentNumber"); got != "A12345678" {
			t.Fatalf("documentNumber = %q", got)
		}
		bodyTxt = "ok"
		_, _ = w.Write([]byte(`{}`))
	})
	defer srv.Close()
	err := c.UploadKycDocument(context.Background(), "usr_demo_1", domain.KycDocument{
		Kind: "identification", Data: []byte("fakejpeg"), Name: "id.jpg",
		Type: "passport", DocumentNumber: "A12345678", IssuingCountry: "NGA",
	})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if bodyTxt != "ok" {
		t.Fatalf("upload did not reach the handler")
	}
}

func TestUploadKycDocumentRejectsMissingMetadata(t *testing.T) {
	srv, c := newMockRail(t, nil)
	defer srv.Close()
	err := c.UploadKycDocument(context.Background(), "usr_demo_1", domain.KycDocument{Kind: "identification", Data: []byte("x"), Type: "passport"})
	if err == nil || !strings.Contains(err.Error(), "documentNumber") {
		t.Fatalf("expected missing documentNumber error, got %v", err)
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

func TestCreateWalletReusesExistingCNGN(t *testing.T) {
	calledCreateManaged := false
	srv, c := newMockRail(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/smart-wallets/account/wallets"):
			_, _ = w.Write([]byte(`[{"id":"wal_existing","currency":"NGN","walletAddress":"0xExisting","isActive":true}]`))
		case strings.HasSuffix(r.URL.Path, "/create-managed"):
			calledCreateManaged = true
			_, _ = w.Write([]byte(`{}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()

	walletID, addr, err := c.CreateWallet(context.Background(), "usr_demo_1")
	if err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	if walletID != "wal_existing" || addr != "0xExisting" {
		t.Fatalf("did not reuse existing wallet: %q %q", walletID, addr)
	}
	if calledCreateManaged {
		t.Fatalf("create-managed called despite existing wallet (would 409)")
	}
}

func TestCreateWalletProceedsWhenNoWalletGroup(t *testing.T) {
	srv, c := newMockRail(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/smart-wallets/account/wallets"):
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":["No embedded smart wallet group found for this user. Call POST .../owner-proof-challenges first."],"error":"Bad Request","statusCode":400}`))
		case strings.HasSuffix(r.URL.Path, "/owner-proof-challenges"):
			_, _ = w.Write([]byte(`{"challengeId":"ch_1","message":"prove it"}`))
		case strings.HasSuffix(r.URL.Path, "/create-managed"):
			_, _ = w.Write([]byte(`{"id":"wal_new","currency":"NGN","walletAddress":"0xNew"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer srv.Close()

	walletID, addr, err := c.CreateWallet(context.Background(), "usr_demo_1")
	if err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	if walletID != "wal_new" || addr != "0xNew" {
		t.Fatalf("wallet mismatch: %q %q", walletID, addr)
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
