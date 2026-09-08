// Package bmoni implements ports.Money against the BMONI Embedded REST API.
// Each instructor gets a real BMONI user + CNGN smart wallet (owner-proof
// challenge -> create-managed -> KYC -> NGN rail onboarding), and prize money
// moves from the host wallet to a winner's Nigerian bank account via
// verify -> register -> offramp (proposal -> approve -> sign). The owner key
// is held server-side and never leaves the process.
package bmoni

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to BMONI Embedded.
type Client struct {
	baseURL string
	apiKey  string

	// ownerKey is the hex secp256k1 private key that signs owner-proof
	// challenges and proposal digests for every provisioned wallet.
	ownerKey string

	// Operator-provided KYC document image paths (JPEG/PNG). Empty means the
	// provisioning flow stops before the upload step (sandbox NGN completes
	// without them; a real rail should configure all three).
	docIdentification string
	docProofOfAddress string
	docBiometric      string

	http *http.Client
}

// New builds a BMONI client. ownerKey is hex without 0x.
func New(baseURL, apiKey, ownerKey string) *Client {
	return &Client{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		apiKey:   apiKey,
		ownerKey: ownerKey,
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

// ErrWalletNotProvisioned is returned by payout/funding helpers when the host
// wallet has no BMONI identity yet.
var ErrWalletNotProvisioned = errors.New("bmoni: wallet not provisioned on the rail")
