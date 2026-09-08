package domain

import "time"

// Instructor is an account that hosts quizzes. Signup provisions one and a
// NGN wallet is issued for it immediately.
type Instructor struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone,omitempty"` // E.164; the host's own BMONI user identity
	PasswordHash string    `json:"-"`
	Avatar       string    `json:"avatar"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Session is a bearer token bound to an instructor.
type Session struct {
	Token        string
	InstructorID string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

// Wallet is an instructor's NGN balance ledger identity. The demo treats it as
// the BMONI-managed balance; `credit` funding is the sandbox test-credit path.
type Wallet struct {
	ID           string    `json:"id"`
	InstructorID string    `json:"instructorId"`
	Balance      Amount    `json:"balanceNaira"` // serialized as naira string
	CreatedAt    time.Time `json:"createdAt"`

	// External BMONI identity, set as the wallet moves through the strict
	// provisioning flow (create-user → KYC → wallet → rail).
	BmoniUserID         string `json:"bmoniUserId,omitempty"`
	BmoniKYCSubmitted   bool   `json:"bmoniKycSubmitted,omitempty"`
	BmoniWalletID       string `json:"bmoniWalletId,omitempty"`
	BmoniWalletAddr     string `json:"bmoniWalletAddress,omitempty"`
	BmoniRailActive     bool   `json:"bmoniRailActive,omitempty"`
}

// WalletTxKind classifies a wallet ledger entry.
type WalletTxKind string

const (
	TxFund   WalletTxKind = "fund"
	TxPool   WalletTxKind = "pool"
	TxPayout WalletTxKind = "payout"
	TxCredit WalletTxKind = "credit"
)

// WalletExternal is the subset of BMONI provisioning results persisted on a
// wallet row.
type WalletExternal struct {
	UserID   string
	WalletID string
	Address  string
}

// SetupStage marks where an instructor's wallet is in the strict BMONI
// provisioning flow: create-user → KYC → wallet → rail → ready.
type SetupStage string

const (
	// SetupUnprovisioned means no BMONI user exists yet (rail off, or
	// create-user failed at signup).
	SetupUnprovisioned SetupStage = "unprovisioned"
	// SetupKYC means the user exists but the host has not submitted KYC.
	SetupKYC SetupStage = "kyc"
	// SetupWallet means KYC is done but the smart wallet is not created.
	SetupWallet SetupStage = "wallet"
	// SetupRail means the wallet exists but the NGN rail is not active.
	SetupRail SetupStage = "rail"
	// SetupReady means the wallet is provisioned and ready to fund.
	SetupReady SetupStage = "ready"
)

// SetupStatus is the wallet provisioning state + next action, surfaced by the
// setup wizard.
type SetupStatus struct {
	Stage               SetupStage
	BmoniUserID         string
	KYCSubmitted        bool
	BmoniWalletID       string
	BmoniWalletAddr     string
	RailActive          bool
	DepositAccountNumber string
	DepositBank         string
}

// UserIdentity is the real identity a host signs up with; it becomes their
// BMONI user. Names come from the user, never from a shared persona config.
type UserIdentity struct {
	FirstName string
	LastName  string
	Email     string
	Phone     string // E.164
}

// KYCProfile is the host-submitted KYC data for the NGN rail. In the sandbox
// these values must match the test persona to resolve; in production they are
// the host's own legal identity.
type KYCProfile struct {
	FirstName   string
	LastName    string
	Phone       string // E.164, carried from the instructor
	DateOfBirth string // YYYY-MM-DD
	Gender      string // male | female
	BVN         string // 11 digits
	Street      string
	City        string
	State       string // valid Nigerian state name
	PostalCode  string // 6 digits
}

// WalletTransaction is one wallet ledger row.
type WalletTransaction struct {
	ID        string       `json:"id"`
	WalletID  string       `json:"walletId"`
	Kind      WalletTxKind `json:"kind"`
	Amount    Amount       `json:"amountNaira"` // +credit / -debit
	Note      string       `json:"note"`
	CreatedAt time.Time    `json:"createdAt"`
}
