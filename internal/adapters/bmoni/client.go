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
		return &APIError{Method: method, Path: path, Status: resp.StatusCode, Body: string(raw)}
	}
	if out != nil && len(raw) > 0 {
		return json.Unmarshal(raw, out)
	}
	return nil
}

// APIError carries the HTTP status so callers can implement documented
// retry/recovery behaviour (e.g. a 409 on create-user = recover the existing
// user rather than retry).
type APIError struct {
	Method string
	Path   string
	Status int
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("bmoni %s %s: status %d: %s", e.Method, e.Path, e.Status, truncate(e.Body, 300))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// BalanceNGN reads the platform wallet NGN balance in kobo. The live sandbox
// reports balances under GET /smart-wallets/account/balances as a list keyed by
// smartWalletId with display currency "NGN".
func (c *Client) BalanceNGN(ctx context.Context) (domain.Amount, error) {
	var resp struct {
		Balances []struct {
			SmartWalletID string `json:"smartWalletId"`
			Currency      string `json:"currency"`
			Balance       string `json:"balance"`
		} `json:"balances"`
	}
	if c.userID == "" {
		return 0, errors.New("bmoni: no user id configured")
	}
	err := c.do(ctx, http.MethodGet,
		fmt.Sprintf("/v1/users/%s/smart-wallets/account/balances", c.userID), nil, &resp)
	if err != nil {
		return 0, err
	}
	for _, b := range resp.Balances {
		if b.Currency == "NGN" || b.Currency == "CNGN" {
			if b.Balance == "" {
				return 0, nil
			}
			return domain.FromNairaString(b.Balance)
		}
	}
	return 0, errors.New("bmoni: no NGN balance in wallet response")
}

// PayWinner transfers prize money from the platform wallet to a recipient
// BMONI user (winner persona). The recipient must hold an active wallet in the
// currency. Returns the settlement reference (proposal id).
func (c *Client) PayWinner(ctx context.Context, toUserID string, amount domain.Amount) (string, error) {
	if c.userID == "" || c.walletID == "" {
		return "", errors.New("bmoni: source wallet not configured")
	}
	if c.ownerKey == "" {
		return "", errors.New("bmoni: owner key required to send")
	}
	return c.sendProposal(ctx, c.userID, c.walletID, map[string]any{
		"type": "TRANSFER", "toUserId": toUserID,
		"amount": amount.NairaString(), "currency": "CNGN", "description": "kweeks prize payout",
	})
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
