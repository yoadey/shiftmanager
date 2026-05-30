# ============================================================
# ShiftManager — Multi-stage Dockerfile
# Stage 1: Build React/Vite frontend
# Stage 2: Build Go binary (with embedded frontend assets)
# Stage 3: Minimal distroless runtime image
# ============================================================

# ---- Stage 1: Frontend build --------------------------------
FROM node:20-alpine AS frontend-builder

WORKDIR /app

# Copy only package files first for better layer caching
COPY frontend/package.json frontend/package-lock.json* ./frontend/
RUN cd frontend && npm ci

# Copy the rest of the frontend source and build
COPY frontend/ ./frontend/
RUN cd frontend && npm run build
# Output: /app/frontend/dist


# ---- Stage 2: Go build --------------------------------------
FROM golang:1.22-alpine AS go-builder

# Install git so `go mod download` can fetch VCS-tagged modules
RUN apk add --no-cache git

WORKDIR /app

# Download dependencies first (cached unless go.mod/go.sum change)
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire Go source tree
COPY cmd/       ./cmd/
COPY internal/  ./internal/
COPY migrations/ ./migrations/

# Embed the compiled frontend into the Go source tree so embed.FS picks it up.
# The path must match the //go:embed directive in the source code.
COPY --from=frontend-builder /app/frontend/dist ./internal/infrastructure/static/dist

# Build a statically linked binary; strip debug info to reduce size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -ldflags="-s -w" \
      -trimpath \
      -o /shiftmanager \
      ./cmd/server


# ---- Stage 3: Distroless runtime ----------------------------
# gcr.io/distroless/static-debian12 contains no shell, no package manager,
# only CA certs and timezone data — minimal attack surface.
FROM gcr.io/distroless/static-debian12 AS runtime

# Copy the self-contained binary
COPY --from=go-builder /shiftmanager /shiftmanager

EXPOSE 8080

# Run as the built-in nonroot user (uid 65532) provided by distroless
USER nonroot:nonroot

ENTRYPOINT ["/shiftmanager"]
