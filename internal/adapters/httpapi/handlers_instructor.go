package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/osugodbless/kweeks/internal/app"
	"github.com/osugodbless/kweeks/internal/domain"
)

type ctxKey int

const instructorKey ctxKey = 1

// instructorFrom resolves the acting instructor. When the auth service is
// wired (production) it returns the identity requireAuth placed in the
// context. In unit tests (no auth service) it returns the demo instructor,
// preserving the pre-auth route behaviour.
func instructorFrom(r *http.Request) string {
	if id, ok := r.Context().Value(instructorKey).(string); ok && id != "" {
		return id
	}
	return "instructor-demo"
}

// requireAuth wraps an instructor-scoped handler. Without an auth service the
// route is open to the demo instructor (test compatibility); with one, a
// Bearer token must resolve to a valid session.
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.auth == nil {
			next(w, r)
			return
		}
		token := bearerToken(r)
		instructor, _, err := s.auth.Resolve(r.Context(), token)
		if err != nil {
			writeErr(w, domain.ErrUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), instructorKey, instructor.ID)
		next(w, r.WithContext(ctx))
	}
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if len(h) > 7 && strings.EqualFold(h[:7], "Bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

// ---- Auth ----

type credentialsReq struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Password  string `json:"password"`
}

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	var req credentialsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, domain.ErrBadCredentials)
		return
	}
	res, err := s.auth.Signup(r.Context(), req.FirstName, req.LastName, req.Email, req.Phone, req.Password)
	if err != nil {
		writeErr(w, err)
		return
	}
	s.writeAuthResult(w, res)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req credentialsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, domain.ErrBadCredentials)
		return
	}
	res, err := s.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeErr(w, err)
		return
	}
	s.writeAuthResult(w, res)
}

func (s *Server) writeAuthResult(w http.ResponseWriter, res *app.SignupResult) {
	writeJSON(w, http.StatusOK, map[string]any{
		"token": res.Token,
		"instructor": map[string]any{
			"id": res.Instructor.ID, "firstName": res.Instructor.FirstName,
			"lastName": res.Instructor.LastName, "name": res.Instructor.Name,
			"email": res.Instructor.Email, "phone": res.Instructor.Phone,
			"avatar": res.Instructor.Avatar,
		},
		"wallet": walletJSON(res.Wallet),
	})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	instructor, wallet, err := s.auth.Resolve(r.Context(), bearerToken(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"instructor": map[string]any{
			"id": instructor.ID, "firstName": instructor.FirstName,
			"lastName": instructor.LastName, "name": instructor.Name,
			"email": instructor.Email, "phone": instructor.Phone,
			"avatar": instructor.Avatar,
		},
		"wallet": walletJSON(wallet),
	})
}

// ---- Wallet ----

func (s *Server) handleWallet(w http.ResponseWriter, r *http.Request) {
	instructor, wallet, err := s.auth.Resolve(r.Context(), bearerToken(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	txs, err := s.wallet.Transactions(r.Context(), instructor.ID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"wallet":       walletJSON(wallet),
		"transactions": txs,
	})
}

type fundReq struct {
	AmountNaira string `json:"amountNaira"`
	Method      string `json:"method"`
}

func (s *Server) handleFundWallet(w http.ResponseWriter, r *http.Request) {
	instructor, _, err := s.auth.Resolve(r.Context(), bearerToken(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	var req fundReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, errors.New("invalid funding request"))
		return
	}
	amount, err := domain.FromNairaString(req.AmountNaira)
	if err != nil {
		writeErr(w, err)
		return
	}
	wallet, err := s.wallet.Fund(r.Context(), instructor.ID, amount, req.Method)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"wallet": walletJSON(wallet)})
}

// handleCreateUser re-runs strict-flow step 1 (create the BMONI user) when it
// failed at signup, so the wizard can retry. Idempotent: already-provisioned
// wallets are left untouched.
func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if s.wallet == nil {
		writeErr(w, errors.New("wallet service not configured"))
		return
	}
	wallet, err := s.wallet.CreateBmoniUser(r.Context(), instructorFrom(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"wallet": walletJSON(wallet)})
}

// handleWalletSetup reports where the wallet is in the strict provisioning
// flow and what the wizard should ask for next.
func (s *Server) handleWalletSetup(w http.ResponseWriter, r *http.Request) {
	if s.wallet == nil {
		writeErr(w, errors.New("wallet service not configured"))
		return
	}
	instructor, _, err := s.auth.Resolve(r.Context(), bearerToken(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	st, err := s.wallet.SetupStatus(r.Context(), instructor.ID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"stage": st.Stage, "bmoniUserId": st.BmoniUserID,
		"kycSubmitted": st.KYCSubmitted, "bmoniWalletId": st.BmoniWalletID,
		"bmoniWalletAddress": st.BmoniWalletAddr, "railActive": st.RailActive,
		"railConfigured": s.wallet.RailConfigured(),
		"depositAccount": map[string]string{
			"accountNumber": st.DepositAccountNumber, "bankName": st.DepositBank,
		},
	})
}

// handleCreateBmoniUser (re)runs the strict-flow step 1: register the host's
// BMONI user from their signup identity. Idempotent. It exists so a host whose
// user creation failed at signup (rail hiccup, bad key, etc.) can retry from
// the wallet wizard without re-signing up.
func (s *Server) handleCreateBmoniUser(w http.ResponseWriter, r *http.Request) {
	if s.wallet == nil {
		writeErr(w, errors.New("wallet service not configured"))
		return
	}
	instructor, _, err := s.auth.Resolve(r.Context(), bearerToken(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	wallet, err := s.wallet.CreateBmoniUser(r.Context(), instructor.ID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"wallet": walletJSON(wallet)})
}

type bvnLookupReq struct {
	BVN string `json:"bvn"`
}

// handleKYCLookup resolves a BVN to its holder record so the KYC form
// pre-fills for the host to confirm (lifecycle stage 3 helper; writes nothing).
func (s *Server) handleKYCLookup(w http.ResponseWriter, r *http.Request) {
	if s.wallet == nil {
		writeErr(w, errors.New("wallet service not configured"))
		return
	}
	instructor, _, err := s.auth.Resolve(r.Context(), bearerToken(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	var req bvnLookupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, domain.ErrBadCredentials)
		return
	}
	rec, err := s.wallet.LookupBVN(r.Context(), instructor.ID, req.BVN)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"bvn": rec.BVN, "firstName": rec.FirstName, "lastName": rec.LastName,
		"middleName": rec.MiddleName, "dateOfBirth": rec.DateOfBirth, "gender": rec.Gender,
		"email": rec.Email, "phoneNumber": rec.PhoneNumber,
		"residentialAddress": rec.ResidentialAddress, "stateOfResidence": rec.StateOfResidence,
		"nin": rec.NIN,
	})
}

type kycReq struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DateOfBirth string `json:"dateOfBirth"`
	Gender      string `json:"gender"`
	BVN         string `json:"bvn"`
	Street      string `json:"street"`
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postalCode"`
}

// handleSubmitKYC submits the host's KYC profile (strict flow step 2).
func (s *Server) handleSubmitKYC(w http.ResponseWriter, r *http.Request) {
	if s.wallet == nil {
		writeErr(w, errors.New("wallet service not configured"))
		return
	}
	instructor, _, err := s.auth.Resolve(r.Context(), bearerToken(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	var req kycReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, domain.ErrBadCredentials)
		return
	}
	wallet, err := s.wallet.SubmitKYC(r.Context(), instructor.ID, domain.KYCProfile{
		FirstName: req.FirstName, LastName: req.LastName, DateOfBirth: req.DateOfBirth,
		Gender: req.Gender, BVN: req.BVN, Street: req.Street,
		City: req.City, State: req.State, PostalCode: req.PostalCode,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"wallet": walletJSON(wallet)})
}

// handleCreateWallet provisions the CNGN smart wallet (strict flow steps 3-4).
func (s *Server) handleCreateWallet(w http.ResponseWriter, r *http.Request) {
	if s.wallet == nil {
		writeErr(w, errors.New("wallet service not configured"))
		return
	}
	instructor, _, err := s.auth.Resolve(r.Context(), bearerToken(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	wallet, err := s.wallet.CreateWallet(r.Context(), instructor.ID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"wallet": walletJSON(wallet)})
}

type activateRailReq struct {
	BVN string `json:"bvn"`
}

// handleActivateRail activates the NGN rail (strict flow step 5).
func (s *Server) handleActivateRail(w http.ResponseWriter, r *http.Request) {
	if s.wallet == nil {
		writeErr(w, errors.New("wallet service not configured"))
		return
	}
	instructor, _, err := s.auth.Resolve(r.Context(), bearerToken(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	var req activateRailReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, domain.ErrBadCredentials)
		return
	}
	wallet, err := s.wallet.ActivateRail(r.Context(), instructor.ID, req.BVN)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"wallet": walletJSON(wallet)})
}

// handleUploadKYC forwards a KYC document image (identification /
// proof-of-address / biometric) with its document-type metadata to the rail.
func (s *Server) handleUploadKYC(w http.ResponseWriter, r *http.Request) {
	if s.wallet == nil {
		writeErr(w, errors.New("wallet service not configured"))
		return
	}
	kind := r.PathValue("kind")
	switch kind {
	case "identification", "proof-of-address", "biometric":
	default:
		writeErr(w, domain.ErrBadCredentials)
		return
	}
	instructor, _, err := s.auth.Resolve(r.Context(), bearerToken(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	if err := r.ParseMultipartForm(6 << 20); err != nil {
		writeErr(w, errors.New("upload a document image (JPEG/PNG)"))
		return
	}
	// The image arrives under the `files` multipart field; document-type
	// metadata rides alongside (identification also needs documentNumber +
	// issuingCountry).
	file, hdr, err := r.FormFile("files")
	if err != nil {
		writeErr(w, errors.New("upload a document image (JPEG/PNG)"))
		return
	}
	defer file.Close()
	data := make([]byte, hdr.Size)
	if _, err := io.ReadFull(file, data); err != nil {
		writeErr(w, errors.New("could not read the uploaded document"))
		return
	}
	if err := s.wallet.UploadKYC(r.Context(), instructor.ID, domain.KycDocument{
		Kind:           kind,
		Data:           data,
		Name:           hdr.Filename,
		Type:           r.FormValue("type"),
		DocumentNumber: r.FormValue("documentNumber"),
		IssuingCountry: r.FormValue("issuingCountry"),
	}); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleDepositAccount returns the host's NGN virtual bank account (number +
// bank) that bank transfers to it fund the wallet.
func (s *Server) handleDepositAccount(w http.ResponseWriter, r *http.Request) {
	if s.wallet == nil || s.auth == nil {
		writeErr(w, domain.ErrUnauthorized)
		return
	}
	instructor, _, err := s.auth.Resolve(r.Context(), bearerToken(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	number, bank, err := s.wallet.DepositAccount(r.Context(), instructor.ID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"accountNumber": number, "bankName": bank})
}

func walletJSON(w *domain.Wallet) map[string]any {
	if w == nil {
		return map[string]any{}
	}
	return map[string]any{
		"id": w.ID, "balanceNaira": w.Balance.DisplayString(),
		"bmoniUserId": w.BmoniUserID, "bmoniKycSubmitted": w.BmoniKYCSubmitted,
		"bmoniWalletId": w.BmoniWalletID, "bmoniWalletAddress": w.BmoniWalletAddr,
		"bmoniRailActive": w.BmoniRailActive,
	}
}

// ---- Dashboard / history ----

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	instructorID := instructorFrom(r)
	stat, err := s.game.Dashboard(r.Context(), instructorID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stat)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	instructorID := instructorFrom(r)
	items, err := s.game.History(r.Context(), instructorID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}
