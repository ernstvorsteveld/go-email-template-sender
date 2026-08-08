package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aymerick/raymond"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/in"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/out"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
	"github.com/oliveagle/jsonpath"
)

type deliveryService struct {
	bindingRepo  out.BindingRepository
	contextRepo  out.ContextRepository
	templatePort in.TemplateUseCase
	emailSender  out.EmailSender
}

// NewDeliveryService creates a new Delivery application service orchestrator.
func NewDeliveryService(
	bindingRepo out.BindingRepository,
	contextRepo out.ContextRepository,
	templatePort in.TemplateUseCase,
	emailSender out.EmailSender,
) in.DeliveryUseCase {
	return &deliveryService{
		bindingRepo:  bindingRepo,
		contextRepo:  contextRepo,
		templatePort: templatePort,
		emailSender:  emailSender,
	}
}

func (s *deliveryService) Dispatch(ctx context.Context, templateID domain.IDType, bindingID domain.IDType) error {
	slog.Info("Starting delivery dispatch", 
		slog.String("template_id", uuid.UUID(templateID).String()), 
		slog.String("binding_id", uuid.UUID(bindingID).String()),
	)

	// 1. Fetch the Binding
	binding, err := s.bindingRepo.FindByID(ctx, bindingID)
	if err != nil {
		slog.Error("Failed to fetch binding", slog.Any("error", err))
		return fmt.Errorf("failed to fetch binding: %w", err)
	}

	// 2. Resolve Context Data using Binding Query (Bulk Operation)
	contexts, err := s.contextRepo.ExecuteQuery(ctx, binding.Query)
	if err != nil {
		slog.Error("Failed to execute binding query", slog.Any("error", err))
		return fmt.Errorf("failed to execute binding query: %w", err)
	}
	slog.Info("Resolved contexts for delivery", slog.Int("count", len(contexts)))

	// 3a. Fetch the Template to get the Subject
	tmpl, err := s.templatePort.GetTemplate(ctx, templateID)
	if err != nil {
		slog.Error("Failed to fetch template", slog.Any("error", err))
		return fmt.Errorf("failed to fetch template: %w", err)
	}

	// 3b. Render base Template (this uses goquery internally to inject CSS)
	renderedHTML, err := s.templatePort.RenderTemplate(ctx, templateID)
	if err != nil {
		return fmt.Errorf("failed to render base template: %w", err)
	}

	// 4 & 5. Parse JSON, merge with Handlebars, and dispatch for EACH resolved Context
	for _, c := range contexts {
		var payloadData interface{}
		if err := json.Unmarshal([]byte(c.Payload), &payloadData); err != nil {
			return fmt.Errorf("invalid json payload for context %v: %w", c.ID, err)
		}

		// Extract email address using JSONPath
		emailRes, err := jsonpath.JsonPathLookup(payloadData, string(c.EmailAddress))
		if err != nil {
			return fmt.Errorf("failed to extract email using jsonpath '%s': %w", c.EmailAddress, err)
		}
		emailStr, ok := emailRes.(string)
		if !ok {
			return fmt.Errorf("extracted email is not a string for context %v", c.ID)
		}

		// Merge context payload into HTML using Handlebars (raymond)
		finalHTML, err := raymond.Render(string(renderedHTML), payloadData)
		if err != nil {
			return fmt.Errorf("failed to merge handlebars template: %w", err)
		}

		// Dispatch via email sender
		err = s.emailSender.Send(ctx, domain.EmailAddressType(emailStr), tmpl.Subject, domain.HTMLType(finalHTML))
		if err != nil {
			slog.Error("Failed to send email", slog.String("email", emailStr), slog.Any("error", err))
			return fmt.Errorf("failed to send email to %s: %w", emailStr, err)
		}

		slog.Debug("Successfully dispatched email", slog.String("context_id", uuid.UUID(c.ID).String()), slog.String("email", emailStr))
	}

	slog.Info("Delivery dispatch complete", slog.Int("total_sent", len(contexts)))
	return nil
}
