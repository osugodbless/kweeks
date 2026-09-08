package httpapi

import (
	"context"
	"encoding/json"
	"testing"
)

// Two instructors must never see each other's dashboard data: quizzes, wallet
// balances, funding, or history. This guards the multi-tenant isolation that a
// shared BMONI persona must not leak across kweeks accounts.
func TestInstructorDataIsolation(t *testing.T) {
	api, st := buildAuthServer(t)

	// --- Instructor A signs up, funds, and hosts a quiz + room via its own
	// authed session (so the quiz is scoped to A's real instructor id). ---
	tokenA, _ := authSignup(t, api)
	if rr := authDo(api, "POST", "/api/wallet/fund", map[string]any{"amountNaira": "50000", "method": "credit"}, tokenA); rr.Code != 200 {
		t.Fatalf("A fund: %d %s", rr.Code, rr.Body.String())
	}
	ctx := context.Background()
	rrCreate := authDo(api, "POST", "/api/quizzes", map[string]any{
		"title": "A's private quiz", "poolNaira": "20000", "winnerCount": 1, "pacing": "manual",
		"defaultDurationMs": 60000,
		"questions":         []map[string]any{{"id": "q1", "prompt": "Q?", "options": []string{"x", "y"}, "correctIndex": 1, "durationMs": 60000}},
	}, tokenA)
	if rrCreate.Code != 201 {
		t.Fatalf("A create quiz: %d %s", rrCreate.Code, rrCreate.Body.String())
	}
	var quizA map[string]any
	decodeBody(t, rrCreate, &quizA)
	if rr := authDo(api, "POST", "/api/rooms", map[string]string{"quizId": quizA["id"].(string)}, tokenA); rr.Code != 201 {
		t.Fatalf("A open room: %d %s", rr.Code, rr.Body.String())
	}
	_ = ctx

	// --- Instructor B signs up fresh with its own email. ---
	rrB := authDo(api, "POST", "/api/auth/signup", map[string]string{
		"name": "Bisi Baker", "email": "b@kweeks.ng", "password": "secret1",
	}, "")
	if rrB.Code != 200 {
		t.Fatalf("B signup: %d %s", rrB.Code, rrB.Body.String())
	}
	var bbody map[string]any
	decodeBody(t, rrB, &bbody)
	tokenB := bbody["token"].(string)
	walletIDB := bbody["wallet"].(map[string]any)["id"].(string)

	// B's wallet must be its own, and B's dashboard must not show A's quiz.
	rr := authDo(api, "GET", "/api/wallet", nil, tokenB)
	if rr.Code != 200 {
		t.Fatalf("B wallet: %d", rr.Code)
	}
	var wbody map[string]any
	decodeBody(t, rr, &wbody)
	w := wbody["wallet"].(map[string]any)
	if w["id"] != walletIDB {
		t.Fatalf("B sees wallet %v, expected own %v", w["id"], walletIDB)
	}
	if w["balanceNaira"] != "0" {
		t.Fatalf("B wallet balance = %v, want 0 (A's credit leaked)", w["balanceNaira"])
	}

	rr2 := authDo(api, "GET", "/api/instructor/dashboard", nil, tokenB)
	if rr2.Code != 200 {
		t.Fatalf("B dashboard: %d", rr2.Code)
	}
	var stat map[string]any
	decodeBody(t, rr2, &stat)
	if stat["availableNaira"] != "0" {
		t.Fatalf("B available = %v, want 0 (A's funding leaked)", stat["availableNaira"])
	}
	if n := stat["quizzesHosted"]; n != float64(0) {
		t.Fatalf("B quizzesHosted = %v, want 0 (A's quiz leaked)", n)
	}
	if quizzes, ok := stat["quizzes"].([]any); ok && len(quizzes) != 0 {
		t.Fatalf("B dashboard shows A's quizzes: %v", quizzes)
	}

	// B's quiz list must be empty.
	rr3 := authDo(api, "GET", "/api/quizzes", nil, tokenB)
	var list []map[string]any
	decodeBody(t, rr3, &list)
	if len(list) != 0 {
		t.Fatalf("B sees quizzes: %+v", list)
	}

	// B's history must be empty (A's funding/hosting must not appear).
	rr4 := authDo(api, "GET", "/api/instructor/history", nil, tokenB)
	var hist []map[string]any
	decodeBody(t, rr4, &hist)
	if len(hist) != 0 {
		t.Fatalf("B history not empty (A's activity leaked): %+v", hist)
	}

	// A must still see its own quiz.
	rrA := authDo(api, "GET", "/api/quizzes", nil, tokenA)
	if rrA.Code != 200 {
		t.Fatalf("A quizzes: %d", rrA.Code)
	}
	var listA []map[string]any
	decodeBody(t, rrA, &listA)
	if len(listA) != 1 || listA[0]["title"] != "A's private quiz" {
		t.Fatalf("A quiz list wrong: %+v", listA)
	}

	_ = st
	_ = json.NewEncoder
}
