.PHONY: generate build test lint e2e e2e-ui dev-backend

# ── Code generation (single source of truth: api/openapi.yaml) ───────────────
#
# Backend Go types + Chi server interface:
#   go generate ./internal/generated/api/
#   (requires: go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest)
#
# Frontend TypeScript SDK:
#   cd frontend && npx @hey-api/openapi-ts
#   (config: frontend/openapi-ts.config.ts)
#
# Sync the embedded spec copy after editing api/openapi.yaml:
#   cp api/openapi.yaml internal/adapter/http/handler/spec/openapi.yaml
#
# Run all three in one shot:
generate: generate-backend generate-frontend generate-spec-copy

generate-backend:
	go generate ./internal/generated/api/

generate-frontend:
	cd frontend && npx @hey-api/openapi-ts

generate-spec-copy:
	cp api/openapi.yaml internal/adapter/http/handler/spec/openapi.yaml

# ── Build ─────────────────────────────────────────────────────────────────────
build: frontend/dist
	go build -o bin/shiftmanager ./cmd/server

frontend/dist: frontend/src frontend/package.json
	cd frontend && npm run build
	cp -r frontend/dist/* internal/infrastructure/static/dist/

# ── Tests ─────────────────────────────────────────────────────────────────────
test:
	go test ./...
	cd frontend && npm test -- --run

lint:
	go vet ./...
	cd frontend && npx tsc --noEmit

# ── E2E tests (Playwright) ────────────────────────────────────────────────────
#
# Prerequisites (running inside the devcontainer):
#   1. The Go backend must be running:  make dev-backend   (or go run ./cmd/server)
#   2. PostgreSQL is available via the devcontainer compose service (postgres:5432)
#
# The Playwright webServer config auto-starts the Vite dev server if needed.
#
e2e:
	cd frontend && npm run e2e

e2e-ui:
	cd frontend && npm run e2e:ui

# Starts the Go backend in the foreground (Ctrl-C to stop).
dev-backend:
	go run ./cmd/server
