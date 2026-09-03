//go:build integration

// Integration test for the Postgres store. Requires a running Postgres reachable
// at DATABASE_URL. Run with:
//
//	make db-up
//	DATABASE_URL=postgres://kweeks:kweeks@localhost:5432/kweeks \
//	    go test -tags integration ./internal/adapters/store/postgres/
package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/osugodbless/kweeks/internal/domain"
)

// randHex returns n random bytes as lowercase hex, for unique test rows.
func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func testDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping Postgres integration test")
	}
	return dsn
}

func TestPostgresRoundTrip(t *testing.T) {
	ctx := context.Background()
	st, err := OpenMigrated(ctx, testDSN(t))
	if err != nil {
		t.Fatalf("open+migrate: %v", err)
	}
	defer st.Close()

	// Unique per-run suffix so re-runs never collide with leftover rows from an
	// earlier run in the same second.
	run := fmt.Sprintf("%d-%s", time.Now().UnixNano(), randHex(8))

	// --- Quiz ---
	quiz := &domain.Quiz{
		ID:              "pg-quiz-" + run,
		InstructorID:    "pg-instructor",
		Title:           "PG Round Trip",
		Pool:            125000,
		WinnerCount:     2,
		Pacing:          domain.PacingManual,
		DefaultDuration: 30 * time.Second,
		Questions: []domain.Question{
			{ID: "q1", Prompt: "Capital of Nigeria?", Options: []string{"Lagos", "Abuja"}, CorrectIndex: 1, Duration: 30 * time.Second},
			{ID: "q2", Prompt: "2+2?", Options: []string{"3", "4", "5"}, CorrectIndex: 1, Duration: 20 * time.Second},
		},
	}
	if err := st.CreateQuiz(ctx, quiz); err != nil {
		t.Fatalf("CreateQuiz: %v", err)
	}
	got, err := st.GetQuiz(ctx, quiz.ID)
	if err != nil {
		t.Fatalf("GetQuiz: %v", err)
	}
	if got.Pool != 125000 || got.WinnerCount != 2 || got.Pacing != domain.PacingManual {
		t.Fatalf("quiz round-trip mismatch: %+v", got)
	}
	if len(got.Questions) != 2 || got.Questions[0].CorrectIndex != 1 || got.Questions[1].Options[1] != "4" {
		t.Fatalf("question round-trip mismatch: %+v", got.Questions)
	}
	if got.Questions[0].Duration != 30*time.Second {
		t.Fatalf("duration round-trip mismatch: %v", got.Questions[0].Duration)
	}

	// --- Room ---
	room := &domain.Room{ID: "pg-room-" + run, QuizID: quiz.ID, HostID: "pg-instructor", State: domain.RoomLobby}
	if err := st.CreateRoom(ctx, room); err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	if err := st.SaveRoom(ctx, &domain.Room{ID: room.ID, QuizID: quiz.ID, HostID: "pg-instructor", State: domain.RoomLive, CurrentQuestionIdx: 0}); err != nil {
		t.Fatalf("SaveRoom: %v", err)
	}
	room2, err := st.GetRoom(ctx, room.ID)
	if err != nil {
		t.Fatalf("GetRoom: %v", err)
	}
	if room2.State != domain.RoomLive || room2.CurrentQuestionIdx != 0 {
		t.Fatalf("room round-trip mismatch: %+v", room2)
	}
	liveRooms, err := st.ListLiveRooms(ctx)
	if err != nil {
		t.Fatalf("ListLiveRooms: %v", err)
	}
	found := false
	for _, r := range liveRooms {
		if r.ID == room.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("listlive did not include room %q", room.ID)
	}

	// --- Participant + duplicate rejection ---
	joined := time.Now()
	p1ID := "pg-p1-" + run
	p, err := st.JoinParticipant(ctx, &domain.Participant{ID: p1ID, RoomID: room.ID, Email: "p1@x.com", Nickname: "alice", Avatar: "🦁", JoinedAt: joined})
	if err != nil {
		t.Fatalf("JoinParticipant: %v", err)
	}
	if p.ID != p1ID {
		t.Fatalf("participant id mismatch: %s", p.ID)
	}
	if _, err := st.JoinParticipant(ctx, &domain.Participant{ID: "pg-p2-" + run, RoomID: room.ID, Email: "p1@x.com", Nickname: "dup", Avatar: "🐱", JoinedAt: joined}); err != domain.ErrDuplicateParticipant {
		t.Fatalf("expected duplicate participant error, got %v", err)
	}
	byEmail, err := st.GetParticipant(ctx, room.ID, "p1@x.com")
	if err != nil || byEmail.ID != p1ID {
		t.Fatalf("GetParticipant: %v / %+v", err, byEmail)
	}

	// --- Answer + already-answered ---
	if err := st.RecordAnswer(ctx, &domain.Answer{
		ID: "pg-a1-" + run, RoomID: room.ID, ParticipantID: p1ID, QuestionID: "q1",
		OptionIndex: 1, Correct: true, QuestionStartedAt: joined, ReceivedAt: joined,
	}); err != nil {
		t.Fatalf("RecordAnswer: %v", err)
	}
	if err := st.RecordAnswer(ctx, &domain.Answer{
		ID: "pg-a2-" + run, RoomID: room.ID, ParticipantID: p1ID, QuestionID: "q1",
		OptionIndex: 0, Correct: false, QuestionStartedAt: joined, ReceivedAt: joined,
	}); err != domain.ErrAlreadyAnswered {
		t.Fatalf("expected already-answered, got %v", err)
	}

	// --- Claim ---
	claim := &domain.Claim{
		ID: "pg-c1-" + run, QuizID: quiz.ID, RoomID: room.ID, Email: "p1@x.com",
		Amount: 62500, ClaimCode: "ABCD-1234", State: domain.ClaimCreated, CreatedAt: joined,
	}
	if err := st.CreateClaim(ctx, claim); err != nil {
		t.Fatalf("CreateClaim: %v", err)
	}
	if _, err := st.GetClaimByCode(ctx, quiz.ID, "ABCD-1234"); err != nil {
		t.Fatalf("GetClaimByCode: %v", err)
	}
	byEmailC, err := st.GetClaimByEmail(ctx, quiz.ID, "p1@x.com")
	if err != nil || byEmailC.Amount != 62500 {
		t.Fatalf("GetClaimByEmail: %v / %+v", err, byEmailC)
	}
	if err := st.UpdateClaimState(ctx, "pg-c1-"+run, domain.ClaimPaid); err != nil {
		t.Fatalf("UpdateClaimState: %v", err)
	}
	final, err := st.GetClaimByEmail(ctx, quiz.ID, "p1@x.com")
	if err != nil || final.State != domain.ClaimPaid {
		t.Fatalf("claim state after update: %v / %+v", err, final)
	}
}
