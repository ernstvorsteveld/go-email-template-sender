package main

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -config oapi-codegen.yaml openapi.yaml

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	http_adapter "github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/in/http"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/out/email"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/out/postgres"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Initialize standard structured logger (slog) available in Go 1.21+
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	slog.Info("Starting Go Email Template Sender API...")

	ctx := context.Background()

	// 1. Initialize Postgres connection pool (Driven Adapter)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/emaildb"
	}

	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("Unable to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer dbPool.Close()

	// 2. Instantiate Outbound Ports (Repositories & Services)
	contextRepo := postgres.NewContextRepository(dbPool)
	stylesheetRepo := postgres.NewStylesheetRepository(dbPool)
	templateRepo := postgres.NewTemplateRepository(dbPool)
	bindingRepo := postgres.NewBindingRepository(dbPool)

	emailHost := os.Getenv("SMTP_HOST")
	if emailHost == "" {
		emailHost = "localhost"
	}
	emailSender := email.NewSMTPSender(emailHost, 1025)

	// 3. Instantiate Inbound Ports (Application Use Cases)
	contextSvc := service.NewContextService(contextRepo)
	stylesheetSvc := service.NewStylesheetService(stylesheetRepo)
	templateSvc := service.NewTemplateService(templateRepo, stylesheetRepo)
	bindingSvc := service.NewBindingService(bindingRepo)
	deliverySvc := service.NewDeliveryService(bindingRepo, contextRepo, templateSvc, emailSender)

	// 4. Instantiate Driving Adapters (HTTP Router)
	router := http_adapter.NewRouter(contextSvc, stylesheetSvc, templateSvc, bindingSvc, deliverySvc)

	// 5. Start HTTP Server
	srv := &http.Server{
		Addr:         ":8180",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	slog.Info("Server listening on :8180")
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("Server failed", slog.Any("error", err))
		os.Exit(1)
	}
}
