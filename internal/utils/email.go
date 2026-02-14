package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"

	"github.com/nodus-protocol/backend/internal/config"
	"go.uber.org/zap"
)

// disposableDomains is a curated list of known throwaway email providers.
var disposableDomains = map[string]struct{}{
	"mailinator.com":    {},
	"guerrillamail.com": {},
	"tempmail.com":      {},
	"10minutemail.com":  {},
	"trashmail.com":     {},
	"fakeinbox.com":     {},
	"yopmail.com":       {},
	"throwam.com":       {},
	"sharklasers.com":   {},
	"dispostable.com":   {},
}

// IsDisposableEmail returns true if the email's domain is a known
// disposable mail provider.
func IsDisposableEmail(email string) bool {
	parts := strings.SplitN(strings.ToLower(email), "@", 2)
	if len(parts) != 2 {
		return false
	}
	_, blocked := disposableDomains[parts[1]]
	return blocked
}

// NormalizeEmail lowercases and trims whitespace from an email address.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Mailer sends transactional emails.
type Mailer struct {
	cfg config.EmailConfig
	log *zap.Logger
}

// NewMailer creates a new Mailer.
func NewMailer(cfg config.EmailConfig, log *zap.Logger) *Mailer {
	return &Mailer{cfg: cfg, log: log}
}

// SendVerificationEmail sends an email with a verification link.
func (m *Mailer) SendVerificationEmail(toEmail, toName, verifyURL string) error {
	subject := "Verify your email — " + "Nodus Protocol"
	body := fmt.Sprintf(`
Hi %s,

Welcome to Nodus Protocol! Please verify your email address by clicking the link below:

%s

This link expires in 24 hours.

If you did not create an account, you can safely ignore this email.

— The Nodus Protocol Team
`, toName, verifyURL)

	return m.send(toEmail, subject, body)
}

// SendPasswordResetEmail sends a password reset link.
func (m *Mailer) SendPasswordResetEmail(toEmail, toName, resetURL string) error {
	subject := "Reset your password — Nodus Protocol"
	body := fmt.Sprintf(`
Hi %s,

We received a request to reset your password. Click the link below to set a new password:

%s

This link expires in 1 hour. If you did not request this, please ignore this email.

— The Nodus Protocol Team
`, toName, resetURL)

	return m.send(toEmail, subject, body)
}

// send is the internal SMTP dispatch method.
func (m *Mailer) send(to, subject, body string) error {
	if m.cfg.Host == "" {
		m.log.Warn("SMTP not configured — skipping email send", zap.String("to", to), zap.String("subject", subject))
		return nil
	}

	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	msg := buildMessage(m.cfg.From, m.cfg.FromName, to, subject, body)
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)

	if err := smtp.SendMail(addr, auth, m.cfg.From, []string{to}, []byte(msg)); err != nil {
		m.log.Error("failed to send email", zap.String("to", to), zap.Error(err))
		return fmt.Errorf("email send failed: %w", err)
	}

	m.log.Info("email sent", zap.String("to", to), zap.String("subject", subject))
	return nil
}

func buildMessage(from, fromName, to, subject, body string) string {
	var buf bytes.Buffer
	t := template.Must(template.New("email").Parse(
		"From: {{.FromName}} <{{.From}}>\r\n" +
			"To: {{.To}}\r\n" +
			"Subject: {{.Subject}}\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			"{{.Body}}",
	))
	_ = t.Execute(&buf, map[string]string{
		"From":     from,
		"FromName": fromName,
		"To":       to,
		"Subject":  subject,
		"Body":     body,
	})
	return buf.String()
}
