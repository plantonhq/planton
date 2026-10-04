# Service Hub Legacy Pages Deep Content Rewrite

**Date**: February 13, 2026
**Type**: Content
**Components**: Documentation

## Summary

Rewrote the three remaining Service Hub legacy pages — `getting-started.md`, `deployment-stage.md`, and `deployment-environments.md` — to bring all 50 documentation pages to the project quality bar. This completes the Week 1+2 documentation overhaul: every page across all 8 sections now meets the established standard.

## Problem Statement / Motivation

Three Service Hub pages remained untouched from the original documentation. They contained:

- "Planton" references throughout (should be "Planton")
- Marketing language ("magic", "fun part", "the club")
- Fabricated CLI commands that don't exist (`planton service create`, `planton service logs`, `planton service status`, `planton service url`, `planton service webhook-status`)
- Invented YAML manifests with unverified `apiVersion`/`kind` values
- Author attribution blocks
- Protobuf code blocks in user-facing documentation
- Transcript and implementation path citations
- Repetitive pattern examples that padded page length without adding value

### Pain Points

- `getting-started.md` was a 316-line fabricated tutorial — six of its CLI commands don't exist in the codebase
- `deployment-stage.md` overlapped heavily with already-rewritten `deployment-targets.md` and `pipelines.md`
- `deployment-environments.md` had real content buried under marketing prose and repetitive examples

## Solution / What's New

### getting-started.md — Clean-Slate Rewrite

Replaced the fabricated tutorial with a concise 76-line navigation page following the established pattern from `infra-hub/getting-started.md`. Each step links to the deep-content page that already documents the topic in full:

1. Connect Git provider → [Git Providers](/docs/connect/git-providers)
2. Create service → [What is a Service?](/docs/service-hub/what-is-a-service)
3. Monitor pipeline → [Pipelines](/docs/service-hub/pipelines)
4. Push code → natural iteration loop
5. Explore next → links to all Service Hub deep-content pages

### deployment-stage.md — Clean-Slate Rewrite

Rewrote as a focused deployment execution deep dive (207 lines). Establishes clear scope boundaries with already-rewritten pages:

- **This page owns**: Manifest resolution paths, Kustomize model in depth, local overlay for `.env` generation, deployment task execution flow
- **Pipelines owns**: High-level pipeline model, triggers, cancellation
- **Deployment Targets owns**: Supported platforms, Git-based vs inline choice

Key content: dual manifest resolution paths (Git-based Kustomize vs inline spec), base + overlay patching model with real YAML examples, `metadata.env` field mapping, `$secrets-group/` and `$variables-group/` reference syntax, local overlay purpose, and verified CLI commands.

### deployment-environments.md — Targeted Rewrite

Tightened from 523 lines to 218 lines. Preserved the real content while eliminating quality issues:

- **Preserved**: Three inline images, Cloudflare Stream video embed, git-branch label documentation with glob patterns, precedence rules, troubleshooting section
- **Removed**: Author block, "Planton" references, marketing introduction, protobuf code block, transcript citation, "What You'll Learn" filler section
- **Consolidated**: Four repetitive pattern sections (Real-World Use Cases, Common Patterns, Environment Groups, Gradual Rollout) reduced to three focused examples

## Implementation Details

### Source Verification

All content verified against:

- `ServiceSpec` proto — `deployment_environments`, `deployment_config_source`, `deployment_targets` fields
- `ServicePipelineConfiguration` proto — `kustomize_base_directory` field
- `PipelineDeploymentStage` and `PipelineDeploymentTask` proto — task structure, env field, manual gates
- CLI Go source — `planton service dot-env` (flags: `--version`, `--set`), `planton service kustomize init` (flag: `--new`), `planton service kustomize build`, `planton service deploy` (flags: `--project`, `--version-message`, `--set`)
- Web console — `WizardOrchestrator.tsx` (10-step wizard), `ServiceDeploymentEnvironments` component (modal, checkbox list)
- ADR: UI-based service deployment configuration (dual-path logic, `DeploymentConfigSource` enum)
- Product docs: `what-is-the-role-of-kustomize-in-service-hub.md` (why Kustomize was chosen)

### Cross-Reference Verification

- 0 broken links across all 50 documentation pages
- All internal cross-references from/to the three rewritten pages verified
- Only remaining "Planton" reference in docs is in `cli.md` (outside Phase 13 scope)

## Benefits

- All 50 documentation pages now meet the established quality bar
- No fabricated CLI commands remain in any documentation page
- No "Planton" references remain in Service Hub section
- No author attribution blocks remain in any documentation page
- No protobuf code blocks remain in user-facing documentation
- No transcript citations remain in any documentation page

## Impact

### Pages Changed

| File | Action | Lines Before | Lines After | Net |
|------|--------|-------------|-------------|-----|
| `service-hub/getting-started.md` | Clean-slate rewrite | 316 | 76 | -240 |
| `service-hub/deployment-stage.md` | Clean-slate rewrite | 403 | 207 | -196 |
| `service-hub/deployment-environments.md` | Targeted rewrite | 523 | 218 | -305 |
| **Total** | | **1,242** | **501** | **-741** |

### Documentation Milestone

This session completes the Planton documentation overhaul:
- **50 pages** across 8 sections, all at quality bar
- **0 broken links** across the entire documentation site
- **0 pages** with marketing language, fabricated content, or "Planton" references (except `cli.md`)

## Related Work

- Phase 12: Infra Hub legacy pages deep content (`_changelog/2026-02/2026-02-13-194719`)
- Phase 11: Platform section deep content (`_changelog/2026-02/2026-02-13-174032`)
- Phase 4.5b: Retrospective audit (`_changelog/2026-02/2026-02-13-153447`)

---

**Status**: Live
**Timeline**: Session 15 of the documentation overhaul project (Week 2, Day 1)
