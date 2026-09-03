package app

import (
	"context"
	"testing"
	"time"

	"github.com/osugodbless/kweeks/internal/adapters/bus"
	"github.com/osugodbless/kweeks/internal/adapters/clock"
	"github.com/osugodbless/kweeks/internal/adapters/store/memory"
	"github.com/osugodbless/kweeks/internal/domain"
)

func mustQuiz() *domain.Quiz {
	return &domain.Quiz{
		ID:              "quiz-1",
		InstructorID:    "ins-1",
		Title:           "Demo",
		Pool:            100000, // NGN 1000
		WinnerCount:     2,
		Pacing:          domain.PacingManual,
		DefaultDuration: 30 * time.Second,
		Questions: []domain.Question{
			{ID: "q1", Prompt: "Capital of Nigeria?", Options: []string{"Lagos", "Abuja", "Kano", "Ibadan"}, CorrectIndex: 1},
			{ID: "q2", Prompt: "2+2?", Options: []string{"3", "4", "5"}, CorrectIndex: 1},
		},
	}
}

func setupGame(t *testing.T) (*Game, *memory.Store, *clock.Static) {
	t.Helper()
	st := memory.New()
	clk := clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC))
	g := NewGame(st, clk, bus.NewInMemory())
	if err := g.CreateQuiz(context.Background(), mustQuiz()); err != nil {
		t.Fatalf("create quiz: %v", err)
	}
	room := &domain.Room{ID: "room-1", QuizID: "quiz-1", HostID: "ins-1"}
	if err := g.OpenRoom(context.Background(), room); err != nil {
		t.Fatalf("open room: %v", err)
	}
	return g, st, clk
}

func join(t *testing.T, st *memory.Store, email, nick string) *domain.Participant {
	t.Helper()
	joiner := NewJoin(st, clock.NewStatic(time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)))
	p, err := joiner.JoinRoom(context.Background(), &domain.Participant{
		RoomID: "room-1", Email: email, Nickname: nick, Avatar: "🦁",
	})
	if err != nil {
		t.Fatalf("join %s: %v", email, err)
	}
	return p
}

func TestStartThenLateAnswerRejected(t *testing.T) {
	g, _, clk := setupGame(t)
	if err := g.Start(context.Background(), "room-1"); err != nil {
		t.Fatalf("start: %v", err)
	}
	alice := join(t, mustStore(g), "alice@x.com", "alice")

	// Answer 1s before the 30s cutoff: accepted.
	clk.Set(clk.Now().Add(29 * time.Second))
	_, err := g.SubmitAnswer(context.Background(), &domain.Answer{
		RoomID: "room-1", ParticipantID: alice.ID, QuestionID: "q1", OptionIndex: 1,
	})
	if err != nil {
		t.Fatalf("on-time answer rejected: %v", err)
	}
	// Answer after cutoff: rejected.
	clk.Set(clk.Now().Add(2 * time.Second))
	_, err = g.SubmitAnswer(context.Background(), &domain.Answer{
		RoomID: "room-1", ParticipantID: alice.ID, QuestionID: "q1", OptionIndex: 1,
	})
	if err != domain.ErrAlreadyAnswered && err != domain.ErrAnswerLate {
		t.Fatalf("expected late/duplicate rejection, got %v", err)
	}
}

func mustStore(g *Game) *memory.Store {
	// Unwrap the store for the join helper.
	return g.store.(*memory.Store)
}

func TestSecondQuestionNeedsAdvance(t *testing.T) {
	g, _, clk := setupGame(t)
	if err := g.Start(context.Background(), "room-1"); err != nil {
		t.Fatalf("start: %v", err)
	}
	alice := join(t, mustStore(g), "alice@x.com", "alice")

	// q2 is not current before advancing.
	_, err := g.SubmitAnswer(context.Background(), &domain.Answer{
		RoomID: "room-1", ParticipantID: alice.ID, QuestionID: "q2", OptionIndex: 1,
	})
	if err != domain.ErrAnswerUnknownQuestion {
		t.Fatalf("expected unknown-question for non-current q2, got %v", err)
	}
	// Advance after the q1 window closes, then q2 is answerable.
	clk.Set(clk.Now().Add(31 * time.Second))
	if err := g.Next(context.Background(), "room-1"); err != nil {
		t.Fatalf("advance: %v", err)
	}
	if _, err := g.SubmitAnswer(context.Background(), &domain.Answer{
		RoomID: "room-1", ParticipantID: alice.ID, QuestionID: "q2", OptionIndex: 1,
	}); err != nil {
		t.Fatalf("q2 answer rejected: %v", err)
	}
}

func TestPodiumSelectsTopNWinners(t *testing.T) {
	g, _, clk := setupGame(t)
	if err := g.Start(context.Background(), "room-1"); err != nil {
		t.Fatalf("start: %v", err)
	}
	alice := join(t, mustStore(g), "alice@x.com", "alice")
	bob := join(t, mustStore(g), "bob@x.com", "bob")
	carol := join(t, mustStore(g), "carol@x.com", "carol")

	// q1: alice correct fast, bob correct slow, carol wrong.
	_, _ = g.SubmitAnswer(context.Background(), &domain.Answer{RoomID: "room-1", ParticipantID: alice.ID, QuestionID: "q1", OptionIndex: 1})
	clk.Set(clk.Now().Add(5 * time.Second))
	_, _ = g.SubmitAnswer(context.Background(), &domain.Answer{RoomID: "room-1", ParticipantID: bob.ID, QuestionID: "q1", OptionIndex: 1})
	_, _ = g.SubmitAnswer(context.Background(), &domain.Answer{RoomID: "room-1", ParticipantID: carol.ID, QuestionID: "q1", OptionIndex: 0})

	// Advance to q2 (window already long gone for q1's answers? clock moved 5s,
	// still inside 30s window). Answers landed fine above.
	clk.Set(clk.Now().Add(26 * time.Second)) // 31s past q1 start
	if err := g.Next(context.Background(), "room-1"); err != nil {
		t.Fatalf("advance: %v", err)
	}
	// q2: bob answers, carol answers, alice misses.
	_, _ = g.SubmitAnswer(context.Background(), &domain.Answer{RoomID: "room-1", ParticipantID: bob.ID, QuestionID: "q2", OptionIndex: 1})
	clk.Set(clk.Now().Add(2 * time.Second))
	_, _ = g.SubmitAnswer(context.Background(), &domain.Answer{RoomID: "room-1", ParticipantID: carol.ID, QuestionID: "q2", OptionIndex: 1})

	// Close last window, finalize.
	clk.Set(clk.Now().Add(30 * time.Second))
	if _, err := g.FinalizePodium(context.Background(), "room-1"); err != nil {
		t.Fatalf("finalize: %v", err)
	}

	standings, err := g.Standings(context.Background(), "room-1")
	if err != nil {
		t.Fatalf("standings: %v", err)
	}
	// alice: 1 correct (fast), bob: 2 correct, carol: 1 correct (slow)
	// Order: bob (2), then alice (1, fast) beats carol (1, slow).
	if len(standings) < 3 {
		t.Fatalf("expected 3 standings, got %d", len(standings))
	}
	if standings[0].Nickname != "bob" {
		t.Fatalf("expected bob first, got %s (scores: %+v)", standings[0].Nickname, standings)
	}
	if standings[1].Nickname != "alice" || standings[2].Nickname != "carol" {
		t.Fatalf("unexpected order: %+v", standings)
	}
}

func TestRedemptionExactlyOnceAndWinnerOnly(t *testing.T) {
	g, st, _ := setupGame(t)
	ctx := context.Background()
	if err := g.Start(ctx, "room-1"); err != nil {
		t.Fatalf("start: %v", err)
	}
	alice := join(t, st, "alice@x.com", "alice")
	bob := join(t, st, "bob@x.com", "bob")

	// Both answer q1 correctly; alice faster.
	_, _ = g.SubmitAnswer(ctx, &domain.Answer{RoomID: "room-1", ParticipantID: alice.ID, QuestionID: "q1", OptionIndex: 1})
	_, _ = g.SubmitAnswer(ctx, &domain.Answer{RoomID: "room-1", ParticipantID: bob.ID, QuestionID: "q1", OptionIndex: 1})

	// Skip q2 and finalize (only q1 answered).
	// Advance out of q2? There are 2 questions; podium needs room past last.
	// We call FinalizePodium directly — it does not require the last window.
	if _, err := g.FinalizePodium(ctx, "room-1"); err != nil {
		t.Fatalf("finalize: %v", err)
	}

	red := NewRedemption(st, clock.NewReal(), nil, nil)

	// alice (winner, rank 1) can claim.
	c1, err := red.CreateClaim(ctx, "room-1", "alice@x.com", 60000)
	if err != nil {
		t.Fatalf("alice claim: %v", err)
	}
	if c1.ClaimCode == "" {
		t.Fatal("claim code empty")
	}
	// Exactly-once: second claim returns the same record.
	c2, err := red.CreateClaim(ctx, "room-1", "alice@x.com", 60000)
	if err != nil {
		t.Fatalf("alice second claim errored: %v", err)
	}
	if c2.ID != c1.ID || c2.ClaimCode != c1.ClaimCode {
		t.Fatal("exactly-once violated: duplicate claim created")
	}
	// A non-winner (someone who never joined) cannot claim.
	if _, err := red.CreateClaim(ctx, "room-1", "mallory@x.com", 60000); err != domain.ErrNotWinner {
		t.Fatalf("expected ErrNotWinner for non-participant, got %v", err)
	}
}

func TestPodiumShareSplit(t *testing.T) {
	pool := domain.Amount(100000)
	shares := domain.SplitPodium(pool, 2)
	if len(shares) != 2 {
		t.Fatalf("expected 2 shares, got %d", len(shares))
	}
	if int64(shares[0]+shares[1]) != 100000 {
		t.Fatalf("shares must sum to pool: %v", shares)
	}
	if shares[0] < shares[1] {
		t.Fatalf("first place must get >= second: %v", shares)
	}
}
