package domain

import (
	"encoding/json"
	"time"
)

// Standing is one row of the live leaderboard. Ordering is server-owned and
// deterministic: CorrectCount desc, then TotalLatency asc (earlier correct
// answers win), then JoinedAt asc (earliest joiner wins the tie).
type Standing struct {
	ParticipantID string
	Nickname      string
	Avatar        string
	CorrectCount  int
	TotalLatency  time.Duration // sum of accepted-correct latency (tie-break order)
	JoinedAt      time.Time
}

// MarshalJSON renders TotalLatency as milliseconds on the wire; a time.Duration
// would otherwise serialize as raw nanoseconds.
func (s Standing) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ParticipantID string    `json:"participantId"`
		Nickname      string    `json:"nickname"`
		Avatar        string    `json:"avatar"`
		CorrectCount  int       `json:"correctCount"`
		TotalLatency  int64     `json:"totalLatencyMs"`
		JoinedAt      time.Time `json:"joinedAt"`
	}{
		ParticipantID: s.ParticipantID,
		Nickname:      s.Nickname,
		Avatar:        s.Avatar,
		CorrectCount:  s.CorrectCount,
		TotalLatency:  s.TotalLatency.Milliseconds(),
		JoinedAt:      s.JoinedAt,
	})
}
