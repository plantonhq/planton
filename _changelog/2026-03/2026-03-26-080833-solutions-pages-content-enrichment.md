# Solutions Pages Content Enrichment and Resource Count Correction

**Date**: March 26, 2026
**Type**: Enhancement
**Components**: Solutions Pages, Features Pages, Documentation

## Summary

Enriched all 10 solutions pages with source-verified content using visually distinct kinds (FlowSteps, AnimatedTerminal, CodeTabs, BentoGrid, MetricsStrip, numbered capability blocks, compliance mapping tables), breaking the cookie-cutter "Hero + 6 cards + CTA" pattern. Also corrected stale resource/provider counts across the entire website from the definitive proto source of truth (`catalog_kind.proto`: 361 kinds across 17 providers).

## Problem Statement / Motivation

The solutions pages had a clear quality divide. Three "use case" pages (IDP, Multi-Cloud, Self-Hosted DevOps) used multiple visual techniques and structural variety. The remaining seven pages (4 by-role, 3 by-size) all followed an identical structure — Hero, 6 flat FeatureCards, CTA — with no terminal demos, code snippets, flow diagrams, comparison data, or visual variety. Each persona got the same layout with different words.

Additionally, resource and provider counts were stale and inconsistent across the site: some pages said "150+", others said "350+", and the actual count from `catalog_kind.proto` was 361 across 17 providers.

### Pain Points

- All 7 Tier 2 pages were structurally identical (~124-128 lines each)
- No visual differentiation between personas — a startup founder and an enterprise CISO saw the same layout
- Some FeatureCards had icons, others didn't — inconsistent visual quality
- Resource counts were wrong: "150+" (from stale docs), "350+" (from earlier website copy), actual: 361
- Provider count was wrong: "11+" shown, actual: 17 (including OpenStack, Scaleway, Alibaba Cloud, OCI, Hetzner Cloud)
- Valuable product details from source docs and demo transcripts were not represented on the marketing pages

## Solution / What's New

### Resource Count Correction (9 files)

Updated all instances across the website to match the proto source of truth:
- `350+` / `150+` → `360+` resource types
- `11+` / `10+` → `17` cloud providers
- Expanded the InfraHub hero provider tag list from 11 to all 17 providers

Files: `infra-hub/capabilities.tsx`, `infra-hub/hero.tsx`, `open-source/capabilities.tsx`, `overview/modules-grid.tsx`, `features/infra-hub/page.tsx`, `agents/technology-section.tsx`, `public/docs/infrastructure/infra-components.md`

### Per-Page Enrichment

Each page received one or two new sections using a **different primary visual kind** so no two pages feel the same:

| Page | Added sections | Visual technique |
|---|---|---|
| Developers | "Your workflow with Planton" flow + terminal demo | FlowSteps, AnimatedTerminal |
| Platform Engineers | "You define vs developers self-serve" split + Infra Chart code example | Two-column Card, CodeTabs |
| Engineering Leaders | Governance metrics bar + governance bento grid | MetricsStrip, BentoGrid |
| Startup Founders | "Zero to production" flow + first deployment terminal | FlowSteps, AnimatedTerminal |
| Startups | "What you get out of the box" bento grid | BentoGrid |
| Growing Teams | Metrics bar + 4 numbered capability blocks with details | MetricsStrip, numbered blocks |
| Enterprises | 3-level security posture + compliance mapping table | Numbered cards, compliance table |
| Self-Hosted DevOps | 3-level security posture + secret backend options | Security level cards |
| IDP | "From zero to self-service IDP" flow | FlowSteps |
| Multi-Cloud | No change needed | Already well-structured |

### Missing Icons Fix

Ensured every FeatureCard across all solution pages has an icon — 14 cards across 6 pages were missing icons.

## Implementation Details

### Source Verification

Every piece of new content maps to a verified source:
- **361 resource kinds, 17 providers** — `openmcf/apis/org/openmcf/shared/catalogkind/catalog_kind.proto`
- **3-level security posture** — Thingularity demo notes (March 10, 2026)
- **Secret backend options** — Thingularity demo notes + Connections docs
- **Pipeline stages** — `what-is-service-hub.md`
- **CloudOps (pod access)** — `what-is-planton-runner.md`
- **Infra Chart structure** — `what-is-an-infra-chart.md`
- **Infra Job audit trail** — `what-is-a-infra-job.md`
- **Custom module contract** — Thingularity demo notes

### Kind Reuse

All new sections use existing shared kinds from `src/components/product/shared/` — no new kinds were created:
- `AnimatedTerminal`, `CodeTabs`, `FlowSteps`, `MetricsStrip`, `BentoGrid/BentoItem`, `ScrollReveal`, `StaggerContainer/StaggerItem`

## Benefits

- Each solution page now has a unique visual identity matching the persona it serves
- Resource and provider counts are now accurate and consistent site-wide
- Source-verified content adds credibility without speculation
- Reusing existing kinds means zero new maintenance burden
- Every FeatureCard now has an icon for visual consistency

## Impact

- **10 solution pages** enriched with additional sections
- **9 files** corrected for stale resource/provider counts
- **14 FeatureCards** across 6 pages had missing icons added
- **1 public doc** updated (`infra-components.md`)
- Build passes cleanly with zero lint warnings

## Related Work

- Previous session: [Website Copywriting Overhaul](_changelog/2026-03/2026-03-25-162529-website-copywriting-overhaul.md)
- Previous session: [Website Accuracy and Quality Overhaul](_changelog/2026-03/2026-03-26-070410-website-accuracy-and-quality-overhaul.md)
- Source of truth: `openmcf/apis/org/openmcf/shared/catalogkind/catalog_kind.proto`
- Additional source: `planton/_meetings/2026/2026-03/2026-03-10-160000.thingularity-planton-platform-demo.notes.md`

---

**Status**: ✅ Live
**Timeline**: Single session
