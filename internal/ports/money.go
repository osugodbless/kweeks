package ports

import (
	"context"

	"github.com/osugodbless/kweeks/internal/domain"
)

// Money is the BMONI boundary. The application layer only knows it can
// provision a user + CNGN wallet for an instructor, and pay a winner's prize
// to a Nigerian bank account through the host's wallet (verify -> register ->
// offramp). How that happens (owner-proof, create-managed,
// proposal/approve/sign/send) is the adapter's problem. A nil Money on a
// service means the rail is disabled and callers degrade to the local ledger.
type Money interface {
	// Provision creates a BMONI user + CNGN smart wallet + active NGN rail for
	// an instructor and returns the external identity. The rail must be
	// configured.
	Provision(ctx context.Context, p domain.BmoniPersona) (*domain.WalletExternal, error)
	// ListNigerianBanks returns the supported Nigerian banks (CBN code + full
	// name) for the claim form. Scoped to the host user.
	ListNigerianBanks(ctx context.Context, userID string) ([]domain.NigerianBank, error)
	// VerifyNigerianAccount resolves an account number + bank code to the
	// registered account holder name (exact name the registration requires).
	VerifyNigerianAccount(ctx context.Context, userID, accountNumber, bankCode string) (string, error)
	// RegisterNigerianWithdrawalAccount get-or-creates the withdrawal account
	// and returns its bankAccountId for the offramp call.
	RegisterNigerianWithdrawalAccount(ctx context.Context, userID string, acct domain.NigerianAccount) (string, error)
	// PayWinnerToNigerianBank offramps prize money from the host wallet to a
	// registered Nigerian bank account. Returns the settlement reference.
	PayWinnerToNigerianBank(ctx context.Context, from *domain.WalletExternal, bankAccountID string, amount domain.Amount) (string, error)
	// DepositAccount returns the host's NGN virtual bank account (number +
	// bank) that bank transfers can be sent to in order to fund the wallet.
	DepositAccount(ctx context.Context, userID, walletID string) (accountNumber, bankName string, err error)
}
