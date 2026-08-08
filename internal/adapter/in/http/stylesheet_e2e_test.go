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

func TestStylesheetAPI_E2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, cleanup, err := testutils.SetupTestDatabase(ctx)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	styleRepo := postgres.NewStylesheetRepository(pool)
	styleSvc := service.NewStylesheetService(styleRepo)
	router := adapter_http.NewRouter(nil, styleSvc, nil, nil, nil)

	// 1. Create Stylesheet via POST /stylesheets
	createBody := `{"name": "Corporate Blue", "code": "STYLE_CORP_1", "css_content": "h1 { color: #003366; }"}`
	req := httptest.NewRequest(http.MethodPost, "/stylesheets", bytes.NewBufferString(createBody))
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

	// 2. List Stylesheets via GET /stylesheets
	req = httptest.NewRequest(http.MethodGet, "/stylesheets?name=Corporate", nil)
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

	// 3. Get Stylesheet via GET /stylesheets/{id}
	req = httptest.NewRequest(http.MethodGet, "/stylesheets/"+id, nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rr.Code)
	}

	var getResp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&getResp)
	if getResp["code"] != "STYLE_CORP_1" {
		t.Errorf("expected code 'STYLE_CORP_1', got %v", getResp["code"])
	}

	// 4. Update Stylesheet via PUT /stylesheets/{id}
	updateBody := `{"name": "Corporate Blue Dark", "code": "STYLE_CORP_1", "css_content": "h1 { color: #001133; }"}`
	req = httptest.NewRequest(http.MethodPut, "/stylesheets/"+id, bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 No Content, got %d", rr.Code)
	}
}
