package httpapi

import (
	"context"
	"encoding/json"
	"errors"
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
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	var req credentialsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, domain.ErrBadCredentials)
		return
	}
	res, err := s.auth.Signup(r.Context(), req.Name, req.Email, req.Password)
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
			"id": res.Instructor.ID, "name": res.Instructor.Name,
			"email": res.Instructor.Email, "avatar": res.Instructor.Avatar,
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
			"id": instructor.ID, "name": instructor.Name,
			"email": instructor.Email, "avatar": instructor.Avatar,
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

func (s *Server) handleProvisionWallet(w http.ResponseWriter, r *http.Request) {
	if s.wallet == nil {
		writeErr(w, errors.New("wallet rail not configured on this server"))
		return
	}
	instructorID := instructorFrom(r)
	if !s.wallet.PersonaConfigured() {
		writeErr(w, errors.New("BMONI persona not configured — cannot provision a wallet"))
		return
	}
	wallet, err := s.wallet.Provision(r.Context(), instructorID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"wallet": walletJSON(wallet)})
}

func walletJSON(w *domain.Wallet) map[string]any {
	if w == nil {
		return map[string]any{}
	}
	return map[string]any{
		"id": w.ID, "balanceNaira": w.Balance.DisplayString(),
		"bmoniUserId": w.BmoniUserID, "bmoniWalletId": w.BmoniWalletID,
		"bmoniWalletAddress": w.BmoniWalletAddr,
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
