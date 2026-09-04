package domain

import "time"

// Instructor is an account that hosts quizzes. Signup provisions one and a
// NGN wallet is issued for it immediately.
type Instructor struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
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

	// External BMONI identity, set when the wallet is provisioned on the rail.
	// Empty until provisioning succeeds.
	BmoniUserID     string `json:"bmoniUserId,omitempty"`
	BmoniWalletID   string `json:"bmoniWalletId,omitempty"`
	BmoniWalletAddr string `json:"bmoniWalletAddress,omitempty"`
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

// BmoniPersona is the identity used to provision a BMONI user + NGN rail.
// Sandbox persona values come from config (test personas match exactly).
type BmoniPersona struct {
	FirstName string
	LastName  string
	Email     string
	Phone     string
	BVN       string
	DOB       string
	Address   string
	City      string
	State     string
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
