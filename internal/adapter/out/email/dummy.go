package email

import (
	"context"
	"log"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/out"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
)

type dummyEmailSender struct{}

// NewDummyEmailSender creates a mock email sender that logs to stdout instead of sending real emails.
func NewDummyEmailSender() out.EmailSender {
	return &dummyEmailSender{}
}

func (s *dummyEmailSender) Send(ctx context.Context, to domain.EmailAddressType, subject domain.SubjectType, htmlBody domain.HTMLType) error {
	log.Printf("[DUMMY EMAIL SENDER] Sending email to: %s\nSubject: %s\nBody Length: %d bytes\n", to, subject, len(htmlBody))
	return nil
}
