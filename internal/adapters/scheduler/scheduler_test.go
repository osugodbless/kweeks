package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/osugodbless/kweeks/internal/adapters/bus"
	"github.com/osugodbless/kweeks/internal/adapters/clock"
	"github.com/osugodbless/kweeks/internal/adapters/store/memory"
	"github.com/osugodbless/kweeks/internal/app"
	"github.com/osugodbless/kweeks/internal/domain"
)

// autoQuiz has two questions at 1s each in AUTO mode.
func autoQuiz() *domain.Quiz {
	return &domain.Quiz{
		ID:              "quiz-auto",
		InstructorID:    "ins-1",
		Title:           "Auto",
		Pool:            50000,
		WinnerCount:     1,
		Pacing:          domain.PacingAuto,
		DefaultDuration: time.Second,
		Questions: []domain.Question{
			{ID: "q1", Prompt: "A?", Options: []string{"a", "b"}, CorrectIndex: 1},
			{ID: "q2", Prompt: "B?", Options: []string{"a", "b"}, CorrectIndex: 0},
		},
	}
}

// TestSchedulerAdvancesAutoRooms proves the room advances WITHOUT any client
// tap: the scheduler advances q1->q2 after the 1s window, then finalizes the
// podium after q2's window.
func TestSchedulerAdvancesAutoRooms(t *testing.T) {
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	game := app.NewGame(st, clk, bus.NewInMemory())

	ctx := context.Background()
	if err := game.CreateQuiz(ctx, autoQuiz()); err != nil {
		t.Fatalf("create quiz: %v", err)
	}
	room := &domain.Room{ID: "room-auto", QuizID: "quiz-auto", HostID: "ins-1"}
	if err := game.OpenRoom(ctx, room); err != nil {
		t.Fatalf("open room: %v", err)
	}
	if err := game.Start(ctx, "room-auto"); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Join a player and answer both questions within their windows so the
	// podium has a winner when the scheduler finalizes.
	joiner := app.NewJoin(st, clk)
	player, err := joiner.JoinRoom(ctx, &domain.Participant{RoomID: "room-auto", Email: "p@x.com", Nickname: "p"})
	if err != nil {
		t.Fatalf("join: %v", err)
	}
	if _, err := game.SubmitAnswer(ctx, &domain.Answer{
		RoomID: "room-auto", ParticipantID: player.ID, QuestionID: "q1", OptionIndex: 1,
	}); err != nil {
		t.Fatalf("q1 answer: %v", err)
	}

	// Manually drive the scheduler's tick with a fake clock.
	s := New(game, st, nil, time.Millisecond)

	// t=0: q1 just started, window 1s. tick should NOT advance yet.
	room1, _ := st.GetRoom(ctx, "room-auto")
	if room1.CurrentQuestionIdx != 0 {
		t.Fatalf("expected q1, got idx %d", room1.CurrentQuestionIdx)
	}

	// t=1500ms: q1 window closed; scheduler must advance to q2.
	clk.Set(clk.Now().Add(1500 * time.Millisecond))
	s.tick(ctx, clk.Now())
	room2, _ := st.GetRoom(ctx, "room-auto")
	if room2.CurrentQuestionIdx != 1 {
		t.Fatalf("expected q2 after scheduler tick, got idx %d (state %s)", room2.CurrentQuestionIdx, room2.State)
	}

	// Answer q2 inside its window.
	if _, err := game.SubmitAnswer(ctx, &domain.Answer{
		RoomID: "room-auto", ParticipantID: player.ID, QuestionID: "q2", OptionIndex: 0,
	}); err != nil {
		t.Fatalf("q2 answer: %v", err)
	}

	// t=3000ms: q2 window closed; scheduler must finalize the podium.
	clk.Set(clk.Now().Add(1500 * time.Millisecond))
	s.tick(ctx, clk.Now())
	room3, _ := st.GetRoom(ctx, "room-auto")
	if room3.State != domain.RoomPodium {
		t.Fatalf("expected podium after scheduler tick, got state %s", room3.State)
	}
}

// TestSchedulerNeverAdvancesManualRooms proves manual rooms are untouched by
// the scheduler; only the instructor next tap moves them.
func TestSchedulerNeverAdvancesManualRooms(t *testing.T) {
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	game := app.NewGame(st, clk, bus.NewInMemory())
	ctx := context.Background()

	q := autoQuiz()
	q.Pacing = domain.PacingManual
	q.ID = "quiz-manual"
	if err := game.CreateQuiz(ctx, q); err != nil {
		t.Fatalf("create quiz: %v", err)
	}
	room := &domain.Room{ID: "room-manual", QuizID: q.ID, HostID: "ins-1"}
	if err := game.OpenRoom(ctx, room); err != nil {
		t.Fatalf("open room: %v", err)
	}
	if err := game.Start(ctx, "room-manual"); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Let minutes pass; the scheduler must not move a manual room.
	clk.Set(clk.Now().Add(10 * time.Minute))
	s := New(game, st, nil, time.Millisecond)
	s.tick(ctx, clk.Now())

	got, _ := st.GetRoom(ctx, "room-manual")
	if got.CurrentQuestionIdx != 0 {
		t.Fatalf("manual room advanced by scheduler to idx %d", got.CurrentQuestionIdx)
	}
	if got.State != domain.RoomLive {
		t.Fatalf("manual room state changed to %s", got.State)
	}
}

// A past-last AUTO room with no eligible winner must reach the ended terminal
// state, not linger live and retry finalize forever.
func TestSchedulerEndsNoWinnerRoom(t *testing.T) {
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	game := app.NewGame(st, clk, bus.NewInMemory())
	ctx := context.Background()

	q := autoQuiz() // 2 questions at 1s, AUTO, winnerCount 1
	if err := game.CreateQuiz(ctx, q); err != nil {
		t.Fatalf("create: %v", err)
	}
	room := &domain.Room{ID: "room-no-win", QuizID: q.ID, HostID: "ins-1"}
	if err := game.OpenRoom(ctx, room); err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := game.Start(ctx, room.ID); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Player joins but never answers -> 0 correct.
	joiner := app.NewJoin(st, clk)
	player, err := joiner.JoinRoom(ctx, &domain.Participant{RoomID: room.ID, Email: "n@x.com", Nickname: "n"})
	if err != nil {
		t.Fatalf("join: %v", err)
	}
	_ = player

	s := New(game, st, nil, time.Millisecond)

	// Advance past both questions (4s total), letting the scheduler tick at
	// each boundary.
	clk.Set(clk.Now().Add(1500 * time.Millisecond)) // close q1 window
	s.tick(ctx, clk.Now())                          // scheduler -> q2
	clk.Set(clk.Now().Add(1500 * time.Millisecond)) // close q2 window
	s.tick(ctx, clk.Now())                          // scheduler -> resolve (no winners)
	clk.Set(clk.Now().Add(1500 * time.Millisecond)) // resolve again
	s.tick(ctx, clk.Now())

	got, _ := st.GetRoom(ctx, room.ID)
	if got.State != domain.RoomEnded {
		t.Fatalf("no-winner room must end, got state %s", got.State)
	}
}
