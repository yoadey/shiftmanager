# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ShiftManager is a volunteer-shift management platform for a German sports club (TSC Schwarz-Gelb Aachen). It is a **modular monolith**: the Go backend embeds the compiled React SPA at build time via `embed.FS`; there is no separate static server. The backend exposes all API routes under `/api/v1`; everything else is served as the SPA.

## Commands

### Backend

```bash
go run ./cmd/server          # Start backend (port 8080)
go test ./...                # Run all Go tests
go vet ./...                 # Lint backend
```

Single package test:
```bash
go test ./internal/usecase/...
go test ./internal/testmode/...    # Integration tests (SQLite in-memory, no DB needed)
```

### Frontend

```bash
cd frontend
npm run dev          # Vite dev server (port 5173), proxies /api → localhost:8080
npm test -- --run    # Unit tests (vitest, one-shot)
npm run test:watch   # Vitest in watch mode
npx tsc --noEmit     # TypeScript type-check
npm run e2e          # Playwright E2E (requires backend running: make dev-backend)
npm run e2e:ui       # Playwright with interactive UI
```

### Build

```bash
make build           # Build frontend then embed into Go binary → bin/shiftmanager
make test            # Run both Go and frontend tests
make lint            # go vet + tsc --noEmit
make generate        # Regenerate API types from api/openapi.yaml (see Code Generation)
```

### Database (golang-migrate, production only)

```bash
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" down 1
migrate create -ext sql -dir migrations -seq <name>
```

## Architecture

### Backend — Hexagonal / Clean Architecture

```
cmd/server/                  — Entry point: wires all dependencies and starts HTTP server
internal/
  domain/                    — Pure domain entities and value objects (no framework deps)
  port/                      — Repository and service interfaces (the hexagon boundary)
  usecase/                   — Business logic; depends only on port interfaces
  adapter/
    db/                      — GORM repositories (PostgreSQL prod, SQLite for tests)
    http/
      handler/               — One handler file per domain area; all use oapi-codegen types
      middleware/             — JWT auth, role checks, rate limiting
      router.go              — Single chi router wiring all handlers
    email/                   — SMTP email service
    oidc/                    — go-oidc v3 adapter
    cache/                   — In-memory cache (Redis optional)
  infrastructure/
    logger/                  — zerolog request logger
    scheduler/               — Cron-style job scheduler (reminders, billing)
    static/                  — embed.FS serving compiled frontend dist
  generated/api/             — oapi-codegen output (types + chi server interface)
  testmode/                  — Self-contained httptest.Server backed by SQLite; used by integration tests
migrations/                  — SQL migration files (golang-migrate; only for PostgreSQL production)
```

**Dependency flow:** `handler → usecase → port ← adapter/db`

The `testmode` package wires the entire stack (all usecases and handlers) with a GORM SQLite in-memory DB and seeded fixture data. Integration tests in `internal/testmode/integration_test.go` use this — no PostgreSQL or OIDC provider required.

Usecase unit tests use lightweight in-memory fakes defined in `internal/usecase/mocks_test.go`; they do **not** use testify/mock.

### Database

GORM is used as the ORM. The `adapter/db/open.go` dispatcher selects SQLite when the DSN starts with `sqlite:`, PostgreSQL otherwise. `db.Migrate()` runs GORM AutoMigrate (safe for repeat calls). SQL migration files in `migrations/` are only used for production PostgreSQL (via golang-migrate), not in tests.

### Code Generation (single source of truth: `api/openapi.yaml`)

Two generated artefacts must be kept in sync whenever the OpenAPI spec changes:

1. **Backend Go types + chi server interface** (`internal/generated/api/types.gen.go`):
   ```bash
   make generate-backend   # go generate ./internal/generated/api/
   ```
2. **Frontend TypeScript SDK** (`frontend/src/api/` generated files):
   ```bash
   make generate-frontend  # cd frontend && npx @hey-api/openapi-ts
   ```
3. Sync the embedded spec copy: `make generate-spec-copy` (copies `api/openapi.yaml` → `internal/adapter/http/handler/spec/openapi.yaml`)

Run all three at once: `make generate`

### Frontend Architecture

```
frontend/src/
  api/           — Typed API clients (generated + hand-written mappers)
  components/    — Shared UI components
  screens/
    admin/       — Admin/Vorstand views
    auth/        — Login, OIDC callback
    events/      — Event timeline views
    kiosk/       — Unauthenticated kiosk flow
    member/      — Member dashboard
  store/
    app.store.ts  — Zustand global UI state (role view, navigation stack, toasts, branding tweaks)
    auth.store.ts — JWT token storage
  hooks/         — Custom React hooks
  types/         — Shared TypeScript types
```

State management: **Zustand** for global UI state; **TanStack Query** for server state. The Vite dev server proxies `/api/*` to `localhost:8080`.

### Authentication Flow

1. Frontend redirects to `/api/v1/auth/login` → OIDC provider.
2. OIDC callback hits `/api/v1/auth/callback` → backend exchanges code, issues JWT, redirects to `LOGIN_REDIRECT_URL` (default `/auth/callback`) with token in URL fragment.
3. `CallbackPage` stores the JWT; all subsequent requests send `Authorization: Bearer <token>`.
4. `BOOTSTRAP_ADMIN_EMAIL` auto-promotes the first admin on login (one-time bootstrap).

### Roles

Five roles with increasing privilege: `kiosk` → `mitglied` → `veranstaltungsleiter` → `vorstand` → `admin`. Role-based access is enforced via `middleware.RequireRole()` wrappers on chi router groups.

## Configuration

All config is via environment variables (see README for the full table). Key variables: `DATABASE_URL`, `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_REDIRECT_URL`, `JWT_SECRET`.

`TEST_MODE=true` (or `testmode.New()` in Go) disables rate limiting and exposes `/api/v1/dev/token` for issuing test JWTs with any role — never enabled in production.

## Specification (OpenSpec) — mandatory workflow

ShiftManager's functional specification lives under `openspec/`, managed with the [OpenSpec](https://github.com/Fission-AI/OpenSpec) CLI (`@fission-ai/openspec`) — not as free-form Markdown.

**Using OpenSpec for every change is mandatory, not optional**, for every contributor (human or AI):

- **Any change that alters or adds user-facing/system behavior** — new feature, changed business rule, modified API contract, UI behavior change — MUST start from an OpenSpec change under `openspec/changes/<change-name>/` (`proposal.md`, `design.md` when warranted, `tasks.md`, `specs/<capability>/spec.md` deltas) **before** code is written. Do not write the code first and document it afterwards.
- Work through the change with `openspec validate <name> --strict` passing before implementation starts, then implement by following `tasks.md`, checking tasks off as they land.
- Once implementation is complete and tasks are done, run `openspec archive <name>` to fold the delta into `openspec/specs/<capability>/spec.md` and move the change to `openspec/changes/archive/`. A pull request that changes behavior without a corresponding archived (or in-flight) OpenSpec change is incomplete.
- **Exception:** pure refactors, tooling, dependency bumps, and other changes with zero user/system-behavior impact don't need a spec delta — but still go through a change with `skip_specs: true` in `.openspec.yaml` if raised as a formal change, or can be a plain commit when trivial. When in doubt, treat it as behavior-changing and write the spec.
- In Claude Code, use the `/opsx:propose`, `/opsx:apply`, `/opsx:sync`, and `/opsx:archive` skills (installed under `.claude/skills/openspec-*` and `.claude/commands/opsx/`) to work through the propose → specs → design → tasks → apply → archive workflow instead of hand-writing change proposals or skipping straight to code.

Structure:

- `openspec/specs/<capability>/spec.md` — the current, already-implemented behavior, one file per capability (`auth`, `veranstaltungen`, `schichten`, `kiosk`, `branding`, `bedienkonzept`, …), each requirement tagged with its legacy ID from `project/requirements_extracted.txt` (e.g. `V-003`, `SC-010`) where one exists.
- `openspec/changes/<change-name>/` — proposed, not-yet-implemented (or in-progress) work. Archived automatically into `openspec/specs/` once implemented (`openspec archive <name>`).
- `openspec/config.yaml` — project context handed to the AI when drafting artifacts (tech stack, German-language convention, ID-prefix rules).

Useful commands: `openspec list --specs`, `openspec list` (in-flight changes), `openspec show <capability> --type spec`, `openspec validate --all`.

## Devcontainer

A `.devcontainer/` setup provides PostgreSQL and a Dex OIDC provider. E2E tests (Playwright) require the backend running (`make dev-backend`) and PostgreSQL available via the devcontainer compose service.
