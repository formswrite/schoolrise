# Installation

This document is for ministry IT operators deploying SchoolRise on their own infrastructure (a VPS, an on-prem server, or a cloud VM).

There are two install paths:

- **Web installer** (recommended) — a guided 8-step wizard in your browser. Use this for production ministry deployments.
- **Headless** — env-var-driven, no UI. Use this for CI, automated provisioning, smoke tests.

## Prerequisites

- Linux server with Docker 24+ and Docker Compose v2
- Public DNS name pointing at the server (for production; localhost is fine for evaluation)
- Outbound HTTPS to `api.resend.com` (email) and `api.openai.com` (AI), unless you bring your own SMTP relay and self-hosted LLM
- ~4 GB RAM, 2 vCPU minimum for an evaluation deploy

## Web installer (recommended)

### 1. Clone the repository

```bash
git clone https://github.com/formswrite/schoolrise.git
cd schoolrise
```

### 2. Generate a minimal `.env`

```bash
make env-stub
```

This creates `.env` with two values you need before any container can start:

| Variable | Why |
|---|---|
| `POSTGRES_PASSWORD` | Random 32-byte password for the bundled Postgres container. |
| `AUTH_SECRET` | Random 32-byte hex used to sign sessions and HMAC the install token. |

You do **not** need to set anything else in `.env` for the web installer path. All other settings (admin account, integrations, hierarchy, etc.) are configured in the wizard.

### 3. Start the stack

```bash
docker compose -f deploy/docker-compose.yml up -d
```

This starts 13 containers: Postgres, the Encore application (12 services + gateway), and the SvelteKit web frontend.

### 4. Retrieve the install token

On first boot, the `setup` service generates a single-use install token and prints it to stderr:

```bash
docker compose logs app | grep -A 2 "install token"
```

You'll see:

```
SchoolRise install token (one-time, single-use):
    7q3X9L2pVkN5sDtY4RmH8WaQbZ6cFePj
The first visitor to /setup with this token claims the admin account.
```

Copy the token. **Treat it like a password** — anyone with it can claim the admin account.

### 5. Open the wizard

Visit your server's URL in a browser. SchoolRise will redirect you to `/setup/welcome`. Walk through the 8 steps:

| # | Step | What you set |
|---|---|---|
| 1 | Welcome | UI language (English / French) |
| 2 | Unlock | Paste the install token from step 4 |
| 3 | Admin | Email, full name, password for the initial administrator |
| 4 | System | Instance name, default locale, public BASE URL, time zone |
| 5 | Levels | Define your administrative hierarchy (e.g., `region → prefecture → school` or `district → school → class`) |
| 6 | Schools (optional) | Bulk-import schools from CSV — `parent_code,level,code,label` per row |
| 7 | Integrations (optional) | Resend, OpenAI, Anthropic, S3 keys + SMTP relay |
| 8 | Review | Confirm and **Finalize** |

**Resume on restart**: each step is persisted server-side. You can quit halfway through, restart the container, and pick up where you left off. The install-token cookie session expires after 30 minutes of inactivity — re-paste the token and continue.

### 6. After Finalize

The wizard redirects you to `/login`. Sign in with the admin credentials you chose in step 3. The `/setup/*` routes are now permanently locked (HTTP 410).

## Headless install (for CI / automated provisioning)

Skip the wizard entirely:

```bash
SCHOOLRISE_HEADLESS=1 \
ADMIN_EMAIL=admin@example.gov \
ADMIN_PASSWORD="$(openssl rand -base64 24)" \
ADMIN_FULL_NAME="Initial Administrator" \
INSTANCE_NAME="My Ministry" \
DEFAULT_LOCALE=en \
BASE_URL=https://schoolrise.example.gov \
TIME_ZONE=UTC \
COUNTRY_PACK=TEMPLATE \
docker compose -f deploy/docker-compose.yml up -d
```

The `setup` service detects `SCHOOLRISE_HEADLESS=1` and runs the same domain code paths the wizard does, then sets `setup_complete_at`. All `/v1/setup/*` endpoints are locked from this point.

`COUNTRY_PACK` references a JSON pack from `pkg/seed/countries/` (only `TEMPLATE` ships by default — adds no levels). To preseed levels in headless mode, drop a JSON file in that directory and rebuild the image.

`SCHOOLRISE_HEADLESS=1` is a no-op once `setup_complete_at` is set. The container warn-logs and continues normally.

## Recovery — re-running the wizard

The wizard is single-shot by design. To re-run it, a database administrator must explicitly reset the setup state:

```bash
docker compose exec postgres psql -U schoolrise -d setup -c \
  "UPDATE setup_state SET setup_complete_at = NULL, install_token_hash = NULL, install_token_consumed_at = NULL;"
docker compose restart app
```

A new install token will be printed to the logs. The wizard reopens. Existing data (admin user, schools, etc.) is **not** wiped — only the lockout flag is cleared.

To wipe everything and start clean:

```bash
docker compose down
docker volume rm <project>_pgdata
docker compose up -d
```

## Environment variables reference

| Variable | Required | Default | Purpose |
|---|---|---|---|
| `POSTGRES_PASSWORD` | yes | (random via `make env-stub`) | Postgres admin password |
| `AUTH_SECRET` | yes | (random via `make env-stub`) | Signs sessions, HMACs install token |
| `TAG` | no | `latest` | Image tag (pin to a release like `v0.1.0`) |
| `SCHOOLRISE_HEADLESS` | no | unset | Set to `1` to use headless install |
| `ADMIN_EMAIL` | headless only | — | Initial admin email |
| `ADMIN_PASSWORD` | headless only | — | Initial admin password |
| `ADMIN_FULL_NAME` | no | `Initial Administrator` | Admin display name |
| `INSTANCE_NAME` | no | `SchoolRise` | Instance display name |
| `DEFAULT_LOCALE` | no | `en` | UI default locale (`en`/`fr`) |
| `BASE_URL` | no | `http://localhost:3000` | Public URL for assignment-link emails |
| `TIME_ZONE` | no | `UTC` | IANA timezone |
| `COUNTRY_PACK` | headless only | (none) | Preseed hierarchy levels from a country pack code (`TEMPLATE` ships by default) |
| `SENTRY_DSN` | no | unset | Optional error reporting |
| `PROMETHEUS_SCRAPE_TOKEN` | no | unset | Optional metrics auth token |

Integration secrets (Resend, OpenAI, S3, SMTP) are no longer configured via `.env`. They live in `system_settings` after the wizard runs (or admin sets them later in `/admin/settings`).
