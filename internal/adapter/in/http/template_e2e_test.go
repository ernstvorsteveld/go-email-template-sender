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

func TestTemplateAPI_E2E(t *testing.T) {
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
	styleSvc := service.NewStylesheetService(styleRepo)
	router := adapter_http.NewRouter(nil, styleSvc, tmplSvc, nil, nil)

	// 1. Create Stylesheet to link to Template
	styleBody := `{"name": "Template Theme", "code": "TMPL_STYLE_1", "css_content": "h1 { color: green; }"}`
	req := httptest.NewRequest(http.MethodPost, "/stylesheets", bytes.NewBufferString(styleBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var styleResp map[string]string
	json.NewDecoder(rr.Body).Decode(&styleResp)
	styleID := styleResp["id"]

	// 2. Create Template via POST /templates
	createBody := `{"name": "Invoice Template", "code": "TMPL_INV_1", "html_content": "<html><head></head><body><h1>Invoice</h1></body></html>", "subject": "Monthly Statement", "stylesheet_id": "` + styleID + `"}`
	req = httptest.NewRequest(http.MethodPost, "/templates", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", rr.Code, rr.Body.String())
	}

	var createResp map[string]string
	json.NewDecoder(rr.Body).Decode(&createResp)
	id := createResp["id"]

	// 3. List Templates via GET /templates
	req = httptest.NewRequest(http.MethodGet, "/templates?name=Invoice", nil)
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

	// 4. Get Template via GET /templates/{id}
	req = httptest.NewRequest(http.MethodGet, "/templates/"+id, nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rr.Code)
	}

	var getResp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&getResp)
	if getResp["code"] != "TMPL_INV_1" {
		t.Errorf("expected code 'TMPL_INV_1', got %v", getResp["code"])
	}

	// 5. Render Template via GET /templates/{id}/render
	req = httptest.NewRequest(http.MethodGet, "/templates/"+id+"/render", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rr.Code)
	}

	renderedHTML := rr.Body.String()
	if !bytes.Contains([]byte(renderedHTML), []byte("h1 { color: green; }")) {
		t.Errorf("expected rendered HTML to contain injected CSS, got %s", renderedHTML)
	}

	// 6. Update Template via PUT /templates/{id}
	updateBody := `{"name": "Invoice Template V2", "code": "TMPL_INV_1", "html_content": "<h1>Updated Invoice</h1>", "subject": "Updated Statement", "stylesheet_id": null}`
	req = httptest.NewRequest(http.MethodPut, "/templates/"+id, bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 No Content, got %d", rr.Code)
	}
}
