<h1 align="center">SchoolRise</h1>

<p align="center">
  <strong>The AI-native, open-source EMIS for ministries of education.</strong><br>
  A real form builder + assessment campaigns + LLM-assisted authoring, self-hosted with one <code>make</code> command.
</p>

<p align="center">
  <a href="https://github.com/formswrite/schoolrise/blob/main/LICENSE"><img alt="License: AGPL-3.0" src="https://img.shields.io/badge/License-AGPL--3.0-blue.svg"></a>
  <a href="https://github.com/formswrite/schoolrise/actions/workflows/test.yml"><img alt="CI" src="https://github.com/formswrite/schoolrise/actions/workflows/test.yml/badge.svg"></a>
  <a href="https://github.com/formswrite/schoolrise/actions/workflows/security.yml"><img alt="Security" src="https://github.com/formswrite/schoolrise/actions/workflows/security.yml/badge.svg"></a>
  <a href="https://github.com/formswrite/schoolrise/issues"><img alt="GitHub issues" src="https://img.shields.io/github/issues/formswrite/schoolrise"></a>
  <a href="https://github.com/formswrite/schoolrise/commits/main"><img alt="Last commit" src="https://img.shields.io/github/last-commit/formswrite/schoolrise"></a>
  <a href="http://makeapullrequest.com"><img alt="PRs welcome" src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg"></a>
</p>

<p align="center">
  <a href="#quickstart">Quickstart</a> ·
  <a href="docs/architecture.md">Architecture</a> ·
  <a href="docs/why-schoolrise.md">Why SchoolRise</a> ·
  <a href="docs/roadmap.md">Roadmap</a> ·
  <a href="CONTRIBUTING.md">Contributing</a> ·
  <a href="https://github.com/formswrite/schoolrise/issues">Issues</a> ·
  <a href="https://github.com/formswrite/schoolrise/discussions">Discussions</a>
</p>

> ⚠️ **Status: pre-release.** The form editor and renderer are shipped (34 e2e tests, 31 of 32 field types). Production deploy is in roll-out. Track the gaps in [What's shipped](#whats-shipped).

---

SchoolRise is an **Education Management Information System** in the same category as [OpenEMIS](https://www.openemis.org) and DHIS2-Education, with two structural differences: a first-class **assessment-authoring layer** (form builder, campaigns, snapshot dashboards) and a first-class **AI layer** (LLM contracts via [BAML](https://github.com/BoundaryML/baml) for item generation, distractor synthesis, rubric drafting, and essay grading). See **[docs/why-schoolrise.md](docs/why-schoolrise.md)** for the full positioning.

## Quickstart

Prerequisites: Docker Desktop (or compatible), `make`.

```bash
git clone https://github.com/formswrite/schoolrise.git
cd schoolrise
make compose-up-local
```

`make compose-up-local` auto-creates `.env` on first run (generates `AUTH_SECRET` and `POSTGRES_PASSWORD` via `openssl rand`, sets `ADMIN_EMAIL=admin@local.test` and `ADMIN_PASSWORD=ChangeMe123!`). Edit `.env` afterwards if you want different values.

Open <http://localhost:3001/login> and sign in as `admin@local.test` / `ChangeMe123!`. Form editor at `/admin/forms`. Tear down with `make compose-down-local`.

## What's shipped

- [x] **Form editor** — drag-reorder, click-to-edit, inline previews for 31 field types, conditional show/hide rules, validation, grading, per-question translations
- [x] **Public renderer** — KaTeX equations, drag-ordering, fill-in-blank cloze, matching pairs, table grids, country-region cascades, signature pad, file upload
- [x] **Assessment data flow** — form versioning + immutable snapshots, scoped campaigns, score aggregation into precomputed snapshots (1.2 s for 101 K rows)
- [x] **MinIO uploads** — same-origin proxy, no public bucket exposure
- [x] **AI layer** — `SuggestItems`, `DraftRubric`, `GenerateDistractors`, `GradeEssay` via BAML; provider-agnostic
- [x] **Production stack** — multi-compose Caddy + Postgres + MinIO + app + web, one DNS record, automatic Let's Encrypt
- [ ] Auto-grading pipeline for ORDERING / MATCHING / FILL_IN_BLANK / EQUATION (editor + renderer ship; scorer is the gap)
- [ ] First-boot password rotation polish
- [ ] Automated backups (`pg_dump` + MinIO mirror)
- [ ] French i18n + UI chrome translation switcher
- [ ] HOTSPOT field type

Full backlog: **[docs/roadmap.md](docs/roadmap.md)**.

## Architecture

Two compute services + two storage services + one TLS edge: SvelteKit (admin + public renderer) and Encore Go (12 microservices) sit behind Caddy; Postgres via PgBouncer and a private MinIO bucket sit behind them. Detailed walk-through: **[docs/architecture.md](docs/architecture.md)**.

## Tech stack

[Encore.go](https://encore.dev) · PostgreSQL 16 + PgBouncer · [SvelteKit 2](https://kit.svelte.dev) (Svelte 5 runes) · Tailwind 4 + bits-ui · MinIO · [Caddy 2](https://caddyserver.com) · [BAML](https://github.com/BoundaryML/baml) · Playwright + Go test + svelte-check.

## Production deployment

Single-host single-domain Caddy + multi-compose topology. One DNS A record, one `make prod-up`. → **[deploy/README-prod.md](deploy/README-prod.md)**

## Contributing

We welcome contributors who care about education infrastructure or self-hostable government tools. Read **[CONTRIBUTING.md](CONTRIBUTING.md)** (dev-container, encore-go fork build, test policy, AGPL implications), pick an issue tagged `good-first-issue`, fork, branch, PR.

Code of conduct: **[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)**. Security disclosures: **[SECURITY.md](SECURITY.md)**.

## Community & support

- 💬 [GitHub Discussions](https://github.com/formswrite/schoolrise/discussions) — questions, deployment war-stories
- 🐛 [GitHub Issues](https://github.com/formswrite/schoolrise/issues) — bug reports, reproducible regressions
- 🔒 [SECURITY.md](SECURITY.md) — responsible disclosure (do not file public issues for vulnerabilities)
- 🤝 Commercial support / OEM licensing — contact the maintainers if AGPL-3.0 doesn't fit your deployment model

## License

AGPL-3.0 — see [LICENSE](LICENSE). For ministries or organizations that need a different arrangement (proprietary integrations, OEM redistribution), contact the maintainers.

## Acknowledgements

[Encore.dev](https://encore.dev) · [SvelteKit](https://kit.svelte.dev) · [MinIO](https://min.io) · [Caddy](https://caddyserver.com) · [BoundaryML / BAML](https://github.com/BoundaryML/baml).
