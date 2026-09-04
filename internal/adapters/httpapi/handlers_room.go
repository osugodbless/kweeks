package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/osugodbless/kweeks/internal/domain"
)

// publicQuestion is the never-correct-answers view a player needs.
type publicQuestion struct {
	ID          string   `json:"id"`
	Index       int      `json:"index"`
	Prompt      string   `json:"prompt"`
	Options     []string `json:"options"`
	StartedAt   string   `json:"startedAt"`
	DurationMs  int64    `json:"durationMs"`
	RemainingMs int64    `json:"remainingMs"`
}

func (s *Server) buildRoomState(ctx context.Context, room *domain.Room, quiz *domain.Quiz, participants []domain.Participant) map[string]any {
	state := map[string]any{
		"id": room.ID, "code": room.Code, "quizId": room.QuizID, "title": quiz.Title,
		"poolNaira": quiz.Pool.DisplayString(), "winnerCount": quiz.WinnerCount,
		"pacing": quiz.Pacing, "state": room.State,
		"questionCount": len(quiz.Questions), "currentIndex": room.CurrentQuestionIdx,
		"participantCount": len(participants), "stateAt": time.Now(),
		"host":            map[string]any{},
		"currentQuestion": nil,
		"winners":         nil,
	}
	plist := make([]map[string]any, 0, len(participants))
	for _, p := range participants {
		plist = append(plist, map[string]any{"id": p.ID, "nickname": p.Nickname, "avatar": p.Avatar})
	}
	state["participants"] = plist

	if room.State == domain.RoomLive && room.CurrentQuestionIdx >= 0 && room.CurrentQuestionIdx < len(quiz.Questions) {
		q := quiz.Questions[room.CurrentQuestionIdx]
		elapsed := time.Since(room.QuestionStartedAt)
		remaining := time.Duration(quiz.QuestionDuration(q)) - elapsed
		if remaining < 0 {
			remaining = 0
		}
		state["currentQuestion"] = publicQuestion{
			ID: q.ID, Index: room.CurrentQuestionIdx, Prompt: q.Prompt, Options: q.Options,
			StartedAt:   room.QuestionStartedAt.Format(time.RFC3339Nano),
			DurationMs:  quiz.QuestionDuration(q).Milliseconds(),
			RemainingMs: remaining.Milliseconds(),
		}
	}
	if room.State == domain.RoomPodium {
		if winners, err := s.game.Winners(ctx, room.ID); err == nil {
			state["winners"] = winners
		}
	}
	return state
}

func (s *Server) handleRoomState(w http.ResponseWriter, r *http.Request) {
	room, err := s.game.GetRoom(r.Context(), r.PathValue("roomID"))
	if err != nil {
		writeErr(w, err)
		return
	}
	quiz, err := s.game.GetQuiz(r.Context(), room.QuizID)
	if err != nil {
		writeErr(w, err)
		return
	}
	participants, err := s.game.RoomParticipants(r.Context(), room.ID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s.buildRoomState(r.Context(), room, quiz, participants))
}

// handleRoomStateByCode is the player entry point: look a room up by its
// 4-letter code and return the public state.
func (s *Server) handleRoomStateByCode(w http.ResponseWriter, r *http.Request) {
	room, err := s.game.GetRoomByCode(r.Context(), r.PathValue("code"))
	if err != nil {
		writeErr(w, err)
		return
	}
	quiz, err := s.game.GetQuiz(r.Context(), room.QuizID)
	if err != nil {
		writeErr(w, err)
		return
	}
	participants, err := s.game.RoomParticipants(r.Context(), room.ID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s.buildRoomState(r.Context(), room, quiz, participants))
}
