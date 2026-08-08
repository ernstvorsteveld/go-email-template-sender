.PHONY: all build run generate test test/http docker/up docker/down clean

APP_NAME=email-sender
MAIN_PATH=./cmd/server

all: generate test build

generate:
	@echo "Generating API server and types from openapi.yaml..."
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -config oapi-codegen.yaml openapi.yaml

build:
	@echo "Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

run:
	@echo "Starting $(APP_NAME)..."
	go run $(MAIN_PATH)

test:
	@echo "Running tests..."
	go test -v ./...

test/http:
	@echo "Running HTTP handler tests..."
	go test -v ./internal/adapter/in/http/...

docker/up:
	@echo "Starting Docker infrastructure (Postgres & Mailpit)..."
	docker compose up -d

docker/down:
	@echo "Stopping Docker infrastructure..."
	docker compose down

clean:
	@echo "Cleaning up..."
	rm -rf bin/
