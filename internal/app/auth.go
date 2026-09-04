// Package app contains the application services that orchestrate domain rules
// against ports. Auth + wallet provisioning live here, alongside the game.
package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/osugodbless/kweeks/internal/domain"
	"github.com/osugodbless/kweeks/internal/ports"
)

// Auth provisions instructors (multi-user), hashes passwords with bcrypt, and
// issues bearer sessions.
type Auth struct {
	store ports.Store
	clock ports.Clock

	// SessionTTL controls how long a token stays valid.
	SessionTTL time.Duration

	// onProvision, when set, is invoked after an instructor's wallet is issued
	// so the app can provision a real BMONI wallet on the rail. Errors are
	// logged by the caller, never fatal to signup (ledger wallet still works).
	onProvision func(ctx context.Context, instructorID string) error
}

func NewAuth(store ports.Store, clock ports.Clock) *Auth {
	a := &Auth{store: store, clock: clock, SessionTTL: 30 * 24 * time.Hour}
	if clock == nil {
		a.clock = nil
	}
	return a
}

// WithWalletProvisioning installs the post-signup BMONI wallet provisioner.
func (a *Auth) WithWalletProvisioning(provision func(ctx context.Context, instructorID string) error) *Auth {
	a.onProvision = provision
	return a
}

func (a *Auth) nowTime() time.Time {
	if a.clock != nil {
		return a.clock.Now()
	}
	return time.Now()
}

// SignupResult carries what signup/login produce.
type SignupResult struct {
	Instructor *domain.Instructor
	Token      string
	Wallet     *domain.Wallet

	// ProvisionErr is non-nil when BMONI wallet provisioning was attempted but
	// failed; signup still succeeds with the local ledger wallet.
	ProvisionErr error
}

// Signup creates an instructor and immediately issues a NGN wallet.
func (a *Auth) Signup(ctx context.Context, name, email, password string) (*SignupResult, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	if name == "" || !strings.Contains(email, "@") || len(password) < 6 {
		return nil, domain.ErrBadCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	instructor := &domain.Instructor{
		ID:           newID(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Avatar:       initials(name),
		CreatedAt:    a.nowTime(),
	}
	if err := a.store.CreateInstructor(ctx, instructor); err != nil {
		return nil, err
	}
	wallet, err := a.issueWallet(ctx, instructor.ID)
	if err != nil {
		return nil, err
	}
	if a.onProvision != nil {
		if perr := a.onProvision(ctx, instructor.ID); perr != nil {
			// Provisioning failure must never block signup: the local ledger
			// wallet is the fallback and can be provisioned later.
			return &SignupResult{Instructor: instructor, Token: "", Wallet: wallet, ProvisionErr: perr}, nil
		}
	}
	token, err := a.createSession(ctx, instructor.ID)
	if err != nil {
		return nil, err
	}
	return &SignupResult{Instructor: instructor, Token: token, Wallet: wallet}, nil
}

// Login verifies credentials and issues a fresh session.
func (a *Auth) Login(ctx context.Context, email, password string) (*SignupResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	instructor, err := a.store.GetInstructorByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrInstructorNotFound) {
			return nil, domain.ErrBadCredentials
		}
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(instructor.PasswordHash), []byte(password)) != nil {
		return nil, domain.ErrBadCredentials
	}
	wallet, err := a.store.GetWalletByInstructor(ctx, instructor.ID)
	if err != nil && !errors.Is(err, domain.ErrWalletNotFound) {
		return nil, err
	}
	if wallet == nil {
		wallet, err = a.issueWallet(ctx, instructor.ID)
		if err != nil {
			return nil, err
		}
	}
	token, err := a.createSession(ctx, instructor.ID)
	if err != nil {
		return nil, err
	}
	return &SignupResult{Instructor: instructor, Token: token, Wallet: wallet}, nil
}

// Resolve maps a bearer token to its instructor + wallet. Returns
// ErrSessionNotFound / ErrUnauthorized for missing/expired tokens.
func (a *Auth) Resolve(ctx context.Context, token string) (*domain.Instructor, *domain.Wallet, error) {
	if token == "" {
		return nil, nil, domain.ErrUnauthorized
	}
	session, err := a.store.GetSession(ctx, token)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return nil, nil, domain.ErrUnauthorized
		}
		return nil, nil, err
	}
	if !session.ExpiresAt.After(a.nowTime()) {
		return nil, nil, domain.ErrUnauthorized
	}
	instructor, err := a.store.GetInstructor(ctx, session.InstructorID)
	if err != nil {
		return nil, nil, err
	}
	wallet, err := a.store.GetWalletByInstructor(ctx, instructor.ID)
	if err != nil {
		if errors.Is(err, domain.ErrWalletNotFound) {
			wallet, err = a.issueWallet(ctx, instructor.ID)
			if err != nil {
				return nil, nil, err
			}
		} else {
			return nil, nil, err
		}
	}
	return instructor, wallet, nil
}

func (a *Auth) issueWallet(ctx context.Context, instructorID string) (*domain.Wallet, error) {
	wallet := &domain.Wallet{
		ID:           fmt.Sprintf("kweeks_ngn_%s", randomHex(6)),
		InstructorID: instructorID,
		Balance:      0,
		CreatedAt:    a.nowTime(),
	}
	if err := a.store.CreateWallet(ctx, wallet); err != nil {
		if errors.Is(err, domain.ErrWalletExists) {
			return a.store.GetWalletByInstructor(ctx, instructorID)
		}
		return nil, err
	}
	return wallet, nil
}

func (a *Auth) createSession(ctx context.Context, instructorID string) (string, error) {
	session := &domain.Session{
		Token:        randomHex(24),
		InstructorID: instructorID,
		CreatedAt:    a.nowTime(),
		ExpiresAt:    a.nowTime().Add(a.SessionTTL),
	}
	if err := a.store.CreateSession(ctx, session); err != nil {
		return "", err
	}
	return session.Token, nil
}

// initials derives an avatar from the name, e.g. "Adeola Peters" -> "AP".
func initials(name string) string {
	parts := strings.Fields(name)
	switch len(parts) {
	case 0:
		return "AP"
	case 1:
		return strings.ToUpper(parts[0][:1])
	default:
		return strings.ToUpper(parts[0][:1] + parts[1][:1])
	}
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
