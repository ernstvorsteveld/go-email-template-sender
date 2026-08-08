package out

import (
	"context"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
)

type EmailSender interface {
	Send(ctx context.Context, to domain.EmailAddressType, subject domain.SubjectType, htmlBody domain.HTMLType) error
}
