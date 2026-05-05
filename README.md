# SchoolRise

[![CI](https://github.com/formswrite/schoolrise/actions/workflows/test.yml/badge.svg)](https://github.com/formswrite/schoolrise/actions/workflows/test.yml)
[![Security](https://github.com/formswrite/schoolrise/actions/workflows/security.yml/badge.svg)](https://github.com/formswrite/schoolrise/actions/workflows/security.yml)
[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)

**The AI-native, open-source EMIS for ministries of education** — built around a real form builder with conditional logic, rich question types, and LLM-assisted item authoring + auto-grading. Self-host with one `make` command.

SchoolRise is an **Education Management Information System** (the same category as OpenEMIS, DHIS2-Education, and most country-built EMIS stacks) but with two structural differences: a first-class **assessment-authoring layer** (form builder, campaigns, scoped delivery, snapshot dashboards) and a **first-class AI layer** (LLM contracts via [BAML](https://github.com/BoundaryML/baml) for item generation, distractor synthesis, rubric drafting, and essay grading).

Designed for ministry-scale deployments: tens of thousands of primary schools, millions of students, multiple administrative tiers, locale-agnostic UI. Hierarchy levels and locale strings are seed data — adopters fork and configure without touching code.

> ⚠️ **Status: pre-release.** The form editor (drag-reorder, click-to-edit, logic rules, 31 of 32 field types, MinIO file uploads, signature capture, KaTeX equations) is shipped and covered by 34 e2e tests. Production deployment is in roll-out. See [What works today](#what-works-today) below.

---

## What it does

A teacher creates a reading assessment with conditional logic ("show passage 2 only if passage 1 was scored ≥ 60"). A regional inspector publishes it to thousands of schools. Students respond on tablets with file uploads + signature capture. A snapshot system aggregates millions of rows of scoring into sub-50 ms region-level dashboards for the minister.

The platform handles:

| Surface | What's there |
|---|---|
| **Form editor** | 3-panel layout (palette / canvas / settings drawer), drag-reorder, click-to-edit, inline previews for 31 field types, conditional show/hide rules, validation (regex, length, range), grading (auto + rubric), per-question translations |
| **Public renderer** | All 31 types render natively (KaTeX equations, drag-ordering, fill-in-blank with `[[N]]` cloze, matching pairs, table grids, country-region cascades, signature pad, file upload with image preview) |
| **Assessment data flow** | Form versioning + immutable snapshots, campaign assignment per scope, score aggregation into precomputed snapshots (1.2 s for 101 K rows) |
| **Hierarchy + people** | Country → region → district → school closure tables, ~30 K teachers, 4 M+ students, role-scoped access (admin / inspector / teacher) |
| **File storage** | MinIO via SvelteKit proxy — uploads stored privately, served same-origin through `/api/uploads/[key]` |
| **AI** | LLM contracts via [BAML](https://github.com/BoundaryML/baml): inspectors author items via natural language, draft rubrics from scale codes, generate distractors for multiple-choice, and auto-grade essays against a rubric. Provider-agnostic (OpenAI, Anthropic, or local models). |
| **Deployment** | Multi-compose topology (postgres / minio / app / web / caddy each its own project), Caddy + Let's Encrypt edge, single A record, no public MinIO exposure |

---

## 60-second local quickstart

Prerequisites: Docker Desktop (or compatible), `make`.

```bash
git clone https://github.com/formswrite/schoolrise.git
cd schoolrise
cp .env.example .env
# edit .env — fill in the four __choose_a_local_*__ placeholders
make compose-up-local
```

Open <http://localhost:3001/login> and sign in with the `ADMIN_EMAIL` / `ADMIN_PASSWORD` you set in `.env`. The form editor lives at `/admin/forms`.

To tear down: `make compose-down-local`.

---

## Architecture, in one diagram

```
                        Browser (admin / public respondent)
                                       │
                                       ▼
                            ┌────────────────────┐
                            │   Caddy (TLS)      │   ← Let's Encrypt, HTTP/2+3
                            │   single hostname  │      HSTS + CSP headers
                            └─────────┬──────────┘
                                      │
                  ┌───────────────────┴───────────────────┐
                  │                                       │
                  ▼ /v1/*                                 ▼ /*
        ┌──────────────────┐                  ┌──────────────────┐
        │ Encore (Go)      │                  │ SvelteKit (Node) │
        │ 12 microservices │                  │ admin + /r/[token]│
        │ ─ auth           │                  │ + /api/uploads    │
        │ ─ tenancy        │                  └────────┬─────────┘
        │ ─ people         │                           │
        │ ─ academics      │                           │
        │ ─ enrollment     │                           │
        │ ─ forms          │                           │
        │ ─ assessment     │                           │
        │ ─ progression    │                           │
        │ ─ ai · imports   │                           │
        │ ─ notifications  │                           │
        │ ─ setup          │                           │
        └────────┬─────────┘                           │
                 │                                     │
                 ▼                                     ▼
            ┌─────────┐                          ┌──────────┐
            │PgBouncer│                          │  MinIO   │  ← S3-compatible,
            │+Postgres│                          │ (private)│     internal-only
            └─────────┘                          └──────────┘
```

Detailed walkthrough: **[docs/architecture.md](docs/architecture.md)**.

---

## What works today

### ✅ Shipped + tested (34 e2e tests, type-clean)

- **Form editor** — Phase 1 (3-panel + drag/click/preview), Phase 2 (logic rules), Phase 3 (validation/grading + 12 rich types end-to-end including KaTeX, FILL_IN_BLANK cloze, MATCHING, ORDERING)
- **MinIO uploads** — `/api/uploads` POST + same-origin proxy + `<FileUploadInput>` + canvas-based `<SignaturePad>` (Phase 4u)
- **Snapshot-based progression dashboard** — 4.3 M-row aggregation in ~1.2 s
- **Multi-compose production stack** — `make prod-up` brings up 5 single-concern compose projects on a shared external Docker network
- **TLS edge** — Caddy + Let's Encrypt, single hostname, no fake DNS
- **Realistic example seed** — 8 regions, 46 districts, ~21 K schools, ~4 M students, age-appropriate DOBs, locale-realistic teacher names

### 🚧 In progress / known gaps

- **Auto-grading wiring for ORDERING/MATCHING/FILL_IN_BLANK/EQUATION** — editor + renderer + DB are wired; the assessment-scoring pipeline still needs to evaluate these types
- **First-boot password rotation** — bootstrap flow exists but needs polish
- **Backups** — no automated pg_dump / MinIO mirror yet
- **i18n** — French content + UI chrome translation switcher (Phase 4) not yet built
- **HOTSPOT** — single field type still placeholder

See [docs/roadmap.md](docs/roadmap.md) for the prioritized list.

---

## How SchoolRise differs from existing EMIS systems

The dominant open-source EMIS today is **[OpenEMIS](https://www.openemis.org)** (UNESCO + Community Systems Foundation, used by 17+ ministries). It does EMIS records — schools, students, staff, infrastructure, finance — well. But there's **no form authoring, no assessment campaigns, no AI assist**: items, rubrics, and dashboards are static templates.

SchoolRise targets the same buyer (ministries of education) but a different surface:

| Capability | OpenEMIS | SchoolRise |
|---|---|---|
| School / student / staff records | ✅ mature, 17 years of refinement | ✅ |
| Form authoring with conditional logic + 30+ question types | ❌ static reports only | ✅ drag-reorder editor with show/hide rules |
| Multi-million-row dashboards | ⚠️ scales by hardware | ✅ snapshot-based aggregation, 1.2 s for 101 K rows |
| AI-assisted item generation, distractor synthesis, essay grading | ❌ not in scope | ✅ LLM contracts via BAML, provider-agnostic |
| Stack | PHP + MySQL + CakePHP | Go + Encore + SvelteKit + PostgreSQL |
| Self-host quickstart | "see the knowledge base" | `make compose-up-local` |
| License | GPL-2.0 | AGPL-3.0 |

We're not trying to replace OpenEMIS's records modules; many ministries already use them. SchoolRise is the **assessment-and-AI layer** that gov-tech teams have been building from scratch in spreadsheets and one-off PHP forms because nothing in the EMIS category provides it.

---

## Production deployment

Single-host single-domain Caddy + multi-compose topology. One DNS A record, one `make prod-up`, automatic Let's Encrypt cert.

→ **[deploy/README-prod.md](deploy/README-prod.md)**

---

## Tech stack

| Layer | Choice | Why |
|---|---|---|
| Backend | [Encore.go](https://encore.dev) | 12 microservices with auto-generated tracing + DB clients + service-to-service RPC |
| Database | PostgreSQL 16 + PgBouncer | Tx-pooling, ~400 connections handle 4.3 M-row dashboards |
| Frontend | SvelteKit 2 + Svelte 5 (runes) | Server-rendered admin + public renderer, type-safe `+page.server.ts` |
| UI | Tailwind 4 + bits-ui (shadcn-svelte) | No design-system lock-in |
| Object storage | MinIO (S3-compatible) | Self-host parity with AWS S3, swap envs for prod |
| Edge proxy | [Caddy 2](https://caddyserver.com) | One-line Let's Encrypt, HTTP/3 by default |
| AI | [BAML](https://github.com/BoundaryML/baml) | Type-safe LLM contracts: `SuggestItems`, `DraftRubric`, `GenerateDistractors`, `GradeEssay`. Provider-agnostic (OpenAI, Anthropic, local). Stub-mode for tests. Job log persisted with token + latency metrics per call. |
| Tests | Playwright + Go test + svelte-check | 34 e2e + ~50 Go tests covering services |
| License | [AGPL-3.0](LICENSE) | Strong copyleft for ministries who need source-available guarantees |

---

## Repository layout

```
.
├── apps/
│   ├── auth/ tenancy/ people/ academics/ enrollment/      # 12 Encore services
│   ├── forms/ assessment/ progression/ imports/
│   ├── notifications/ ai/ setup/
│   └── web/                                                # SvelteKit
├── pkg/                          # shared Go: domain, hierarchy, seed, …
├── deploy/
│   ├── docker-compose.local.yml  # all-in-one local-dev stack
│   ├── compose/                  # 5 single-concern prod compose projects
│   └── Caddyfile                 # TLS edge config
├── infra-config/                 # Encore infra templates (selfhost / gcp / aws)
├── docs/
│   ├── architecture.md           # one-page system overview
│   ├── operations/install.md     # admin install handbook
│   └── encore-go-build.md        # custom Encore fork rebuild steps
├── examples/seed-template/       # realistic seed dataset template
└── .github/workflows/            # CI: test, lint, security, release
```

---

## Contributing

We welcome contributors who care about education infrastructure or self-hostable government tools.

1. Read **[CONTRIBUTING.md](CONTRIBUTING.md)** — covers the dev-container approach, the encore-go fork build, the test-required policy, and AGPL implications.
2. Pick an issue tagged `good-first-issue` or `help-wanted`.
3. Fork, branch, commit, open a PR. CI must be green: `lint`, `test`, `security`.

Code of conduct: **[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)**. Security disclosures: **[SECURITY.md](SECURITY.md)**.

---

## License

AGPL-3.0. If you deploy SchoolRise as a network service, the AGPL requires you to make your source modifications available to the users of that service. See [LICENSE](LICENSE).

For ministries or organizations that need a different license arrangement (e.g. proprietary integrations, OEM redistribution), contact the maintainers.

---

## Acknowledgements

Built on the shoulders of:
- [Encore.dev](https://encore.dev) for the backend framework
- [SvelteKit](https://kit.svelte.dev) for the frontend
- [MinIO](https://min.io) for object storage
- [Caddy](https://caddyserver.com) for the TLS edge
- [BoundaryML / BAML](https://github.com/BoundaryML/baml) for type-safe LLM contracts
- The teams behind [FormSG](https://github.com/opengovsg/FormSG) and [Kolibri](https://kolibri.learningequality.org) — public-good education tools that informed our design choices
