.PHONY: help run test test-pkg test-race test-cover lint lint-fix fmt vet tidy \
        build-docker build-docker-local web-dev web-build web-lint web-install \
        compose-up compose-down compose-build-local compose-up-local \
        compose-down-local compose-logs-local compose-smoke compose-restart-local \
        env-stub clean

help:
	@echo "Local development (Docker):"
	@echo "  env-stub              create .env with safe local stubs (idempotent)"
	@echo "  compose-build-local   build all images locally (encore + web)"
	@echo "  compose-up-local      bring up the full local stack"
	@echo "  compose-down-local    tear down the local stack"
	@echo "  compose-logs-local    tail logs from all containers"
	@echo "  compose-smoke         curl every service health endpoint via gateway"
	@echo "  compose-restart-local rebuild + up (in one go)"
	@echo ""
	@echo "Local development (host):"
	@echo "  run                   encore run (all 11 services + dashboard, no Docker)"
	@echo "  web-dev               SvelteKit dev server"
	@echo ""
	@echo "Tests and quality:"
	@echo "  test                  encore test ./..."
	@echo "  test-pkg              go test ./pkg/..."
	@echo "  test-race             encore test -race ./..."
	@echo "  test-cover            encore test -coverprofile=coverage.out ./..."
	@echo "  lint                  golangci-lint run ./..."
	@echo "  lint-fix              golangci-lint run --fix ./..."
	@echo "  fmt                   golangci-lint fmt ./..."
	@echo "  vet                   go vet ./..."
	@echo "  tidy                  go mod tidy"
	@echo ""
	@echo "Production builds:"
	@echo "  build-docker          build per-service images tagged for GHCR"
	@echo "  compose-up            docker compose up using GHCR images (prod)"
	@echo "  compose-down          docker compose down (prod)"
	@echo ""
	@echo "  clean                 remove build artefacts"

run:
	encore run

test:
	encore test ./...

test-pkg:
	go test ./pkg/...

test-race:
	encore test -race ./...

test-cover:
	encore test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

fmt:
	golangci-lint fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

generate: baml-generate
	sqlc generate

generate-check: baml-generate
	sqlc generate
	git diff --exit-code apps

baml-generate:
	@command -v baml-cli >/dev/null 2>&1 || { \
		echo "baml-cli not found. Install: go install github.com/boundaryml/baml/baml-cli@latest"; exit 1; }
	@command -v goimports >/dev/null 2>&1 || { \
		echo "goimports not found. Install: go install golang.org/x/tools/cmd/goimports@latest"; exit 1; }
	cd apps/ai && baml-cli generate

SERVICES := auth tenancy people academics enrollment forms assessment progression imports notifications ai setup
TAG ?= dev

build-docker:
	@for svc in $(SERVICES); do \
		echo ">> building ghcr.io/formswrite/schoolrise-$$svc:$(TAG)"; \
		encore build docker ghcr.io/formswrite/schoolrise-$$svc:$(TAG) --services=$$svc --config=infra-config/selfhost.json; \
	done
	encore build docker ghcr.io/formswrite/schoolrise-gateway:$(TAG) --gateways=api --config=infra-config/selfhost.json

LOCAL_ARCH ?= $(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')

build-docker-local:
	@echo ">> pulling ghcr.io/formswrite/schoolrise-app:latest (built by .github/workflows/build-images.yml)"
	docker pull ghcr.io/formswrite/schoolrise-app:latest

build-web-pull:
	@echo ">> pulling ghcr.io/formswrite/schoolrise-web:latest"
	docker pull ghcr.io/formswrite/schoolrise-web:latest

web-install:
	cd apps/web && npm install --engine-strict=false

web-dev:
	cd apps/web && npm run dev

web-build:
	cd apps/web && npm run build

web-lint:
	cd apps/web && npm run lint

build-web-local: build-web-pull

compose-up:
	docker compose -f deploy/docker-compose.yml up -d

compose-down:
	docker compose -f deploy/docker-compose.yml down

env-stub:
	@if [ -f .env ]; then \
		echo ".env already exists — leaving untouched"; \
	else \
		cp .env.example .env; \
		AUTH_SECRET=$$(openssl rand -hex 32); \
		PG_PASS=$$(openssl rand -hex 16); \
		MINIO_PASS=$$(openssl rand -hex 16); \
		sed -i.bak "s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=$$PG_PASS|" .env; \
		sed -i.bak "s|^ADMIN_EMAIL=.*|ADMIN_EMAIL=admin@local.test|" .env; \
		sed -i.bak "s|^ADMIN_PASSWORD=.*|ADMIN_PASSWORD=ChangeMe123!|" .env; \
		sed -i.bak "s|^AUTH_SECRET=.*|AUTH_SECRET=$$AUTH_SECRET|" .env; \
		sed -i.bak "s|^MINIO_ROOT_PASSWORD=.*|MINIO_ROOT_PASSWORD=$$MINIO_PASS|" .env; \
		sed -i.bak "s|^BASE_URL=.*|BASE_URL=http://localhost:3000|" .env; \
		sed -i.bak "s|^RESEND_API_KEY=.*|RESEND_API_KEY=re_local_stub|" .env; \
		sed -i.bak "s|^EMAIL_FROM=.*|EMAIL_FROM=schoolrise@local.test|" .env; \
		sed -i.bak "s|^OPENAI_API_KEY=.*|OPENAI_API_KEY=sk-local-stub|" .env; \
		rm -f .env.bak; \
		echo ".env created with local-dev stubs."; \
		echo "WARNING: RESEND_API_KEY and OPENAI_API_KEY are stubs — replace before exercising email/AI features."; \
	fi

compose-build-local: build-docker-local build-web-local

compose-up-local: env-stub
	docker compose --env-file .env -f deploy/docker-compose.local.yml up -d
	@echo ""
	@echo "Stack is up. Endpoints:"
	@echo "  web:      http://localhost:3001"
	@echo "  gateway:  http://localhost:8080"
	@echo "  postgres: localhost:5432 (user: schoolrise, db: schoolrise)"
	@echo ""
	@echo "Run 'make compose-smoke' to verify all 11 services."

compose-down-local:
	docker compose --env-file .env -f deploy/docker-compose.local.yml down

compose-logs-local:
	docker compose --env-file .env -f deploy/docker-compose.local.yml logs -f --tail=200

compose-restart-local: compose-down-local compose-build-local compose-up-local

compose-smoke:
	@echo "Smoke-testing health endpoints via gateway..."
	@for svc in $(SERVICES); do \
		printf "  %-15s " "$$svc"; \
		code=$$(curl -s -o /dev/null -w '%{http_code}' "http://localhost:8080/v1/$$svc/health"); \
		body=$$(curl -s "http://localhost:8080/v1/$$svc/health"); \
		echo "HTTP $$code  $$body"; \
	done

e2e:
	@echo "Running Playwright e2e suite (assumes stack is up)..."
	cd apps/web && npm run test:e2e

e2e-personas:
	@echo "Running multi-persona Playwright suite (admin, teacher, inspector, student)..."
	cd apps/web && npx playwright test 05-personas

e2e-report:
	cd apps/web && npm run test:e2e:report

e2e-install:
	cd apps/web && npm run test:e2e:install

clean:
	rm -rf bin build coverage .encore
	rm -rf apps/web/.svelte-kit apps/web/build apps/web/.vite

# =====================================================================
# Production lifecycle (multi-compose, single-host TLS via Caddy)
# =====================================================================

prod-network:
	@docker network create schoolrise-net 2>/dev/null || true

prod-up: prod-network
	docker compose --env-file .env.production -f deploy/compose/postgres.yml up -d
	docker compose --env-file .env.production -f deploy/compose/minio.yml up -d
	docker compose --env-file .env.production -f deploy/compose/app.yml up -d
	docker compose --env-file .env.production -f deploy/compose/web.yml up -d
	docker compose --env-file .env.production -f deploy/compose/caddy.yml up -d

prod-down:
	-docker compose -f deploy/compose/caddy.yml down
	-docker compose -f deploy/compose/web.yml down
	-docker compose -f deploy/compose/app.yml down
	-docker compose -f deploy/compose/minio.yml down
	-docker compose -f deploy/compose/postgres.yml down

prod-logs:
	docker compose -f deploy/compose/caddy.yml logs --tail 50 -f caddy

prod-smoke-http: prod-network
	docker compose --env-file .env.production -f deploy/compose/postgres.yml up -d
	docker compose --env-file .env.production -f deploy/compose/minio.yml up -d
	docker compose --env-file .env.production -f deploy/compose/app.yml up -d
	docker compose --env-file .env.production -f deploy/compose/web.yml up -d
	@echo "--- network attached ---"
	@docker network inspect schoolrise-net --format '{{range .Containers}}{{.Name}} {{end}}'
	@echo "--- web reachable from inside the network ---"
	@docker run --rm --network schoolrise-net curlimages/curl:latest -sI http://web:3000/login | head -3

# =====================================================================
# Test split — pure unit tests (pkg/) vs Encore service tests (need DB)
# =====================================================================

test-unit:
	go test ./pkg/... -race -shuffle=on

test-integration:
	@if [ -z "$$AUTH_SECRET" ]; then \
		echo "AUTH_SECRET must be set; export AUTH_SECRET=$$(openssl rand -base64 32)"; exit 1; \
	fi
	encore test ./apps/... -race -shuffle=on -count=1
