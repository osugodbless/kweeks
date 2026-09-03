package domain

import "time"

// Participant is a player in a room. Email is the identity anchor for claims;
// a rejoin with the same email resumes the SAME participant.
type Participant struct {
	ID       string    `json:"id"`
	RoomID   string    `json:"roomId"`
	Email    string    `json:"email"`
	Nickname string    `json:"nickname"`
	Avatar   string    `json:"avatar"` // emoji from the fixed set
	JoinedAt time.Time `json:"joinedAt"`
}

// Answer is a participant's submission for one question. ReceivedAt is the
// server acceptance time and is the only timing that counts.
type Answer struct {
	ID            string
	RoomID        string
	ParticipantID string
	QuestionID    string
	OptionIndex   int
	Correct       bool
	// ReceivedAt is the server clock time of acceptance.
	ReceivedAt time.Time
	// QuestionStartedAt is copied from the room for speed scoring.
	QuestionStartedAt time.Time
}

// AnswerReceipt is what an accepted answer write produces. Score is 0 unless
// the answer is correct; latencyMs measures time from question start to the
// server acceptance instant.
type AnswerReceipt struct {
	ID        string `json:"id"`
	Correct   bool   `json:"correct"`
	Score     int    `json:"score"`
	LatencyMs int64  `json:"latencyMs"`
}
