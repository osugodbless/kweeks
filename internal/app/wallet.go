package app

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/osugodbless/kweeks/internal/domain"
	"github.com/osugodbless/kweeks/internal/ports"
)

// Wallet is the instructor wallet service: provisioning on the BMONI rail,
// funding, balance, and ledger.
type Wallet struct {
	store ports.Store
	clock ports.Clock

	// money, when set, is the BMONI rail used to provision wallets and settle
	// credit/payout moves. A nil money means the rail is disabled and funding
	// stays on the local ledger.
	money ports.Money

	// persona is the identity used to provision a BMONI user + NGN rail.
	persona domain.BmoniPersona
	// provisionOnSignup provisions a real BMONI wallet when an instructor
	// signs up and the rail is configured.
	provisionOnSignup bool
}

func NewWallet(store ports.Store, clock ports.Clock, money ports.Money) *Wallet {
	return &Wallet{store: store, clock: clock, money: money}
}

// WithProvisioning configures the rail persona + auto-provisioning on signup.
func (w *Wallet) WithProvisioning(persona domain.BmoniPersona, provisionOnSignup bool) *Wallet {
	w.persona = persona
	w.provisionOnSignup = provisionOnSignup
	return w
}

func (w *Wallet) nowTime() time.Time {
	if w.clock != nil {
		return w.clock.Now()
	}
	return time.Now()
}

// PersonaConfigured reports whether a full persona is available to provision.
func (w *Wallet) PersonaConfigured() bool {
	p := w.persona
	return p.FirstName != "" && p.LastName != "" && p.Phone != "" && p.BVN != ""
}

// Provision creates a real BMONI user + CNGN wallet for the instructor and
// records the external ids on their wallet row. Idempotent: a wallet that is
// already provisioned is left untouched.
//
// Every instructor provisions a DISTINCT BMONI user: the BMONI user is created
// with the instructor's own email + phone (falling back to a deterministic
// unique phone derived from their id) and the persona's name/BVN/DOB/address so
// sandbox KYC verification matches. The shared persona phone must never be used
// for create-user, or every signup on a shared key resolves to the first
// user's BMONI identity and all wallets share one money identity.
func (w *Wallet) Provision(ctx context.Context, instructorID string) (*domain.Wallet, error) {
	if w.money == nil {
		return nil, errors.New("wallet provisioning unavailable: money rail not configured")
	}
	if !w.PersonaConfigured() {
		return nil, errors.New("wallet provisioning unavailable: BMONI persona not configured")
	}
	wallet, err := w.store.GetWalletByInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	if wallet.BmoniUserID != "" {
		return wallet, nil // already provisioned
	}
	persona := w.persona
	if instructor, err := w.store.GetInstructor(ctx, instructorID); err == nil {
		if instructor.Email != "" {
			persona.Email = instructor.Email
		}
		if instructor.Phone != "" {
			persona.Phone = instructor.Phone
		} else {
			persona.Phone = uniquePhoneFor(instructor.ID)
		}
	}
	ext, err := w.money.Provision(ctx, persona)
	if err != nil {
		return nil, err
	}
	if err := w.store.SetWalletBmoni(ctx, wallet.ID, ext); err != nil {
		return nil, err
	}
	return w.store.GetWalletByInstructor(ctx, instructorID)
}

// uniquePhoneFor derives a deterministic E.164 phone from an instructor id so
// repeated provisioning of the same instructor reuses the same BMONI user while
// different instructors never collide on the shared persona phone.
func uniquePhoneFor(instructorID string) string {
	h := sha256.Sum256([]byte("kweeks-phone:" + instructorID))
	n := binary.BigEndian.Uint64(h[:8]) % 1000000000
	return fmt.Sprintf("+2347%09d", n)
}

// Fund credits an instructor's wallet. The `credit` method is the instant
// local-ledger credit (the sandbox/demo funding path). `card`/`transfer` are
// external rails: real production funding happens by sending money to the
// host's NGN virtual bank account (see DepositAccount), which BMONI credits to
// the wallet; the local ledger mirrors it. Without a live onramp webhook the
// ledger is only updated by `credit`, so card/transfer return a clear error
// directing the host to the deposit account instead of silently minting naira.
func (w *Wallet) Fund(ctx context.Context, instructorID string, amount domain.Amount, method string) (*domain.Wallet, error) {
	if amount <= 0 {
		return nil, domain.ErrBadCredentials // reuse: amount must be positive
	}
	wallet, err := w.store.GetWalletByInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	m := strings.ToLower(strings.TrimSpace(method))
	if m == "" {
		m = "credit"
	}

	if m != "credit" {
		return nil, errors.New("external funding settles via your NGN bank account — use 'wallet credit' for instant demo funding")
	}

	tx := &domain.WalletTransaction{
		ID:        newID(),
		WalletID:  wallet.ID,
		Kind:      domain.TxFund,
		Amount:    amount,
		Note:      fundingNote(m),
		CreatedAt: w.nowTime(),
	}
	if err := w.store.ApplyWalletTx(ctx, wallet.ID, tx); err != nil {
		return nil, err
	}
	return w.store.GetWalletByInstructor(ctx, instructorID)
}

// DepositAccount returns the host's NGN virtual bank account (account number
// + bank) that a bank transfer can be sent to in order to fund the wallet. It
// requires the wallet to be provisioned on the rail.
func (w *Wallet) DepositAccount(ctx context.Context, instructorID string) (accountNumber, bankName string, err error) {
	if w.money == nil {
		return "", "", errors.New("money rail not configured")
	}
	ext, err := w.External(ctx, instructorID)
	if err != nil {
		return "", "", err
	}
	if ext == nil {
		return "", "", errors.New("wallet is not provisioned on the money rail")
	}
	return w.money.DepositAccount(ctx, ext.UserID, ext.WalletID)
}

// Balance returns the instructor's wallet.
func (w *Wallet) Balance(ctx context.Context, instructorID string) (*domain.Wallet, error) {
	return w.store.GetWalletByInstructor(ctx, instructorID)
}

// Transactions returns the wallet ledger, newest first.
func (w *Wallet) Transactions(ctx context.Context, instructorID string) ([]domain.WalletTransaction, error) {
	wallet, err := w.store.GetWalletByInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	return w.store.ListWalletTransactions(ctx, wallet.ID)
}

// External returns the wallet's external (provisioned) identity, or nil when
// not provisioned.
func (w *Wallet) External(ctx context.Context, instructorID string) (*domain.WalletExternal, error) {
	wallet, err := w.store.GetWalletByInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	if wallet.BmoniUserID == "" {
		return nil, nil
	}
	return &domain.WalletExternal{
		UserID: wallet.BmoniUserID, WalletID: wallet.BmoniWalletID, Address: wallet.BmoniWalletAddr,
	}, nil
}

func fundingNote(method string) string {
	switch method {
	case "card":
		return "Funded wallet · card"
	case "transfer":
		return "Funded wallet · transfer"
	default:
		return "Funded wallet · instant credit"
	}
}
