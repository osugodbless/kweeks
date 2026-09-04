package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/osugodbless/kweeks/internal/adapters/bus"
	"github.com/osugodbless/kweeks/internal/adapters/clock"
	"github.com/osugodbless/kweeks/internal/adapters/store/memory"
	"github.com/osugodbless/kweeks/internal/app"
	"github.com/osugodbless/kweeks/internal/domain"
)

// buildAuthServer wires the auth + wallet services so routes enforce Bearer.
func buildAuthServer(t *testing.T) (*Server, *memory.Store) {
	t.Helper()
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	game := app.NewGame(st, clk, bus.NewInMemory())
	join := app.NewJoin(st, clk)
	red := app.NewRedemption(st, clk, nil, nil)
	auth := app.NewAuth(st, clk)
	wallet := app.NewWallet(st, clk, nil)
	api := New(game, join, red).WithServices(auth, wallet)
	return api, st
}

func reqReader(body any) *bytes.Reader {
	var rdr *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	} else {
		rdr = bytes.NewReader(nil)
	}
	return rdr
}

func authDo(api *Server, method, path string, body any, token string) *httptest.ResponseRecorder {
	rdr := reqReader(body)
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	mux := http.NewServeMux()
	api.Routes(mux)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func decodeBody(t *testing.T, rr *httptest.ResponseRecorder, out any) {
	t.Helper()
	if err := json.NewDecoder(rr.Body).Decode(out); err != nil {
		t.Fatalf("decode: %v (body=%s)", err, rr.Body.String())
	}
}

func authSignup(t *testing.T, api *Server) (token, walletID string) {
	t.Helper()
	rr := authDo(api, "POST", "/api/auth/signup", map[string]string{
		"name": "Adeola Peters", "email": "host@kweeks.ng", "password": "secret1",
	}, "")
	if rr.Code != 200 {
		t.Fatalf("signup status %d: %s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	decodeBody(t, rr, &body)
	token = body["token"].(string)
	wallet := body["wallet"].(map[string]any)
	walletID = wallet["id"].(string)
	return
}

func TestAuthSignupLoginWallet(t *testing.T) {
	api, _ := buildAuthServer(t)
	token, walletID := authSignup(t, api)

	// Login with the same credentials returns a fresh token + same identity.
	if rr := authDo(api, "POST", "/api/auth/login", map[string]string{
		"email": "host@kweeks.ng", "password": "secret1",
	}, ""); rr.Code != 200 {
		t.Fatalf("login status %d", rr.Code)
	}

	// me returns the wallet.
	if rr := authDo(api, "GET", "/api/auth/me", nil, token); rr.Code != 200 {
		t.Fatalf("me status %d", rr.Code)
	}

	// fund the wallet via instant credit.
	if rr := authDo(api, "POST", "/api/wallet/fund", map[string]any{
		"amountNaira": "200000", "method": "credit",
	}, token); rr.Code != 200 {
		t.Fatalf("fund status %d: %s", rr.Code, rr.Body.String())
	}

	// wallet ledger reflects it.
	rr := authDo(api, "GET", "/api/wallet", nil, token)
	if rr.Code != 200 {
		t.Fatalf("wallet status %d", rr.Code)
	}
	var wbody map[string]any
	decodeBody(t, rr, &wbody)
	w := wbody["wallet"].(map[string]any)
	if w["id"] != walletID {
		t.Fatalf("wallet id mismatch: %v", w["id"])
	}
	if w["balanceNaira"] != "200000" {
		t.Fatalf("balance = %v", w["balanceNaira"])
	}
	if txs, ok := wbody["transactions"].([]any); !ok || len(txs) != 1 {
		t.Fatalf("expected 1 tx, got %v", wbody["transactions"])
	}
}

func TestAuthRejectsWrongPasswordAndDuplicateEmail(t *testing.T) {
	api, _ := buildAuthServer(t)
	authSignup(t, api)

	if rr := authDo(api, "POST", "/api/auth/login", map[string]string{
		"email": "host@kweeks.ng", "password": "wrongpw1",
	}, ""); rr.Code != 401 {
		t.Fatalf("bad login expected 401 got %d", rr.Code)
	}
	if rr := authDo(api, "POST", "/api/auth/signup", map[string]string{
		"name": "Adeola2", "email": "host@kweeks.ng", "password": "secret1",
	}, ""); rr.Code != 409 {
		t.Fatalf("dup signup expected 409 got %d", rr.Code)
	}
	if rr := authDo(api, "GET", "/api/auth/me", nil, ""); rr.Code != 401 {
		t.Fatalf("anon me expected 401 got %d", rr.Code)
	}
}

func TestDashboardHistoryReflectQuizAndFunding(t *testing.T) {
	api, _ := buildAuthServer(t)
	token, _ := authSignup(t, api)
	authDo(api, "POST", "/api/wallet/fund", map[string]any{"amountNaira": "50000", "method": "credit"}, token)

	ctx := context.Background()
	quiz := &domain.Quiz{
		ID: "q-dash", InstructorID: "instructor-demo", Title: "Dash Quiz",
		Pool: 50000, WinnerCount: 1, Pacing: domain.PacingManual,
		DefaultDuration: 60 * time.Second,
		Questions:       []domain.Question{{ID: "q1", Prompt: "P?", Options: []string{"a", "b"}, CorrectIndex: 0}},
	}
	if err := api.game.CreateQuiz(ctx, quiz); err != nil {
		t.Fatal(err)
	}

	rr := authDo(api, "GET", "/api/instructor/dashboard", nil, token)
	if rr.Code != 200 {
		t.Fatalf("dashboard status %d", rr.Code)
	}
	var stat map[string]any
	decodeBody(t, rr, &stat)
	if stat["availableNaira"] != "50000" {
		t.Fatalf("available = %v", stat["availableNaira"])
	}

	hr := authDo(api, "GET", "/api/instructor/history", nil, token)
	if hr.Code != 200 {
		t.Fatalf("history status %d", hr.Code)
	}
	var hist []map[string]any
	decodeBody(t, hr, &hist)
	if len(hist) == 0 {
		t.Fatalf("expected history entries")
	}
}

func TestRoomOpenReturnsCodeAndStateByCode(t *testing.T) {
	api, _ := buildAuthServer(t)
	token, _ := authSignup(t, api)

	ctx := context.Background()
	quiz := &domain.Quiz{
		ID: "q-state", InstructorID: "instructor-demo", Title: "State Quiz",
		Pool: 100000, WinnerCount: 2, Pacing: domain.PacingManual,
		DefaultDuration: 60 * time.Second,
		Questions: []domain.Question{
			{ID: "q1", Prompt: "P?", Options: []string{"a", "b", "c"}, CorrectIndex: 1},
			{ID: "q2", Prompt: "Q?", Options: []string{"x", "y"}, CorrectIndex: 0},
		},
	}
	if err := api.game.CreateQuiz(ctx, quiz); err != nil {
		t.Fatal(err)
	}

	rr := authDo(api, "POST", "/api/rooms", map[string]string{"quizId": "q-state"}, token)
	if rr.Code != 201 {
		t.Fatalf("open status %d: %s", rr.Code, rr.Body.String())
	}
	var room map[string]any
	decodeBody(t, rr, &room)
	code, _ := room["code"].(string)
	if len(code) != 4 {
		t.Fatalf("code length %d (%q)", len(code), code)
	}

	// Public state by code: lobby, no current question, participants empty.
	rr2 := authDo(api, "GET", "/api/lookup/"+code, nil, "")
	if rr2.Code != 200 {
		t.Fatalf("lookup status %d", rr2.Code)
	}
	var state map[string]any
	decodeBody(t, rr2, &state)
	if state["code"] != code || state["state"] != "lobby" {
		t.Fatalf("lookup mismatch: %v %v", state["code"], state["state"])
	}
	if state["currentQuestion"] != nil {
		t.Fatalf("lobby should have no current question")
	}
}
