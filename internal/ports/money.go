package ports

import (
	"context"

	"github.com/osugodbless/kweeks/internal/domain"
)

// Money is the BMONI boundary. It mirrors the strict server-side flow from the
// docs: create user → submit KYC → owner-proof → create-managed → start-nigeria
// → (documents) → wallet ready to fund. The application layer drives each step
// with user-supplied data; the adapter handles signing and proposal mechanics.
type Money interface {
	// CreateUser registers a BMONI user from the host's real identity and
	// returns the bmoniUserId used as {userId} for every later call. A 409
	// (already exists) is surfaced for the caller to recover by phone/email.
	CreateUser(ctx context.Context, id domain.UserIdentity) (string, error)
	// SubmitKYC writes the host's KYC profile (personalInfo + address + bvn).
	SubmitKYC(ctx context.Context, userID string, k domain.KYCProfile) error
	// LookupBVN resolves a BVN to its holder record (GET /kyc/bvn-lookup/{bvn})
	// so the KYC form can pre-fill and the host confirm. Writes nothing.
	LookupBVN(ctx context.Context, userID, bvn string) (*domain.BVNRecord, error)
	// UploadKycDocument submits one KYC document image (multipart) for the
	// user. The image is sent under the `files` multipart field (selfie for
	// biometric) with the required document-type metadata.
	UploadKycDocument(ctx context.Context, userID string, doc domain.KycDocument) error
	// CreateWallet provisions a CNGN smart wallet for the user via the
	// owner-proof challenge + create-managed handshake. Returns wallet id +
	// on-chain address.
	CreateWallet(ctx context.Context, userID string) (walletID, addr string, err error)
	// ActivateRail runs POST /onboarding/start-nigeria: activates the NGN rail
	// against the wallet address using the host's BVN.
	ActivateRail(ctx context.Context, userID, walletAddr, bvn string) error
	// DepositAccount returns the host's NGN virtual bank account (number +
	// bank) that bank transfers to it fund the wallet.
	DepositAccount(ctx context.Context, userID, walletID string) (accountNumber, bankName string, err error)
	// ListNigerianBanks returns the supported Nigerian banks for the payout
	// form.
	ListNigerianBanks(ctx context.Context, userID string) ([]domain.NigerianBank, error)
	// VerifyNigerianAccount resolves an account number + bank code to the
	// registered account holder name.
	VerifyNigerianAccount(ctx context.Context, userID, accountNumber, bankCode string) (string, error)
	// RegisterNigerianWithdrawalAccount get-or-creates the withdrawal account
	// and returns its bankAccountId for the offramp call.
	RegisterNigerianWithdrawalAccount(ctx context.Context, userID string, acct domain.NigerianAccount) (string, error)
	// PayWinnerToNigerianBank offramps prize money from the host wallet to a
	// registered Nigerian bank account. Returns the settlement reference.
	PayWinnerToNigerianBank(ctx context.Context, from *domain.WalletExternal, bankAccountID string, amount domain.Amount) (string, error)
}
