package domain

import (
	"testing"
	"time"
)

func TestAnswerGate(t *testing.T) {
	start := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	dur := 30 * time.Second

	// Just before cutoff: accepted.
	if !AnswerGate(start.Add(29*time.Second), start, dur) {
		t.Fatal("answer 1s before cutoff should be accepted")
	}
	// Exactly at cutoff: rejected (now < started+duration is strict).
	if AnswerGate(start.Add(30*time.Second), start, dur) {
		t.Fatal("answer at exact cutoff should be rejected")
	}
	// After cutoff: rejected.
	if AnswerGate(start.Add(31*time.Second), start, dur) {
		t.Fatal("answer after cutoff should be rejected")
	}
	// Zero startedAt: always rejected (no question live).
	if AnswerGate(time.Now(), time.Time{}, dur) {
		t.Fatal("answer with zero startedAt must be rejected")
	}
}

func TestLatencyMs(t *testing.T) {
	start := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	got := LatencyMs(start.Add(1500*time.Millisecond), start)
	if got != 1500 {
		t.Fatalf("expected 1500ms, got %d", got)
	}
}

func TestStandingLessOrdering(t *testing.T) {
	base := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		a, b Standing
		want bool // is a above b
	}{
		{
			"more correct wins",
			Standing{CorrectCount: 5, TotalLatency: 9 * time.Second},
			Standing{CorrectCount: 4, TotalLatency: 1 * time.Second},
			true,
		},
		{
			"equal correct, lower latency wins",
			Standing{CorrectCount: 5, TotalLatency: 9 * time.Second},
			Standing{CorrectCount: 5, TotalLatency: 12 * time.Second},
			true,
		},
		{
			"equal correct and latency, earlier join wins",
			Standing{CorrectCount: 5, TotalLatency: 9 * time.Second, JoinedAt: base},
			Standing{CorrectCount: 5, TotalLatency: 9 * time.Second, JoinedAt: base.Add(time.Minute)},
			true,
		},
		{
			"lower-correct does not rank above",
			Standing{CorrectCount: 3},
			Standing{CorrectCount: 5},
			false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := StandingLess(tc.a, tc.b); got != tc.want {
				t.Fatalf("StandingLess = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSortAndSelectWinners(t *testing.T) {
	base := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	in := []Standing{
		{ParticipantID: "slow", CorrectCount: 5, TotalLatency: 12 * time.Second, JoinedAt: base},
		{ParticipantID: "fast", CorrectCount: 5, TotalLatency: 9 * time.Second, JoinedAt: base},
		{ParticipantID: "also", CorrectCount: 4, TotalLatency: 1 * time.Second, JoinedAt: base},
	}
	sorted := SortStandings(in)
	if sorted[0].ParticipantID != "fast" {
		t.Fatalf("expected fast on top, got %s", sorted[0].ParticipantID)
	}
	winners := SelectWinners(sorted, 2)
	if len(winners) != 2 || winners[0].ParticipantID != "fast" || winners[1].ParticipantID != "slow" {
		t.Fatalf("unexpected winners: %+v", winners)
	}
	// Input not mutated.
	if in[0].ParticipantID != "slow" {
		t.Fatal("SortStandings mutated its input")
	}
}

// A participant with zero correct answers never takes the pool, even if they
// are the only player. This is the money-beat guardrail.
func TestSelectWinnersExcludesZeroCorrect(t *testing.T) {
	base := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	solo := []Standing{{ParticipantID: "only", CorrectCount: 0, JoinedAt: base}}
	if winners := SelectWinners(solo, 1); len(winners) != 0 {
		t.Fatalf("zero-correct sole participant must not win, got %+v", winners)
	}

	mixed := []Standing{
		{ParticipantID: "zero", CorrectCount: 0, JoinedAt: base},
		{ParticipantID: "one", CorrectCount: 1, JoinedAt: base},
		{ParticipantID: "two", CorrectCount: 2, JoinedAt: base},
	}
	winners := SelectWinners(SortStandings(mixed), 2)
	if len(winners) != 2 || winners[0].ParticipantID != "two" || winners[1].ParticipantID != "one" {
		t.Fatalf("expected the two correct players, got %+v", winners)
	}
}

func TestFromNairaString(t *testing.T) {
	cases := map[string]int64{
		"25.00": 2500,
		"500":   50000,
		"0.01":  1,
		"10.5":  1050,
		"0":     0,
		"0007":  700,
	}
	for in, want := range cases {
		a, err := FromNairaString(in)
		if err != nil {
			t.Fatalf("FromNairaString(%q): %v", in, err)
		}
		if int64(a) != want {
			t.Fatalf("FromNairaString(%q) = %d, want %d", in, int64(a), want)
		}
	}
	if _, err := FromNairaString("1.234"); err == nil {
		t.Fatal("expected error for three decimal places")
	}
	if _, err := FromNairaString("abc"); err == nil {
		t.Fatal("expected error for non-numeric input")
	}
}

func TestNairaString(t *testing.T) {
	cases := map[int64]string{
		2500:  "25.00",
		50000: "500.00",
		1:     "0.01",
		1050:  "10.50",
		0:     "0.00",
	}
	for in, want := range cases {
		got := Amount(in).NairaString()
		if got != want {
			t.Fatalf("Amount(%d).NairaString() = %q, want %q", in, got, want)
		}
	}
}

func TestSplitPodiumSumsToPool(t *testing.T) {
	for _, pool := range []int64{2500, 100000, 33333, 100000000} {
		for n := 1; n <= 5; n++ {
			shares := SplitPodium(Amount(pool), n)
			if len(shares) != n {
				t.Fatalf("pool %d n %d: got %d shares", pool, n, len(shares))
			}
			var sum int64
			var prev int64 = int64(1) << 62
			for i, s := range shares {
				v := int64(s)
				if v < 0 {
					t.Fatalf("negative share %d", v)
				}
				if i > 0 && v > prev {
					t.Fatalf("shares not descending: %v", shares)
				}
				prev = v
				sum += v
			}
			if sum != pool {
				t.Fatalf("pool %d n %d: shares sum to %d", pool, n, sum)
			}
		}
	}
}

func TestValidateQuiz(t *testing.T) {
	good := &Quiz{
		ID:              "q1",
		Pool:            100000,
		WinnerCount:     3,
		Pacing:          PacingManual,
		DefaultDuration: time.Second,
		Questions: []Question{
			{ID: "a", Prompt: "p?", Options: []string{"x", "y"}, CorrectIndex: 1},
		},
	}
	if err := ValidateQuiz(good); err != nil {
		t.Fatalf("valid quiz rejected: %v", err)
	}

	badCases := []struct {
		name   string
		mutate func(*Quiz)
		want   error
	}{
		{"no questions", func(q *Quiz) { q.Questions = nil }, ErrNoQuestions},
		{"zero winners", func(q *Quiz) { q.WinnerCount = 0 }, ErrInvalidWinnerCount},
		{"bad pacing", func(q *Quiz) { q.Pacing = PacingMode("x") }, ErrInvalidPacing},
		{"duplicate id", func(q *Quiz) { q.Questions = append(q.Questions, q.Questions[0]) }, ErrDuplicateQuestionID},
		{"correct idx OOR", func(q *Quiz) { q.Questions[0].CorrectIndex = 5 }, ErrCorrectOptionInvalid},
	}
	for _, tc := range badCases {
		t.Run(tc.name, func(t *testing.T) {
			q := *good
			q.Questions = []Question{{ID: "a", Prompt: "p?", Options: []string{"x", "y"}, CorrectIndex: 1}}
			tc.mutate(&q)
			if err := ValidateQuiz(&q); err != tc.want {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

func TestClaimTransitions(t *testing.T) {
	valid := []struct{ from, to ClaimState }{
		{ClaimCreated, ClaimInvited},
		{ClaimCreated, ClaimOnboarded},
		{ClaimCreated, ClaimFailed},
		{ClaimInvited, ClaimOnboarded},
		{ClaimOnboarded, ClaimPaid},
	}
	for _, tr := range valid {
		if !CanTransition(tr.from, tr.to) {
			t.Fatalf("expected %s -> %s to be valid", tr.from, tr.to)
		}
	}
	invalid := []struct{ from, to ClaimState }{
		{ClaimCreated, ClaimPaid}, // cannot skip onboarding
		{ClaimPaid, ClaimCreated}, // terminal state
		{ClaimOnboarded, ClaimCreated},
	}
	for _, tr := range invalid {
		if CanTransition(tr.from, tr.to) {
			t.Fatalf("expected %s -> %s to be invalid", tr.from, tr.to)
		}
	}
}
