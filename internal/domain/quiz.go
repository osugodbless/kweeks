package domain

import "time"

// PacingMode controls how a room advances between questions.
type PacingMode string

const (
	// PacingAuto advances questions on a server schedule using each
	// question's duration from its started_at.
	PacingAuto PacingMode = "auto"
	// PacingManual advances only when the instructor taps next.
	PacingManual PacingMode = "manual"
)

// Quiz is a fully-authored quiz. Questions and correct answers are stored
// server-side at authoring time and never shipped to clients.
type Quiz struct {
	ID           string
	InstructorID string
	Title        string

	Pool Amount // NGN prize pool in kobo
	// WinnerCount is the number of podium winners funded from Pool.
	WinnerCount int

	Pacing PacingMode
	// DefaultDuration applies to every question unless overridden.
	DefaultDuration time.Duration

	Questions []Question
	CreatedAt time.Time
}

// Question is a single authored question. CorrectIndex is the server-side
// answer key; it must never be rendered to players.
type Question struct {
	ID           string
	Prompt       string
	Options      []string
	CorrectIndex int
	// Duration overrides DefaultDuration when > 0.
	Duration time.Duration
}

// ValidateQuiz enforces authoring rules before a quiz is saved.
func ValidateQuiz(q *Quiz) error {
	if len(q.Questions) == 0 {
		return ErrNoQuestions
	}
	if q.WinnerCount < 1 {
		return ErrInvalidWinnerCount
	}
	if q.Pool < 0 {
		return ErrPoolInsufficient
	}
	if q.Pacing != PacingAuto && q.Pacing != PacingManual {
		return ErrInvalidPacing
	}
	if q.Pacing == PacingAuto && q.DefaultDuration <= 0 {
		return ErrInvalidQuestion
	}
	seen := map[string]bool{}
	for _, question := range q.Questions {
		if question.ID == "" {
			return ErrDuplicateQuestionID
		}
		if seen[question.ID] {
			return ErrDuplicateQuestionID
		}
		seen[question.ID] = true
		if question.Prompt == "" || len(question.Options) < 2 {
			return ErrInvalidQuestion
		}
		if question.CorrectIndex < 0 || question.CorrectIndex >= len(question.Options) {
			return ErrCorrectOptionInvalid
		}
	}
	return nil
}

// QuestionDuration returns the effective duration for a question, falling
// back to the quiz default when the question has no override.
func (q *Quiz) QuestionDuration(question Question) time.Duration {
	if question.Duration > 0 {
		return question.Duration
	}
	return q.DefaultDuration
}

// CurrentQuestion returns the question at idx, or the zero Question if idx is
// out of range.
func (q *Quiz) CurrentQuestion(idx int) Question {
	if idx < 0 || idx >= len(q.Questions) {
		return Question{}
	}
	return q.Questions[idx]
}
