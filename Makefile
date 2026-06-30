# Cloud Log Access — developer commands.
# Run `make help` for the list.

COMPOSE := docker compose

.DEFAULT_GOAL := help
.PHONY: help up down clean logs build run test lint tidy fe-install fe-dev fe-build fe-lint fe-typecheck

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

up: ## Build and start the full stack (detached)
	$(COMPOSE) up --build -d

down: ## Stop the stack
	$(COMPOSE) down

clean: ## Stop the stack and remove volumes
	$(COMPOSE) down -v

logs: ## Tail stack logs
	$(COMPOSE) logs -f

build: ## Compile the Go binaries
	cd backend && go build ./...

run: ## Run the BFF locally (requires JWT_SECRET in the environment)
	cd backend && go run ./cmd/server

test: ## Run Go tests with the race detector
	cd backend && go test -race ./...

lint: ## Run golangci-lint
	cd backend && golangci-lint run

tidy: ## Tidy and verify Go modules
	cd backend && go mod tidy

# ── Frontend ────────────────────────────────────────────────────────────────
fe-install: ## Install frontend dependencies
	cd frontend && npm ci

fe-dev: ## Run the Vite dev server (proxies /api → localhost:8080)
	cd frontend && npm run dev

fe-build: ## Type-check + production build of the SPA
	cd frontend && npm run build

fe-lint: ## ESLint the frontend
	cd frontend && npm run lint

fe-typecheck: ## tsc --noEmit on the frontend
	cd frontend && npm run typecheck
