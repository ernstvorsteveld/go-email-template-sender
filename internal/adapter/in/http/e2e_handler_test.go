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

func TestE2E_HTTPHandlers(t *testing.T) {
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

	// 3. Setup Handlers and Router
	tmplHandler := adapter_http.NewTemplateHandler(tmplSvc)
	styleHandler := adapter_http.NewStylesheetHandler(styleSvc)
	ctxHandler := adapter_http.NewContextHandler(ctxSvc)
	bindHandler := adapter_http.NewBindingHandler(bindSvc)
	delHandler := adapter_http.NewDeliveryHandler(delSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /templates", tmplHandler.Create)
	mux.HandleFunc("GET /templates/{id}", tmplHandler.Get)
	mux.HandleFunc("POST /stylesheets", styleHandler.Create)
	mux.HandleFunc("POST /contexts", ctxHandler.Create)
	mux.HandleFunc("POST /bindings", bindHandler.Create)
	mux.HandleFunc("POST /deliveries", delHandler.Create)

	sendRequest := func(method, path, body string) *httptest.ResponseRecorder {
		var req *http.Request
		if body != "" {
			req = httptest.NewRequest(method, path, bytes.NewBuffer([]byte(body)))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(method, path, nil)
		}
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		return rr
	}

	// ==========================================
	// 4. Run Complete End-to-End User Journey
	// ==========================================

	// STEP A: Create Stylesheet
	rr := sendRequest("POST", "/stylesheets", `{"name": "Style", "code": "S_01", "css_content": "h1 {color: blue;}"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("failed to create stylesheet: %s", rr.Body.String())
	}
	var styleResp map[string]string
	json.NewDecoder(rr.Body).Decode(&styleResp)
	styleID := styleResp["id"]

	// STEP B: Create Template linked to the Stylesheet
	rr = sendRequest("POST", "/templates", `{"name": "E2E Template", "code": "E2E_01", "html_content": "<h1>Hi {{user}}!</h1>", "subject": "E2E Test", "stylesheet_id": "`+styleID+`"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("failed to create template: %s", rr.Body.String())
	}
	var tmplResp map[string]string
	json.NewDecoder(rr.Body).Decode(&tmplResp)
	tmplID := tmplResp["id"]

	// STEP C: Create Customer Context with JSON data
	rr = sendRequest("POST", "/contexts", `{"reference_id": "U-1", "customer_name": "Acme", "payload": "{\"user\": \"Bob\", \"email\": \"bob@e2e.local\"}", "email_jsonpath": "$.email"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("failed to create context: %s", rr.Body.String())
	}

	// STEP D: Create a SQL Binding to fetch Contexts and map to Template
	rr = sendRequest("POST", "/bindings", `{"name": "Acme Binding", "query": "SELECT id, reference_id, customer_name, payload::text, email_jsonpath FROM contexts", "template_id": "`+tmplID+`"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("failed to create binding: %s", rr.Body.String())
	}
	var bindResp map[string]string
	json.NewDecoder(rr.Body).Decode(&bindResp)
	bindID := bindResp["id"]

	// STEP E: Trigger Delivery Orchestrator via API
	rr = sendRequest("POST", "/deliveries", `{"template_id": "`+tmplID+`", "binding_id": "`+bindID+`"}`)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("failed to dispatch delivery: %s", rr.Body.String())
	}

	// STEP F: Wait and verify SMTP traffic hit Mailpit correctly
	time.Sleep(1 * time.Second)
	mpResp, err := http.Get(mailpit.APIURL + "/messages")
	if err != nil {
		t.Fatalf("failed to fetch from mailpit: %v", err)
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
	if msg.Subject != "E2E Test" {
		t.Errorf("expected subject 'E2E Test', got %q", msg.Subject)
	}
	if len(msg.To) == 0 || msg.To[0].Address != "bob@e2e.local" {
		t.Errorf("expected recipient 'bob@e2e.local', got %q", msg.To)
	}
}
