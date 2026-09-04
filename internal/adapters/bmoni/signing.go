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
	Data struct {
		Proposal struct {
			ID string `json:"id"`
		} `json:"proposal"`
		ProposalID string `json:"proposalId"`
		ID         string `json:"id"`
	} `json:"data"`
	ProposalID string `json:"proposalId"`
	ID         string `json:"id"`
}

// CreditTo funds a provisioned user's CNGN wallet from the platform wallet.
func (c *Client) CreditTo(ctx context.Context, toUserID, toWalletID, amountNaira string) (string, error) {
	return c.transferToWallet(ctx, c.userID, c.walletID, toUserID, toWalletID, amountNaira)
}

// PayTo sends CNGN from a provisioned wallet to a recipient wallet address.
func (c *Client) PayTo(ctx context.Context, fromUserID, fromWalletID, toAddr, amountNaira string) (string, error) {
	return c.transferToAddress(ctx, fromUserID, fromWalletID, toAddr, amountNaira, "kweeks prize payout")
}

// transferToWallet credits toUserID's wallet (recipient must hold an active
// CNGN wallet). This is the funding path used by CreditNGN.
func (c *Client) transferToWallet(ctx context.Context, fromUserID, fromWalletID, toUserID, _ string, amountNaira string) (string, error) {
	return c.sendProposal(ctx, fromUserID, fromWalletID, map[string]any{
		"toUserId": toUserID, "amount": amountNaira, "currency": "CNGN", "description": "kweeks wallet credit",
	})
}

// transfer creates+approves+signs+sends a proposal debiting fromUserID's
// wallet and crediting toUserID (which must hold an active wallet in CNGN).
func (c *Client) transfer(ctx context.Context, fromUserID, fromWalletID, toUserID, _ string, amountNaira, description string) (string, error) {
	return c.sendProposal(ctx, fromUserID, fromWalletID, map[string]any{
		"toUserId": toUserID, "amount": amountNaira, "currency": "CNGN", "description": description,
	})
}

// transferToAddress is the reliable sandbox path: recipient address that holds
// the wallet (no active-rail requirement on the recipient).
func (c *Client) transferToAddress(ctx context.Context, fromUserID, fromWalletID, toAddr, amountNaira, description string) (string, error) {
	return c.sendProposal(ctx, fromUserID, fromWalletID, map[string]any{
		"toAddress": toAddr, "amount": amountNaira, "currency": "CNGN", "description": description,
	})
}

func (c *Client) sendProposal(ctx context.Context, fromUserID, fromWalletID string, proposal map[string]any) (string, error) {
	if c.ownerKey == "" {
		return "", errors.New("bmoni: owner key required to send")
	}
	// 1. Create the proposal.
	var created proposalReply
	if err := c.do(ctx, http.MethodPost,
		fmt.Sprintf("/v1/users/%s/smart-wallets/%s/proposals", fromUserID, fromWalletID),
		map[string]any{"proposal": proposal}, &created); err != nil {
		return "", err
	}
	proposalID := firstNonEmpty(created.Data.Proposal.ID, created.Data.ProposalID, created.Data.ID, created.ProposalID, created.ID)
	if proposalID == "" {
		return "", errors.New("bmoni: proposal returned no id")
	}

	// 2. Approve.
	if err := c.do(ctx, http.MethodPost,
		fmt.Sprintf("/v1/users/%s/smart-wallets/proposals/%s/approve", fromUserID, proposalID),
		nil, nil); err != nil {
		return "", err
	}

	// 3. Fetch the signing payload (raw 32-byte digest, no prefix).
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

	// 4. Sign the digest with the owner key and submit.
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
