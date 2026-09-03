package clock

import "time"

// Real is the production clock: wall-clock time.
type Real struct{}

func NewReal() *Real { return &Real{} }

func (Real) Now() time.Time { return time.Now() }

// Static is a test clock frozen at a fixed instant.
type Static struct {
	t time.Time
}

func NewStatic(t time.Time) *Static { return &Static{t: t} }

func (s *Static) Now() time.Time { return s.t }

// Set moves the static clock (used to advance time in tests).
func (s *Static) Set(t time.Time) { s.t = t }
