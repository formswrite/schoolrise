# SchoolRise — production deployment

Single-host TLS topology with multi-compose lifecycle.

## Topology

- **One public hostname** (`<DOMAIN>`). No subdomains.
- **MinIO is internal-only** — never exposed to the internet. All file fetches go through SvelteKit's `/api/uploads/[key]` proxy.
- **Five docker-compose projects**, joined by one external network (`schoolrise-net`):
  - `postgres.yml` — postgres + pgbouncer
  - `minio.yml` — MinIO + bucket-init (private bucket)
  - `app.yml` — Encore Go service
  - `web.yml` — SvelteKit
  - `caddy.yml` — TLS edge (Let's Encrypt + path routing)

## 1. DNS

Point ONE A record at your server's public IP:

| Record | Type | Value |
|---|---|---|
| `<DOMAIN>` (e.g. `schoolrise`) | A | server-IP |

Verify: `dig +short <DOMAIN>` returns your server IP.

## 2. Server prep (Ubuntu 22.04+ / Debian 12+)

```bash
# Docker + Compose
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER  # log out + back in

# Open ports
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # ACME HTTP-01 challenge + redirect
sudo ufw allow 443/tcp   # HTTPS
sudo ufw allow 443/udp   # HTTP/3 (QUIC)
sudo ufw enable
```

## 3. Repo + secrets

```bash
git clone https://github.com/<org>/school-rise.git
cd school-rise
cp .env.production.example .env.production
# Edit .env.production — fill DOMAIN, LE_EMAIL, all __GENERATE...__ secrets
chmod 600 .env.production
```

Generate strong secrets: `openssl rand -base64 32`

## 4. Build the app images

The Encore app image requires a custom build (BAML cgo + patched Encore CLI). Two options:

**Option A — pull from CI** (preferred once GitHub Actions is wired):
```bash
docker pull ghcr.io/<org>/schoolrise-app:patched
docker pull ghcr.io/<org>/schoolrise-web:local
```

**Option B — build on the server**:
```bash
docker build -t schoolrise-web:local apps/web
# schoolrise-app:patched is the encore-go yak shave; see CONTRIBUTING.md
```

## 5. First boot

```bash
make prod-up
```

This runs, in order:
1. `docker network create schoolrise-net` (idempotent)
2. `docker compose -f deploy/compose/postgres.yml up -d`
3. `docker compose -f deploy/compose/minio.yml up -d`
4. `docker compose -f deploy/compose/app.yml up -d`
5. `docker compose -f deploy/compose/web.yml up -d`
6. `docker compose -f deploy/compose/caddy.yml up -d`

Watch Caddy obtain the cert (~30 s the first time):
```bash
make prod-logs
# Look for: "certificate obtained successfully"
```

Check service health:
```bash
docker ps | grep schoolrise-
curl -I https://<DOMAIN>/login
```

## 6. Verify

| URL | Expected |
|---|---|
| `https://<DOMAIN>/login` | SchoolRise login page over TLS |
| `https://<DOMAIN>/admin/dashboard` | redirects to /login if not authenticated |
| `https://<DOMAIN>/v1/forms/field-types` (with `Authorization: Bearer ...`) | JSON of registered field types |
| `https://<DOMAIN>/api/uploads/<key>` | streams the MinIO object via the proxy |

Quick TLS check:
```bash
curl -I https://<DOMAIN>/login | grep -i strict-transport
# Expect: strict-transport-security: max-age=15552000; includeSubDomains; preload
```

## 7. MinIO console (admin only)

The MinIO web console is **not** publicly exposed. To reach it for ops tasks:

```bash
# From your laptop:
ssh -L 9001:schoolrise-minio:9001 user@<server>
# Then visit http://localhost:9001 in your browser.
# Login with MINIO_ROOT_USER / MINIO_ROOT_PASSWORD from .env.production.
```

## 8. Cert renewal

Caddy renews 30 days before expiry, automatically. No cron needed. Verify:

```bash
docker compose -f deploy/compose/caddy.yml logs caddy --since 24h | grep -i renew
```

## 9. Rollback

```bash
make prod-down                    # stops all 5 projects in reverse order
docker tag schoolrise-web:previous schoolrise-web:local
make prod-up
```

Data persists in named volumes (`pgdata-prod`, `minio-data-prod`, `caddy-data`, `caddy-config`) across `down`/`up`.

## 10. Lifecycle commands (Makefile)

| Target | What it does |
|---|---|
| `make prod-network` | creates `schoolrise-net` (idempotent) |
| `make prod-up` | brings up all 5 services in order |
| `make prod-down` | tears down in reverse order; volumes preserved |
| `make prod-logs` | tails Caddy logs |
| `make prod-smoke-http` | brings up the 4 backend services on schoolrise-net WITHOUT Caddy and verifies network plumbing — used during dev to sanity-check the multi-compose split |

## 11. Adding more projects to the same MinIO

Other apps on the same box (or different compose projects) can share this MinIO instance by joining `schoolrise-net`:

```yaml
services:
  other-app:
    networks: [schoolrise-net]
networks:
  schoolrise-net:
    external: true
```

They reach MinIO at `http://minio:9000` with the same `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD`. Each app should use a separate bucket.

## 12. What this gives you

- ✅ TLS on `<DOMAIN>` (Let's Encrypt, auto-renewed)
- ✅ HSTS, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, Permissions-Policy on every response
- ✅ HTTP/2 + HTTP/3 (QUIC)
- ✅ Path-based API routing (`/v1/*` → encore, everything else → SvelteKit)
- ✅ Private MinIO (no public bucket, no anonymous downloads)
- ✅ Independently lifecycled services (you can restart MinIO without touching Encore, etc.)

## 13. What this does NOT yet give you

These are the next blockers from the production roadmap:
- 🚧 Backups for postgres + MinIO
- 🚧 First-boot password rotation flow
- 🚧 Secret store (currently file-based; should move to Vault/SSM/1Password)
- 🚧 Error tracking (Sentry)
- 🚧 Log aggregation (Loki)
- 🚧 Uptime monitoring
