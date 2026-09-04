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
	"os"
	"path/filepath"

	"github.com/osugodbless/kweeks/internal/domain"
)

// Onboarding holds the external BMONI ids created for a kweeks instructor.
type Onboarding struct {
	UserID          string
	SmartWalletID   string
	SmartWalletAddr string
	OwnerAddress    string
}

// CreateUser registers a user with BMONI and returns the bmoniUserId.
// A 409 (already exists) is treated as success from a prior attempt per the
// retries-and-duplicates guidance: the caller recovers the existing user via
// the same phone.
func (c *Client) CreateUser(ctx context.Context, p domain.BmoniPersona) (string, error) {
	var resp struct {
		BmoniUserID string `json:"bmoniUserId"`
		Data        struct {
			UserID string `json:"bmoniUserId"`
		} `json:"data"`
	}
	err := c.do(ctx, http.MethodPost, "/v1/users", map[string]any{
		"firstName": p.FirstName, "lastName": p.LastName,
		"email": p.Email, "phoneNumber": p.Phone,
	}, &resp)
	if err != nil {
		return "", err
	}
	if resp.BmoniUserID != "" {
		return resp.BmoniUserID, nil
	}
	if resp.Data.UserID != "" {
		return resp.Data.UserID, nil
	}
	return "", errors.New("bmoni: create-user returned no user id")
}

// SubmitKYC posts the persona KYC profile ahead of rail activation.
func (c *Client) SubmitKYC(ctx context.Context, userID string, p domain.BmoniPersona) error {
	body := map[string]any{
		"personalInfo": map[string]any{
			"firstName": p.FirstName, "lastName": p.LastName,
			"dateOfBirth": p.DOB, "gender": "male",
		},
		"addressDetails": map[string]any{
			"street": p.Address, "city": p.City, "state": p.State,
			"countryCode": "NGA",
		},
	}
	return c.do(ctx, http.MethodPatch, "/v1/users/"+userID+"/kyc", body, nil)
}

// ownerAddress derives the EVM address for c.ownerKey.
func (c *Client) ownerAddress() (string, error) {
	addr, err := pubkeyToAddress(c.ownerKey)
	if err != nil {
		return "", err
	}
	return addr, nil
}

// ProvisionWallet runs the owner-proof handshake + create-managed for a CNGN
// smart wallet owned by c.ownerKey. Returns the smart wallet id + address.
func (c *Client) ProvisionWallet(ctx context.Context, userID string) (walletID, addr string, err error) {
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

	// 3. create-managed CNGN wallet.
	var wallet struct {
		SmartWalletID string `json:"smartWalletId"`
		Address       string `json:"address"`
		Data          struct {
			ID      string `json:"smartWalletId"`
			Address string `json:"address"`
			Wallet  struct {
				ID      string `json:"smartWalletId"`
				Address string `json:"address"`
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
	walletID = firstNonEmpty(wallet.SmartWalletID, wallet.Data.ID, wallet.Data.Wallet.ID)
	addr = firstNonEmpty(wallet.Address, wallet.Data.Address, wallet.Data.Wallet.Address)
	if walletID == "" {
		return "", "", errors.New("bmoni: create-managed returned no wallet id")
	}
	return walletID, addr, nil
}

// StartNigeria activates the NGN rail against the provisioned wallet. BVN is
// verified during the flow and auto-populates KYC.
func (c *Client) StartNigeria(ctx context.Context, userID, walletAddr, bvn string) error {
	if err := c.do(ctx, http.MethodPost,
		"/v1/users/"+userID+"/onboarding/start-nigeria",
		map[string]any{
			"bvn": bvn, "ngnWalletAddress": walletAddr, "ngnWalletIndex": 0,
		}, nil); err != nil {
		return err
	}
	return nil
}

// OnboardingStatus returns whether the NGN rail is active for the user.
func (c *Client) OnboardingStatus(ctx context.Context, userID string) (string, error) {
	var resp struct {
		Status string `json:"status"`
		Data   struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := c.do(ctx, http.MethodGet, "/v1/users/"+userID+"/onboarding/status", nil, &resp); err != nil {
		return "", err
	}
	status := firstNonEmpty(resp.Status, resp.Data.Status)
	if status == "" {
		// Some sandbox responses nest per-currency activity under "currencies".
		return "pending", nil
	}
	return status, nil
}

// UploadDocument submits one KYC document image (multipart) for the user.
// Uploads are the operator step of NGN onboarding; skip by not configuring
// file paths.
func (c *Client) UploadDocument(ctx context.Context, userID, kind, filePath string) error {
	if filePath == "" {
		return nil
	}
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("bmoni: open %s doc %s: %w", kind, filePath, err)
	}
	defer f.Close()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, f); err != nil {
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
