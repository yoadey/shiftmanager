# ShiftManager Helm Chart

Deploys the ShiftManager application (Go backend + embedded React SPA, single
image) to Kubernetes, with **optional bundled PostgreSQL and Redis** (Bitnami
subcharts) for a turnkey install.

## TL;DR

```bash
# from the repo root
helm dependency update ./deploy/helm/shiftmanager

helm upgrade --install shiftmanager ./deploy/helm/shiftmanager \
  --namespace shiftmanager --create-namespace \
  --set app.jwtSecret="$(openssl rand -hex 32)" \
  --set postgresql.auth.password="$(openssl rand -hex 16)" \
  --set app.bootstrapAdminEmail="vorstand@mein-verein.de" \
  --set app.oidc.issuer="https://id.example/realms/verein" \
  --set app.oidc.clientId="shiftmanager" \
  --set app.oidc.clientSecret="••••" \
  --set app.baseUrl="https://schichten.example.de" \
  --set app.oidc.redirectUrl="https://schichten.example.de/api/v1/auth/callback"
```

> `helm dependency update` requires network access to
> `oci://registry-1.docker.io/bitnamicharts`. In air-gapped clusters, vendor the
> subchart `.tgz` files into `charts/` or disable them (see below).

## Architecture

- **One application image** serves the API under `/api/v1` and the SPA for all
  other paths. Container port **8080**; probes `/health` (liveness) and
  `/readyz` (readiness).
- **Migrations** run as an init container (`/shiftmanager migrate`) before the
  app starts. Toggle with `migration.enabled`.
- **Uploads** (club logos, B-004) are written to `app.uploadDir` and persisted
  via a PVC (`persistence.*`).
- **Config** is split into a ConfigMap (non-secret env) and a Secret (JWT,
  DB/Redis URLs, OIDC secret, SMTP password). Bring your own with
  `existingSecret`.

## Database & Redis options

| Scenario | Settings |
|---|---|
| Bundled Postgres (default) | `postgresql.enabled=true`, set `postgresql.auth.password` |
| External Postgres | `postgresql.enabled=false`, set `database.external.url` (or provide `DATABASE_URL` via `existingSecret`) |
| Bundled Redis | `redis.enabled=true`, set `redis.auth.password` |
| External Redis | `redis.enabled=false`, set `redis.external.url` |
| No Redis (in-memory cache) | `redis.enabled=false`, leave `redis.external.url` empty |

When the bundled charts are enabled, `DATABASE_URL` / `REDIS_URL` are derived
automatically from `postgresql.auth.*` / `redis.auth.*`. When you use
`existingSecret`, the chart does **not** derive them — provide them in your
Secret.

## Secrets

Either let the chart create the Secret from values (`app.jwtSecret`,
`app.oidc.clientSecret`, `app.smtp.password`, derived `DATABASE_URL`/`REDIS_URL`),
or manage it yourself:

```bash
kubectl create secret generic shiftmanager-secrets -n shiftmanager \
  --from-literal=JWT_SECRET="$(openssl rand -hex 32)" \
  --from-literal=DATABASE_URL="postgres://user:pass@db:5432/shiftmanager?sslmode=require" \
  --from-literal=OIDC_CLIENT_SECRET="••••" \
  --from-literal=SMTP_PASS="••••" \
  --from-literal=REDIS_URL="redis://:pass@redis:6379/0"   # optional
# then: --set existingSecret=shiftmanager-secrets
```

## Key parameters

| Key | Default | Description |
|---|---|---|
| `replicaCount` | `2` | App replicas (ignored if autoscaling) |
| `image.repository` / `image.tag` | `ghcr.io/yoadey/shiftmanager` / chart appVersion | Container image |
| `app.baseUrl` | `http://localhost:8080` | Public URL (e-mail links) |
| `app.jwtSecret` | `""` (**required**) | JWT signing secret (≥32 chars) |
| `app.bootstrapAdminEmail` | `""` | First-admin bootstrap e-mail |
| `app.oidc.*` | – | Issuer, clientId, clientSecret, redirectUrl, loginRedirectUrl |
| `app.smtp.*` | – | host/port/user/password/from |
| `existingSecret` | `""` | Use an externally managed Secret |
| `postgresql.enabled` | `true` | Bundle Bitnami PostgreSQL |
| `database.external.url` | `""` | DSN when bundled Postgres is off |
| `redis.enabled` | `false` | Bundle Bitnami Redis |
| `redis.external.url` | `""` | Redis DSN when bundled Redis is off |
| `persistence.enabled` / `persistence.size` | `true` / `1Gi` | Uploads PVC |
| `ingress.enabled` | `false` | Create an Ingress |
| `autoscaling.enabled` | `false` | HorizontalPodAutoscaler |
| `podDisruptionBudget.enabled` | `true` | PDB (`minAvailable: 1`) |
| `networkPolicy.enabled` | `false` | Restrict pod traffic |
| `backup.enabled` | `false` | Daily `pg_dump` CronJob |
| `migration.enabled` | `true` | Run migrations as init container |

See [`values.yaml`](./values.yaml) for the full, commented list and
[`values-production.yaml`](./values-production.yaml) for a production example.

## Validate, test, upgrade

```bash
helm lint ./deploy/helm/shiftmanager
helm template t ./deploy/helm/shiftmanager --set app.jwtSecret=x | kubectl apply --dry-run=client -f -
helm test shiftmanager -n shiftmanager        # runs /health, /readyz, openapi.json checks
helm upgrade shiftmanager ./deploy/helm/shiftmanager -n shiftmanager -f values-production.yaml
```

## Notes on the Bitnami subcharts

Bitnami has been restructuring its free image catalogue. If the bundled
PostgreSQL/Redis pods fail to pull images, either pin working image coordinates
(e.g. `--set postgresql.image.repository=bitnamilegacy/postgresql`) or, for
production, prefer an external managed database and set `postgresql.enabled=false`.
