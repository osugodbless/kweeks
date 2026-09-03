package ports

import (
	"context"

	"github.com/osugodbless/kweeks/internal/domain"
)

// Money is the BMONI boundary. The application layer only knows it can move
// an amount of NGN to a recipient identity; how that happens (proposal,
// approve, sign, send) is the adapter's problem.
type Money interface {
	// BalanceNGN returns the current NGN balance of the platform wallet in
	// kobo. Callers treat this as a cached value, never a live dependency.
	BalanceNGN(ctx context.Context) (domain.Amount, error)
	// PayWinner sends amount to the recipient identified by their BMONI
	// identity (email or user id). Returns the settlement reference.
	PayWinner(ctx context.Context, toEmail string, amount domain.Amount) (string, error)
}
