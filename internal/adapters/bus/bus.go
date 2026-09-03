// Package bus provides realtime fan-out implementations of ports.Bus.
package bus

import (
	"context"
	"sync"

	"github.com/osugodbless/kweeks/internal/ports"
)

// InMemory is a ports.Bus that keeps per-room subscriber channels. It is the
// unit-test double and the demo fallback when no websocket hub is wired.
type InMemory struct {
	mu   sync.Mutex
	subs map[string][]chan ports.Event
}

func NewInMemory() *InMemory {
	return &InMemory{subs: map[string][]chan ports.Event{}}
}

func (b *InMemory) Publish(ctx context.Context, e ports.Event) error {
	b.mu.Lock()
	subs := make([]chan ports.Event, 0, len(b.subs[e.RoomID]))
	for _, ch := range b.subs[e.RoomID] {
		select {
		case ch <- e:
			subs = append(subs, ch) // still subscribed
		default:
			// Slow consumer: drop for this event but keep the channel.
			subs = append(subs, ch)
		}
	}
	b.subs[e.RoomID] = subs
	b.mu.Unlock()
	return nil
}

func (b *InMemory) Subscribe(roomID string) (<-chan ports.Event, func()) {
	ch := make(chan ports.Event, 64)
	b.mu.Lock()
	b.subs[roomID] = append(b.subs[roomID], ch)
	b.mu.Unlock()

	cancel := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		remaining := b.subs[roomID][:0]
		for _, c := range b.subs[roomID] {
			if c != ch {
				remaining = append(remaining, c)
			}
		}
		b.subs[roomID] = remaining
		close(ch)
	}
	return ch, cancel
}
