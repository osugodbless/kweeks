// Package httpapi exposes the kweeks REST surface: instructor auth + wallet,
// quiz authoring and room control, player join/answer, room state, and winner
// redemption.
package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/osugodbless/kweeks/internal/app"
	"github.com/osugodbless/kweeks/internal/domain"
)

// Server routes HTTP requests to the application services.
type Server struct {
	game *app.Game
	join *app.Join
	red  *app.Redemption

	auth   *app.Auth
	wallet *app.Wallet

	// wsHub, when set, serves the realtime room socket. Nil disables the
	// websocket route (unit tests and REST-only deployments).
	wsHub interface {
		ServeWS(w http.ResponseWriter, r *http.Request, roomID string)
	}
}

// New builds an httpapi Server.
func New(game *app.Game, join *app.Join, red *app.Redemption) *Server {
	return &Server{game: game, join: join, red: red}
}

// WithWS attaches a websocket hub so player sockets can stream room events.
func (s *Server) WithWS(hub interface {
	ServeWS(w http.ResponseWriter, r *http.Request, roomID string)
}) *Server {
	s.wsHub = hub
	return s
}

// WithServices attaches the auth + wallet services (instructor surface).
func (s *Server) WithServices(auth *app.Auth, wallet *app.Wallet) *Server {
	s.auth = auth
	s.wallet = wallet
	return s
}

// Routes registers every handler on the mux.
func (s *Server) Routes(mux *http.ServeMux) {
	// Auth + wallet
	mux.HandleFunc("POST /api/auth/signup", s.handleSignup)
	mux.HandleFunc("POST /api/auth/login", s.handleLogin)
	mux.HandleFunc("GET /api/auth/me", s.requireAuth(s.handleMe))
	mux.HandleFunc("GET /api/wallet", s.requireAuth(s.handleWallet))
	mux.HandleFunc("POST /api/wallet/fund", s.requireAuth(s.handleFundWallet))
	mux.HandleFunc("GET /api/instructor/dashboard", s.requireAuth(s.handleDashboard))
	mux.HandleFunc("GET /api/instructor/history", s.requireAuth(s.handleHistory))

	// Quiz authoring
	mux.HandleFunc("POST /api/quizzes", s.requireAuth(s.handleCreateQuiz))
	mux.HandleFunc("GET /api/quizzes", s.requireAuth(s.handleListQuizzes))
	mux.HandleFunc("GET /api/quizzes/{quizID}", s.requireAuth(s.handleGetQuiz))
	mux.HandleFunc("PUT /api/quizzes/{quizID}", s.requireAuth(s.handleUpdateQuiz))

	// Rooms
	mux.HandleFunc("POST /api/rooms", s.requireAuth(s.handleOpenRoom))
	mux.HandleFunc("POST /api/rooms/{roomID}/start", s.handleStartRoom)
	mux.HandleFunc("POST /api/rooms/{roomID}/next", s.handleNextQuestion)
	mux.HandleFunc("GET /api/rooms/{roomID}/standings", s.handleStandings)
	mux.HandleFunc("POST /api/rooms/{roomID}/podium", s.handleFinalizePodium)
	mux.HandleFunc("GET /api/rooms/{roomID}", s.handleRoomState)
	mux.HandleFunc("GET /api/lookup/{code}", s.handleRoomStateByCode)

	// Player join/answer
	mux.HandleFunc("POST /api/rooms/{roomID}/join", s.handleJoin)
	mux.HandleFunc("POST /api/rooms/{roomID}/answer", s.handleAnswer)

	// Winner redemption
	mux.HandleFunc("POST /api/rooms/{roomID}/redeem", s.handleRedeem)

	if s.wsHub != nil {
		mux.HandleFunc("GET /api/rooms/{roomID}/ws", s.handleWS)
	}
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("roomID")
	s.wsHub.ServeWS(w, r, roomID)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	switch {
	case errors.Is(err, domain.ErrQuizNotFound),
		errors.Is(err, domain.ErrRoomNotFound),
		errors.Is(err, domain.ErrParticipantNotFound),
		errors.Is(err, domain.ErrClaimNotFound),
		errors.Is(err, domain.ErrInstructorNotFound),
		errors.Is(err, domain.ErrSessionNotFound),
		errors.Is(err, domain.ErrWalletNotFound):
		code = http.StatusNotFound
	case errors.Is(err, domain.ErrUnauthorized), errors.Is(err, domain.ErrBadCredentials):
		code = http.StatusUnauthorized
	case errors.Is(err, domain.ErrEmailTaken), errors.Is(err, domain.ErrWalletExists),
		errors.Is(err, domain.ErrRoomCodeTaken):
		code = http.StatusConflict
	case errors.Is(err, domain.ErrInsufficientBalance):
		code = http.StatusUnprocessableEntity
	case errors.Is(err, domain.ErrNotWinner), errors.Is(err, domain.ErrBadClaimCode):
		code = http.StatusForbidden
	case errors.Is(err, domain.ErrAnswerLate), errors.Is(err, domain.ErrAlreadyAnswered),
		errors.Is(err, domain.ErrAnswerUnknownQuestion), errors.Is(err, domain.ErrInvalidOptionIndex),
		errors.Is(err, domain.ErrRoomWrongState), errors.Is(err, domain.ErrRoomNotLive),
		errors.Is(err, domain.ErrQuizAlreadyStarted), errors.Is(err, domain.ErrNoQuestions),
		errors.Is(err, domain.ErrInvalidQuestion), errors.Is(err, domain.ErrDuplicateQuestionID),
		errors.Is(err, domain.ErrCorrectOptionInvalid), errors.Is(err, domain.ErrInvalidWinnerCount),
		errors.Is(err, domain.ErrInvalidPacing), errors.Is(err, domain.ErrClaimExists),
		errors.Is(err, domain.ErrNoWinners), errors.Is(err, domain.ErrDuplicateParticipant),
		errors.Is(err, domain.ErrInvalidTransition):
		code = http.StatusBadRequest
	}
	writeJSON(w, code, map[string]string{"error": err.Error()})
}
