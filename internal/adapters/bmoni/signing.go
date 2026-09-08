package bmoni

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
)

// signMessage signs a text message with the EIP-191 prefix (personal_sign).
// This is the correct method for the owner-proof challenge at wallet creation.
func signMessage(ownerKeyHex, message string) (string, error) {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(ownerKeyHex, "0x"))
	if err != nil {
		return "", fmt.Errorf("bmoni: bad owner key: %w", err)
	}
	hash := accounts.TextHash([]byte(message))
	sig, err := crypto.Sign(hash, key)
	if err != nil {
		return "", err
	}
	sig[64] += 27
	return "0x" + hex.EncodeToString(sig), nil
}

// signDigest signs a raw 32-byte digest (no prefix) with the owner key. This
// is the correct method for proposal signing (the opposite of the owner-proof
// challenge above).
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

// pubkeyToAddress derives the 0x address for an owner private key.
func pubkeyToAddress(ownerKeyHex string) (string, error) {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(ownerKeyHex, "0x"))
	if err != nil {
		return "", fmt.Errorf("bmoni: bad owner key: %w", err)
	}
	return crypto.PubkeyToAddress(key.PublicKey).Hex(), nil
}

// proposalReply captures the various proposal id response shapes.
type proposalReply struct {
	ProposalID string `json:"proposalId"`
	ID         string `json:"id"`
	Proposal   struct {
		ID string `json:"id"`
	} `json:"proposal"`
	Data struct {
		ProposalID string `json:"proposalId"`
		ID         string `json:"id"`
		Proposal   struct {
			ID string `json:"id"`
		} `json:"proposal"`
	} `json:"data"`
}

// approveAndSign drives a created proposal to settlement: approve, fetch the
// signing payload, sign the raw digest with the owner key, and submit.
func (c *Client) approveAndSign(ctx context.Context, fromUserID, proposalID string) (string, error) {
	if c.ownerKey == "" {
		return "", errors.New("bmoni: owner key required to send")
	}
	// 1. Approve.
	if err := c.do(ctx, http.MethodPost,
		fmt.Sprintf("/v1/users/%s/smart-wallets/proposals/%s/approve", fromUserID, proposalID),
		nil, nil); err != nil {
		return "", err
	}

	// 2. Fetch the signing payload (raw 32-byte digest, no prefix).
	var payload struct {
		Data struct {
			HashToSign string `json:"hashToSign"`
		} `json:"data"`
	}
	if err := c.do(ctx, http.MethodGet,
		fmt.Sprintf("/v1/users/%s/smart-wallets/proposals/%s/sign-payload", fromUserID, proposalID),
		nil, &payload); err != nil {
		return "", err
	}

	// 3. Sign the digest with the owner key and submit.
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
		fmt.Sprintf("/v1/users/%s/smart-wallets/proposals/%s/sign", fromUserID, proposalID),
		map[string]string{"signature": sig}, &submit); err != nil {
		return "", err
	}
	pid := firstNonEmpty(submit.Data.Proposal.ID, proposalID)
	return pid, nil
}
