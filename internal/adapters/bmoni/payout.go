package bmoni

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/osugodbless/kweeks/internal/domain"
)

// ListNigerianBanks returns the supported Nigerian banks (CBN code + full
// name) for the winner's payout form. The user id only scopes the call; the
// list itself is partner-wide.
func (c *Client) ListNigerianBanks(ctx context.Context, userID string) ([]domain.NigerianBank, error) {
	var resp struct {
		Banks []struct {
			Code string `json:"code"`
			Name string `json:"name"`
			// The live sandbox returns bankName/bankCode.
			BankCode string `json:"bankCode"`
			BankName string `json:"bankName"`
		} `json:"banks"`
		Data struct {
			Banks []struct {
				Code string `json:"code"`
				Name string `json:"name"`
				// The live sandbox returns bankName/bankCode.
				BankCode string `json:"bankCode"`
				BankName string `json:"bankName"`
			} `json:"banks"`
		} `json:"data"`
	}
	if err := c.do(ctx, http.MethodGet,
		"/v1/users/"+userID+"/bank-accounts/nigerian-banks", nil, &resp); err != nil {
		return nil, err
	}
	src := resp.Banks
	if len(src) == 0 {
		src = resp.Data.Banks
	}
	out := make([]domain.NigerianBank, 0, len(src))
	for _, b := range src {
		code, name := b.Code, b.Name
		if code == "" {
			code = b.BankCode
		}
		if name == "" {
			name = b.BankName
		}
		if code == "" || name == "" {
			continue
		}
		out = append(out, domain.NigerianBank{Code: code, Name: name})
	}
	if len(out) == 0 {
		return nil, errors.New("bmoni: nigerian-banks returned no banks")
	}
	return out, nil
}

// VerifyNigerianAccount resolves an account number + bank code to the
// registered account holder name (exact name the registration requires).
func (c *Client) VerifyNigerianAccount(ctx context.Context, userID, accountNumber, bankCode string) (string, error) {
	var resp struct {
		AccountHolderName string `json:"accountHolderName"`
		AccountName       string `json:"accountName"`
		Name              string `json:"name"`
		Data              struct {
			AccountHolderName string `json:"accountHolderName"`
			AccountName       string `json:"accountName"`
			Name              string `json:"name"`
		} `json:"data"`
	}
	if err := c.do(ctx, http.MethodPost,
		"/v1/users/"+userID+"/bank-accounts/verify-nigerian-account",
		map[string]string{"accountNumber": accountNumber, "bankCode": bankCode}, &resp); err != nil {
		return "", err
	}
	name := firstNonEmpty(resp.AccountHolderName, resp.AccountName, resp.Name,
		resp.Data.AccountHolderName, resp.Data.AccountName, resp.Data.Name)
	if name == "" {
		return "", errors.New("bmoni: verify-nigerian-account returned no account holder name")
	}
	return name, nil
}

// RegisterNigerianWithdrawalAccount get-or-creates the withdrawal account and
// returns its bankAccountId for the offramp call.
func (c *Client) RegisterNigerianWithdrawalAccount(ctx context.Context, userID string, acct domain.NigerianAccount) (string, error) {
	var resp struct {
		ID             string `json:"id"`
		BankAccountID  string `json:"bankAccountId"`
		AccountID      string `json:"accountId"`
		BankAccount    struct{ ID string `json:"id"` } `json:"bankAccount"`
		Data           struct {
			ID            string `json:"id"`
			BankAccountID string `json:"bankAccountId"`
			AccountID     string `json:"accountId"`
			BankAccount   struct{ ID string `json:"id"` } `json:"bankAccount"`
		} `json:"data"`
	}
	if err := c.do(ctx, http.MethodPost,
		"/v1/users/"+userID+"/bank-accounts/withdrawal-accounts/nigeria",
		map[string]string{
			"accountNumber": acct.AccountNumber, "bankCode": acct.BankCode,
			"bankName": acct.BankName, "accountHolderName": acct.AccountHolderName,
		}, &resp); err != nil {
		return "", err
	}
	id := firstNonEmpty(resp.ID, resp.BankAccountID, resp.AccountID, resp.BankAccount.ID,
		resp.Data.ID, resp.Data.BankAccountID, resp.Data.AccountID, resp.Data.BankAccount.ID)
	if id == "" {
		return "", errors.New("bmoni: withdrawal-accounts/nigeria returned no account id")
	}
	return id, nil
}

// PayWinnerToNigerianBank offramps prize money from the host wallet to a
// registered Nigerian bank account (proposal -> approve -> sign). Returns the
// settlement reference (proposal id).
func (c *Client) PayWinnerToNigerianBank(ctx context.Context, from *domain.WalletExternal, bankAccountID string, amount domain.Amount) (string, error) {
	if from == nil || from.UserID == "" || from.WalletID == "" {
		return "", ErrWalletNotProvisioned
	}
	if c.ownerKey == "" {
		return "", errors.New("bmoni: owner key required to send")
	}
	var created struct {
		Data struct {
			ProposalID string `json:"proposalId"`
			Proposal   struct {
				ID string `json:"id"`
			} `json:"proposal"`
		} `json:"data"`
	}
	if err := c.do(ctx, http.MethodPost,
		fmt.Sprintf("/v1/users/%s/smart-wallets/%s/offramp/nigeria", from.UserID, from.WalletID),
		map[string]string{"bankAccountId": bankAccountID, "fromAmount": amount.NairaString()}, &created); err != nil {
		return "", err
	}
	proposalID := firstNonEmpty(created.Data.ProposalID, created.Data.Proposal.ID)
	if proposalID == "" {
		return "", errors.New("bmoni: offramp/nigeria returned no proposal id")
	}
	return c.approveAndSign(ctx, from.UserID, proposalID)
}

// DepositAccount returns the host's NGN virtual bank account (account number
// + bank) that a bank transfer to it funds the wallet. The VBA is issued
// during Nigeria onboarding; this reads it back.
func (c *Client) DepositAccount(ctx context.Context, userID, walletID string) (accountNumber, bankName string, err error) {
	if userID == "" || walletID == "" {
		return "", "", ErrWalletNotProvisioned
	}
	var resp struct {
		Accounts []struct {
			ID             string `json:"id"`
			AccountNumber  string `json:"accountNumber"`
			BankName       string `json:"bankName"`
			BankCode       string `json:"bankCode"`
			Currency       string `json:"currency"`
			TargetCurrency string `json:"targetCurrency"`
		} `json:"accounts"`
		Data struct {
			Accounts []struct {
				ID             string `json:"id"`
				AccountNumber  string `json:"accountNumber"`
				BankName       string `json:"bankName"`
				BankCode       string `json:"bankCode"`
				Currency       string `json:"currency"`
				TargetCurrency string `json:"targetCurrency"`
			} `json:"accounts"`
		} `json:"data"`
	}
	if err := c.do(ctx, http.MethodGet,
		"/v1/users/"+userID+"/bank-accounts/deposit-accounts/NGN", nil, &resp); err != nil {
		return "", "", err
	}
	src := resp.Accounts
	if len(src) == 0 {
		src = resp.Data.Accounts
	}
	for _, a := range src {
		if a.AccountNumber == "" {
			continue
		}
		// Pick the NGN VBA; tolerate currency reported on either field.
		if strings.EqualFold(a.Currency, "NGN") || strings.EqualFold(a.TargetCurrency, "NGN") {
			return a.AccountNumber, a.BankName, nil
		}
	}
	// Fall back to the first account when the currency field is absent.
	if len(src) > 0 {
		return src[0].AccountNumber, src[0].BankName, nil
	}
	return "", "", errors.New("bmoni: deposit-accounts/NGN returned no account")
}
