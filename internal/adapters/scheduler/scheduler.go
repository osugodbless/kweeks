// Package scheduler advances AUTO-paced rooms on a server schedule. No client
// tab is the authority; the scheduler polls live rooms and advances any whose
// current question window has elapsed.
package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/osugodbless/kweeks/internal/app"
	"github.com/osugodbless/kweeks/internal/domain"
	"github.com/osugodbless/kweeks/internal/ports"
)

// Scheduler advances AUTO rooms and finalizes podiums when the last question
// window closes.
type Scheduler struct {
	game   *app.Game
	store  ports.Store
	logger *slog.Logger
	period time.Duration
}

// New builds a scheduler that wakes every period (e.g. 500ms) to advance rooms.
func New(game *app.Game, store ports.Store, logger *slog.Logger, period time.Duration) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}
	if period <= 0 {
		period = 500 * time.Millisecond
	}
	return &Scheduler{game: game, store: store, logger: logger, period: period}
}

// Run blocks, advancing eligible AUTO rooms until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.period)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case now := <-ticker.C:
			s.tick(ctx, now)
		}
	}
}

// tick advances every live AUTO room whose current window has closed. Rooms in
// manual pacing are never advanced here; only the instructor's next tap moves
// them.
func (s *Scheduler) tick(ctx context.Context, now time.Time) {
	// The store exposes room lookup by id; for the demo the scheduler needs an
	// index of live rooms. LiveRooms returns rooms in a live state.
	rooms, err := s.store.ListLiveRooms(ctx)
	if err != nil {
		s.logger.Warn("scheduler: cannot list live rooms", "err", err)
		return
	}
	for _, room := range rooms {
		quiz, err := s.store.GetQuiz(ctx, room.QuizID)
		if err != nil {
			s.logger.Warn("scheduler: missing quiz for room", "room", room.ID, "err", err)
			continue
		}
		// Pacing is a quiz attribute, not persisted on the room, so the
		// scheduler decides from the quiz. Auto rooms advance here; manual
		// rooms only move on the instructor's next tap.
		if quiz.Pacing != domain.PacingAuto {
			continue
		}
		q := quiz.CurrentQuestion(room.CurrentQuestionIdx)
		if q.ID == "" {
			// Room past its last question: resolve to podium, or end it when
			// nobody took the pool.
			s.resolve(ctx, room.ID)
			continue
		}
		if now.Before(room.QuestionStartedAt.Add(quiz.QuestionDuration(q))) {
			continue // window still open
		}
		if err := s.game.Next(ctx, room.ID); err != nil {
			if err == domain.ErrRoomWrongState {
				// No next question: the last window closed, so resolve.
				s.resolve(ctx, room.ID)
				continue
			}
			s.logger.Warn("scheduler: advance", "room", room.ID, "err", err)
		}
	}
}

// resolve finalizes the podium for a past-last room. When no eligible winner
// exists (e.g. nobody answered correctly), it ends the room so it reaches a
// terminal state instead of lingering live and retrying forever.
func (s *Scheduler) resolve(ctx context.Context, roomID string) {
	if _, err := s.game.FinalizePodium(ctx, roomID); err == nil {
		return
	} else if err != domain.ErrNoWinners {
		s.logger.Warn("scheduler: finalize", "room", roomID, "err", err)
		return
	}
	// No winners: end the room. Errors here are logged; the room may already
	// be ended by a concurrent tick, which is fine.
	if err := s.game.EndRoom(ctx, roomID); err != nil {
		s.logger.Warn("scheduler: end no-winner room", "room", roomID, "err", err)
	}
}
