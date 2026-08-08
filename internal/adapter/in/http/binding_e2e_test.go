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

func TestBindingAPI_E2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, cleanup, err := testutils.SetupTestDatabase(ctx)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	tmplRepo := postgres.NewTemplateRepository(pool)
	styleRepo := postgres.NewStylesheetRepository(pool)
	bindRepo := postgres.NewBindingRepository(pool)

	tmplSvc := service.NewTemplateService(tmplRepo, styleRepo)
	bindSvc := service.NewBindingService(bindRepo)
	router := adapter_http.NewRouter(nil, nil, tmplSvc, bindSvc, nil)

	// 1. Create a Template to link to Binding
	tmplBody := `{"name": "Binding Target Tmpl", "code": "TMPL_BIND_1", "html_content": "<h1>Hello</h1>", "subject": "Binding Test Subject", "stylesheet_id": null}`
	req := httptest.NewRequest(http.MethodPost, "/templates", bytes.NewBufferString(tmplBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var tmplResp map[string]string
	json.NewDecoder(rr.Body).Decode(&tmplResp)
	tmplID := tmplResp["id"]

	// 2. Create Binding via POST /bindings
	createBody := `{"name": "VIP Customers Query", "query": "SELECT id, reference_id, customer_name, payload::text, email_jsonpath FROM contexts WHERE payload->>'status' = 'ACTIVE'", "template_id": "` + tmplID + `"}`
	req = httptest.NewRequest(http.MethodPost, "/bindings", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", rr.Code, rr.Body.String())
	}

	var createResp map[string]string
	json.NewDecoder(rr.Body).Decode(&createResp)
	id := createResp["id"]

	// 3. List Bindings via GET /bindings
	req = httptest.NewRequest(http.MethodGet, "/bindings?name=VIP", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rr.Code)
	}

	var listResp []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&listResp)
	if len(listResp) != 1 {
		t.Fatalf("expected 1 item in list, got %d", len(listResp))
	}

	// 4. Get Binding via GET /bindings/{id}
	req = httptest.NewRequest(http.MethodGet, "/bindings/"+id, nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rr.Code)
	}

	var getResp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&getResp)
	if getResp["name"] != "VIP Customers Query" {
		t.Errorf("expected name 'VIP Customers Query', got %v", getResp["name"])
	}

	// 5. Update Binding via PUT /bindings/{id}
	updateBody := `{"name": "All Customers Query", "query": "SELECT id, reference_id, customer_name, payload::text, email_jsonpath FROM contexts", "template_id": "` + tmplID + `"}`
	req = httptest.NewRequest(http.MethodPut, "/bindings/"+id, bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 No Content, got %d", rr.Code)
	}
}
