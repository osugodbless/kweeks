package app

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/osugodbless/kweeks/internal/domain"
	"github.com/osugodbless/kweeks/internal/ports"
)

// Wallet is the instructor wallet service: funding, balance, ledger.
type Wallet struct {
	store ports.Store
	clock ports.Clock

	// money, when set, is the BMONI rail used to settle card/transfer funding.
	// credit funding never needs it (instant platform credit / sandbox credit).
	money ports.Money
}

func NewWallet(store ports.Store, clock ports.Clock, money ports.Money) *Wallet {
	return &Wallet{store: store, clock: clock, money: money}
}

func (w *Wallet) nowTime() time.Time {
	if w.clock != nil {
		return w.clock.Now()
	}
	return time.Now()
}

// Fund credits an instructor's wallet. method "credit" is the instant platform
// credit (BMONI sandbox test-credit); card/transfer settle through the money
// rail when configured.
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
			return nil, errors.New("external funding unavailable: money rail not configured")
		}
		// A real settle would invoke the BMONI deposit rail here. For the demo
		// contract we gate on the rail being configured and surface a clear
		// error otherwise; credit is the always-available path.
		if _, err := w.money.BalanceNGN(ctx); err != nil {
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
