package ports

import (
	"context"

	"github.com/osugodbless/kweeks/internal/domain"
)

// Money is the BMONI boundary. The application layer only knows it can
// provision a user + CNGN wallet, read the platform balance, and move NGN to
// a recipient; how that happens (owner-proof, create-managed,
// proposal/approve/sign/send) is the adapter's problem. A nil Money on a
// service means the rail is disabled and callers degrade to the local ledger.
type Money interface {
	// BalanceNGN returns the current NGN balance of the platform wallet in
	// kobo. Callers treat this as a cached value, never a live dependency.
	BalanceNGN(ctx context.Context) (domain.Amount, error)
	// PayWinner sends amount to the recipient identified by their BMONI
	// identity (email or user id). Returns the settlement reference.
	PayWinner(ctx context.Context, toEmail string, amount domain.Amount) (string, error)
	// Provision creates a BMONI user + CNGN smart wallet for an instructor and
	// returns the external identity. The rail must be configured.
	Provision(ctx context.Context, p domain.BmoniPersona) (*domain.WalletExternal, error)
	// CreditNGN funds a provisioned CNGN wallet from the platform funding
	// wallet. Returns the settlement reference.
	CreditNGN(ctx context.Context, external *domain.WalletExternal, amount domain.Amount) (string, error)
	// PayWinnerFrom sends prize money from an instructor's wallet to a winner
	// wallet address. Returns the settlement reference.
	PayWinnerFrom(ctx context.Context, from *domain.WalletExternal, toAddr string, amount domain.Amount) (string, error)
}
