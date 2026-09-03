// Package app contains the application services that orchestrate domain rules
// against ports. It has no knowledge of HTTP, websockets, Postgres, or BMONI.
package app

import (
	"context"
	"errors"
	"time"

	"github.com/osugodbless/kweeks/internal/domain"
	"github.com/osugodbless/kweeks/internal/ports"
)

// Join handles player entry: email, nickname, avatar. Rejoining with the same
// email resumes the SAME participant (score and claim slot preserved).
type Join struct {
	store ports.Store
	clock ports.Clock
}

func NewJoin(store ports.Store, clock ports.Clock) *Join {
	return &Join{store: store, clock: clock}
}

func (j *Join) nowTime() time.Time {
	if j.clock != nil {
		return j.clock.Now()
	}
	return time.Now()
}

// JoinRoom merges a join into the existing participant row for (room, email)
// when one exists. The returned participant is always the canonical row. A new
// participant receives a fresh id and join time; rejoins return the original.
func (j *Join) JoinRoom(ctx context.Context, p *domain.Participant) (*domain.Participant, error) {
	if p == nil || p.Email == "" || p.RoomID == "" {
		return nil, domain.ErrParticipantNotFound
	}
	existing, err := j.store.GetParticipant(ctx, p.RoomID, p.Email)
	if err == nil && existing != nil {
		return existing, nil
	}
	if err != nil && !errors.Is(err, domain.ErrParticipantNotFound) {
		return nil, err
	}
	if p.ID == "" {
		p.ID = newID()
	}
	if p.JoinedAt.IsZero() {
		p.JoinedAt = j.nowTime()
	}
	return j.store.JoinParticipant(ctx, p)
}
