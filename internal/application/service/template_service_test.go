package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/out/postgres"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/service"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/testutils"
	"github.com/google/uuid"
)

func TestTemplateService_RenderTemplate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, cleanup, err := testutils.SetupTestDatabase(ctx)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()
	
	tmplRepo := postgres.NewTemplateRepository(pool)
	styleRepo := postgres.NewStylesheetRepository(pool)

	// 1. Setup linked stylesheet
	styleID := domain.IDType(uuid.New())
	style := domain.Stylesheet{
		ID:      styleID,
		Name:    "Test Style",
		Code:    "TS_1",
		Content: domain.CSSType("h1 { color: red; }"),
	}
	if err := styleRepo.Save(ctx, style); err != nil {
		t.Fatalf("failed to save stylesheet: %v", err)
	}

	// 2. Setup base template
	tmplID := domain.IDType(uuid.New())
	tmpl := domain.Template{
		ID:         tmplID,
		Name:       "Test Tmpl",
		Code:       "TT_1",
		Version:    1,
		Stylesheet: &styleID,
		Content:    domain.HTMLType("<html><head><title>Test</title></head><body><h1>Hello</h1></body></html>"),
		Subject:    "Test Subject",
	}
	if err := tmplRepo.Save(ctx, tmpl); err != nil {
		t.Fatalf("failed to save template: %v", err)
	}

	// 3. Execute
	svc := service.NewTemplateService(tmplRepo, styleRepo)
	html, err := svc.RenderTemplate(ctx, tmplID)
	
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 4. Assert goquery successfully injected the CSS into the <head>
	expectedCSS := "<style>\nh1 { color: red; }\n</style>"
	if !strings.Contains(string(html), expectedCSS) {
		t.Errorf("expected HTML to contain injected CSS %q, got: %s", expectedCSS, html)
	}
}

func TestTemplateService_CreateTemplate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, cleanup, err := testutils.SetupTestDatabase(ctx)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	tmplRepo := postgres.NewTemplateRepository(pool)
	styleRepo := postgres.NewStylesheetRepository(pool)

	svc := service.NewTemplateService(tmplRepo, styleRepo)
	id, err := svc.CreateTemplate(ctx, domain.NameType("Test"), domain.CodeType("TST_01"), domain.HTMLType("<p>test</p>"), nil, domain.SubjectType("Test Subject"))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	saved, err := tmplRepo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("expected template to be saved to DB")
	}

	if saved.Name != "Test" {
		t.Errorf("expected name Test, got %v", saved.Name)
	}
	if saved.Version != 1 {
		t.Errorf("expected initial version to be 1, got %d", saved.Version)
	}
}
