# Planton Docs Overhaul Phase 4.5b: Retrospective Audit of Medium-Priority Pages

**Date**: February 13, 2026
**Type**: Content
**Components**: Documentation

## Summary

Retrospective audit and rewrite of 6 medium-priority documentation pages to eliminate protobuf leakage, add "why" context, remove system-level implementation details, and bring all pages to the quality bar established by the Connect section. This is part of the ongoing documentation overhaul to achieve feature parity between Planton's documentation and its platform capabilities.

## Problem Statement / Motivation

Six documentation pages written in the project's early sessions (Sessions 1-3) contained violations against the revised documentation philosophy established in Session 4:

- **Protobuf leakage**: Internal field names, message types, and enum values exposed to users who never interact with these identifiers
- **Missing "why" context**: Features described without explaining why they exist or what problem they solve
- **System-level details**: Architecture internals that don't help users accomplish their goals
- **Unresolved TODO comments**: 5 TODO markers left from earlier sessions

### Pain Points

- `infra-charts.md` exposed `InfraChartParam`, `ValueFromRef`, and `ApiResourceSelector` in user-facing prose
- `infra-stacks.md` had field name tables that mirrored protobuf message definitions
- `what-is-a-service.md` listed protobuf field names as table column headers
- `build-methods.md` referenced internal image construction patterns
- `monorepo-support.md` named protobuf message types directly in section headers
- `deployment-targets.md` used `ServiceEnvironmentDeploymentTarget` in descriptions

## Solution / What's New

### Clean-Slate Rewrites (2 pages)

**infra-hub/infra-charts.md** — Rewritten from scratch with a three-problem motivation opening (repetitive patterns, hidden dependencies, no standardization), real-world AWS ECS chart example, and CLI commands verified against Go source.

**infra-hub/infra-stacks.md** — Rewritten from scratch with persistent-record motivation, simplified lifecycle operations, and corrected CLI paths (install lives under `chart`, not `infra-stack`).

### Targeted Rewrites (4 pages)

**service-hub/what-is-a-service.md** — Added developer-pain "why" opening, replaced field name tables with capability descriptions, translated enum values to user-facing language.

**service-hub/build-methods.md** — Added convenience-vs-control framing, simplified image tagging section, removed internal construction patterns.

**service-hub/monorepo-support.md** — Added first-class-feature "why" opening, removed `ServiceGitRepo` and `ServicePipelineConfiguration` message type names, rephrased section headers to user-facing language.

**service-hub/deployment-targets.md** — Added historical "why" context from ADR (user feedback led to inline model), removed message type names, replaced field name table with capability descriptions.

## Implementation Details

- All CLI commands verified against Go source at `client-apps/cli/`
- Product documentation (`_team/agents/_knowledge-base/context/product/`) used as primary source for "why" context — elevated to Tier 2b reference source
- ADR `2026-01-17-145248-ui-based-service-deployment-configuration.md` provided historical context for deployment targets
- YAML field names preserved in code blocks where users write them; only prose references translated

## Benefits

- **Zero protobuf leakage**: No internal identifiers remain in user-facing prose across all 6 pages
- **Motivated documentation**: Every page now opens with why the feature exists and what problem it solves
- **Verified accuracy**: All CLI commands match actual Go source code (2 corrections made)
- **Clean pages**: All 5 TODO comments resolved

## Impact

- 6 documentation pages improved, covering Infra Hub (charts, Infra Stacks) and Service Hub (services, build methods, monorepo support, deployment targets)
- Approximately half of the Session 1-3 pages have now been audited (high-priority in Phase 4.5, medium-priority in this phase)
- Lower-priority pages remain acceptable given their contexts (advanced topics, architectural focus)

## Related Work

- Phase 4.5 (CP09): Retrospective audit of 4 high-priority pages
- Phase 4 (CP04): Guidelines evolution that established the revised philosophy
- Connect section (CP04): Quality exemplar that the audited pages are measured against

---

**Status**: Live
**Timeline**: Single session
