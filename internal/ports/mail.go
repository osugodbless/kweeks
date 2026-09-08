package ports

import "context"

// Mail is the redemption-email boundary. Email is a recovery artifact, never
// the critical path: adapters must not fail the claim flow when mail fails.
type Mail interface {
	// SendRedemptionEmail is best-effort. claimURL is the public /claim page
	// pre-filled with the claim code so the winner can redeem even after
	// leaving the podium screen. Errors are logged by the caller and the claim
	// proceeds.
	SendRedemptionEmail(ctx context.Context, to, claimCode, amountNaira, claimURL string) error
}
