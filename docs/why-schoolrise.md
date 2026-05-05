# Why SchoolRise

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

## The two structural differences

1. **First-class assessment-authoring layer** — form builder with conditional logic, campaigns scoped to admin tiers, immutable form-version snapshots, sub-50 ms region-level dashboards via precomputed snapshots.
2. **First-class AI layer** — type-safe LLM contracts via [BAML](https://github.com/BoundaryML/baml). Inspectors author items in natural language; the system drafts rubrics, generates distractors for multiple-choice, and auto-grades essays against a rubric. Provider-agnostic (OpenAI, Anthropic, or local models).

## What SchoolRise is *not*

- Not a Firebase-style BaaS — it's a vertical EMIS, not a platform you build apps on top of.
- Not a replacement for OpenEMIS's records modules; ministries already deployed on those can adopt SchoolRise as the assessment + AI layer alongside.
- Not multi-tenant SaaS — each ministry self-hosts its own instance for data residency.
