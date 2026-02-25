.PHONY: run build test lint tidy migrate-up migrate-down docker-up docker-down

APP_NAME    := todo-app
BINARY_DIR  := bin
BINARY      := $(BINARY_DIR)/$(APP_NAME)
MIGRATIONS  := migrations
DB_URL      ?= postgres://postgres:postgres@localhost:5432/todo_db?sslmode=disable

## ── Build ───────────────────────────────────────────────────────────────────

build:
	@echo "→ Building $(APP_NAME)..."
	@mkdir -p $(BINARY_DIR)
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/api

run:
	go run ./cmd/api

## ── Test & Quality ──────────────────────────────────────────────────────────

test:
	go test ./... -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

test-verbose:
	go test ./... -race -v

lint:
	golangci-lint run ./...

tidy:
	go mod tidy
	go mod verify

vet:
	go vet ./...

## ── Database ────────────────────────────────────────────────────────────────

migrate-up:
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" down

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS) -seq $$name

db-seed: build
	$(BINARY) seed

## ── Docker ──────────────────────────────────────────────────────────────────

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-build:
	docker compose build

## ── Helpers ─────────────────────────────────────────────────────────────────

help:
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:' Makefile | awk -F: '{print "  " $$1}' | sort
	@echo ""
