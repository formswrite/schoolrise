# SchoolRise — Production Architecture PRD (v3, Encore.go)

> **Status.** Customer-scoped to the 6-section French specification. Backend rebuilt around **Encore.go** (Go, MPL-2.0, 11.9k★, active April 2026). One Encore application starting as a single service, with package-per-domain structure that lets us split into multiple services later without a rewrite. Two-container default deploy: Encore-built app image + Postgres. Frontend is Next.js 15. Designed to be open-sourced and self-hosted by ministries with 1–3 person IT teams.

---

## 1. Context

**Customer.** Republic of Guinea — Ministère de l'Éducation Nationale et de l'Alphabétisation. The customer's French specification describes a **plateforme de gestion scolaire** — administrative + pedagogical follow-up of education establishments and student performance tracking. 6 sections in scope; modules 7–10 of the original `analysius` vision (exams/certificates/commissions/LMS/careers) are parked until the customer expands scope.

**Backend stack.** Encore.go. Open-source MPL-2.0 framework that defines services as Go packages, APIs as annotated Go functions, and infrastructure (Postgres, pub/sub, secrets, cron, caches) as code-declared resources. Self-hostable: `encore build docker` emits a `FROM scratch` Linux image (~30 MB).

**Why Encore.go.** Same justification chain as before — small memory footprint, single static binary, easy ministry self-host — but Encore eliminates the boilerplate we'd otherwise hand-roll (logger, metrics, OpenAPI, tracing, RPC plumbing, Dockerfile, local-dev compose). It also gives us **elastic service granularity**: start as one service, split a package into its own service later when load actually demands it. No "monolith vs microservices" argument; we just move code between packages.

**Foundation.** Empty Git repo already provisioned at **`https://github.com/formswrite/schoolrise.git`** — `git clone` it into `/Users/dvira/Desktop/projects/formswrite/school-rise/` and scaffold the Encore app + Next.js frontend inside it. **Form building is delivered by an in-house `forms` service** — a sibling Encore service inside the SchoolRise app. No external SaaS, no AI calls, no `api.formswrite.com`. The `forms` service mirrors **Formswrite's form-builder mechanics** (full 26-field-type catalogue, discriminated-union question model, three-panel builder UX) and is consumed by the `assessment` service when an inspector creates a campaign tied to a published form version. We do **not** replicate Formswrite's AI document-import pipeline.

**Reference codebases (read for patterns; do not import code from):**
- Formswrite frontend: `/Users/dvira/Desktop/projects/formswrite/formswrite-frontend/` — port the form-editor three-panel layout, field-types catalogue, grading logic, conditional-logic engine
- Formswrite backend: `/Users/dvira/Desktop/projects/formswrite/formswrite-backend/` — port the question/answer/response data model and the `formswrite_question.model.js` discriminated-union shape

**Source-of-truth references** (we read but do not import code from these):
- Customer spec (French): the 6-section document in this conversation
- Long-term vision: `/Users/dvira/Desktop/projects/formswrite/school-rise/analysius`
- Architecture inspiration: `/Users/dvira/Desktop/projects/scrapengine/scrapengine-backend/` (patterns only — its NestJS libs are no longer needed; Encore replaces them)
- Domain inspiration: `/Users/dvira/core/` (OpenEMIS Core — for the domain model only, not deployment patterns)

---

## 2. Customer Requirements (French Specification)

| § | Customer requirement | Encore service / package |
|---|---|---|
| 1 | **Informations administratives et pédagogiques** — IRE / DPE / DSEE / École, personnel enseignant et encadrant, organisation pédagogique, effectifs scolaires globaux | `tenancy`, `people`, `academics` |
| 2 | **Niveaux d'étude et organisation des groupes** — CE1, CE2, CM1, groupes pédagogiques, élèves par groupe, répartition selon besoins | `academics` |
| 3 | **Suivi des effectifs et des évaluations** — élèves testés, garçons, filles, total | `enrollment` |
| 4 | **Compétences en Français** — Débutant / Lettres / Mots / Paragraphe / Histoire | `assessment` |
| 5 | **Compétences en Mathématiques** — Débutant / 1 chiffre / 2 chiffres / Soustraction / Division | `assessment` |
| 6 | **Tableau de progression dynamique** — time-based, comparisons par groupe/classe/école, French + Maths | `progression` |

**Hierarchy (customer terminology, kept verbatim from the French specification — these are *configurable seed-data labels*, not hardcoded code identifiers).**
```
National
└── IRE  (Inspection Régionale de l'Éducation)
    └── DPE  (Direction Préfectorale de l'Éducation)
        └── DSEE  (Délégation Scolaire de l'Enseignement Élémentaire)
            └── École
                └── Classe (CE1 / CE2 / CM1 …)
                    └── Groupe pédagogique
                        └── Élève
```

In code, these are referenced by neutral identifiers (`region`, `prefecture`, `delegation`, `institution`, `class`, `group`, `student`). Display labels are loaded from the active locale pack — Guinea ships with the French labels above; an English-default deployment shows "Region / Prefecture / Delegation / School / Class / Group / Student"; another country renames them in their own seed file.

**Default UI language: English.** The customer's spec is in French, but the platform itself ships English-first; French (and other ministry languages) are translation packs delivered as JSON locale files.

---

## 3. Architecture: One Encore Application, Many Service-Shaped Packages

### 3.1 Topology

```
┌──────────────────────────────────────────────────────────────┐
│  schoolrise (Encore.go application — single deployable)      │
│                                                              │
│  Services (Go packages, each with //encore:api endpoints):   │
│    auth · tenancy · people · academics · enrollment ·        │
│    forms · assessment · progression · imports · notifications│
│                                                              │
│  Infrastructure declared in code:                            │
│    sqldb.NewDatabase("schoolrise", ...)                      │
│    pubsub.NewTopic("score.finalized", ...) (NSQ in self-host)│
│    secrets, cron, cache (Redis) — all in-code                │
└──────────────────────────────────────────────────────────────┘
                            │
                            ▼
                     ┌────────────┐
                     │ PostgreSQL │
                     │  16-alpine │
                     │  per-svc   │
                     │  schemas   │
                     └────────────┘
                            │
                            ▼
            ┌────────────────────────────────┐
            │  Next.js 15 (App Router, RSC)  │  apps/web
            │  Calls Encore-generated TS     │  (separate
            │  client (auto-typed)           │  Node image)
            └────────────────────────────────┘
```

**Deployment shape: per-service images from day one.** Each Encore service ships as its own `FROM scratch` Docker image (~30 MB each, produced via `encore build docker <service> --services=<service>`). This lets ministries scale individual services independently — e.g., during a national campaign close, run 5 replicas of `progression` while keeping `auth` at 1. Inter-service calls between separate images become network RPC; within an image they stay in-process. Encore handles the rewiring automatically based on the `--services` flag at build time — no application-code changes required.

**Default v1 deploy = 13 containers:**
- `postgres` (one DB, schema-per-service)
- `web` (Next.js)
- 11 Encore service images: `auth`, `tenancy`, `people`, `academics`, `enrollment`, `forms`, `assessment`, `progression`, `imports`, `notifications`, `ai`
- One additional `gateway` (Encore-generated routing layer that fronts the per-service backends and exposes a single `:8080` to the frontend)

Pure in-house — no external services beyond Resend (email) and OpenAI (AI), no phone-home, no SaaS dependencies. Air-gap-capable when those two egress destinations are firewalled or replaced.

### 3.2 Why this shape

- **Encore lets us start as one service and split later for free.** v1 ships as a single Encore service named `schoolrise` containing all packages. If `progression` becomes load-heavy in Phase 2, we make it its own Encore service in the same app — same code, separate process, Encore handles the wiring.
- **One Postgres database, one schema per package.** Encore's `sqldb.NewDatabase` per package produces logical separation; cross-schema reads are confined to `progression` (which materialises views from the others).
- **No hand-rolled platform code.** Encore provides tracing, metrics, structured logs, OpenAPI, RPC, secrets, local dev — we don't lift libs from scrapengine; Encore replaces them.
- **Built-in service catalog + auto-generated architecture diagram.** Important for an open-source ministry-grade project: anyone forking the repo opens the dev dashboard and sees what's in front of them.
- **Frontend stays orthogonal.** Next.js 15 in `apps/web` consumes the Encore-generated TypeScript client. Type-safe end-to-end without us writing API clients.
- **Form-building is in-house, no SaaS.** The `forms` service is a sibling of `assessment` inside the SchoolRise app. Inspectors build assessment questionnaires (questions, options, validation) directly in the platform. No call to `api.formswrite.com`, no LLM, no AI grading — everything stays in the ministry's data plane.

### 3.3 Repo layout

```
schoolrise/                                 (cloned from https://github.com/formswrite/schoolrise.git)
├── encore.app                              Encore app manifest
├── go.mod
│
├── auth/                                   service: identity, sessions, RBAC scopes
│   ├── auth.go                             //encore:api endpoints
│   ├── service.go                          encore:service struct
│   └── migrations/                         schema/auth.*
│
├── tenancy/                                service: IRE/DPE/DSEE/École hierarchy + schools
│   ├── institutions.go
│   ├── hierarchy.go                        closure-table queries
│   ├── service.go
│   └── migrations/
│
├── people/                                 service: students, staff, guardians
│   ├── students.go
│   ├── staff.go
│   ├── customfields.go                     port of OpenEMIS CustomField pattern
│   ├── service.go
│   └── migrations/
│
├── academics/                              service: periods, niveaux, groupes, classes
├── enrollment/                             service: rosters, coverage rollups
├── forms/                                  service: form builder — questionnaires, fields, validation, response capture
├── assessment/                             service: scales, campaigns, assignments + signed tokens, deterministic scoring
├── progression/                            service: materialised views, drilldown API
├── imports/                                service: CSV/XLSX bulk-load with templates
├── notifications/                          service: email (Resend) + alert rules
├── ai/                                     service: LLM chokepoint — item suggestion, rubric drafting, free-text grading, distractor gen
│
├── pkg/                                    plain Go shared code (no encore-decorated symbols)
│   ├── domain/                             Hierarchy, Period, Scale, RoleScope
│   ├── customfields/                       per-entity JSONB engine (port from OpenEMIS)
│   ├── hierarchy/                          closure-table helpers
│   ├── importtmpl/                         CSV template parser + validator
│   └── seed/                               Guinea seed data (regions, scales, RBAC)
│
├── infra-config/                           per-environment Encore infra config
│   ├── selfhost.json                       NSQ + local Postgres (default for ministries)
│   ├── gcp.json                            GCP Cloud SQL + GCP Pub/Sub
│   └── aws.json                            RDS + SNS/SQS
│
├── apps/
│   └── web/                                Next.js 15 (separate Node image)
│       ├── src/app/
│       └── package.json
│
├── deploy/
│   ├── docker-compose.yml                  default 3-container ministry deploy
│   ├── docker-compose.formswrite.yml       overlay enabling AI extraction
│   ├── docker-compose.scale.yml            overlay adding NSQ for high-load campaigns
│   └── helm/schoolrise/                    K8s chart (Phase 2, customer-driven)
│
├── examples/
│   ├── seed-guinea/                        IRE/DPE/DSEE rows + 8 region names
│   └── seed-senegal/                       sample alternative hierarchy seed
│
├── docs/                                   Docusaurus or MkDocs site
├── .github/workflows/                      CI: lint, vet, test, encore build, GHCR push
└── README.md
```

### 3.4 Encore service patterns (canonical examples)

**Defining a service + an API endpoint.**

```go
// tenancy/institutions.go
package tenancy

import "context"

type CreateInstitutionParams struct {
    Name        string
    DseeID      int64
    Type        string  // "primary" | "secondary"
    Code        string
}

//encore:api auth method=POST path=/v1/institutions
func CreateInstitution(ctx context.Context, p *CreateInstitutionParams) (*Institution, error) {
    // ... persist via tenancyDB
}

//encore:api auth method=GET path=/v1/institutions/:id
func GetInstitution(ctx context.Context, id int64) (*Institution, error) { /* ... */ }
```

**Declaring a Postgres database (per service).**

```go
// tenancy/db.go
package tenancy

import "encore.dev/storage/sqldb"

var tenancyDB = sqldb.NewDatabase("tenancy", sqldb.DatabaseConfig{
    Migrations: "./migrations",
})
```

**Calling another service (typed function call, Encore handles RPC).**

```go
// progression/refresh.go
package progression

import "encore.app/tenancy"

func refreshFor(ctx context.Context, institutionID int64) error {
    inst, err := tenancy.GetInstitution(ctx, institutionID)
    if err != nil {
        return err
    }
    // … materialise rollups for inst.Hierarchy
}
```

**Async pub/sub (NSQ in self-host, GCP/AWS in cloud — backend chosen at deploy time).**

```go
// assessment/events.go
package assessment

import "encore.dev/pubsub"

type ScoreFinalized struct {
    StudentID  int64
    CampaignID int64
    Band       string
    Subject    string  // "french" | "maths"
}

var ScoreFinalizedTopic = pubsub.NewTopic[*ScoreFinalized]("score.finalized",
    pubsub.TopicConfig{ DeliveryGuarantee: pubsub.AtLeastOnce })

// progression/subscribers.go
var _ = pubsub.NewSubscription(assessment.ScoreFinalizedTopic, "progression-refresh",
    pubsub.SubscriptionConfig[*assessment.ScoreFinalized]{ Handler: HandleScoreFinalized })
```

**Auth handler (gateway-level, runs before any `auth=true` endpoint).**

```go
// auth/handler.go
package auth

//encore:authhandler
func AuthHandler(ctx context.Context, token string) (auth.UID, *AuthData, error) {
    // validate session/JWT, load RoleScope, return principal
}
```

### 3.5 Service responsibilities & customer-section mapping

| Service / package | Customer § | Owns | Public endpoints |
|---|---|---|---|
| `auth` | all | users, sessions, api_keys, roles, role_assignments | `POST /v1/auth/login`, `POST /v1/sessions`, `POST /v1/api-keys` |
| `tenancy` | 1, 2 | regions (IRE), prefectures (DPE), delegations (DSEE), institutions (École), hierarchy_closure | `GET/POST /v1/{regions,prefectures,delegations,institutions}` |
| `people` | 1 | persons, students, staff, guardians, person_custom_fields | `GET/POST /v1/{students,staff,guardians}`, `POST /v1/people/custom-fields` |
| `academics` | 1, 2 | academic_periods, niveaux, groupes_pedagogiques, classes, class_students, class_staff | `GET/POST /v1/{periods,niveaux,classes}`, `POST /v1/classes/:id/enroll` |
| `enrollment` | 3 | enrollments, enrollment_events, coverage_snapshots | `POST /v1/enrollment/transfers`, `GET /v1/enrollment/coverage` |
| `forms` | 4, 5 | forms, form_versions, form_fields, form_field_options, form_validation_rules, form_responses | `GET/POST /v1/forms`, `POST /v1/forms/:id/publish`, `POST /v1/forms/:id/responses` |
| `assessment` | 4, 5 | scales (seeded: french_5level, maths_5level), scale_bands, campaigns, responses, scores | `POST /v1/campaigns`, `POST /v1/responses`, `GET /v1/scales` |
| `progression` | 6 | mv_*_progression materialised views, Redis cache | `GET /v1/progression/drilldown?level=...&subject=...` |
| `imports` | cross | import_templates, import_runs, import_errors | `POST /v1/imports/runs` |
| `notifications` | cross | notifications_outbox, alert_rules | `POST /v1/notifications/test` |
| `ai` | cross | ai_jobs, ai_job_results, prompt registry | `POST /v1/ai/suggest-items`, `/v1/ai/draft-rubric`, `/v1/ai/grade-essay`, `/v1/ai/generate-distractors` |

### 3.6 Async events (Encore Pub/Sub)

NSQ backend in self-host (only OSS option Encore supports for self-host). v1 uses pub/sub minimally — most flows are direct typed calls. Event surface:

- `tenancy.institution.created|updated|archived` → `progression`, `enrollment`
- `enrollment.enrollment.created|transferred|dropped` → `progression`, `notifications`
- `assessment.score.finalized` → `progression` (refresh debounce), `notifications` (low-band remediation alerts)
- `assessment.campaign.closed` → `progression` (full refresh)
- `imports.run.completed` → `notifications`

If a deployment doesn't enable NSQ, Encore degrades pub/sub to in-process delivery — fine for pilot deployments.

---

## 4. Data Model

**One Postgres database, one schema per Encore service.** Each `sqldb.NewDatabase("name", ...)` creates a separate logical database in Encore's local cluster (and a separate schema in shared self-host Postgres via the infra config). Cross-schema reads are confined to `progression`.

```
db: schoolrise
schemas:
  auth.*           tenancy.*        people.*         academics.*
  enrollment.*     forms.*          assessment.*     progression.*
  imports.*        notifications.*  ai.*
```

### 4.1 Core tables (essentials)

**`tenancy`** — `regions`, `prefectures`, `delegations`, `institutions`, `institution_types`, `hierarchy_closure` (closure table for fast subtree queries).

**`people`** — `persons` (PII root), `students` (FK person), `staff` (FK person), `guardians`; `person_custom_field_definitions` + `person_custom_fields` (JSONB).

**`academics`** — `academic_periods`, `niveaux`, `groupes_pedagogiques`, `classes`, `class_staff`, `class_students`.

**`enrollment`** — `enrollments`, `coverage_snapshots` (precomputed: tested/garçons/filles/total per scope per campaign).

**`forms`** (mirrors Formswrite's form-builder data model — see `/Users/dvira/Desktop/projects/formswrite/formswrite-backend/src/sequelize/models/formswrite_question.model.js`):

- `forms` (id, public_id (12-char nanoid), owner_id, title, description, status: `draft`|`published`|`closed`, settings JSONB {collect_email, limit_one_response, accepting_responses, confirmation_message}, response_count, view_count, published_at, created_at, updated_at)
- `form_versions` (immutable once published — campaigns lock to a version_id)
- `questions` (id, form_id, version_id, client_id (16-char nanoid for ephemeral UI ID), order, title, description, type VARCHAR(64), options JSONB, scale_min, scale_max, scale_labels JSONB, required, image JSONB, grading JSONB, extra JSONB (assessment-type-specific data), translations JSONB, deleted_at)
- `responses` (id, form_id, version_id, student_id (FK people.students), campaign_id (FK assessment.campaigns), respondent_email, submitted_at, encrypted_payload, created_at)
- `answers` (id, response_id, question_id (stable across edits), value text, array_value JSONB)

**Field-type catalogue (26 types, discriminated by `questions.type`)** — lifted from Formswrite's `lib/field-types.js`:

| Category | Types |
|---|---|
| **Text** | `SHORT_ANSWER`, `PARAGRAPH`, `EMAIL`, `PHONE`, `HOME_NUMBER`, `NUMBER`, `DECIMAL` |
| **Choice** | `MULTIPLE_CHOICE`, `CHECKBOX`, `DROPDOWN`, `RADIO`, `YES_NO`, `COUNTRY_REGION` |
| **Scale & Rating** | `LINEAR_SCALE`, `RATING` |
| **Date & Time** | `DATE`, `TIME` |
| **Media & Files** | `FILE_UPLOAD`, `ATTACHMENT`, `IMAGE`, `SIGNATURE` |
| **Compound** | `ADDRESS`, `TABLE` |
| **Assessment** (gradable) | `ORDERING`, `MATCHING`, `FILL_IN_BLANK`, `EQUATION`, `ESSAY`, `HOTSPOT`, `CODE_BLOCK` |
| **Layout** (non-submittable) | `SECTION`, `STATEMENT` |

Validation lives on the question itself (not a separate table) as `validation JSONB { min, max }` — applies to `SHORT_ANSWER`, `PARAGRAPH`, `NUMBER`, `DECIMAL`, `EMAIL`, `PHONE`. Type-specific config (e.g., `allowedDomains[]` for `EMAIL`, `scale_min`/`scale_max`/`scale_labels` for `LINEAR_SCALE`, `pairs[]` for `MATCHING`, `template` + `blanks[]` for `FILL_IN_BLANK`, `regions[]` for `HOTSPOT`) goes in `extra JSONB`. Conditional logic (`show field B if field A == X`) is supported via a separate `form_logic_rules` table porting Formswrite's `lib/logic.js` operators (`equals`, `either`, `lte`, `gte`).

**`assessment`** — `scales`, `scale_bands`, `campaigns` (links to a `form_version_id` for the items, plus `scale_id` for the scoring band), `campaign_targets`, `assignments` (campaign_id × student_id × access_token: each student gets a unique signed link), `scores` (student × campaign × band, deterministic). Campaigns reference a form version; the `forms` service handles item rendering and answer capture, the `assessment` service handles assignment generation and band scoring.

**`progression`** — materialised views only: `mv_student_progression` → `mv_class_progression` → `mv_school_progression` → `mv_dsee_progression` → `mv_dpe_progression` → `mv_ire_progression` → `mv_national_progression`. Refresh strategy: `REFRESH MATERIALIZED VIEW CONCURRENTLY` debounced 30 s after the last `score.finalized` event in a campaign window.

### 4.2 The two scales (seeded in migration, not config)

```sql
-- assessment/migrations/2_seed_scales.up.sql
INSERT INTO scales (code, label) VALUES
  ('french_5level', 'Compétences en Français'),
  ('maths_5level',  'Compétences en Mathématiques');

INSERT INTO scale_bands (scale_code, ordinal, code, label) VALUES
  ('french_5level', 1, 'debutant',   'Débutant'),
  ('french_5level', 2, 'lettres',    'Lettres'),
  ('french_5level', 3, 'mots',       'Mots'),
  ('french_5level', 4, 'paragraphe', 'Paragraphe'),
  ('french_5level', 5, 'histoire',   'Histoire'),
  ('maths_5level',  1, 'debutant',     'Débutant'),
  ('maths_5level',  2, 'un_chiffre',   '1 chiffre'),
  ('maths_5level',  3, 'deux_chiffres','2 chiffres'),
  ('maths_5level',  4, 'soustraction', 'Soustraction'),
  ('maths_5level',  5, 'division',     'Division');
```

The schema is generic, so a future EGRA / CEFR scale is a new migration, not a code branch.

---

## 5. Frontend (Next.js 15)

`apps/web/` is a separate Node container. It consumes the Encore-generated TypeScript client (`encore gen client typescript`) which produces fully-typed wrappers around every `//encore:api` endpoint.

### 5.1 Page map (matches the customer's 6 sections)

| § | Next.js route | Purpose |
|---|---|---|
| 1 | `/admin/institutions` | Browse hierarchy IRE → DPE → DSEE → École, edit schools |
| 1 | `/admin/staff` | Personnel enseignant et encadrant |
| 2 | `/admin/niveaux` | CE1 / CE2 / CM1 + groupes pédagogiques per école |
| 2 | `/admin/classes` | Class roster, students per group, learning-needs distribution |
| 3 | `/enrollment` | Roster per class |
| 3 | `/enrollment/coverage` | Tested counts: garçons / filles / total per scope |
| 4 | `/assessments/french` | Run + monitor French campaigns; band positioning |
| 5 | `/assessments/maths` | Run + monitor Maths campaigns; band positioning |
| 4, 5 | `/forms` | Form-builder list (own forms + shared) |
| 4, 5 | `/forms/:id/edit` | Three-panel form editor (port Formswrite layout: outline sidebar / WYSIWYG canvas / settings panel + QuickFieldPicker for the 26 field types) |
| 4, 5 | `/forms/:id/preview` | Renderer preview (what the student will see) |
| 4, 5 | `/r/:assignmentToken` | **Public student-facing form renderer** — student opens their unique signed link, completes the form, submits. No login required for students |
| 6 | `/dashboard` | Tableau de progression — time series + group/class/école comparisons |

**Stack.** Next.js 15 App Router, RSC, TypeScript, Tailwind + shadcn/ui in `apps/web/src/lib/ui`, `next-intl` for locale strings, **English as the v1 canonical locale**. All UI strings, error messages, validation messages, email templates, and DB seed labels live as i18n keys with English values in `apps/web/src/locales/en.json` (and a sibling `pkg/seed/locales/en.json` for backend-emitted strings). Adding French (or Wolof, Malinké, Soussou, Portuguese) is a Phase 2 task: drop a `fr.json` locale pack alongside, no code changes required.

### 5.2 Auth bridge

NextAuth session → Encore's auth handler over a signed cookie. The Encore client is configured in `apps/web/src/server/encore.ts` with the `Authorization` header set per request from the session.

### 5.3 Delivery model — teacher creates, students complete

The customer-confirmed delivery model is **teacher-creates / student-completes** (not teacher-captures-on-behalf):

1. **Teacher / inspector builds the form** in `/forms/:id/edit` using the three-panel builder (outline / canvas / settings).
2. **Teacher publishes the form** (creates an immutable `form_versions` row).
3. **Inspector creates a campaign** linked to the form version, scoped to specific classes / écoles. On open, the `assessment` service generates one signed assignment token per targeted student (`assessment.assignments`).
4. **Tokens are delivered to students** via one of two paths:
   - **Email** (when the student has a guardian-supplied or institution-issued email): `notifications` service sends the link via Resend
   - **Printed sheet / shared link**: the teacher prints a per-student sheet with a QR code + token, or shares a class link the teacher unlocks per-student in class
5. **Student opens their unique link** at `/r/:assignmentToken` — no login. The form renders (port Formswrite's renderer); student answers; submits.
6. **Submission writes to** `forms.responses` + `forms.answers`; `assessment` deterministically scores the answers against the scale; `assessment.scores` row is written; `score.finalized` event fires; `progression` materialised views refresh.

Tokens are **single-use** by default (configurable per form: `limit_one_response`), expire on campaign close, and are revocable by the inspector. No student account, no password — the signed token *is* the identity for the duration of the campaign.

For schools with no per-student device or connectivity (the rural reality in many of Guinea's écoles), the teacher can fall back to opening each student's link on a single shared device — a degenerate case of the same flow, no separate code path.

---

## 6. Deployment — "Git Clone, Docker Compose Up, Done"

We ship a default 3-container compose that genuinely works on first try. No installer wizard, no SQL dump to import.

### 6.1 Default deployment (ministries / pilot)

Per-service images, one container per Encore service plus the gateway, web, and postgres. Each service can be independently scaled by raising its replica count without touching the others.

```yaml
# deploy/docker-compose.yml
x-service-defaults: &service-defaults
  env_file: .env
  environment:
    ENCORE_RUNTIME_CONFIG: /infra/selfhost.json
  volumes:
    - ./infra-config/selfhost.json:/infra/selfhost.json:ro
  depends_on:
    postgres: { condition: service_healthy }
  restart: unless-stopped

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: schoolrise
      POSTGRES_USER: schoolrise
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes: [pgdata:/var/lib/postgresql/data]
    healthcheck: { test: ["CMD-SHELL", "pg_isready -U schoolrise"], interval: 5s }

  # ─── 11 Encore services (one image each) ──────────────────────
  auth:          { <<: *service-defaults, image: ghcr.io/formswrite/schoolrise-auth:${TAG:-latest} }
  tenancy:       { <<: *service-defaults, image: ghcr.io/formswrite/schoolrise-tenancy:${TAG:-latest} }
  people:        { <<: *service-defaults, image: ghcr.io/formswrite/schoolrise-people:${TAG:-latest} }
  academics:     { <<: *service-defaults, image: ghcr.io/formswrite/schoolrise-academics:${TAG:-latest} }
  enrollment:    { <<: *service-defaults, image: ghcr.io/formswrite/schoolrise-enrollment:${TAG:-latest} }
  forms:         { <<: *service-defaults, image: ghcr.io/formswrite/schoolrise-forms:${TAG:-latest} }
  assessment:    { <<: *service-defaults, image: ghcr.io/formswrite/schoolrise-assessment:${TAG:-latest} }
  progression:   { <<: *service-defaults, image: ghcr.io/formswrite/schoolrise-progression:${TAG:-latest} }
  imports:       { <<: *service-defaults, image: ghcr.io/formswrite/schoolrise-imports:${TAG:-latest} }
  notifications: { <<: *service-defaults, image: ghcr.io/formswrite/schoolrise-notifications:${TAG:-latest} }
  ai:            { <<: *service-defaults, image: ghcr.io/formswrite/schoolrise-ai:${TAG:-latest} }

  # ─── Gateway: Encore-generated single API entry point ─────────
  gateway:
    image: ghcr.io/formswrite/schoolrise-gateway:${TAG:-latest}
    env_file: .env
    ports: ["8080:8080"]
    depends_on:
      - auth
      - tenancy
      - people
      - academics
      - enrollment
      - forms
      - assessment
      - progression
      - imports
      - notifications
      - ai
    restart: unless-stopped

  # ─── Frontend ────────────────────────────────────────────────
  web:
    image: ghcr.io/formswrite/schoolrise-web:${TAG:-latest}
    env_file: .env
    environment:
      ENCORE_API_URL: http://gateway:8080
    ports: ["3000:3000"]
    depends_on: [gateway]
    restart: unless-stopped

volumes: { pgdata: {} }
```

**Scaling individual services.** With docker-compose: `docker compose up -d --scale progression=5 --scale assessment=3`. With Kubernetes: each service gets its own Deployment + HPA in the Helm chart (`deploy/helm/schoolrise/templates/<service>-deployment.yaml`), targeting per-service CPU/memory thresholds. The `progression` HPA, for example, can scale 1→10 replicas based on CPU > 70% during campaign-close windows, then drift back down.

**Inter-service communication.** Once services are split into separate images, Encore's RPC layer becomes a real network call. Encore handles this transparently — service discovery via the `gateway` container's known addresses, retries, timeouts, distributed tracing across hops. No application code change vs. the bundled-image alternative; it's all driven by `encore build docker --services=<svc>` at CI time.

### 6.1.1 Operator-supplied `.env` (single file, applies to both services)

The deployment **requires the operator to provide a `.env` file** at the repo root before `docker compose up`. We ship `.env.example` documenting every variable in French + English. The contract:

```bash
# ─── REQUIRED — bootstrap & DB ────────────────────────────────
POSTGRES_PASSWORD=<random strong password>     # any strong password
ADMIN_EMAIL=admin@minedu.gov.gn                # initial admin (used once on first boot)
ADMIN_PASSWORD=<temporary password>            # forced change on first login
AUTH_SECRET=<openssl rand -hex 32>             # signs sessions + assignment tokens
BASE_URL=https://schoolrise.minedu.gov.gn      # public base URL for assignment links

# ─── REQUIRED — email delivery (Resend) ───────────────────────
RESEND_API_KEY=re_xxxxxxxxxxxxxxxxxxxxxx       # Resend account API key
EMAIL_FROM=schoolrise@minedu.gov.gn            # verified sender on Resend
EMAIL_FROM_NAME=SchoolRise — Ministry of Education
LOG_LOCALE=en                                  # default: English; switch to fr/es/pt when locale pack present

# ─── REQUIRED — AI provider (SchoolRise is AI-native) ─────────
OPENAI_API_KEY=sk-...                          # required; powers AI features across the platform
OPENAI_MODEL=gpt-4o-mini                       # default model
ANTHROPIC_API_KEY=                             # optional alternative provider; falls back to OpenAI

# ─── OPTIONAL — file storage (FILE_UPLOAD/IMAGE/SIGNATURE) ────
S3_ENDPOINT=                                   # leave blank for local-disk storage
S3_BUCKET=
S3_ACCESS_KEY=
S3_SECRET_KEY=
S3_REGION=auto

# ─── OPTIONAL — observability ────────────────────────────────
PROMETHEUS_SCRAPE_TOKEN=                       # if set, /metrics is gated
SENTRY_DSN=                                    # if set, errors fan out to Sentry
```

**v1 minimum to boot:** `POSTGRES_PASSWORD`, `ADMIN_EMAIL`, `ADMIN_PASSWORD`, `AUTH_SECRET`, `BASE_URL`, `RESEND_API_KEY`, `EMAIL_FROM`, `OPENAI_API_KEY`. SchoolRise is **AI-native** — there is no flag to disable AI; the platform uses LLMs as a first-class capability the same way it uses Postgres. Operators who don't want AI features visible to their users can use a billing-capped key, but the API key itself is required.

**Boot-time env validation:** `pkg/seed/bootstrap.go` validates the env file on first boot, fails fast with an English-language error listing missing/malformed required vars (locale-controlled via `LOG_LOCALE`), and refuses to start the API server. Saves operators from "it's running but emails don't send" surprises.

### 6.2 Encore infra config (self-host)

```json
// infra-config/selfhost.json
{
  "$schema": "https://encore.dev/schemas/infra.schema.json",
  "metadata": {
    "app_id": "schoolrise",
    "env_name": "selfhost",
    "env_type": "production",
    "cloud": "self-hosted",
    "base_url": "${BASE_URL}"
  },
  "sql_servers": [{
    "host": "${DB_HOST}:5432",
    "databases": {
      "auth":          { "username": "schoolrise", "password": {"$env": "DB_PASSWORD"} },
      "tenancy":       { "username": "schoolrise", "password": {"$env": "DB_PASSWORD"} },
      "people":        { "username": "schoolrise", "password": {"$env": "DB_PASSWORD"} },
      "academics":     { "username": "schoolrise", "password": {"$env": "DB_PASSWORD"} },
      "enrollment":    { "username": "schoolrise", "password": {"$env": "DB_PASSWORD"} },
      "forms":         { "username": "schoolrise", "password": {"$env": "DB_PASSWORD"} },
      "assessment":    { "username": "schoolrise", "password": {"$env": "DB_PASSWORD"} },
      "progression":   { "username": "schoolrise", "password": {"$env": "DB_PASSWORD"} },
      "imports":       { "username": "schoolrise", "password": {"$env": "DB_PASSWORD"} },
      "notifications": { "username": "schoolrise", "password": {"$env": "DB_PASSWORD"} },
      "ai":            { "username": "schoolrise", "password": {"$env": "DB_PASSWORD"} }
    }
  }],
  "secrets": {
    "AdminEmail":     {"$env": "ADMIN_EMAIL"},
    "AdminPassword":  {"$env": "ADMIN_PASSWORD"},
    "AuthSecret":     {"$env": "AUTH_SECRET"},
    "BaseURL":        {"$env": "BASE_URL"},
    "ResendAPIKey":   {"$env": "RESEND_API_KEY"},
    "EmailFrom":      {"$env": "EMAIL_FROM"},
    "EmailFromName":  {"$env": "EMAIL_FROM_NAME"},
    "OpenAIAPIKey":   {"$env": "OPENAI_API_KEY"},
    "AnthropicAPIKey":{"$env": "ANTHROPIC_API_KEY"},
    "S3Endpoint":     {"$env": "S3_ENDPOINT"},
    "S3Bucket":       {"$env": "S3_BUCKET"},
    "S3AccessKey":    {"$env": "S3_ACCESS_KEY"},
    "S3SecretKey":    {"$env": "S3_SECRET_KEY"},
    "SentryDSN":      {"$env": "SENTRY_DSN"},
    "LogLocale":      {"$env": "LOG_LOCALE"}
  },
  "metrics": { "type": "prometheus", "endpoint": "/metrics" }
}
```

For deployments that need scale-out async (national campaign close at full Guinea volume), add `nsq` as a service and switch the pub/sub block to `{"type": "nsq", "addr": "nsq:4150"}`.

### 6.3 Image build (per-service matrix)

CI builds 13 images per release tag: 11 Encore service images + 1 gateway + 1 web image. The CI matrix:

```yaml
# .github/workflows/release.yml (excerpt)
strategy:
  matrix:
    service: [auth, tenancy, people, academics, enrollment, forms,
              assessment, progression, imports, notifications, ai]
steps:
  - run: encore build docker ghcr.io/formswrite/schoolrise-${{ matrix.service }}:${{ github.ref_name }} \
           --services=${{ matrix.service }} \
           --config=infra-config/selfhost.json
  - run: docker push ghcr.io/formswrite/schoolrise-${{ matrix.service }}:${{ github.ref_name }}
```

Plus separate jobs for the gateway (`encore build docker ... --gateways=api`) and the Next.js web image (multi-stage `apps/web/Dockerfile`).

Each Encore service image is a `FROM scratch` Linux binary running on port 8080 — typically ~30 MB on disk, ~25 MB RAM idle. The gateway image is similarly small.

**Why this is cheap to do on day one.** Encore controls the build; we don't hand-roll Dockerfiles per service. Adding a 12th service later means adding one entry to the CI matrix — no per-service Dockerfile maintenance, no template duplication.

### 6.4 First-boot bootstrap (no manual installer)

The Encore app's `service.go` files declare initialisation hooks. We add a one-shot startup task in `pkg/seed/bootstrap.go` (called from each service's `init`) that:

1. **Validates `.env`** — required vars present, well-formed (Resend key prefix, URL format, etc.); fails fast in French if anything's missing
2. **Verifies Resend connectivity** — sends a test API call (no email sent); fails fast if `RESEND_API_KEY` invalid
3. Runs Encore migrations automatically (handled by Encore on connect)
4. Seeds the two assessment scales if `scale_bands` is empty
5. Seeds default Guinea hierarchy from `examples/seed-guinea/regions.sql` if `regions` is empty
6. Creates initial admin from `ADMIN_EMAIL` / `ADMIN_PASSWORD` if `users` is empty, marks `must_change_password=true`
7. Logs each step to stdout in English (visible via `docker compose logs app`) — locale-controlled via `LOG_LOCALE` env var (default `en`); future French translation pack switches this to `fr`

**Result:** `git clone https://github.com/formswrite/schoolrise.git`, copy `.env.example` to `.env` and fill in the 7 required vars, `docker compose up -d`, open `https://<BASE_URL>`. No installer wizard.

### 6.5 Resource sizing (per-service topology)

Each Encore service idle ~25 MB RAM × 11 services + gateway = ~325 MB for the full backend at 1 replica each. Postgres still dominates.

| Scale | Schools | Students | Recommended infra | Per-service replicas |
|---|---|---|---|---|
| Pilot (1 IRE, ~50 schools) | 50 | 15 000 | 2 vCPU / 4 GB RAM / 20 GB disk | 1× each |
| Region full (1 IRE) | 1 500 | 450 000 | 4 vCPU / 8 GB RAM / 50 GB disk | 1× each |
| National Guinea | 13 000 | 3 000 000 | 8 vCPU / 32 GB RAM / 500 GB disk on Postgres; per-service replicas tuned for load | 2× `auth`, 2× `gateway`, 5× `progression` during campaign close (HPA), 3× `assessment`, 2× `forms`, 1× others |

The per-service split lets ministries scale just the services under load rather than the whole stack — `progression` can run 5 replicas during the 5-day window after a national campaign closes, then drift back to 1 replica when refresh activity ends. `auth` stays at 2 for HA. Idle services (`imports`, `notifications`) stay at 1.

For the pilot tier, all services run at 1 replica each on a single 2 vCPU / 4 GB box; the per-service split is invisible to the operator until they hit a bottleneck and choose to scale. The 4 GB floor (vs. 2 GB previously) accounts for the 11 service processes plus AI request handling.

---

## 7. Open-Source Posture

- **License: AGPL-3.0 (decided).** The SchoolRise application code is licensed AGPL-3.0. Mirrors OpenEMIS's copyleft philosophy and goes one step stricter — the AGPL network clause closes the SaaS loophole that plain GPL leaves open. Ministries self-hosting are unaffected; commercial vendors cannot fork SchoolRise into closed-source SaaS. Encore itself is MPL-2.0 — compatible. Every contributor signs a **CLA** giving FormsWrite the right to relicense their contributions, which preserves the option to **dual-license** (paid commercial exemption) for vendors who want to embed SchoolRise in proprietary products. `LICENSE` file at the repo root is the AGPL-3.0 text; optional `COMMERCIAL_LICENSE.md` added when the first commercial inquiry arrives.
- **Repo:** GitHub `schoolrise/schoolrise` with `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, French + English issue/PR templates
- **CI:** GitHub Actions — lint, vet, test, `encore build docker`, push to GHCR on tag
- **Docs:** Docusaurus site under `docs/` with sections: Quickstart, Self-hosting, Configure your country (seed files), Custom fields, Bulk import templates, API reference (auto-generated by Encore), Roadmap
- **Auto-generated artefacts** (a real differentiator vs OpenEMIS): Encore's dev dashboard renders an architecture diagram of all services, a service catalog, and OpenAPI specs out of the box. Forkers see the system on first `encore run`
- **Localisation:** **English canonical.** Every UI string + DB seed label + email body + validation message is i18n-keyed against `en.json`. Adding French / Wolof / Malinké / Soussou / Portuguese is a contributor PR adding a sibling locale JSON — no code change. This positions the project for international adoption: the open-source repo speaks English (the lingua franca of GitHub contributors and most ministry IT teams worldwide), and ministry deployments add their working language as a translation pack.
- **Hierarchy levels are seed data.** `examples/seed-guinea/`, `examples/seed-senegal/`, `examples/seed-template/` show how a ministry forks with their own administrative tiers
- **Custom fields engine** ported from OpenEMIS `CustomField` plugin into `pkg/customfields/` — biggest single deferred-flexibility win for international adoption

---

## 8. Phased Delivery (4–5 engineers, 10–12 weeks Phase 1)

Scope: the customer's 6 sections, frontend, bulk import, deployment, docs, v0.1.0 release.

### Engineer ownership

| Engineer | Owns |
|---|---|
| **E1 (platform)** | `encore.app`, repo scaffold, CI, GHCR publish, infra-config files (selfhost / gcp / aws), `deploy/docker-compose.yml`, `pkg/seed/bootstrap.go`, Helm chart skeleton, docs site infrastructure |
| **E2 (foundations)** | `auth` (better-auth-style sessions, RBAC scopes), `tenancy` (IRE→École hierarchy + closure table + Guinea seed), `people` + custom fields engine, `academics` (periods, niveaux, groupes), `imports` (CSV templates) |
| **E3 (assessment + dashboard)** | `assessment` (scales seed, campaigns, responses, scoring), `forms` (form-builder primitives), `enrollment` (rosters + coverage rollups), `progression` (materialised views, drilldown API, debounced refresh, optional Redis cache) |
| **E4 (frontend lead)** | `apps/web` shell, `libs/ui` design system (Tailwind + shadcn/ui), Encore TS client wiring, admin pages (institutions / staff / niveaux / classes), enrollment + coverage pages |
| **E5 (frontend + integrations)** | Assessment campaign pages (French + Maths), progression dashboard (drilldown UI + comparisons + time series), `notifications` (email outbox via Resend, alert rules), `ai` service (LLM chokepoint — item suggestion + rubric drafting + free-text grading) wired into form builder + assessment scoring, Docusaurus docs site, demo seed |

### Week-by-week (Phase 1)

| Weeks | Milestone |
|---|---|
| 1–2 | Encore app scaffold cloned from `https://github.com/formswrite/schoolrise.git`. `auth` service skeleton. Infra-config files (`selfhost.json`). **CI matrix builds 11 per-service scratch images + gateway + web image, pushes all to GHCR on tag.** `apps/web` shell + `libs/ui`. `deploy/docker-compose.yml` (13-container per-service topology) boots green on a clean machine |
| 3–4 | `tenancy` (closure table, Guinea seed in `examples/seed-guinea/`), `people` + custom fields, `academics` (periods, niveaux, groupes). Admin UI pages: institutions, staff. RBAC scopes wired through Encore auth handler |
| 5–6 | `enrollment` (rosters + coverage rollups), `imports` with CSV templates for students/staff/schools (port OpenEMIS Import patterns). Admin UI: classes, niveaux, roster. End-to-end bulk-import test on 1500 students |
| 7–8 | `forms` service ported from Formswrite: 26-field-type catalogue, discriminated-union question model, form versioning, conditional logic; three-panel builder UI in `apps/web` (outline / canvas / settings + QuickFieldPicker); public `/r/:assignmentToken` student renderer. `assessment` — scales seed, campaign lifecycle (links to `form_version_id`), assignment-token generation, deterministic scoring. Campaign UI for French + Maths. Resend email delivery for assignment links |
| 9–10 | `progression` materialised views (national → IRE → DPE → DSEE → école → classe → groupe → élève chain), drilldown API, debounced refresh on `score.finalized` topic. Dashboard UI: time series + comparisons by groupe/classe/école for both subjects |
| 11–12 | `notifications` (Resend email on import done / campaign closed / assignment links). `ai` service v1 surface: AI item suggestion in builder, rubric drafting for `ESSAY`, free-text auto-grading hooked into `assessment` scoring, distractor generation for `MULTIPLE_CHOICE`. `pkg/seed/bootstrap.go` admin auto-create + env validation + Resend + OpenAI connectivity check. End-to-end test suite. Load test at Guinea scale. Docs site (Quickstart + Self-hosting + Form builder + AI features + Custom fields + Configure your country + API reference). Release v0.1.0 to GitHub + GHCR |

**Phase 1 exit criteria:** §10 verification flow runs green on a fresh `docker compose up` with only 4 env vars set.

### Phase 2 (only if customer expands)

Configurable scales editor, more import templates (attendance, results), staff appraisals, bulk SMS adapter, Helm chart for K8s, multi-language UI, optional NSQ overlay for high-load campaigns.

**No examinations / certificates / commissions / LMS / careers until a customer signs for them.**

---

## 9. When To Add Complexity (extraction roadmap)

Encore makes most extractions trivial — splitting a service is a config change, not a refactor.

| Trigger | Add | How (in Encore) |
|---|---|---|
| Progression dashboard p95 > 1 s sustained | Redis cache for top-of-tree drilldown | Add `cache.NewKeyspace(...)` in `progression`, switch infra config to include Redis |
| Refresh latency > 30 s on incremental updates | `pg_ivm` extension for incremental view maintenance | Postgres extension; rewrite materialised views as `CREATE INCREMENTAL MATERIALIZED VIEW`; ~half-day work, 10–100× refresh-latency reduction |
| Single service under sustained load (e.g., `progression` during campaign close) | Scale that service horizontally | `docker compose up -d --scale progression=N`, or raise the per-service HPA `maxReplicas` in Helm. Already supported by the per-service image topology |
| > 10 simultaneous campaigns nationwide | Add NSQ for async scoring + import workers | Switch infra-config pub/sub block to NSQ, add NSQ container to compose |
| Cross-service query patterns become ad-hoc / multi-tenant SaaS | Add ClickHouse as derived OLAP store | Postgres logical replication or Debezium → ClickHouse; `progression` becomes a router (per-student → Postgres, aggregate → ClickHouse) |
| Object storage need (PDFs, content uploads) | Add MinIO | Encore object-storage primitive, `objects.NewBucket(...)` |
| Multi-region HA | Helm + managed Postgres + multi-AZ | Phase 3+, customer-driven |

The rule: every added piece of infrastructure has to point at a real, measured user-facing problem.

---

## 10. Verification Plan

A v1 build is sound if a single engineer can run this against a fresh `docker compose up` and every assertion passes. This is also the demo script.

### Smoke flow: "Guinea pilot — 1 IRE, 50 schools, French + Maths campaigns"

1. **Boot.** `cp .env.example .env`, set `POSTGRES_PASSWORD`, `ADMIN_EMAIL`, `ADMIN_PASSWORD`, `AUTH_SECRET`, `BASE_URL`, `RESEND_API_KEY`, `EMAIL_FROM`, `OPENAI_API_KEY`, then `docker compose up -d`. Within 60 s, the three containers are healthy. `http://localhost:3000` shows the login page in English.
2. **Login.** Authenticate as the bootstrapped admin; forced password change works; admin dashboard loads.
3. **Hierarchy seeded.** Guinea's 8 IRE rows + sample DPE/DSEE seeds are present from `examples/seed-guinea/`.
4. **Bulk import schools.** Upload sample CSV with 50 schools across 3 DSEE; assert all created, error report empty.
5. **Bulk import students.** Upload CSV with 1 500 students into those schools (30/school); assert linked to classes.
6. **Provision niveaux + groupes.** Each school: 3 niveaux (CE1, CE2, CM1), 1 groupe pédagogique each; 1 teacher per niveau. RBAC: teacher login sees only own classes.
7. **Build a French questionnaire.** Open `/forms`, click *Nouveau formulaire*, use the three-panel builder to add 10 questions covering at least 6 of the 26 field types (e.g., `MULTIPLE_CHOICE`, `SHORT_ANSWER`, `LINEAR_SCALE`, `MATCHING`, `FILL_IN_BLANK`, `ORDERING`). Add one conditional rule (`show Q5 only if Q3 == "Oui"`). Publish version 1.
8. **Run French campaign.** Create campaign scoped to all 50 schools, scale = `french_5level`, links to the form's published version, opens immediately, closes in 5 days. `assessment` generates one signed assignment token per targeted student; `notifications` sends each student's link via Resend (or prints them as a class sheet for offline distribution).
9. **Students complete the form.** Open `/r/:assignmentToken` from a sample student link — no login required. Render the form, answer, submit. Deterministic scoring assigns a band (Débutant…Histoire) from the answers; result visible to the teacher and inspector.
10. **Build a Maths questionnaire + run Maths campaign.** Same, scale = `maths_5level`, bands Débutant…Division.
11. **Coverage rollups.** `GET /v1/enrollment/coverage?scope=ire_1` returns counts: tested / garçons / filles / total. Numbers match raw query.
12. **Progression dashboard.** Within 60 s of campaign close, `/dashboard` shows time series for both subjects. Drill: national → IRE → DPE → DSEE → école → classe → groupe → élève. Sum of children = parent ± rounding.
13. **Comparisons.** Filter dashboard by groupe / classe / école — comparison view renders for French and Maths.
14. **Permission boundaries.** Teacher token can't read another school. DSEE inspector sees all schools in delegation, not in another. Verified by automated e2e test. **Assignment tokens** are single-use and reject reuse with HTTP 410.
15. **Trace propagation.** One request frontend → Encore API → Postgres produces a single trace ID visible in Encore's local dashboard and in production via Prometheus + traces.
16. **Reboot test.** `docker compose down && docker compose up -d` — bootstrap detects existing data, skips seed, app comes back up clean.
17. **Self-host test.** From a different machine, `git clone https://github.com/formswrite/schoolrise.git`, follow `docs/quickstart.md` exactly — works end-to-end in < 10 minutes.
18. **Egress audit.** Run `tcpdump`/`iptables` audit during the smoke flow: outbound traffic is restricted to `api.resend.com` (email delivery) and `api.openai.com` (AI features). All other traffic is internal to the docker network. Document this as the platform's egress contract.
19. **Architecture diagram.** `encore run` locally; the dev dashboard renders a service catalog + auto-generated architecture diagram of `auth / tenancy / people / academics / enrollment / forms / assessment / progression / imports / notifications`. This becomes a screenshot in the README.

If any assertion fails, v1 is not ready. If all pass, tag v0.1.0, push image to GHCR, publish on GitHub.

### Automated suites

- Encore-native unit tests per service (`encore test`) — target 80% coverage on domain logic
- e2e suite mirroring §10 smoke flow — runs against ephemeral Postgres via `encore run`
- Load test: k6 simulating 13 000 schools × 30 students × 1 campaign close — gate v0.1.0 release on p95 dashboard query < 2 s

---

## 11. Critical Risks

| # | Risk | Mitigation |
|---|---|---|
| 1 | Encore annotation lock-in | MPL-2.0 + Docker export means we're never blocked. Business logic, SQL, tests, frontend are all portable Go/SQL/TS. Migration cost = rewriting `//encore:api` boundaries — concentrated, finite |
| 2 | NSQ is the only OSS pub/sub option for self-host (RabbitMQ + Kafka unsupported) | v1 doesn't need pub/sub at all (in-process delivery is fine). Phase 2: NSQ overlay added only when load demands. NSQ is OSS, mature, low-ops |
| 3 | Customer expands scope mid-build to add modules 7–10 | Encore packages slot in cleanly. Resist scope creep until v1 exit criteria met |
| 4 | Progression dashboard slow at national scale | Materialised views chained per hierarchy level, debounced refresh, indexed `(scope_id, period_id, scale_code)`. **Per-service image topology already lets `progression` scale horizontally without touching other services.** Escape hatches in order: `pg_ivm` for incremental refresh → Redis cache for hot scopes → ClickHouse OLAP store |
| 5 | Bootstrap failure on first run leaves operator confused | `pkg/seed/bootstrap.go` writes French-language step-by-step status to stdout. Idempotent. Health endpoint exposes `bootstrap_status` for `docker compose ps` checks |
| 6 | Country-specific hierarchy (other ministries don't use IRE/DPE/DSEE) | Hierarchy levels are seed-data rows. Each deployment ships its own `examples/seed-<country>/` |
| 7 | Form-builder UX is non-trivial in 2 weeks (week 7–8) | Lock scope to the field types the assessment service actually needs (single_select, multi_select, text, number, scale_band). Defer drag-and-drop reorder, conditional logic, and file upload to Phase 2. Use shadcn/ui form primitives — no custom rendering engine |
| 8 | Encore migrations are up-only by default | Down migrations are documented in `docs/operations/migrations.md` as a manual procedure. We accept this — it's the standard operational stance |

---

## 12. Critical Files (templates & references)

**Read for patterns (do not import code from):**
- `/Users/dvira/Desktop/projects/formswrite/formswrite-backend/src/sequelize/models/formswrite_question.model.js` → `forms/` (discriminated-union question model — port directly to Encore + Postgres schema)
- `/Users/dvira/Desktop/projects/formswrite/formswrite-backend/src/sequelize/models/formswrite_form.model.js` → `forms/` (Form entity, settings, status enum, public_id pattern)
- `/Users/dvira/Desktop/projects/formswrite/formswrite-backend/src/utils/upsert-form-questions.js` → `forms/` (question upsert + version-bump logic)
- `/Users/dvira/Desktop/projects/formswrite/formswrite-frontend/lib/field-types.js` → `apps/web/src/lib/field-types.ts` (the 26 field types, configs, defaults — port verbatim minus the AI hooks)
- `/Users/dvira/Desktop/projects/formswrite/formswrite-frontend/lib/grading.ts` → `assessment/` (deterministic scoring of gradable types)
- `/Users/dvira/Desktop/projects/formswrite/formswrite-frontend/lib/logic.js` → `forms/` + `apps/web/` (conditional logic engine + frontend evaluator)
- `/Users/dvira/Desktop/projects/formswrite/formswrite-frontend/components/form-editor/FormEditorLayout.js` → `apps/web/src/components/form-editor/` (three-panel layout)
- `/Users/dvira/Desktop/projects/formswrite/formswrite-frontend/components/form-editor/QuickFieldPicker.js` → `apps/web/src/components/form-editor/` (categorised field picker)
- `/Users/dvira/core/plugins/CustomField/` → `pkg/customfields/` (per-entity JSONB engine)
- `/Users/dvira/core/plugins/Import/` → `imports/` (CSV templates, error reports, archived runs)
- `/Users/dvira/core/plugins/Area/` + `Institution/` → `tenancy/` (closure-table hierarchy)
- `/Users/dvira/core/plugins/Student/` + `Staff/` → `people/`
- `/Users/dvira/core/plugins/AcademicPeriod/` → `academics/`
- `/Users/dvira/core/plugins/Assessment/` + `Competency/` → `assessment/`
- `/Users/dvira/core/Dockerfile` + `docker-config/init.sh` → reference for the gap we're closing (their installer step is manual; ours isn't)

**Encore documentation we'll lean on:**
- https://encore.dev/go (framework overview)
- https://encore.dev/docs/go/primitives/services (service definition)
- https://encore.dev/docs/go/primitives/databases (sqldb)
- https://encore.dev/docs/go/primitives/pubsub (NSQ self-host)
- https://encore.dev/docs/ts/self-host/build (`encore build docker`)
- https://encore.dev/docs/go/self-host/configure-infra (infra-config JSON schema)

**Customer specification (re-read before each phase):**
- The 6-section French specification in this conversation
- `/Users/dvira/Desktop/projects/formswrite/school-rise/analysius` (long-term vision, parked beyond v1)

---

## 13. Open Decisions (non-blocking)

- **License: AGPL-3.0 vs MIT.** AGPL recommended (mirrors OpenEMIS). Confirm before first public release.
- ~~Single-image vs split web/api images.~~ **Decided: per-service images from day one.** Each Encore service ships as its own ~30 MB scratch image, plus a gateway image and the Next.js web image. Total 13 images per release. Lets ministries scale individual services independently (e.g., 5× `progression` during campaign close) without touching the rest of the stack.
- **Initial admin credentials.** Plan: env vars on first boot only, then forced password change on first login. Confirm acceptable.
- **Hosted demo for the Guinea pitch.** Optional `demo.schoolrise.org` with seeded sample data. Decision before v0.1.0.
- **Encore Cloud for the team's own dev environments.** The OSS framework is enough; Encore Cloud is a separate product (dev environments, preview deploys). Cheap convenience, not required. Decide whether the team uses it.
- **AI surface in v1 vs Phase 2.** SchoolRise is AI-native; `OPENAI_API_KEY` is required at boot. Decide which AI capabilities ship in v1 vs Phase 2. Candidate v1 surface (cheap to implement, high value): item-suggestion in the form builder, rubric drafting for `ESSAY` fields, distractor generation for `MULTIPLE_CHOICE`, free-text response auto-grading against a band rubric. Phase 2: remediation suggestions per low-band student, content recommendation, multi-language item translation.
- **Field types we genuinely ship in v1 vs defer.** All 26 are catalogued; the form builder UI must render and validate every one. v1 minimum: Text + Choice + Scale & Rating + Date & Time categories (~14 types) + `MATCHING`, `FILL_IN_BLANK`, `ORDERING`, `ESSAY` from Assessment. Confirm whether `HOTSPOT`, `EQUATION`, `CODE_BLOCK` are in v1 or Phase 2 (they're the most complex renderers).
- **AI provider.** Plan defaults to OpenAI (`gpt-4o-mini`) for cost/quality balance. `ANTHROPIC_API_KEY` is supported as a fallback. Decide whether to add an LLM router (e.g., LiteLLM) or keep direct SDK calls for simplicity.
