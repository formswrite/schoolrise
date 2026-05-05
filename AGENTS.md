# AGENTS.md

Instructions for AI agents and human contributors working in this repo.

## Code style — no comments, ever

**Do not write comments in code. This is a hard rule, not a guideline.**

- No `//` line comments in Go, TS, JS, or any language that supports them
- No `/* ... */` block comments
- No `#` comments in YAML, Dockerfile, shell, Python, etc., unless the line is a directive that the language requires (e.g. `#!/usr/bin/env bash` shebang, `# syntax=docker/dockerfile:1` directive)
- No JSDoc / GoDoc / docstrings on functions, types, or packages
- No "TODO" / "FIXME" / "XXX" markers
- No section banners like `// ─── Frontend ───`
- No commented-out code

**Encore framework annotations are not comments.** `//encore:api`, `//encore:service`, `//encore:authhandler` are required compiler directives — keep them. Same for build tags (`//go:build linux`), `// +build` lines, and shebangs.

**SQL comments (`-- ...`) follow the same rule** — do not write them. Migration files contain SQL only.

**JSON cannot have comments at all.** Some configs (`encore.app`, `infra-config/*.json`) are strict JSON.

**Reasoning lives elsewhere.** Architectural reasoning belongs in `PLAN.md`, the customer spec, or commit messages. Code uses well-named identifiers to communicate intent.

## TDD — tests first, always

**All development follows test-driven development. This is a hard rule, not a guideline.**

The cycle for every change:

1. **Red** — write a failing test for the next slice of behaviour. The test must compile and run, and it must fail because the implementation does not exist yet (or behaves wrongly). A failing test that fails for the wrong reason (typo, missing import) does not count.
2. **Green** — write the minimum implementation that makes the test pass. Resist the urge to write code beyond what the test requires.
3. **Refactor** — once green, clean up implementation and test code without changing observable behaviour. Tests stay green throughout.

Apply this to every new endpoint, every new domain function, every new schema concern. Edge cases and validation rules are added as **new failing tests first**, then made to pass — not by patching implementation defensively.

**Test placement.** Unit tests live next to the code they test (`auth/sessions_test.go` next to `auth/sessions.go`). Integration tests that need Encore's local infra (Postgres, pubsub) use Encore's test runner via `encore test ./...` and run inside the same package. Cross-service end-to-end tests live under `tests/e2e/` and target the running gateway.

**Coverage target:** 80% on domain logic per PLAN.md §10. Coverage is measured but the rule that wins arguments is "did the failing test come first?". Coverage without TDD is theatre.

**No skipped or `t.Skip` tests in main.** A test that does not run is worse than no test — it gives false confidence. If a test has to be skipped temporarily, open an issue, link it from the test, and put a deadline on the skip.

**Test naming.** `TestXxx` for top-level tests. Subtests via `t.Run("descriptive lowercase phrase", ...)`. Table-driven tests use field names `input` and `expected` (per `samber/cc-skills-golang@golang-naming`).

**No mocks of internal code.** Mock at system boundaries only (Resend API, OpenAI API, S3). Use real Postgres in integration tests via Encore's test runner — mocking the database hides bugs (PLAN.md §10.5 risk).

## Repo conventions

- Module path is `encore.app` (Encore.go convention for self-host) — service packages are imported as `encore.app/apps/<service>`
- All eleven Encore services live under `apps/` (`apps/auth/`, `apps/tenancy/`, ..., `apps/ai/`) alongside the SvelteKit frontend at `apps/web/`. The `apps/` folder is the home of every deployable. Encore discovers services by scanning all packages in the module — the directory path does not matter to Encore; the service name comes from the package declaration (`package auth`, etc.) so `--services=auth` still works
- `pkg/` (at the repo root) contains plain Go shared code (no `//encore:*` decorators)
- Migrations are up-only by Encore's default; document down-procedure in `docs/operations/migrations.md`
- Service tests run via `encore test ./...`, not `go test`. Plain `go test` panics on `sqldb.NewDatabase` outside Encore runtime. `pkg/` tests run via plain `go test ./pkg/...`.

## Frontend stack — SvelteKit (overrides PLAN.md §5)

PLAN.md §5 specifies Next.js 15. That is **superseded** by this AGENTS.md entry. SchoolRise uses:

- **SvelteKit 2** (Svelte 5 runes) under `apps/web/`
- **TypeScript** strict mode
- **Tailwind CSS v4**
- **shadcn-svelte** for component primitives (Svelte port of shadcn/ui)
- **paraglide-js** for type-safe i18n; English (`en.json`) is canonical, French / Wolof / Malinké / Soussou / Portuguese ship as locale packs
- **Vitest** for unit tests, **Playwright** for e2e
- **`@sveltejs/adapter-node`** for the self-host Docker image
- File-based routing under `src/routes/` (e.g. `/admin/institutions` → `src/routes/admin/institutions/+page.svelte`)
- Auth bridge via `src/hooks.server.ts` — sets `event.locals.user` from a signed Encore session cookie

**Reference codebase ports** (PLAN.md §12) need this adjustment:

- `formswrite-frontend/lib/field-types.js`, `grading.ts`, `logic.js` — **pure logic, ports cleanly to TS in SvelteKit**
- `formswrite-frontend/components/form-editor/FormEditorLayout.js`, `QuickFieldPicker.js` — these are React components. They are **inspiration, not source-to-port**. Re-implement in Svelte preserving the three-panel layout and 26-field-type picker UX.

## When making changes

- Read `PLAN.md` first — it is the source of truth for scope, architecture, customer mapping, and phasing
- Do not introduce dependencies beyond what `PLAN.md` calls for (Encore, Postgres, NSQ, Redis, Resend, OpenAI, Sentry)
- Do not add modules 7–10 from `analysius` until the customer signs scope expansion
