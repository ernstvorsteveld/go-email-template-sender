package email

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/out"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
)

type smtpSender struct {
	host string
	port int
}

// NewSMTPSender creates a real SMTP email sender.
func NewSMTPSender(host string, port int) out.EmailSender {
	return &smtpSender{
		host: host,
		port: port,
	}
}

func (s *smtpSender) Send(ctx context.Context, to domain.EmailAddressType, subject domain.SubjectType, htmlBody domain.HTMLType) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	// Minimal SMTP MIME headers for HTML email
	msg := []byte(strings.Join([]string{
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		string(htmlBody),
	}, "\r\n"))

	// For local testing (Mailpit), we don't need auth.
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = "noreply@localhost"
	}

	err := smtp.SendMail(addr, nil, from, []string{string(to)}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	return nil
}
