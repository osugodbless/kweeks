package domain

import "time"

// RoomState is the coarse phase of a live room.
type RoomState string

const (
	RoomLobby  RoomState = "lobby"
	RoomLive   RoomState = "live"
	RoomPodium RoomState = "podium"
	RoomEnded  RoomState = "ended"
)

// Room is a single live run of a quiz. The room row is the authoritative
// source of current question + started_at, so any client can resync by
// reading it.
type Room struct {
	ID     string
	QuizID string
	State  RoomState
	Pacing PacingMode
	HostID string

	// CurrentQuestionIdx is -1 before the first question starts.
	CurrentQuestionIdx int
	// QuestionStartedAt is the server epoch when the current question was
	// broadcast. Answer cutoff = QuestionStartedAt + question duration.
	QuestionStartedAt time.Time
	// StartedAt is when the room left the lobby.
	StartedAt time.Time
}
