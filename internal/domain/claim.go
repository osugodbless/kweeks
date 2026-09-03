package domain

import "time"

// ClaimState tracks one winner's redemption lifecycle.
type ClaimState string

const (
	// ClaimCreated means the winner tapped redeem on their screen; a claim
	// row exists and is idempotent per (winner email, quiz).
	ClaimCreated ClaimState = "created"
	// ClaimInvited means the redemption email (recovery artifact) was sent.
	ClaimInvited ClaimState = "invited"
	// ClaimOnboarded means the winner completed the BMONI no-app onboarding.
	ClaimOnboarded ClaimState = "onboarded"
	// ClaimPaid means the BMONI transfer reached COMPLETED.
	ClaimPaid ClaimState = "paid"
	// ClaimFailed means settlement errored and needs attention.
	ClaimFailed ClaimState = "failed"
)

// Claim is a winner's exactly-once redemption request. ClaimCode is issued to
// the winner's session only and is the capability that authorizes the claim.
type Claim struct {
	ID        string
	QuizID    string
	RoomID    string
	Email     string
	Amount    Amount
	ClaimCode string
	State     ClaimState
	CreatedAt time.Time
	PaidAt    *time.Time
}

// ValidTransitions returns the states reachable from a given state.
func ValidTransitions(from ClaimState) map[ClaimState]bool {
	switch from {
	case ClaimCreated:
		return map[ClaimState]bool{ClaimInvited: true, ClaimOnboarded: true, ClaimFailed: true}
	case ClaimInvited:
		return map[ClaimState]bool{ClaimOnboarded: true, ClaimFailed: true}
	case ClaimOnboarded:
		return map[ClaimState]bool{ClaimPaid: true, ClaimFailed: true}
	case ClaimPaid, ClaimFailed:
		return map[ClaimState]bool{}
	default:
		return map[ClaimState]bool{}
	}
}

// CanTransition reports whether to is a legal successor of from.
func CanTransition(from, to ClaimState) bool {
	return ValidTransitions(from)[to]
}
