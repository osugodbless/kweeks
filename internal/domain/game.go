package domain

import (
	"sort"
	"time"
)

// AnswerGate decides whether an answer submitted at `now` is accepted for a
// question that started at `startedAt` with the given duration. The server is
// the sole authority: acceptance is purely a function of these timestamps,
// never of client receipt.
func AnswerGate(now, startedAt time.Time, duration time.Duration) bool {
	if startedAt.IsZero() {
		return false
	}
	cutoff := startedAt.Add(duration)
	return now.Before(cutoff)
}

// ScoreFor returns the score of a single answer: 1 when correct, else 0.
func ScoreFor(correct bool) int {
	if correct {
		return 1
	}
	return 0
}

// LatencyMs returns the milliseconds between a question start and an answer's
// server acceptance. Used for the deterministic tie-break (earlier correct
// answers win).
func LatencyMs(receivedAt, startedAt time.Time) int64 {
	return receivedAt.Sub(startedAt).Milliseconds()
}

// StandingLess reports whether a ranks above b on the podium ordering.
func StandingLess(a, b Standing) bool {
	if a.CorrectCount != b.CorrectCount {
		return a.CorrectCount > b.CorrectCount
	}
	if a.TotalLatency != b.TotalLatency {
		return a.TotalLatency < b.TotalLatency
	}
	return a.JoinedAt.Before(b.JoinedAt)
}

// SortStandings returns a new slice of standings ordered by the podium
// ordering (best first). The input is not mutated.
func SortStandings(in []Standing) []Standing {
	out := make([]Standing, len(in))
	copy(out, in)
	sort.SliceStable(out, func(i, j int) bool { return StandingLess(out[i], out[j]) })
	return out
}

// WinningsEligible reports whether a standing actually placed: it must have at
// least one correct answer. A participant who answered nothing (or nothing
// correctly) does not take the pool, even if they are the only player.
func WinningsEligible(s Standing) bool {
	return s.CorrectCount > 0
}

// SelectWinners returns the top n eligible standings by podium ordering. Only
// standings with at least one correct answer qualify for a payout.
func SelectWinners(sorted []Standing, n int) []Standing {
	if n <= 0 || len(sorted) == 0 {
		return nil
	}
	eligible := sorted[:0:0]
	for _, s := range sorted {
		if WinningsEligible(s) {
			eligible = append(eligible, s)
		}
	}
	if len(eligible) == 0 {
		return nil
	}
	if n > len(eligible) {
		n = len(eligible)
	}
	return eligible[:n]
}
