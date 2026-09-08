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

// Wallet is the instructor wallet service: BMONI user creation, KYC submission,
// smart-wallet provisioning, NGN rail activation, funding, balance, and ledger.
// Every step uses user-supplied identity data (never a shared persona config).
type Wallet struct {
	store ports.Store
	clock ports.Clock

	// money, when set, is the BMONI rail used to provision wallets and settle
	// credit/payout moves. A nil money means the rail is disabled and funding
	// stays on the local ledger.
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

// RailConfigured reports whether a real money rail is available.
func (w *Wallet) RailConfigured() bool {
	return w.money != nil
}

// identity derives the BMONI user identity from the instructor's real signup
// data (program-generated from user input, never from env persona). The
// explicit first/last name parts the host typed are used verbatim; only
// pre-existing accounts created before name parts were stored fall back to a
// best-effort split of the display name.
func identityFor(instructor domain.Instructor) domain.UserIdentity {
	first, last := instructor.FirstName, instructor.LastName
	if first == "" && last == "" {
		first, last = splitName(instructor.Name)
	}
	return domain.UserIdentity{
		FirstName: first, LastName: last,
		Email: instructor.Email, Phone: instructor.Phone,
	}
}

func splitName(name string) (first, last string) {
	parts := strings.Fields(strings.TrimSpace(name))
	switch len(parts) {
	case 0:
		return "", ""
	case 1:
		return parts[0], ""
	default:
		return parts[0], strings.Join(parts[1:], " ")
	}
}

// CreateBmoniUser registers the instructor's BMONI user (step 1 of the strict
// flow) and records the bmoniUserId on their wallet. Idempotent.
func (w *Wallet) CreateBmoniUser(ctx context.Context, instructorID string) (*domain.Wallet, error) {
	if w.money == nil {
		return nil, errors.New("money rail not configured")
	}
	wallet, err := w.store.GetWalletByInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	if wallet.BmoniUserID != "" {
		return wallet, nil
	}
	instructor, err := w.store.GetInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	userID, err := w.money.CreateUser(ctx, identityFor(*instructor))
	if err != nil {
		return nil, err
	}
	if err := w.store.SetWalletBmoniUser(ctx, wallet.ID, userID); err != nil {
		return nil, err
	}
	return w.store.GetWalletByInstructor(ctx, instructorID)
}

// SubmitKYC writes the host's KYC profile (lifecycle stage 3). The host inputs
// their own identity; in the sandbox, persona values resolve.
func (w *Wallet) SubmitKYC(ctx context.Context, instructorID string, k domain.KYCProfile) (*domain.Wallet, error) {
	if w.money == nil {
		return nil, errors.New("money rail not configured")
	}
	wallet, err := w.store.GetWalletByInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	if wallet.BmoniUserID == "" {
		return nil, errors.New("create your BMONI user before submitting KYC")
	}
	if k.Phone == "" {
		if instructor, err := w.store.GetInstructor(ctx, instructorID); err == nil {
			k.Phone = instructor.Phone
		}
	}
	if err := w.money.SubmitKYC(ctx, wallet.BmoniUserID, k); err != nil {
		return nil, err
	}
	if err := w.store.SetWalletKYC(ctx, wallet.ID, true); err != nil {
		return nil, err
	}
	return w.store.GetWalletByInstructor(ctx, instructorID)
}

// UploadKYC forwards one KYC document image to the rail. Per the lifecycle
// these are uploaded before PATCH /kyc and are required for the NGN profile.
func (w *Wallet) UploadKYC(ctx context.Context, instructorID string, doc domain.KycDocument) error {
	if w.money == nil {
		return errors.New("money rail not configured")
	}
	wallet, err := w.store.GetWalletByInstructor(ctx, instructorID)
	if err != nil {
		return err
	}
	if wallet.BmoniUserID == "" {
		return errors.New("create your BMONI user before uploading KYC documents")
	}
	return w.money.UploadKycDocument(ctx, wallet.BmoniUserID, doc)
}

// CreateWallet provisions the CNGN smart wallet (lifecycle stage 2:
// owner-proof challenge → create-managed) and records the ids.
func (w *Wallet) CreateWallet(ctx context.Context, instructorID string) (*domain.Wallet, error) {
	if w.money == nil {
		return nil, errors.New("money rail not configured")
	}
	wallet, err := w.store.GetWalletByInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	if wallet.BmoniWalletID != "" {
		return wallet, nil
	}
	if wallet.BmoniUserID == "" {
		return nil, errors.New("create your BMONI user before provisioning a wallet")
	}
	walletID, addr, err := w.money.CreateWallet(ctx, wallet.BmoniUserID)
	if err != nil {
		return nil, err
	}
	if err := w.store.SetWalletBmoniWallet(ctx, wallet.ID, walletID, addr); err != nil {
		return nil, err
	}
	return w.store.GetWalletByInstructor(ctx, instructorID)
}

// ActivateRail runs the NGN rail onboarding (lifecycle stage 4:
// start-nigeria with the host's BVN) and marks the wallet ready to fund.
func (w *Wallet) ActivateRail(ctx context.Context, instructorID, bvn string) (*domain.Wallet, error) {
	if w.money == nil {
		return nil, errors.New("money rail not configured")
	}
	wallet, err := w.store.GetWalletByInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	if wallet.BmoniRailActive {
		return wallet, nil
	}
	if wallet.BmoniUserID == "" || wallet.BmoniWalletAddr == "" {
		return nil, errors.New("create your BMONI user + wallet before activating the rail")
	}
	if err := w.money.ActivateRail(ctx, wallet.BmoniUserID, wallet.BmoniWalletAddr, bvn); err != nil {
		return nil, err
	}
	if err := w.store.SetWalletRailActive(ctx, wallet.ID, true); err != nil {
		return nil, err
	}
	return w.store.GetWalletByInstructor(ctx, instructorID)
}

// SetupStatus reports where the wallet is in the strict provisioning flow and
// the next action needed. When ready, it also returns the NGN deposit account
// the host funds by bank transfer.
func (w *Wallet) SetupStatus(ctx context.Context, instructorID string) (*domain.SetupStatus, error) {
	wallet, err := w.store.GetWalletByInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	st := &domain.SetupStatus{
		BmoniUserID:     wallet.BmoniUserID,
		KYCSubmitted:    wallet.BmoniKYCSubmitted,
		BmoniWalletID:   wallet.BmoniWalletID,
		BmoniWalletAddr: wallet.BmoniWalletAddr,
		RailActive:      wallet.BmoniRailActive,
	}
	if !w.RailConfigured() {
		st.Stage = domain.SetupUnprovisioned
		return st, nil
	}
	switch {
	case wallet.BmoniUserID == "":
		st.Stage = domain.SetupUnprovisioned
	case wallet.BmoniWalletID == "":
		// Lifecycle stage 2: the smart wallet precedes KYC.
		st.Stage = domain.SetupWallet
	case !wallet.BmoniKYCSubmitted:
		// Lifecycle stage 3: verify identity (KYC) after the wallet exists.
		st.Stage = domain.SetupKYC
	case !wallet.BmoniRailActive:
		// Lifecycle stage 4: activate the rail.
		st.Stage = domain.SetupRail
	default:
		st.Stage = domain.SetupReady
		if number, bank, err := w.money.DepositAccount(ctx, wallet.BmoniUserID, wallet.BmoniWalletID); err == nil {
			st.DepositAccountNumber = number
			st.DepositBank = bank
		}
	}
	return st, nil
}

// LookupBVN resolves the host's BVN to its holder record so the KYC form can
// pre-fill and the host confirm before submitting (lifecycle stage 3 helper).
func (w *Wallet) LookupBVN(ctx context.Context, instructorID, bvn string) (*domain.BVNRecord, error) {
	if w.money == nil {
		return nil, errors.New("money rail not configured")
	}
	wallet, err := w.store.GetWalletByInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	if wallet.BmoniUserID == "" {
		return nil, errors.New("create your BMONI user before looking up a BVN")
	}
	return w.money.LookupBVN(ctx, wallet.BmoniUserID, bvn)
}

// DepositAccount returns the host's NGN virtual bank account (account number
// + bank) that a bank transfer to it funds the wallet.
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

// Fund credits an instructor's wallet. The `credit` method is the instant
// local-ledger credit (the sandbox/demo funding path). Real production funding
// happens by sending money to the host's NGN virtual bank account (see
// DepositAccount), which BMONI credits to the wallet.
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

// uniquePhoneFor derives a deterministic E.164 phone from an instructor id so
// repeated provisioning of the same instructor reuses the same BMONI user while
// different instructors never collide on a shared persona phone.
func uniquePhoneFor(instructorID string) string {
	h := sha256.Sum256([]byte("kweeks-phone:" + instructorID))
	n := binary.BigEndian.Uint64(h[:8]) % 1000000000
	return fmt.Sprintf("+2347%09d", n)
}
