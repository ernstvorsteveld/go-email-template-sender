package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/out/postgres"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/service"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/testutils"
)

func TestContextService_CreateContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, cleanup, err := testutils.SetupTestDatabase(ctx)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := postgres.NewContextRepository(pool)
	svc := service.NewContextService(repo)

	id, err := svc.CreateContext(
		ctx,
		domain.ReferenceType("REF-123"),
		domain.CustomerType("Acme"),
		domain.JSONPayloadType(`{"key":"value"}`),
		domain.JSONPathType("$.key"),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	saved, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("expected context to be saved to DB")
	}

	if saved.ID != id {
		t.Errorf("expected id %v, got %v", id, saved.ID)
	}
	if saved.Reference != "REF-123" {
		t.Errorf("expected reference REF-123, got %v", saved.Reference)
	}
	if saved.Customer != "Acme" {
		t.Errorf("expected customer Acme, got %v", saved.Customer)
	}
}
