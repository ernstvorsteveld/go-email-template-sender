package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/out/email"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/out/postgres"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/service"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/testutils"
	"github.com/google/uuid"
)

func TestDeliveryService_Dispatch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 1. Setup Postgres Container
	pool, pgCleanup, err := testutils.SetupTestDatabase(ctx)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer pgCleanup()

	mailpit, err := testutils.SetupMailpit(ctx)
	if err != nil {
		t.Fatalf("Failed to setup mailpit: %v", err)
	}
	defer mailpit.Cleanup()

	// 3. Setup Repositories
	tmplRepo := postgres.NewTemplateRepository(pool)
	styleRepo := postgres.NewStylesheetRepository(pool)
	bindRepo := postgres.NewBindingRepository(pool)
	ctxRepo := postgres.NewContextRepository(pool)

	tmplUC := service.NewTemplateService(tmplRepo, styleRepo)
	sender := email.NewSMTPSender(mailpit.SMTPHost, mailpit.SMTPPort)

	// 4. Seed Data
	tmplID := domain.IDType(uuid.New())
	tmpl := domain.Template{
		ID:      tmplID,
		Name:    "Integration Tmpl",
		Code:    "INT_1",
		Version: 1,
		Content: domain.HTMLType("<h1>Hello from Postgres!</h1>"),
		Subject: "Testcontainers Integration Subject",
	}
	if err := tmplRepo.Save(ctx, tmpl); err != nil {
		t.Fatalf("failed to save template: %v", err)
	}

	bindID := domain.IDType(uuid.New())
	bind := domain.Binding{
		ID:         bindID,
		Name:       "All Users",
		Query:      domain.SQLQueryType("SELECT id, reference_id, customer_name, payload::text, email_jsonpath FROM contexts"),
		Template: tmplID,
	}
	if err := bindRepo.Save(ctx, bind); err != nil {
		t.Fatalf("failed to save binding: %v", err)
	}

	ctxID := domain.IDType(uuid.New())
	contextEntry := domain.Context{
		ID:            ctxID,
		Reference:     "INT-123",
		Customer:      "John Doe",
		Payload:       domain.JSONPayloadType(`{"email": "john@testcontainers.local"}`),
		EmailAddress: domain.JSONPathType("$.email"),
	}
	if err := ctxRepo.Save(ctx, contextEntry); err != nil {
		t.Fatalf("failed to save context: %v", err)
	}

	// 5. Run Dispatch
	svc := service.NewDeliveryService(bindRepo, ctxRepo, tmplUC, sender)
	err = svc.Dispatch(ctx, tmplID, bindID)
	if err != nil {
		t.Fatalf("expected no error from Dispatch, got %v", err)
	}

	// 6. Verify via Mailpit API
	time.Sleep(1 * time.Second) // wait for mailpit to process

	resp, err := http.Get(mailpit.APIURL + "/messages")
	if err != nil {
		t.Fatalf("failed to fetch messages from mailpit: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Messages []struct {
			ID      string `json:"ID"`
			Subject string `json:"Subject"`
			To      []struct {
				Address string `json:"Address"`
			} `json:"To"`
		} `json:"messages"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode mailpit response: %v", err)
	}

	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message in mailpit, got %d", len(result.Messages))
	}

	msg := result.Messages[0]
	if msg.Subject != "Testcontainers Integration Subject" {
		t.Errorf("expected subject 'Testcontainers Integration Subject', got %q", msg.Subject)
	}

	if len(msg.To) == 0 || msg.To[0].Address != "john@testcontainers.local" {
		t.Errorf("expected recipient 'john@testcontainers.local', got %v", msg.To)
	}
}

func TestDeliveryService_DispatchWithJSONBFilter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 1. Setup Postgres & Mailpit Containers
	pool, pgCleanup, err := testutils.SetupTestDatabase(ctx)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer pgCleanup()

	mailpit, err := testutils.SetupMailpit(ctx)
	if err != nil {
		t.Fatalf("Failed to setup mailpit: %v", err)
	}
	defer mailpit.Cleanup()

	// 2. Setup Repositories & Services
	tmplRepo := postgres.NewTemplateRepository(pool)
	styleRepo := postgres.NewStylesheetRepository(pool)
	bindRepo := postgres.NewBindingRepository(pool)
	ctxRepo := postgres.NewContextRepository(pool)

	tmplUC := service.NewTemplateService(tmplRepo, styleRepo)
	sender := email.NewSMTPSender(mailpit.SMTPHost, mailpit.SMTPPort)

	// 3. Seed Template
	tmplID := domain.IDType(uuid.New())
	tmpl := domain.Template{
		ID:      tmplID,
		Name:    "Filtered Tmpl",
		Code:    "FILT_1",
		Version: 1,
		Content: domain.HTMLType("<h1>Hello {{user}}!</h1>"),
		Subject: "JSONB Filter Test",
	}
	if err := tmplRepo.Save(ctx, tmpl); err != nil {
		t.Fatalf("failed to save template: %v", err)
	}

	// 4. Seed Binding with Postgres JSONB WHERE filter (payload->>'status' = 'ACTIVE')
	bindID := domain.IDType(uuid.New())
	bind := domain.Binding{
		ID:         bindID,
		Name:       "Active VIP Users Only",
		Query:      domain.SQLQueryType("SELECT id, reference_id, customer_name, payload::text, email_jsonpath FROM contexts WHERE payload->>'status' = 'ACTIVE'"),
		Template: tmplID,
	}
	if err := bindRepo.Save(ctx, bind); err != nil {
		t.Fatalf("failed to save binding: %v", err)
	}

	// 5. Context 1: ACTIVE (should match query)
	ctx1 := domain.Context{
		ID:            domain.IDType(uuid.New()),
		Reference:     "REF-1",
		Customer:      "Alice",
		Payload:       domain.JSONPayloadType(`{"user": "Alice", "email": "alice@active.local", "status": "ACTIVE"}`),
		EmailAddress: domain.JSONPathType("$.email"),
	}
	if err := ctxRepo.Save(ctx, ctx1); err != nil {
		t.Fatalf("failed to save context 1: %v", err)
	}

	// 6. Context 2: INACTIVE (should be filtered OUT by query)
	ctx2 := domain.Context{
		ID:            domain.IDType(uuid.New()),
		Reference:     "REF-2",
		Customer:      "Bob",
		Payload:       domain.JSONPayloadType(`{"user": "Bob", "email": "bob@inactive.local", "status": "INACTIVE"}`),
		EmailAddress: domain.JSONPathType("$.email"),
	}
	if err := ctxRepo.Save(ctx, ctx2); err != nil {
		t.Fatalf("failed to save context 2: %v", err)
	}

	// 7. Dispatch Delivery
	svc := service.NewDeliveryService(bindRepo, ctxRepo, tmplUC, sender)
	err = svc.Dispatch(ctx, tmplID, bindID)
	if err != nil {
		t.Fatalf("expected no error from Dispatch, got %v", err)
	}

	time.Sleep(1 * time.Second)

	// 8. Verify Mailpit received ONLY 1 message (for alice@active.local, NOT bob@inactive.local)
	resp, err := http.Get(mailpit.APIURL + "/messages")
	if err != nil {
		t.Fatalf("failed to fetch messages from mailpit: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Messages []struct {
			To []struct {
				Address string `json:"Address"`
			} `json:"To"`
		} `json:"messages"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode mailpit response: %v", err)
	}

	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message in mailpit, got %d", len(result.Messages))
	}

	if result.Messages[0].To[0].Address != "alice@active.local" {
		t.Errorf("expected recipient 'alice@active.local', got %v", result.Messages[0].To[0].Address)
	}
}

func TestDeliveryService_DispatchWithNestedJSONBFilter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 1. Setup Postgres & Mailpit Containers
	pool, pgCleanup, err := testutils.SetupTestDatabase(ctx)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer pgCleanup()

	mailpit, err := testutils.SetupMailpit(ctx)
	if err != nil {
		t.Fatalf("Failed to setup mailpit: %v", err)
	}
	defer mailpit.Cleanup()

	// 2. Setup Repositories & Services
	tmplRepo := postgres.NewTemplateRepository(pool)
	styleRepo := postgres.NewStylesheetRepository(pool)
	bindRepo := postgres.NewBindingRepository(pool)
	ctxRepo := postgres.NewContextRepository(pool)

	tmplUC := service.NewTemplateService(tmplRepo, styleRepo)
	sender := email.NewSMTPSender(mailpit.SMTPHost, mailpit.SMTPPort)

	// 3. Seed Template with Handlebars navigation over nested JSON
	tmplID := domain.IDType(uuid.New())
	tmpl := domain.Template{
		ID:      tmplID,
		Name:    "Nested Tmpl",
		Code:    "NEST_1",
		Version: 1,
		Content: domain.HTMLType("<h1>Hello {{customer.profile.name}}! Plan: {{customer.account.billing.plan}}</h1>"),
		Subject: "Nested Subobject Test",
	}
	if err := tmplRepo.Save(ctx, tmpl); err != nil {
		t.Fatalf("failed to save template: %v", err)
	}

	// 4. Seed Binding with query selecting on deep subobject attributes:
	// payload->'customer'->'account'->'billing'->>'plan' = 'ENTERPRISE'
	bindID := domain.IDType(uuid.New())
	bind := domain.Binding{
		ID:         bindID,
		Name:       "Enterprise Plan Billing Only",
		Query:      domain.SQLQueryType("SELECT id, reference_id, customer_name, payload::text, email_jsonpath FROM contexts WHERE payload->'customer'->'account'->'billing'->>'plan' = 'ENTERPRISE'"),
		Template: tmplID,
	}
	if err := bindRepo.Save(ctx, bind); err != nil {
		t.Fatalf("failed to save binding: %v", err)
	}

	// 5. Context 1: Enterprise Plan (Matches deep nested query)
	ctx1 := domain.Context{
		ID:        domain.IDType(uuid.New()),
		Reference: "REF-NESTED-1",
		Customer:  "Acme Corp",
		Payload: domain.JSONPayloadType(`{
			"customer": {
				"profile": {
					"name": "Alice Enterprise",
					"contact": {
						"email": "alice@enterprise-corp.com"
					}
				},
				"account": {
					"status": "ACTIVE",
					"billing": {
						"plan": "ENTERPRISE",
						"currency": "EUR"
					}
				}
			}
		}`),
		EmailAddress: domain.JSONPathType("$.customer.profile.contact.email"),
	}
	if err := ctxRepo.Save(ctx, ctx1); err != nil {
		t.Fatalf("failed to save context 1: %v", err)
	}

	// 6. Context 2: Starter Plan (Filtered OUT by deep nested query)
	ctx2 := domain.Context{
		ID:        domain.IDType(uuid.New()),
		Reference: "REF-NESTED-2",
		Customer:  "Startup LLC",
		Payload: domain.JSONPayloadType(`{
			"customer": {
				"profile": {
					"name": "Bob Starter",
					"contact": {
						"email": "bob@startup.com"
					}
				},
				"account": {
					"status": "ACTIVE",
					"billing": {
						"plan": "STARTER",
						"currency": "USD"
					}
				}
			}
		}`),
		EmailAddress: domain.JSONPathType("$.customer.profile.contact.email"),
	}
	if err := ctxRepo.Save(ctx, ctx2); err != nil {
		t.Fatalf("failed to save context 2: %v", err)
	}

	// 7. Dispatch Delivery
	svc := service.NewDeliveryService(bindRepo, ctxRepo, tmplUC, sender)
	err = svc.Dispatch(ctx, tmplID, bindID)
	if err != nil {
		t.Fatalf("expected no error from Dispatch, got %v", err)
	}

	time.Sleep(1 * time.Second)

	// 8. Verify Mailpit received ONLY 1 message (for Alice Enterprise, NOT Bob Starter)
	resp, err := http.Get(mailpit.APIURL + "/messages")
	if err != nil {
		t.Fatalf("failed to fetch messages from mailpit: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Messages []struct {
			To []struct {
				Address string `json:"Address"`
			} `json:"To"`
		} `json:"messages"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode mailpit response: %v", err)
	}

	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message in mailpit, got %d", len(result.Messages))
	}

	if result.Messages[0].To[0].Address != "alice@enterprise-corp.com" {
		t.Errorf("expected recipient 'alice@enterprise-corp.com', got %v", result.Messages[0].To[0].Address)
	}
}


