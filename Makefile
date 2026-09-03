# Variables
BINARY_NAME=kweeks-server
GO=go
GOFLAGS=-v

# Build variables
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

.PHONY: all build clean test test-integration lint fmt run dev db-up db-down db-logs install help

all: fmt test build

## build: Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/kweeks-server

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

## test: Run all tests
test:
	@echo "Running tests..."
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

## test-integration: Run Postgres integration tests (requires db-up)
test-integration:
	DATABASE_URL=postgres://kweeks:kweeks@localhost:5432/kweeks \
		$(GO) test -race -tags integration -count=1 ./internal/adapters/store/postgres/

## lint: Run linter
lint:
	golangci-lint run ./...

## fmt: Format code
fmt:
	$(GO) fmt ./...
	$(GO) vet ./...

## run: Build and run the application
run: build
	./bin/$(BINARY_NAME)

## dev: Build and run reading .env from disk (if present)
dev: build
	@test -f .env || cp .env.example .env; \
	./bin/$(BINARY_NAME)

## db-up: Start the Postgres container (docker compose)
db-up:
	docker compose up -d postgres
	@echo "Waiting for Postgres to accept connections..."; \
	until docker compose exec -T postgres pg_isready -U kweeks -d kweeks >/dev/null 2>&1; do sleep 1; done; \
	echo "Postgres ready at postgres://kweeks:kweeks@localhost:5432/kweeks"

## db-down: Stop the Postgres container
db-down:
	docker compose down

## db-logs: Tail Postgres logs
db-logs:
	docker compose logs -f postgres

## install: Download dependencies
install:
	$(GO) mod download
	$(GO) mod tidy

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
