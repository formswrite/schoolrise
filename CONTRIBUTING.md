# Contributing to SchoolRise

Thank you for considering a contribution. Read this document before opening a pull request.

## Before you start

1. Read **`PLAN.md`** — it is the source of truth for scope, architecture, customer mapping, and the v0.1.0 release contract.
2. Read **`AGENTS.md`** — coding conventions (no comments rule, TDD rule, repo conventions, frontend stack) apply to every contribution. These rules are not optional.
3. Read **`README.md`** for setup instructions.

## Contributor License Agreement

By submitting a pull request you agree that your contribution is licensed under AGPL-3.0 and that the SchoolRise maintainers may relicense it under a commercial license alongside AGPL. The CLA bot will block your first PR until you sign.

This preserves the project's ability to dual-license — ministries self-hosting are unaffected, and commercial vendors who want to embed SchoolRise in proprietary products can purchase a commercial exemption.

## Development workflow

Two paths — pick one based on whether you need the BAML-backed AI features (`apps/ai/`).

### Path A — Full Docker stack (Recommended; covers all 12 services)

```bash
git clone https://github.com/formswrite/schoolrise.git
cd schoolrise
cp .env.example .env
# fill in the four __choose_a_local_*__ placeholders in .env
make compose-up-local
```

Open <http://localhost:3001> (SvelteKit web), <http://localhost:8080> (Encore API gateway), <http://localhost:9001> (MinIO console).

This path uses the prebuilt `schoolrise-app:patched` image. If you don't have it cached, you'll need to build it once — see **[docs/encore-go-build.md](docs/encore-go-build.md)** for the encore-go fork build steps.

### Path B — Native `encore run` (faster reloads; no BAML)

```bash
git clone https://github.com/formswrite/schoolrise.git
cd schoolrise
cp .env.example .env
go mod tidy
encore run
```

In a separate shell:

```bash
cd apps/web
npm install
npm run dev
```

Open <http://localhost:9400> for the Encore dev dashboard, <http://localhost:5173> for the SvelteKit app.

Path B uses the upstream `encore` CLI, which means the patched static-link fix is not applied. Anything that imports BAML will fail to start. Stick to Path A if you're touching `apps/ai/`.

## The encore-go fork

SchoolRise depends on a 6-line patch to Encore's CLI to disable static linking (BAML's cgo runtime is incompatible with the default). If you contribute to anything that uses BAML or rebuilds the Docker image, read **[docs/encore-go-build.md](docs/encore-go-build.md)**. Once GitHub Actions builds + pushes the image to GHCR (production-roadmap item #9), most contributors won't need to do this manually.

## TDD is mandatory

Every behavioural change follows red-green-refactor:

1. Write a failing test for the next slice of behaviour.
2. Make the minimum implementation that turns the test green.
3. Refactor without changing observable behaviour, keeping tests green.

Pull requests that add behaviour without a failing test first will be sent back for revision. See `AGENTS.md` § *TDD* for placement, naming, and integration-test rules.

## No comments

Code does not contain comments. Reasoning lives in `PLAN.md`, commit messages, or markdown docs. Encore framework directives (`//encore:api`, `//encore:service`), build tags, linter directives (`//nolint:...`), and shebangs are not comments — keep them. See `AGENTS.md` § *Code style* for the full rule.

## Branch and commit style

- Work on a topic branch off `main`. Branch name: `<type>/<short-slug>` where `<type>` is `feat`, `fix`, `chore`, `docs`, `test`, `refactor`, or `ci`.
- Commit messages are short imperative summaries on the first line. The body, when present, explains *why* not *what*. Reference `PLAN.md` sections or customer-spec sections by name when relevant.
- Keep PRs scoped — one concern per PR. Mixed concerns are harder to review and harder to revert.

## Required checks before requesting review

```bash
make lint
make test
cd apps/web && npm run check && npm run lint && npx playwright test
```

CI runs these plus `gitleaks` (secrets), `govulncheck` (Go CVE scan), CodeQL (Go + JS), and the per-service Docker build on tag pushes.

## Pre-commit hook (gitleaks)

The repo ships a `.pre-commit-config.yaml` with gitleaks. Install once:

```bash
pip install pre-commit
pre-commit install
```

Every commit then scans staged changes for accidental secret leaks. The CI security workflow runs the same scan on every PR — so even if you skip the local hook, leaked credentials are caught before merge.

## Go style — Encore-flavored exceptions

The codebase mostly follows the conventions in `samber/cc-skills-golang` (linted via `.golangci.yml`). Two exceptions are deliberate Encore-flavored deviations from pure Go style:

1. **`XxxAPI` suffix on Encore handler methods.** Functions like `(s *Service) CreateUserAPI(...)` and `(s *Service) GetFormAPI(...)` keep the `API` suffix. Without it, Encore's codegen produces a wrapper at package level with the same name as the method, which collides with internal business-logic functions sharing the root name (e.g., `apps/auth/users.go::CreateUser` and the method-named wrapper would both be at package level). The `API` suffix disambiguates the HTTP-layer endpoint from the underlying internal function. **Do not strip it.**

2. **`exported` and `package-comments` revive rules are disabled.** The skill's recommended baseline requires godoc comments on every exported symbol. The codebase has a deliberate no-comments policy (reasoning lives in `PLAN.md`, commit messages, and markdown docs). Both rules are off in `.golangci.yml::linters.settings.revive.rules`. Don't re-enable.

Outside these two exceptions, Go style follows `samber/cc-skills-golang/golang-naming` and `golang-code-style`: MixedCaps identifiers, sentinel error form `errors.New("pkg: lowercase msg")`, single-letter receivers, no `Get` prefix on getters except for booleans (`Is*`/`Has*`/`Can*`), early-return error handling, ≤ 4 function params (use options structs beyond that), no `reflect`/`unsafe` in business code, etc.

## AGPL implications you should understand

SchoolRise is AGPL-3.0. The "A" matters:

- **If you fork and self-host SchoolRise privately for your ministry**, AGPL imposes nothing visible (the source you started with is already public).
- **If you fork and run a SaaS that lets external users interact with SchoolRise over a network**, AGPL requires you to publish your modifications and offer them to those users. This applies even if you never distribute a binary.
- **If you embed SchoolRise code in a proprietary product**, AGPL prohibits it without a separate commercial license.

For ministries, this means: you can deploy SchoolRise for your country, modify it freely, and your modifications stay private to your team. For commercial vendors who want to ship a proprietary SaaS based on SchoolRise, contact the maintainers about dual-licensing.

## Adding a new Encore service

Rare during Phase 1 — the eleven services in `PLAN.md` cover the customer's six sections. If you need a new one, open an issue first explaining the customer-spec mapping. The framework lets us split a package into its own service later (`encore build docker --services=<name>`) without a rewrite, so the bias is to add a sub-package inside an existing service rather than a brand-new one.

## Adding a new locale

Drop a sibling JSON file alongside `pkg/seed/locales/en.json` and `apps/web/src/locales/en.json` covering the same keys. No code changes required. Open the PR with both files present.

## Adding a new country seed

Copy `examples/seed-template/` to `examples/seed-<country>/` and replace the hierarchy-level rows with the local administrative tiers. The hierarchy levels are seed data, not hardcoded — every country can fork without touching `tenancy/`.

## Issues

- **Bugs**: include reproduction steps, expected behaviour, and the relevant `encore run` log lines.
- **Features**: explain the customer-spec section it serves. Modules 7–10 from `analysius` (exams/certificates/commissions/LMS/careers) are parked until a paying customer expands scope; do not open feature requests for them.

## Code of conduct

See `CODE_OF_CONDUCT.md`. We use the Contributor Covenant.

## Maintainers

The Formswrite team maintains SchoolRise on behalf of partner ministries. For commercial inquiries (dual-license, paid support, hosted demos), email the address listed in `SECURITY.md`.
