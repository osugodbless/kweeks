package domain

import "time"

// ClaimState tracks one winner's redemption lifecycle.
type ClaimState string

const (
	// ClaimCreated means the winner tapped redeem on their screen; a claim
	// row exists, a claim code was issued, and the recovery email (with the
	// code and a claim URL) was dispatched immediately.
	ClaimCreated ClaimState = "created"
	// ClaimBankSubmitted means the winner submitted Nigerian bank details and
	// the account was verified + registered on the host's rail.
	ClaimBankSubmitted ClaimState = "bank_submitted"
	// ClaimPaying means the offramp proposal was created/signed and the money
	// move is settling.
	ClaimPaying ClaimState = "paying"
	// ClaimPaid means the offramp settled and the bank payout COMPLETED.
	ClaimPaid ClaimState = "paid"
	// ClaimFailed means settlement errored and needs attention.
	ClaimFailed ClaimState = "failed"
)

// Claim is a winner's exactly-once redemption request. ClaimCode is issued to
// the winner's session AND emailed as a recovery artifact; it is the capability
// that authorizes the payout.
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

	// Nigerian bank payout details, populated as the claim moves through
	// bank_submitted -> paying -> paid.
	BankAccountID     string
	PayoutRef         string
	BankAccountNumber string
	BankName          string
	AccountHolderName string
}

// ValidTransitions returns the states reachable from a given state.
func ValidTransitions(from ClaimState) map[ClaimState]bool {
	switch from {
	case ClaimCreated:
		return map[ClaimState]bool{ClaimBankSubmitted: true, ClaimFailed: true}
	case ClaimBankSubmitted:
		return map[ClaimState]bool{ClaimPaying: true, ClaimFailed: true}
	case ClaimPaying:
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

// NigerianBank is one supported Nigerian bank (CBN code + full name) used to
// populate the winner's payout form.
type NigerianBank struct {
	Code string
	Name string
}

// NigerianAccount is the winner's Nigerian bank account, collected on the claim
// page and used to register the withdrawal account before the host offramps.
type NigerianAccount struct {
	AccountNumber     string
	BankCode          string
	BankName          string
	AccountHolderName string
}
