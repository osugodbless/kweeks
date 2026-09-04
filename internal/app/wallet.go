package app

import (
	"context"
	"errors"
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
	ext, err := w.money.Provision(ctx, w.persona)
	if err != nil {
		return nil, err
	}
	if err := w.store.SetWalletBmoni(ctx, wallet.ID, ext); err != nil {
		return nil, err
	}
	return w.store.GetWalletByInstructor(ctx, instructorID)
}

// Fund credits an instructor's wallet. When the rail is configured and the
// wallet is provisioned, the credit settles on the real BMONI rail first and
// the local ledger records it. Without the rail, "credit" is an instant local
// ledger credit (sandbox/dev) and card/transfer return a clear error.
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
		if w.money == nil {
			return nil, errors.New("external funding unavailable: money rail not configured. Try instant wallet credit.")
		}
		if wallet.BmoniUserID == "" {
			return nil, errors.New("wallet is not provisioned on the money rail")
		}
		if _, err := w.money.CreditNGN(ctx, &domain.WalletExternal{
			UserID: wallet.BmoniUserID, WalletID: wallet.BmoniWalletID, Address: wallet.BmoniWalletAddr,
		}, amount); err != nil {
			return nil, err
		}
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
