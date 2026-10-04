# Infra Hub Legacy Pages Deep Content Rewrite

**Date**: February 13, 2026
**Type**: Content
**Components**: Documentation

## Summary

Deep content pass on the 6 remaining Infra Hub legacy pages. Rewrote 3 pages from scratch (infra-jobs, flow-control, getting-started), deleted 3 redundant pages (what-is-an-infra-job, a second catalog page, credentials-and-mappings), and updated cross-references across 4 files. Infra Hub section consolidated from 13 pages to 10, with all pages now at quality bar.

## Problem Statement / Motivation

Six Infra Hub pages remained untouched from the original documentation — the oldest, most marketing-heavy pages in the docs site. They contained fabricated YAML manifests, invented CLI commands, "Planton" branding, author blocks, marketing language, and extensive unverified claims.

### Pain Points

- `what-is-an-infra-job.md` and `infra-jobs.md` were two separate pages covering the same concept (620 combined lines), with the Information Architecture explicitly calling for a merge
- A second catalog page (665 lines) duplicated the already-rewritten `catalog-kinds.md` — both covered the catalog taxonomy and browsing experience
- `credentials-and-mappings.md` (633 lines) duplicated the Connect section (8 pages) — credential management, environment mappings, and default connections were already documented
- `getting-started.md` (463 lines) contained fabricated YAML manifests with invented `apiVersion`/`kind` values, fake CLI commands, and fake resource outputs
- `flow-control.md` (617 lines) contained fabricated YAML examples and unverified policy hierarchy claims

## Solution / What's New

### Pages Rewritten (3 clean-slate)

**`infra-jobs.md`** — Merged two pages into one comprehensive reference. Covers: what an Infra Job is, execution sequence (init/refresh/preview/apply), the four essentials that get resolved before execution (IaC module, provider credentials, state backend, flow control), two deployment paths (direct vs orchestrated), monitoring, controlling execution (pause/cancel/rerun), preflight checks, and full CLI reference with verified flags.

**`flow-control.md`** — Documented the five boolean controls using exact labels from the web console component (`flow-control-display.tsx`): Manual Approval Required, Lifecycle Events Disabled, Skip Refresh, Preview Before Apply, Pause After Preview. Documented the four-level resolution hierarchy (resource > environment > organization > platform, first match wins, no merging). Added three practical patterns (development, production, shared infrastructure).

**`getting-started.md`** — Replaced 463-line fabricated tutorial with a concise 5-step orientation page (68 lines). Each step links to the relevant deep-content page rather than duplicating content with invented examples.

### Pages Deleted (3 redundant)

- `what-is-an-infra-job.md` — Content merged into the new `infra-jobs.md`
- The second catalog page — Content covered by `catalog-kinds.md`
- `credentials-and-mappings.md` — Content covered by the Connect section

### Cross-References Fixed (4 files)

- `infra-hub/index.md` — Removed the second catalog page's and Credentials and Mappings entries, updated Mermaid diagram, updated Getting Started links
- `infra-hub/catalog-kinds.md` — Removed the second catalog page's link from Related Documentation
- `infra-hub/infra-components.md` — Pointed its catalog link at catalog-kinds
- `infra-hub/openmcf.md` — Pointed its catalog link at catalog-kinds

## Implementation Details

### Source Verification

**Infra Jobs page** verified against:
- `apis/ai/planton/infrahub/infrajob/v1/` — `InfraJobOperationType` enum, `InfraJobSpec`, `InfraJobEssentials`, preflight checks
- ADR `2026-01-18-085121-redesign-infrajob-essentials-resolution.md` — Essentials resolution simplification
- CLI: 11 subcommands from Go source (`create-infra-job`, `cancel`, `resume`, `rerun`, `list`, `preflight-checks`, `stream-progress-events`, `stream-status`, `iac-input`, `iac-execute-input`, `execute`) with verified flags
- Web console: Infra Job detail page, log streaming component, flow control display
- Backend README: `backend/services/infra-hub/_module/src/main/java/ai/planton/infrahub/domain/infrajob/README.md`

**Flow Control page** verified against:
- `apis/ai/planton/infrahub/flowcontrolpolicy/v1/` — `InfraJobFlowControl` (5 booleans), `FlowControlPolicySpec` (selector pattern), `FlowControlPolicyQueryController` (getBySelector RPC)
- `flowcontrolpolicy/v1/README.md` — Policy hierarchy, first-match-wins semantics
- Web console: `flow-control-display.tsx` — exact labels and descriptions for each control

### Key Decisions

- **Merged two pages into one**: `what-is-an-infra-job.md` (introductory) and `infra-jobs.md` (deep dive) had significant overlap; IA explicitly called for merge
- **Deleted rather than redirected**: Three redundant pages deleted entirely rather than converted to thin redirect pages — reduces maintenance burden
- **Getting Started as navigation page**: Concise orientation rather than tutorial — each step links to deep-content pages, avoiding duplication and staleness
- **Flow control labels from web console**: Used exact labels from `FlowControlDisplay` component for user-facing consistency

## Benefits

- **Infra Hub section reduced from 13 to 10 pages** — every remaining page at quality bar
- **~2,400 lines of fabricated/marketing content removed** across 6 legacy pages
- **~340 lines of verified, source-backed content** across 3 rewritten pages
- **0 broken cross-references** verified across all docs
- **0 duplicate content** between Infra Hub and other sections (Connect, Catalog Kinds)

## Impact

- All Infra Hub pages now at the quality bar established by the Connect section exemplar
- Infra Jobs documentation matches actual CLI commands and web console behavior
- Flow Control documentation uses exact web console labels for consistency
- Getting Started page links to deep-content pages rather than duplicating with fabricated examples
- Readers no longer encounter two pages about the same concept (Infra Jobs) or three locations for credential documentation

## Related Work

- Phase 2 (Session 2): Created the Infra Hub pages that these legacy pages are now deleted in favor of
- Phase 4 (Session 4): Created the Connect section that supersedes credentials-and-mappings
- Phase 4.5 (Sessions 9-10): Retrospective audit of other Infra Hub pages that established the quality bar

---

**Status**: Live
**Timeline**: Single session
