package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	adapter_http "github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/in/http"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/out/email"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/out/postgres"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/service"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/testutils"
)

func TestDeliveryAPI_E2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 1. Setup Testcontainers (Postgres and Mailpit)
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

	// 2. Setup Adapters and Services
	tmplRepo := postgres.NewTemplateRepository(pool)
	styleRepo := postgres.NewStylesheetRepository(pool)
	ctxRepo := postgres.NewContextRepository(pool)
	bindRepo := postgres.NewBindingRepository(pool)
	sender := email.NewSMTPSender(mailpit.SMTPHost, mailpit.SMTPPort)

	tmplSvc := service.NewTemplateService(tmplRepo, styleRepo)
	styleSvc := service.NewStylesheetService(styleRepo)
	ctxSvc := service.NewContextService(ctxRepo)
	bindSvc := service.NewBindingService(bindRepo)
	delSvc := service.NewDeliveryService(bindRepo, ctxRepo, tmplSvc, sender)

	router := adapter_http.NewRouter(ctxSvc, styleSvc, tmplSvc, bindSvc, delSvc)

	sendRequest := func(method, path, body string) *httptest.ResponseRecorder {
		var req *http.Request
		if body != "" {
			req = httptest.NewRequest(method, path, bytes.NewBuffer([]byte(body)))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(method, path, nil)
		}
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		return rr
	}

	// 3. Create Stylesheet
	rr := sendRequest(http.MethodPost, "/stylesheets", `{"name": "E2E Style", "code": "DELIV_S1", "css_content": "h1 {color: purple;}"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("failed to create stylesheet: %s", rr.Body.String())
	}
	var styleResp map[string]string
	json.NewDecoder(rr.Body).Decode(&styleResp)
	styleID := styleResp["id"]

	// 4. Create Template
	rr = sendRequest(http.MethodPost, "/templates", `{"name": "Delivery Template", "code": "DELIV_T1", "html_content": "<h1>Hello {{user}}!</h1>", "subject": "Delivery E2E Test Subject", "stylesheet_id": "`+styleID+`"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("failed to create template: %s", rr.Body.String())
	}
	var tmplResp map[string]string
	json.NewDecoder(rr.Body).Decode(&tmplResp)
	tmplID := tmplResp["id"]

	// 5. Create Context
	rr = sendRequest(http.MethodPost, "/contexts", `{"reference_id": "DELIV-1", "customer_name": "Delivery Customer", "payload": "{\"user\": \"Charlie\", \"email\": \"charlie@delivery-e2e.local\"}", "email_jsonpath": "$.email"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("failed to create context: %s", rr.Body.String())
	}

	// 6. Create Binding
	rr = sendRequest(http.MethodPost, "/bindings", `{"name": "Delivery Binding", "query": "SELECT id, reference_id, customer_name, payload::text, email_jsonpath FROM contexts", "template_id": "`+tmplID+`"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("failed to create binding: %s", rr.Body.String())
	}
	var bindResp map[string]string
	json.NewDecoder(rr.Body).Decode(&bindResp)
	bindID := bindResp["id"]

	// 7. Trigger Delivery Dispatch via POST /deliveries
	rr = sendRequest(http.MethodPost, "/deliveries", `{"template_id": "`+tmplID+`", "binding_id": "`+bindID+`"}`)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("failed to dispatch delivery: %s", rr.Body.String())
	}

	// 8. Verify SMTP email receipt in Mailpit
	time.Sleep(1 * time.Second)
	mpResp, err := http.Get(mailpit.APIURL + "/messages")
	if err != nil {
		t.Fatalf("failed to fetch messages from mailpit: %v", err)
	}
	defer mpResp.Body.Close()

	var result struct {
		Messages []struct {
			Subject string `json:"Subject"`
			To      []struct {
				Address string `json:"Address"`
			} `json:"To"`
		} `json:"messages"`
	}
	json.NewDecoder(mpResp.Body).Decode(&result)

	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message in mailpit, got %d", len(result.Messages))
	}

	msg := result.Messages[0]
	if msg.Subject != "Delivery E2E Test Subject" {
		t.Errorf("expected subject 'Delivery E2E Test Subject', got %q", msg.Subject)
	}
	if len(msg.To) == 0 || msg.To[0].Address != "charlie@delivery-e2e.local" {
		t.Errorf("expected recipient 'charlie@delivery-e2e.local', got %v", msg.To)
	}
}
