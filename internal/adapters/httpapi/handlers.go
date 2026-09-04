package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/osugodbless/kweeks/internal/domain"
)

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ---- Quiz authoring ----

type createQuizReq struct {
	Title           string        `json:"title"`
	PoolNaira       string        `json:"poolNaira"`
	WinnerCount     int           `json:"winnerCount"`
	Pacing          string        `json:"pacing"` // auto | manual
	DefaultDuration int64         `json:"defaultDurationMs"`
	Questions       []questionReq `json:"questions"`
}

type questionReq struct {
	ID           string   `json:"id"`
	Prompt       string   `json:"prompt"`
	Options      []string `json:"options"`
	CorrectIndex int      `json:"correctIndex"`
	DurationMs   int64    `json:"durationMs"`
}

func (s *Server) handleCreateQuiz(w http.ResponseWriter, r *http.Request) {
	var req createQuizReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, domain.ErrInvalidQuestion)
		return
	}
	quiz, err := quizFromReq(&req, instructorFrom(r), "")
	if err != nil {
		writeErr(w, err)
		return
	}
	if err := s.game.CreateQuiz(r.Context(), quiz); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": quiz.ID})
}

func (s *Server) handleUpdateQuiz(w http.ResponseWriter, r *http.Request) {
	var req createQuizReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, domain.ErrInvalidQuestion)
		return
	}
	quiz, err := quizFromReq(&req, instructorFrom(r), r.PathValue("quizID"))
	if err != nil {
		writeErr(w, err)
		return
	}
	if err := s.game.UpdateQuiz(r.Context(), quiz); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": quiz.ID})
}

func quizFromReq(req *createQuizReq, instructorID, id string) (*domain.Quiz, error) {
	pool, err := domain.FromNairaString(req.PoolNaira)
	if err != nil {
		return nil, domain.ErrInvalidQuestion
	}
	if id == "" {
		id = newID()
	}
	quiz := &domain.Quiz{
		ID:              id,
		InstructorID:    instructorID,
		Title:           req.Title,
		Pool:            pool,
		WinnerCount:     req.WinnerCount,
		Pacing:          domain.PacingMode(req.Pacing),
		DefaultDuration: time.Duration(req.DefaultDuration) * time.Millisecond,
		CreatedAt:       time.Now(),
	}
	for _, q := range req.Questions {
		quiz.Questions = append(quiz.Questions, domain.Question{
			ID:           q.ID,
			Prompt:       q.Prompt,
			Options:      q.Options,
			CorrectIndex: q.CorrectIndex,
			Duration:     time.Duration(q.DurationMs) * time.Millisecond,
		})
	}
	return quiz, nil
}

func (s *Server) handleListQuizzes(w http.ResponseWriter, r *http.Request) {
	quizzes, err := s.game.ListQuizzes(r.Context(), instructorFrom(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	// Never ship correct answers to the dashboard list.
	out := make([]map[string]any, 0, len(quizzes))
	for _, q := range quizzes {
		item := map[string]any{
			"id": q.ID, "title": q.Title, "poolNaira": q.Pool.DisplayString(),
			"winnerCount": q.WinnerCount, "pacing": q.Pacing, "questionCount": len(q.Questions),
			"createdAt": q.CreatedAt,
		}
		if room, err := s.game.LatestLiveRoom(r.Context(), q.ID); err == nil && room != nil {
			item["roomCode"] = room.Code
			item["roomId"] = room.ID
			item["state"] = room.State
		}
		out = append(out, item)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetQuiz(w http.ResponseWriter, r *http.Request) {
	quiz, err := s.game.GetQuiz(r.Context(), r.PathValue("quizID"))
	if err != nil {
		writeErr(w, err)
		return
	}
	// Instructor-only route: full questions (answers) are fine here.
	type qJSON struct {
		ID           string   `json:"id"`
		Prompt       string   `json:"prompt"`
		Options      []string `json:"options"`
		CorrectIndex int      `json:"correctIndex"`
		DurationMs   int64    `json:"durationMs"`
	}
	questions := make([]qJSON, 0, len(quiz.Questions))
	for _, q := range quiz.Questions {
		questions = append(questions, qJSON{q.ID, q.Prompt, q.Options, q.CorrectIndex, q.Duration.Milliseconds()})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id": quiz.ID, "title": quiz.Title, "poolNaira": quiz.Pool.DisplayString(),
		"winnerCount": quiz.WinnerCount, "pacing": quiz.Pacing,
		"defaultDurationMs": quiz.DefaultDuration.Milliseconds(),
		"questions":         questions,
	})
}

// ---- Rooms ----

type openRoomReq struct {
	QuizID string `json:"quizId"`
}

func (s *Server) handleOpenRoom(w http.ResponseWriter, r *http.Request) {
	var req openRoomReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, domain.ErrRoomNotFound)
		return
	}
	room := &domain.Room{ID: newID(), QuizID: req.QuizID, HostID: instructorFrom(r)}
	if err := s.game.OpenRoom(r.Context(), room); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": room.ID, "code": room.Code})
}

func (s *Server) handleStartRoom(w http.ResponseWriter, r *http.Request) {
	if err := s.game.Start(r.Context(), r.PathValue("roomID")); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleNextQuestion(w http.ResponseWriter, r *http.Request) {
	if err := s.game.Next(r.Context(), r.PathValue("roomID")); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleStandings(w http.ResponseWriter, r *http.Request) {
	standings, err := s.game.Standings(r.Context(), r.PathValue("roomID"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, standings)
}

func (s *Server) handleFinalizePodium(w http.ResponseWriter, r *http.Request) {
	winners, err := s.game.FinalizePodium(r.Context(), r.PathValue("roomID"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, winners)
}

// ---- Player join ----

type joinReq struct {
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	var req joinReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, domain.ErrParticipantNotFound)
		return
	}
	p, err := s.join.JoinRoom(r.Context(), &domain.Participant{
		RoomID: r.PathValue("roomID"), Email: req.Email, Nickname: req.Nickname, Avatar: req.Avatar,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// ---- Answer ----

type answerReq struct {
	ParticipantID string `json:"participantId"`
	QuestionID    string `json:"questionId"`
	OptionIndex   int    `json:"optionIndex"`
}

func (s *Server) handleAnswer(w http.ResponseWriter, r *http.Request) {
	var req answerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, domain.ErrAnswerUnknownQuestion)
		return
	}
	receipt, err := s.game.SubmitAnswer(r.Context(), &domain.Answer{
		ID: newID(), RoomID: r.PathValue("roomID"), ParticipantID: req.ParticipantID,
		QuestionID: req.QuestionID, OptionIndex: req.OptionIndex,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, receipt)
}

// ---- Redemption ----

type redeemReq struct {
	Email string `json:"email"`
}

func (s *Server) handleRedeem(w http.ResponseWriter, r *http.Request) {
	var req redeemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, domain.ErrNotWinner)
		return
	}
	roomID := r.PathValue("roomID")
	room, err := s.game.GetRoom(r.Context(), roomID)
	if err != nil {
		writeErr(w, err)
		return
	}
	quiz, err := s.game.GetQuiz(r.Context(), room.QuizID)
	if err != nil {
		writeErr(w, err)
		return
	}
	claim, err := s.red.CreateClaim(r.Context(), roomID, req.Email, quiz.Pool)
	if err != nil {
		writeErr(w, err)
		return
	}
	s.red.SendRedemptionEmail(r.Context(), claim)
	// The claim code is delivered in this response because the caller IS the
	// winner's own session (the code is never broadcast via standings/room
	// state). It is the capability that authorizes the claim.
	writeJSON(w, http.StatusOK, map[string]any{
		"id": claim.ID, "amountNaira": claim.Amount.DisplayString(), "state": claim.State,
		"claimCode": claim.ClaimCode,
	})
}
