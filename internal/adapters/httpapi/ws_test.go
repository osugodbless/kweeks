package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/osugodbless/kweeks/internal/adapters/clock"
	"github.com/osugodbless/kweeks/internal/adapters/store/memory"
	"github.com/osugodbless/kweeks/internal/adapters/ws"
	"github.com/osugodbless/kweeks/internal/app"
)

// craftServer builds a Server wired with a real websocket hub, wrapped in an
// httptest.Server so the dialer can connect over a live ws:// URL.
func craftServer(t *testing.T) *httptest.Server {
	t.Helper()
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	hub := ws.NewHub(nil)
	game := app.NewGame(st, clk, hub)
	join := app.NewJoin(st, clk)
	red := app.NewRedemption(st, clk, nil, nil)
	api := New(game, join, red).WithWS(hub)

	mux := http.NewServeMux()
	api.Routes(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func httpPost(t *testing.T, base, path, body string) (int, map[string]any) {
	t.Helper()
	resp, err := http.Post(base+path, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post %s: %v", path, err)
	}
	defer resp.Body.Close()
	var out map[string]any
	if resp.ContentLength != 0 {
		_ = json.NewDecoder(resp.Body).Decode(&out)
	}
	return resp.StatusCode, out
}

// seedWSQuizRoom creates a quiz and a room over HTTP, returning the roomID and
// the ws URL for that room.
func seedWSQuizRoom(t *testing.T, srv *httptest.Server) (string, string) {
	t.Helper()
	code, _ := httpPost(t, srv.URL, "/api/quizzes", `{"title":"WS","poolNaira":"5000.00","winnerCount":1,"pacing":"manual","defaultDurationMs":60000,"questions":[{"id":"q1","prompt":"Q?","options":["a","b"],"correctIndex":1}]}`)
	if code != 201 {
		t.Fatalf("create quiz got %d", code)
	}
	resp, err := http.Get(srv.URL + "/api/quizzes")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var quizzes []map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&quizzes)
	resp.Body.Close()
	quizID := quizzes[0]["id"].(string)

	_, open := httpPost(t, srv.URL, "/api/rooms", `{"quizId":"`+quizID+`"}`)
	roomID := open["id"].(string)
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/rooms/" + roomID + "/ws"
	return roomID, wsURL
}

func dialWS(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial ws: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func readEvent(t *testing.T, conn *websocket.Conn) (string, []byte) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read ws event: %v", err)
	}
	var ev struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(msg, &ev); err != nil {
		t.Fatalf("decode ws event %q: %v", msg, err)
	}
	return ev.Type, msg
}

// The server must stream a "question" event to a connected player socket the
// moment the instructor starts the room.
func TestRealTimeQuestionEventOverWS(t *testing.T) {
	srv := craftServer(t)
	roomID, wsURL := seedWSQuizRoom(t, srv)
	conn := dialWS(t, wsURL)

	code, _ := httpPost(t, srv.URL, "/api/rooms/"+roomID+"/start", "")
	if code != 204 && code != 200 {
		t.Fatalf("start got %d", code)
	}

	typ, _ := readEvent(t, conn)
	if typ != "question" {
		t.Fatalf("expected 'question' event, got %q", typ)
	}
}

// After the podium is declared, the winner socket must receive a "podium"
// event (the money beat resolves live, not on poll).
func TestRealTimePodiumEventOverWS(t *testing.T) {
	srv := craftServer(t)
	roomID, wsURL := seedWSQuizRoom(t, srv)
	conn := dialWS(t, wsURL)

	httpPost(t, srv.URL, "/api/rooms/"+roomID+"/start", "")

	_, join := httpPost(t, srv.URL, "/api/rooms/"+roomID+"/join", `{"email":"w@x.com","nickname":"w","avatar":"🦁"}`)
	pid := join["id"].(string)
	httpPost(t, srv.URL, "/api/rooms/"+roomID+"/answer", `{"participantId":"`+pid+`","questionId":"q1","optionIndex":1}`)

	// Drain the question event from start.
	if typ, _ := readEvent(t, conn); typ != "question" {
		t.Fatalf("unexpected first event %q", typ)
	}

	httpPost(t, srv.URL, "/api/rooms/"+roomID+"/podium", "")

	if typ, _ := readEvent(t, conn); typ != "podium" {
		t.Fatalf("expected 'podium' event, got %q", typ)
	}

	// Drain the answer'd inactivity: there may be no more events; just assert we
	// got podium which is what matters.
}
