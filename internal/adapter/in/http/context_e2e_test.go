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

func TestContextAPI_E2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, cleanup, err := testutils.SetupTestDatabase(ctx)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	ctxRepo := postgres.NewContextRepository(pool)
	ctxSvc := service.NewContextService(ctxRepo)
	router := adapter_http.NewRouter(ctxSvc, nil, nil, nil, nil)

	// 1. Create Context via POST /contexts
	createBody := `{"reference_id": "REF-E2E-1", "customer_name": "Acme Corp", "payload": "{\"email\":\"alice@acme.com\"}", "email_jsonpath": "$.email"}`
	req := httptest.NewRequest(http.MethodPost, "/contexts", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", rr.Code, rr.Body.String())
	}

	var createResp map[string]string
	json.NewDecoder(rr.Body).Decode(&createResp)
	id := createResp["id"]
	if id == "" {
		t.Fatalf("expected valid id in create response")
	}

	// 2. List Contexts via GET /contexts
	req = httptest.NewRequest(http.MethodGet, "/contexts?customer_name=Acme", nil)
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

	// 3. Get Context via GET /contexts/{id}
	req = httptest.NewRequest(http.MethodGet, "/contexts/"+id, nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rr.Code)
	}

	var getResp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&getResp)
	if getResp["reference_id"] != "REF-E2E-1" {
		t.Errorf("expected reference_id 'REF-E2E-1', got %v", getResp["reference_id"])
	}

	// 4. Update Context via PUT /contexts/{id}
	updateBody := `{"reference_id": "REF-E2E-UPDATED", "customer_name": "Acme Corp Updated", "payload": "{\"email\":\"alice.new@acme.com\"}", "email_jsonpath": "$.email"}`
	req = httptest.NewRequest(http.MethodPut, "/contexts/"+id, bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 No Content, got %d", rr.Code)
	}
}
