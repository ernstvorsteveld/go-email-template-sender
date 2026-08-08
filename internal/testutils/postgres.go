package testutils

import (
	"context"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupTestDatabase(ctx context.Context) (*pgxpool.Pool, func(), error) {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	schemaPath := filepath.Join(basepath, "..", "..", "schema.sql")

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithInitScripts(schemaPath),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(10*time.Second)),
	)
	if err != nil {
		return nil, nil, err
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, err
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		pool.Close()
		postgresContainer.Terminate(context.Background())
	}

	return pool, cleanup, nil
}
