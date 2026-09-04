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

func buildServer(t *testing.T) (*Server, *memory.Store, *clock.Static) {
	t.Helper()
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	game := app.NewGame(st, clk, bus.NewInMemory())
	join := app.NewJoin(st, clk)
	red := app.NewRedemption(st, clk, nil, nil)
	return New(game, join, red), st, clk
}

func seedQuizRoom(t *testing.T, api *Server, st *memory.Store, clk *clock.Static, finalize bool) (roomID string, aliceID string) {
	t.Helper()
	ctx := context.Background()
	// Create a quiz via the service layer.
	quiz := &domain.Quiz{
		ID:              "quiz-http",
		InstructorID:    "instructor-demo",
		Title:           "HTTP Demo",
		Pool:            100000,
		WinnerCount:     1,
		Pacing:          domain.PacingManual,
		DefaultDuration: 60 * time.Second,
		Questions: []domain.Question{
			{ID: "q1", Prompt: "Q?", Options: []string{"a", "b", "c"}, CorrectIndex: 1},
		},
	}
	if err := api.game.CreateQuiz(ctx, quiz); err != nil {
		t.Fatalf("create quiz: %v", err)
	}
	room := &domain.Room{ID: "room-http", QuizID: quiz.ID, HostID: "instructor-demo"}
	if err := api.game.OpenRoom(ctx, room); err != nil {
		t.Fatalf("open room: %v", err)
	}
	if err := api.game.Start(ctx, room.ID); err != nil {
		t.Fatalf("start: %v", err)
	}
	p, err := api.join.JoinRoom(ctx, &domain.Participant{RoomID: room.ID, Email: "alice@x.com", Nickname: "alice", Avatar: "🦁"})
	if err != nil {
		t.Fatalf("join: %v", err)
	}
	if _, err := api.game.SubmitAnswer(ctx, &domain.Answer{
		RoomID: room.ID, ParticipantID: p.ID, QuestionID: "q1", OptionIndex: 1,
	}); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if finalize {
		if _, err := api.game.FinalizePodium(ctx, room.ID); err != nil {
			t.Fatalf("finalize: %v", err)
		}
	}
	return room.ID, p.ID
}

func doJSON(t *testing.T, api *Server, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var rdr *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	} else {
		rdr = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("Content-Type", "application/json")
	mux := http.NewServeMux()
	api.Routes(mux)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestHTTPRedeemWinnerOnly(t *testing.T) {
	api, _, _ := buildServer(t)
	roomID, _ := seedQuizRoom(t, api, nil, nil, true)

	// Winner redeems -> 200 with claim.
	rr := doJSON(t, api, "POST", "/api/rooms/"+roomID+"/redeem", map[string]string{"email": "alice@x.com"})
	if rr.Code != http.StatusOK {
		t.Fatalf("redeem got %d: %s", rr.Code, rr.Body.String())
	}
	var claim struct {
		ID          string `json:"id"`
		AmountNaira string `json:"amountNaira"`
		State       string `json:"state"`
		ClaimCode   string `json:"claimCode"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &claim); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if claim.AmountNaira != "1000" {
		t.Fatalf("unexpected amount %q", claim.AmountNaira)
	}
	if claim.State != "created" {
		t.Fatalf("unexpected state %q", claim.State)
	}
	if len(claim.ClaimCode) == 0 {
		t.Fatalf("claim code must be returned to the winner's own session")
	}

	// Non-winner (not a participant) -> 403.
	rr2 := doJSON(t, api, "POST", "/api/rooms/"+roomID+"/redeem", map[string]string{"email": "mallory@x.com"})
	if rr2.Code != http.StatusForbidden {
		t.Fatalf("non-winner redeem got %d, want 403: %s", rr2.Code, rr2.Body.String())
	}
}

func TestHTTPAnswerLateRejected(t *testing.T) {
	api, st, clk := buildServer(t)
	roomID, aliceID := seedQuizRoom(t, api, st, clk, false)

	// Advance the clock past the 60s question window, then answer -> 400 late.
	clk.Set(clk.Now().Add(61 * time.Second))
	rr := doJSON(t, api, "POST", "/api/rooms/"+roomID+"/answer", map[string]any{
		"participantId": aliceID, "questionId": "q1", "optionIndex": 1,
	})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("late answer got %d, want 400: %s", rr.Code, rr.Body.String())
	}
	var body map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if body["error"] != domain.ErrAnswerLate.Error() {
		t.Fatalf("unexpected error body: %s", rr.Body.String())
	}
}

func TestHTTPJoinMergeOnRejoin(t *testing.T) {
	api, st, _ := buildServer(t)
	// Create a room without a full game to test the join path only.
	ctx := context.Background()
	quiz := &domain.Quiz{
		ID: "qj", InstructorID: "instructor-demo", Pool: 100, WinnerCount: 1,
		Pacing: domain.PacingManual, DefaultDuration: time.Second,
		Questions: []domain.Question{{ID: "q1", Prompt: "Q?", Options: []string{"a", "b"}, CorrectIndex: 1}},
	}
	if err := api.game.CreateQuiz(ctx, quiz); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := api.game.OpenRoom(ctx, &domain.Room{ID: "room-j", QuizID: "qj", HostID: "instructor-demo"}); err != nil {
		t.Fatalf("open: %v", err)
	}

	rr := doJSON(t, api, "POST", "/api/rooms/room-j/join", map[string]string{"email": "a@x.com", "nickname": "alice", "avatar": "🦁"})
	if rr.Code != http.StatusOK {
		t.Fatalf("join got %d", rr.Code)
	}
	var p1 struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &p1)

	// Rejoin same email -> same participant id, not a duplicate.
	rr2 := doJSON(t, api, "POST", "/api/rooms/room-j/join", map[string]string{"email": "a@x.com", "nickname": "alice2", "avatar": "🐱"})
	if rr2.Code != http.StatusOK {
		t.Fatalf("rejoin got %d", rr2.Code)
	}
	var p2 struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(rr2.Body.Bytes(), &p2)
	if p1.ID == "" || p1.ID != p2.ID {
		t.Fatalf("rejoin did not merge: %q vs %q", p1.ID, p2.ID)
	}
	if _, err := st.ListParticipants(ctx, "room-j"); err != nil {
		t.Fatalf("list: %v", err)
	}
}
