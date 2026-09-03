// Package mailer implements ports.Mail. Email is a recovery artifact only:
// failures log and return nil so the claim flow never depends on delivery.
package mailer

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"

	"github.com/osugodbless/kweeks/internal/ports"
)

// Mailer sends redemption emails over SMTP with a logged fallback.
type Mailer struct {
	host   string
	port   int
	user   string
	pass   string
	from   string
	logger *slog.Logger
}

func New(host string, port int, user, pass, from string, logger *slog.Logger) *Mailer {
	if logger == nil {
		logger = slog.Default()
	}
	return &Mailer{host: host, port: port, user: user, pass: pass, from: from, logger: logger}
}

// SendRedemptionEmail is best-effort by contract. On SMTP failure it logs the
// error AND the full redemption payload to the server log so the demo has a
// paper trail even with no mail server configured.
func (m *Mailer) SendRedemptionEmail(ctx context.Context, to, claimCode string, amountNaira string) error {
	if m.host == "" {
		m.logger.Info("mail: no SMTP configured; redemption payload logged",
			"to", to, "amountNaira", amountNaira)
		return nil
	}
	subject := "Your kweeks prize is ready to redeem"
	body := fmt.Sprintf("You won %s NGN on kweeks.\n\nClaim code: %s\nRedeem at your kweeks podium screen.\n", amountNaira, claimCode)
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body + "\r\n")

	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	var auth smtp.Auth
	if m.user != "" {
		auth = smtp.PlainAuth("", m.user, m.pass, m.host)
	}
	if err := smtp.SendMail(addr, auth, m.from, []string{to}, msg); err != nil {
		// Never fail the caller; the claim is the source of truth.
		m.logger.Error("mail: send failed; redemption payload logged",
			"err", err, "to", to, "amountNaira", amountNaira, "claimCode", claimCode)
	}
	return nil
}

// Compile-time assertion that Mailer satisfies ports.Mail.
var _ ports.Mail = (*Mailer)(nil)
