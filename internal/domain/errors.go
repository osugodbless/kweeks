package domain

import "errors"

// Sentinel errors for the domain rules. Adapters map these to transport
// statuses; they never leak as raw strings.
var (
	ErrQuizNotFound        = errors.New("quiz not found")
	ErrRoomNotFound        = errors.New("room not found")
	ErrParticipantNotFound = errors.New("participant not found")
	ErrClaimNotFound       = errors.New("claim not found")
	ErrUnauthorized        = errors.New("unauthorized")

	ErrQuizAlreadyStarted    = errors.New("quiz already started")
	ErrRoomWrongState        = errors.New("room is not in the required state")
	ErrNoQuestions           = errors.New("quiz must contain at least one question")
	ErrInvalidQuestion       = errors.New("invalid question: prompt or options malformed")
	ErrDuplicateQuestionID   = errors.New("duplicate question id")
	ErrCorrectOptionInvalid  = errors.New("correct answer index out of range")
	ErrInvalidWinnerCount    = errors.New("winner count must be at least one")
	ErrInvalidOptionIndex    = errors.New("option index out of range")
	ErrDuplicateParticipant  = errors.New("participant email already joined")
	ErrAnswerLate            = errors.New("answer submitted after the cutoff")
	ErrAnswerUnknownQuestion = errors.New("answer references a question not in this quiz")
	ErrAlreadyAnswered       = errors.New("participant already answered this question")
	ErrRoomNotLive           = errors.New("room is not live")
	ErrQuestionNotCurrent    = errors.New("question is not the current one")
	ErrNoWinners             = errors.New("no winners to declare")
	ErrNotWinner             = errors.New("participant is not a podium winner")
	ErrClaimExists           = errors.New("claim already exists for this winner")
	ErrBadClaimCode          = errors.New("invalid claim code")
	ErrInvalidTransition     = errors.New("invalid claim state transition")
	ErrInvalidPacing         = errors.New("invalid pacing mode")
	ErrPoolInsufficient      = errors.New("pool must be positive to fund a podium")
)
