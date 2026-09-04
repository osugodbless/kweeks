package bmoni

import (
	"context"
	"errors"

	"github.com/osugodbless/kweeks/internal/domain"
)

// Provision implements ports.Money.Provision: create the BMONI user, submit the
// NGN KYC profile, provision a CNGN smart wallet, activate the NGN rail, and
// upload the operator-provided KYC documents (when configured). Returns the
// external ids for the wallet.
func (c *Client) Provision(ctx context.Context, p domain.BmoniPersona) (*domain.WalletExternal, error) {
	if c.apiKey == "" || c.ownerKey == "" {
		return nil, errors.New("bmoni: api key and owner key required to provision")
	}
	userID, err := c.CreateUser(ctx, p)
	if err != nil {
		return nil, err
	}

	// KYC profile must carry the persona name before wallet/rail activation.
	if err := c.SubmitKYC(ctx, userID, p); err != nil {
		return nil, err
	}

	walletID, addr, err := c.ProvisionWallet(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Operator step: the three document image uploads complete verification.
	if err := c.uploadOperatorDocs(ctx, userID); err != nil {
		return nil, err
	}

	if err := c.StartNigeria(ctx, userID, addr, p.BVN); err != nil {
		return nil, err
	}

	return &domain.WalletExternal{UserID: userID, WalletID: walletID, Address: addr}, nil
}

func (c *Client) uploadOperatorDocs(ctx context.Context, userID string) error {
	for _, u := range []struct{ kind, path string }{
		{"identification", c.docIdentification},
		{"proof-of-address", c.docProofOfAddress},
		{"biometric", c.docBiometric},
	} {
		if u.path == "" {
			continue
		}
		if err := c.UploadDocument(ctx, userID, u.kind, u.path); err != nil {
			return err
		}
	}
	return nil
}

// CreditNGN implements ports.Money.CreditNGN: fund the provisioned wallet from
// the platform funding wallet.
func (c *Client) CreditNGN(ctx context.Context, external *domain.WalletExternal, amount domain.Amount) (string, error) {
	if external == nil {
		return "", errors.New("bmoni: wallet not provisioned")
	}
	return c.CreditTo(ctx, external.UserID, external.WalletID, amount.NairaString())
}

// PayWinnerFrom implements ports.Money.PayWinnerFrom: prize money from an
// instructor wallet to a winner wallet address.
func (c *Client) PayWinnerFrom(ctx context.Context, from *domain.WalletExternal, toAddr string, amount domain.Amount) (string, error) {
	if from == nil {
		return "", errors.New("bmoni: source wallet not provisioned")
	}
	return c.PayTo(ctx, from.UserID, from.WalletID, toAddr, amount.NairaString())
}
