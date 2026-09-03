package ports

import "context"

// Event is a realtime room event pushed to connected clients.
type Event struct {
	RoomID string
	Type   string // "question", "standings", "podium", "lobby", "ended", "claim"
	// Payload is transport-neutral: JSON-marshalable struct. The ws adapter
	// decides exact wire shape.
	Payload any
}

// Bus is the realtime fan-out boundary. One room = one logical channel.
type Bus interface {
	// Publish sends an event to every subscriber of a room.
	Publish(ctx context.Context, e Event) error
	// Subscribe returns a channel of events for a room. The caller must
	// Close the returned subscription when done.
	Subscribe(roomID string) (<-chan Event, func())
}
