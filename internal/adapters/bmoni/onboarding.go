package bmoni

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/osugodbless/kweeks/internal/domain"
)

// CreateUser registers a BMONI user from the host's real identity and returns
// the bmoniUserId. A 409 (already exists) is handled per the
// retries-and-duplicates guidance: the existing user is recovered by
// phone/email rather than retried, so re-provisioning an instructor reuses
// their own user (never someone else's).
func (c *Client) CreateUser(ctx context.Context, id domain.UserIdentity) (string, error) {
	var resp struct {
		BmoniUserID string `json:"bmoniUserId"`
		User        struct {
			BmoniUserID string `json:"bmoniUserId"`
		} `json:"user"`
		Data struct {
			BmoniUserID string `json:"bmoniUserId"`
		} `json:"data"`
	}
	err := c.do(ctx, http.MethodPost, "/v1/users", map[string]any{
		"firstName": id.FirstName, "lastName": id.LastName,
		"email": id.Email, "phoneNumber": id.Phone,
	}, &resp)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Status == 409 {
			recovered, rerr := c.findExistingUser(ctx, id.Phone, id.Email)
			if rerr == nil {
				return recovered, nil
			}
		}
		return "", err
	}
	if resp.BmoniUserID != "" {
		return resp.BmoniUserID, nil
	}
	if resp.User.BmoniUserID != "" {
		return resp.User.BmoniUserID, nil
	}
	if resp.Data.BmoniUserID != "" {
		return resp.Data.BmoniUserID, nil
	}
	return "", errors.New("bmoni: create-user returned no user id")
}

// findExistingUser recovers the existing BMONI user for a 409 on create-user
// by searching the partner's users for the phone or email.
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
	return "", errors.New("bmoni: create-user returned 409 but existing user could not be recovered")
}

// SubmitKYC writes the host's KYC profile ahead of rail activation. Field
// names follow the KYC — Nigeria requirements page (personalInfo + address
// with streetLine1/city/state/postalCode/countryCode + identificationNumbers
// bvn). BVN verification during start-nigeria auto-populates the profile; the
// submitted name must match the persona.
func (c *Client) SubmitKYC(ctx context.Context, userID string, k domain.KYCProfile) error {
	if len(k.BVN) != 11 {
		return errors.New("bmoni: bvn must be exactly 11 digits")
	}
	body := map[string]any{
		"personalInfo": map[string]any{
			"firstName": k.FirstName, "lastName": k.LastName,
			"dateOfBirth": k.DateOfBirth, "gender": k.Gender,
		},
		"address": map[string]any{
			"streetLine1": k.Street, "city": k.City, "state": k.State,
			"postalCode": k.PostalCode, "countryCode": "NGA",
		},
		"identificationNumbers": []map[string]any{
			{"type": "bvn", "number": k.BVN, "issuingCountryCode": "NGA"},
		},
		"sourceOfFunds": "salary", "accountPurpose": "personal", "actingAsIntermediary": false,
	}
	if k.Phone != "" {
		body["personalInfo"].(map[string]any)["phoneNumber"] = k.Phone
	}
	return c.do(ctx, http.MethodPatch, "/v1/users/"+userID+"/kyc", body, nil)
}

// UploadKycDocument submits one KYC document image (multipart) for the user.
func (c *Client) UploadKycDocument(ctx context.Context, userID, kind string, data []byte, filename string) error {
	if len(data) == 0 {
		return errors.New("bmoni: empty document upload")
	}
	if filename == "" {
		filename = kind + ".jpg"
	}
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	if _, err := part.Write(data); err != nil {
		return err
	}
	_ = mw.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/v1/users/"+userID+"/kyc/documents/"+kind, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("bmoni upload %s: %w", kind, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("bmoni upload %s: status %d: %s", kind, resp.StatusCode, truncate(string(raw), 300))
	}
	return nil
}

// ownerAddress derives the EVM address for c.ownerKey.
func (c *Client) ownerAddress() (string, error) {
	addr, err := pubkeyToAddress(c.ownerKey)
	if err != nil {
		return "", err
	}
	return addr, nil
}

// CreateWallet provisions a CNGN smart wallet owned by c.ownerKey, following
// the documented handshake: owner-proof challenge → EIP-191 sign → create-managed.
// Returns the smart wallet id + address.
func (c *Client) CreateWallet(ctx context.Context, userID string) (walletID, addr string, err error) {
	if c.ownerKey == "" {
		return "", "", errors.New("bmoni: owner key required to provision a wallet")
	}
	owner, err := c.ownerAddress()
	if err != nil {
		return "", "", err
	}

	// 1. Owner-proof challenge.
	var ch struct {
		ChallengeID string `json:"challengeId"`
		Message     string `json:"message"`
		Data        struct {
			ChallengeID string `json:"challengeId"`
			Message     string `json:"message"`
		} `json:"data"`
	}
	if err := c.do(ctx, http.MethodPost,
		"/v1/users/"+userID+"/smart-wallets/owner-proof-challenges",
		map[string]string{"currency": "CNGN", "userOwnerAddress": owner}, &ch); err != nil {
		return "", "", err
	}
	challengeID := firstNonEmpty(ch.ChallengeID, ch.Data.ChallengeID)
	msg := firstNonEmpty(ch.Message, ch.Data.Message)
	if challengeID == "" || msg == "" {
		return "", "", errors.New("bmoni: owner-proof challenge missing id or message")
	}

	// 2. Sign the challenge text with the EIP-191 prefix (personal_sign).
	sig, err := signMessage(c.ownerKey, msg)
	if err != nil {
		return "", "", err
	}

	// 3. create-managed CNGN wallet. Live response is a flat wallet object;
	// tolerate the documented nested shapes too.
	var wallet struct {
		ID            string `json:"id"`
		SmartWalletID string `json:"smartWalletId"`
		WalletAddress string `json:"walletAddress"`
		Address       string `json:"address"`
		Currency      string `json:"currency"`
		IsActive      bool   `json:"isActive"`
		Data          struct {
			ID            string `json:"id"`
			SmartWalletID string `json:"smartWalletId"`
			WalletAddress string `json:"walletAddress"`
			Address       string `json:"address"`
			Wallet        struct {
				ID      string `json:"id"`
				Address string `json:"walletAddress"`
			} `json:"wallet"`
		} `json:"data"`
	}
	if err := c.do(ctx, http.MethodPost,
		"/v1/users/"+userID+"/smart-wallets/create-managed",
		map[string]string{
			"currency": "CNGN", "userOwnerAddress": owner,
			"ownerProofChallengeId": challengeID, "ownerProofSignature": sig,
		}, &wallet); err != nil {
		return "", "", err
	}
	walletID = firstNonEmpty(wallet.ID, wallet.SmartWalletID, wallet.Data.ID, wallet.Data.SmartWalletID, wallet.Data.Wallet.ID)
	addr = firstNonEmpty(wallet.WalletAddress, wallet.Address, wallet.Data.WalletAddress, wallet.Data.Address, wallet.Data.Wallet.Address)
	if walletID == "" {
		return "", "", errors.New("bmoni: create-managed returned no wallet id")
	}
	return walletID, addr, nil
}

// ActivateRail runs POST /onboarding/start-nigeria: activates the NGN rail
// against the wallet address using the host's BVN. BVN verification happens in
// this call and auto-populates the KYC profile.
func (c *Client) ActivateRail(ctx context.Context, userID, walletAddr, bvn string) error {
	if len(bvn) != 11 {
		return errors.New("bmoni: bvn must be exactly 11 digits")
	}
	if err := c.do(ctx, http.MethodPost,
		"/v1/users/"+userID+"/onboarding/start-nigeria",
		map[string]any{
			"bvn": bvn, "ngnWalletAddress": walletAddr, "ngnWalletIndex": 0,
		}, nil); err != nil {
		return err
	}
	return nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// parseErrorShape tolerates both object and {data:{...}} envelope shapes and
// any nested error objects the proxy returns.
func parseErrorShape(raw []byte) string {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return string(raw)
	}
	if msg, ok := m["message"].(string); ok {
		return msg
	}
	if errs, ok := m["message"].([]any); ok && len(errs) > 0 {
		return fmt.Sprintf("%v", errs)
	}
	return truncate(string(raw), 300)
}
