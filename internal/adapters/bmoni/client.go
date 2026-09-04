// Package bmoni implements ports.Money against the BMONI Embedded REST API
// (create-managed wallet, proposal -> approve -> sign -> send). The owner key
// is held server-side and never leaves the process.
package bmoni

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/osugodbless/kweeks/internal/domain"
)

// Client talks to BMONI Embedded.
type Client struct {
	baseURL string
	apiKey  string

	// ownerKey is the hex secp256k1 private key of the instructor wallet.
	ownerKey string

	// userID / walletID identify the platform wallet whose balance is the pool.
	userID   string
	walletID string

	// Operator-provided KYC document image paths (JPEG/PNG). Empty means the
	// provisioning flow stops before the upload step.
	docIdentification string
	docProofOfAddress string
	docBiometric      string

	http *http.Client
}

// New builds a BMONI client. ownerKey is hex without 0x.
func New(baseURL, apiKey, ownerKey, userID, walletID string) *Client {
	return &Client{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		apiKey:   apiKey,
		ownerKey: ownerKey,
		userID:   userID,
		walletID: walletID,
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

// WithKYCDocuments supplies operator-provided document image paths so
// provisioning can complete the upload step.
func (c *Client) WithKYCDocuments(identification, proofOfAddress, biometric string) *Client {
	c.docIdentification = identification
	c.docProofOfAddress = proofOfAddress
	c.docBiometric = biometric
	return c
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		buf = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, buf)
	if err != nil {
		return err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("bmoni %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("bmoni %s %s: status %d: %s", method, path, resp.StatusCode, truncate(string(raw), 300))
	}
	if out != nil && len(raw) > 0 {
		return json.Unmarshal(raw, out)
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// BalanceNGN reads the platform wallet NGN balance in kobo.
func (c *Client) BalanceNGN(ctx context.Context) (domain.Amount, error) {
	var resp struct {
		Data struct {
			Balance string `json:"balance"`
		} `json:"data"`
	}
	if c.userID == "" {
		return 0, errors.New("bmoni: no user id configured")
	}
	err := c.do(ctx, http.MethodGet,
		fmt.Sprintf("/v1/users/%s/smart-wallets/account/wallets", c.userID), nil, &resp)
	if err != nil {
		return 0, err
	}
	// BMONI reports display currency balance; treat the response's CNGN entry
	// as the pool. The adapter normalizes decimal string -> kobo.
	if resp.Data.Balance == "" {
		return 0, errors.New("bmoni: empty balance in wallet response")
	}
	return domain.FromNairaString(resp.Data.Balance)
}

// PayWinner runs the four-call proposal flow: create proposal, approve, fetch
// sign payload, sign with the owner key, submit. The recipient is resolved by
// the no-app/invite identity in the real build; this adapter signs against a
// configured recipient wallet address for the sandbox.
func (c *Client) PayWinner(ctx context.Context, toEmail string, amount domain.Amount) (string, error) {
	if c.ownerKey == "" || c.userID == "" {
		return "", errors.New("bmoni: owner key and user id required to send")
	}

	// 1. Proposal.
	var proposal struct {
		Data struct {
			Proposal struct {
				ID string `json:"id"`
			} `json:"proposal"`
		} `json:"data"`
	}
	err := c.do(ctx, http.MethodPost,
		fmt.Sprintf("/v1/users/%s/smart-wallets/%s/proposals", c.userID, c.walletID),
		map[string]any{
			"proposal": map[string]any{
				"type":        "TRANSFER",
				"toUserId":    toEmail, // replaced by resolved recipient in the real flow
				"amount":      amount.NairaString(),
				"currency":    "CNGN",
				"description": "kweeks prize payout",
			},
		}, &proposal)
	if err != nil {
		return "", err
	}
	proposalID := proposal.Data.Proposal.ID
	if proposalID == "" {
		return "", errors.New("bmoni: empty proposal id")
	}

	// 2. Approve.
	if err := c.do(ctx, http.MethodPost,
		fmt.Sprintf("/v1/users/%s/smart-wallets/proposals/%s/approve", c.userID, proposalID),
		nil, nil); err != nil {
		return "", err
	}

	// 3. Fetch sign payload (raw 32-byte digest; no EIP-191 prefix).
	var payload struct {
		Data struct {
			HashToSign string `json:"hashToSign"`
		} `json:"data"`
	}
	if err := c.do(ctx, http.MethodGet,
		fmt.Sprintf("/v1/users/%s/smart-wallets/proposals/%s/sign-payload", c.userID, proposalID),
		nil, &payload); err != nil {
		return "", err
	}

	// 4. Sign digest with the owner key and submit.
	sig, err := signDigest(c.ownerKey, payload.Data.HashToSign)
	if err != nil {
		return "", err
	}
	var submit struct {
		Data struct {
			Proposal struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"proposal"`
		} `json:"data"`
	}
	if err := c.do(ctx, http.MethodPost,
		fmt.Sprintf("/v1/users/%s/smart-wallets/proposals/%s/sign", c.userID, proposalID),
		map[string]string{"signature": sig}, &submit); err != nil {
		return "", err
	}
	return submit.Data.Proposal.ID, nil
}

// signDigest signs a raw 32-byte digest (no prefix) with the owner key.
func signDigest(ownerKeyHex, digestHex string) (string, error) {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(ownerKeyHex, "0x"))
	if err != nil {
		return "", fmt.Errorf("bmoni: bad owner key: %w", err)
	}
	digest := strings.TrimPrefix(digestHex, "0x")
	if len(digest) != 64 {
		return "", errors.New("bmoni: hashToSign is not a 32-byte hex digest")
	}
	sig, err := crypto.Sign(decodeHex(digest), key)
	if err != nil {
		return "", err
	}
	sig[64] += 27 // v: 0/1 -> 27/28
	return "0x" + hex.EncodeToString(sig), nil
}

func decodeHex(s string) []byte {
	out, _ := hex.DecodeString(s)
	return out
}
