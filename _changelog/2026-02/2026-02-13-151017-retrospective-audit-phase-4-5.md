# Retrospective Audit: Phase 4.5 Documentation Quality Rewrite

**Date**: February 13, 2026
**Type**: Enhancement
**Components**: Documentation

## Summary

Rewrote 4 high-priority documentation pages from Sessions 1-3 to meet the revised documentation philosophy established during the Connect section work (Session 4). Eliminated protobuf leakage from user-facing prose, added missing "why" context from ADRs and co-located documentation, removed system-level implementation details, resolved all TODO comments, corrected CLI command paths against Go source code, and fixed a broken link to a deleted page.

## Problem Statement / Motivation

The documentation pages written in Sessions 1-3 were correct under the original guidelines, which prioritized source code accuracy. However, the guidelines evolved significantly in Session 4 — introducing the "no protobuf leakage" rule, the three-tier source truth model, and the principle that documentation is for users, not implementors.

### Pain Points

- **Protobuf field names in prose**: Pages exposed internal identifiers like `PipelineDeploymentTask`, `ServicePipelineConfiguration`, `CatalogKind` enum names, `image_build_failure_analysis`, and `progress_status` — none of which users ever type or see
- **Missing motivation**: Pages jumped into "what" and "how" without explaining "why" — Infra Components existed with no explanation of the problem they solve; Infra Pipelines lacked context for why DAG-based orchestration matters
- **System-level internals**: Pulumi stack naming conventions, Terraform state paths, ID format patterns (`ic_<prefix>_<ulid>`), DAG execution algorithms, and Tekton PipelineRun deletion mechanics — implementation details that don't help users accomplish tasks
- **Unresolved TODO comments**: 9 TODO comments across the 4 pages, most referencing internal proto files or backend logic
- **Incorrect CLI commands**: Pipeline commands used `planton pipeline cancel` instead of the correct `planton service pipeline cancel` path

## Solution / What's New

### Pages Rewritten (Clean-Slate)

**`service-hub/pipelines.md`** — Restructured around the user's journey: why pipelines exist, how a pipeline works (two stages), when pipelines trigger, controlling behavior, manual approval gates, cancellation, and monitoring. All protobuf field names replaced with plain language. CLI commands corrected to verified paths (`planton service pipeline` subcommand tree). Tekton and GitHub API internals removed.

**`infra-hub/infra-components.md`** — Led with why Infra Components exist (unified interface across providers), described what an Infra Component represents from the user's perspective, simplified lifecycle to four user-facing operations (create, update, destroy, purge), added web console workflow descriptions from actual UI components. Removed KRM structure details, state management internals, and all 5 TODO comments.

**`infra-hub/catalog-kinds.md`** — Reframed entirely around the Infra Catalog (what users actually browse). Removed enum names, ID prefix metadata, and IaC module mapping internals. Added "Service-Deployable Kinds" section explaining which types can be deployment targets. Removed the "Extending the Catalog" implementation details. Resolved all 4 TODO comments.

**`infra-hub/infra-pipelines.md`** — Added concrete Mermaid diagram showing VPC-to-RDS dependency flow. Added "Why Infra Pipelines Exist" section with the 25-minute-to-12-minute parallel execution motivation from ADRs. Corrected manual gate CLI commands (`resolve-env-manual-gate`, not `resolve-environment-manual-gate`). Kept the Pipeline vs Direct Deployment comparison table (user-useful). Removed DAG algorithm internals and status field names.

### Broken Link Fixed

**`service-hub/what-is-a-service.md`** — Updated the "Related Documentation" link from the deleted `secrets-and-variables` page to the new `secrets-and-config` section.

## Implementation Details

### Source Verification

Each rewrite was verified against:
- **CLI commands**: Go source in `client-apps/cli/cmd/planton/root/domain/` — exact command names, subcommands, and flag names
- **ADRs and README files**: Design rationale for pipeline architecture, Infra Component abstraction, Kind taxonomy, and DAG orchestration
- **Web console kinds**: Pipeline list table columns, status labels, trigger modal, Infra Component creation flow, DAG visualization

### Key Corrections

- Pipeline CLI uses `planton service pipeline <subcommand>`, not `planton pipeline <subcommand>`
- Infra pipeline manual gate commands are `resolve-env-manual-gate` (with env-name argument) and `resolve-node-manual-gate` (with env-name and node-id arguments)
- Infra component creation uses `planton create -f manifest.yaml` (generic create command), not `planton create catalog object manifest.yaml`
- `stream-status` for infra pipelines exists in code but is not registered as a subcommand

## Benefits

- **No protobuf leakage**: Zero internal field names, message types, or enum values in user-facing prose across all 4 pages
- **User motivation**: Every page now opens with "why this exists" — the problem it solves and why the reader should care
- **Verified CLI**: All CLI examples confirmed against Go source code
- **Zero TODOs**: All 9 TODO comments resolved or removed
- **Consistent quality**: Pages now match the quality bar established by the Connect section (Session 4 exemplar)

## Impact

- 5 files changed: 240 insertions, 314 deletions (net reduction of 74 lines — tighter, more focused content)
- 4 high-priority audit violations resolved (from the retrospective-audit-criteria.md scope)
- 1 broken link fixed (prevents 404 for readers navigating from what-is-a-service.md)
- 6 medium-priority pages identified for future audit session

## Related Work

- Phase 1 (Session 1): Initial quality fixes across 25 pages — `checkpoints/CP01_phase1_quality_fixes.md`
- Phase 4 (Session 4): Connect section + guidelines evolution — established the quality bar this audit enforces
- Retrospective audit criteria: `coding-guidelines/retrospective-audit-criteria.md` — the scope document for this work
- Future: Phase 4.5b could address the 6 medium-priority pages (infra-charts, infra-stacks, what-is-a-service, build-methods, monorepo-support, deployment-targets)

---

**Status**: Live
**Timeline**: Single session
