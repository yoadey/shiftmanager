#!/bin/bash
set -e

echo "==> Installing frontend dependencies..."
if [ -f /workspace/frontend/package.json ]; then
    cd /workspace/frontend && npm install
fi

echo "==> Downloading Go modules..."
cd /workspace && go mod download

echo "==> Creating .env file..."
if [ ! -f /workspace/.env ]; then
    cat > /workspace/.env << 'EOF'
# Database
DATABASE_URL=postgres://postgres:postgres@postgres:5432/shiftmanager?sslmode=disable

# Redis
REDIS_URL=redis://redis:6379/0

# JWT
JWT_SECRET=dev-secret-change-in-production
JWT_EXPIRATION=24h

# Server
PORT=8080
BASE_URL=http://localhost:8080
LOG_LEVEL=debug

# OIDC (optional)
# OIDC_ISSUER=
# OIDC_CLIENT_ID=
# OIDC_CLIENT_SECRET=
# OIDC_REDIRECT_URL=http://localhost:8080/api/v1/auth/callback

# SMTP (optional)
# SMTP_HOST=localhost
# SMTP_PORT=587
# SMTP_USER=
# SMTP_PASS=
# SMTP_FROM=noreply@shiftmanager.local
EOF
fi

# postgres is guaranteed healthy at this point via depends_on condition: service_healthy
echo "==> Running database migrations..."
cd /workspace
go run cmd/server/main.go migrate 2>&1 \
    || echo "Warning: migrations skipped. Run manually: go run cmd/server/main.go migrate"

echo ""
echo "==> Dev container ready!"
echo "    Backend:  go run cmd/server/main.go"
echo "    Frontend: cd frontend && npm run dev"
