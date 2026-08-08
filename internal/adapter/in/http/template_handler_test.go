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
	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/out/postgres"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/service"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/testutils"
)

func TestTemplateHandler_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, cleanup, err := testutils.SetupTestDatabase(ctx)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	tmplRepo := postgres.NewTemplateRepository(pool)
	styleRepo := postgres.NewStylesheetRepository(pool)
	tmplSvc := service.NewTemplateService(tmplRepo, styleRepo)
	handler := adapter_http.NewTemplateHandler(tmplSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /templates", handler.Create)
	mux.HandleFunc("GET /templates/{id}", handler.Get)

	// 1. Create a Template via HTTP POST
	reqBody := []byte(`{
		"name": "Integration Template",
		"code": "INT_01",
		"html_content": "<h1>Hello!</h1>",
		"subject": "Integration Test Subject",
		"stylesheet_id": null
	}`)

	req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var createResp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	id, ok := createResp["id"]
	if !ok || id == "" {
		t.Fatalf("expected id in create response")
	}

	// 2. Fetch the created Template via HTTP GET
	req = httptest.NewRequest(http.MethodGet, "/templates/"+id, nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var getResp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&getResp); err != nil {
		t.Fatalf("failed to decode get response: %v", err)
	}

	// Assert payload has all required keys (especially subject)
	expectedKeys := []string{"id", "name", "code", "version", "stylesheet_id", "html_content", "subject"}
	for _, key := range expectedKeys {
		if _, exists := getResp[key]; !exists {
			t.Errorf("missing key in response: %s", key)
		}
	}

	if getResp["name"] != "Integration Template" {
		t.Errorf("expected name 'Integration Template', got %v", getResp["name"])
	}
	if getResp["subject"] != "Integration Test Subject" {
		t.Errorf("expected subject 'Integration Test Subject', got %v", getResp["subject"])
	}
}
