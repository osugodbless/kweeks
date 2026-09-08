package bmoni

import (
	"context"
	"errors"
	"net/http"

	"github.com/osugodbless/kweeks/internal/domain"
)

// Provision implements ports.Money.Provision: create (or recover) the BMONI
// user, submit the NGN KYC profile, provision a CNGN smart wallet, and
// activate the NGN rail. Returns the external ids for the wallet.
func (c *Client) Provision(ctx context.Context, p domain.BmoniPersona) (*domain.WalletExternal, error) {
	if c.apiKey == "" || c.ownerKey == "" {
		return nil, errors.New("bmoni: api key and owner key required to provision")
	}

	userID, err := c.CreateUser(ctx, p)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Status == 409 {
			// Documented recovery: a 409 on create-user means the user already
			// exists under this partner key (email or phone collided with a
			// prior successful create). Recover the existing user rather than
			// retrying.
			userID, err = c.findExistingUser(ctx, p.Phone, p.Email)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// KYC profile must carry the persona name before wallet/rail activation.
	if err := c.SubmitKYC(ctx, userID, p); err != nil {
		return nil, err
	}

	walletID, addr, err := c.ProvisionWallet(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Operator step: the document image uploads complete verification. No-op
	// when no paths are configured (sandbox NGN rail resolves without them).
	if err := c.uploadOperatorDocs(ctx, userID); err != nil {
		return nil, err
	}

	if err := c.StartNigeria(ctx, userID, addr, p.BVN); err != nil {
		return nil, err
	}

	return &domain.WalletExternal{UserID: userID, WalletID: walletID, Address: addr}, nil
}

// findExistingUser recovers the existing BMONI user for a 409 on create-user
// by searching the partner's users for the persona phone or the supplied
// email.
func (c *Client) findExistingUser(ctx context.Context, phone, email string) (string, error) {
	var resp struct {
		Users []struct {
			BmoniUserID string `json:"bmoniUserId"`
			PhoneNumber string `json:"phoneNumber"`
			Email       string `json:"email"`
		} `json:"users"`
		Data struct {
			Users []struct {
				BmoniUserID string `json:"bmoniUserId"`
				PhoneNumber string `json:"phoneNumber"`
				Email       string `json:"email"`
			} `json:"users"`
		} `json:"data"`
	}
	if err := c.do(ctx, http.MethodGet, "/v1/users?limit=100", nil, &resp); err != nil {
		return "", err
	}
	src := resp.Users
	if len(src) == 0 {
		src = resp.Data.Users
	}
	for _, u := range src {
		if u.BmoniUserID == "" {
			continue
		}
		if u.PhoneNumber == phone || (email != "" && u.Email == email) {
			return u.BmoniUserID, nil
		}
	}
	return "", errors.New("bmoni: create-user returned 409 but existing user could not be recovered (phone/email not in list)")
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
