# Service Hub Documentation Overhaul

**Date**: February 13, 2026
**Type**: Content
**Components**: Documentation

## Summary

Rewrote 2 existing Service Hub documentation pages and created 6 new pages, bringing the Service Hub section from 7 pages of marketing-toned content to 13 pages of source-verified, technically precise documentation. Every claim is traceable to protobuf API definitions, CLI source code, or web console kinds.

## Problem Statement / Motivation

The existing Service Hub documentation had critical quality defects that undermined user trust and created confusion:

### Pain Points

- **Fabricated CLI commands**: `planton pipeline logs`, `planton pipeline status`, and `planton pipeline rerun --debug` were documented but don't exist in the CLI codebase. The actual commands are `planton pipeline stream-logs`, `planton pipeline stream-status`, and `planton service rerun-pipeline`.
- **Incorrect field names**: `tekton_pipeline_yaml_directory` was documented, but the actual protobuf field is `tekton_pipeline_yaml_file`.
- **Marketing tone in technical docs**: "Magic Happens", "Auto-Magic", emoji checkmarks, casual FAQ sections — all violating the platform's docs guidelines.
- **Unverified claims**: Azure Container Instances listed as a deployment target (no `is_service_kind` entry exists), specific resource allocation numbers (2-4 cores, 4-8GB RAM), and a notifications feature that doesn't exist in the codebase.
- **Missing feature coverage**: Build methods, self-managed pipelines, deployment targets, secrets/variables, ingress, and monorepo support had zero dedicated pages.

## Solution / What's New

### Pages Rewritten from Scratch

**`what-is-a-service.md`** (310 lines down to 162 lines)
Focused on the Service as a configuration bridge with four pillars: where code lives (GitRepo), how to build (PipelineConfiguration), where to deploy (deployment targets with git/kustomize vs inline models), and how to reach it (Ingress). Removed all marketing language, the FAQ section, and detailed build/monorepo/ingress content that now lives in dedicated pages.

**`pipelines.md`** (365 lines down to 218 lines)
Correct three-stage pipeline model (Creation, Build, Deploy) with verified CLI commands, accurate `PipelineDeploymentTask` structure, and proper cancellation behavior documentation from the command.proto comments.

### 6 New Pages Created

| Page | Lines | Key Content |
|------|-------|-------------|
| `build-methods.md` | 120 | Buildpacks vs Dockerfile, `ImageBuildMethod` enum, image tagging convention, Cloudflare Worker builds |
| `self-managed-pipelines.md` | 150 | `PipelineProvider` enum, platform-injected params, custom params, TektonPipeline/TektonTask registry, web console editor |
| `deployment-targets.md` | 175 | `DeploymentConfigSource` enum (git vs inline), 5 wizard kinds across 4 providers, `ServiceEnvironmentDeploymentTarget`, manual approval gates |
| `secrets-and-variables.md` | 145 | SecretsGroup and VariablesGroup as API resources, `literal_or_ref` oneof, `ValueFromRef`, reference syntax, full CLI command set |
| `ingress.md` | 85 | `ServiceIngressConfiguration`, DnsDomain resource, when to enable/disable — honest about scope |
| `monorepo-support.md` | 150 | `project_root`, `trigger_paths`, `sparse_checkout_directories`, `kustomize_base_directory`, practical monorepo patterns |

## Implementation Details

### Source Verification

All content verified against four layers of source truth:

1. **Protobuf APIs**: `servicehub/service/v1/` (spec.proto, enum.proto, api.proto, command.proto), `servicehub/pipeline/v1/`, `servicehub/secretsgroup/v1/`, `servicehub/variablesgroup/v1/`, `servicehub/dnsdomain/v1/`, `servicehub/tektonpipeline/v1/`
2. **CLI source**: `client-apps/cli/cmd/planton/root/domain/servicehub/` — verified every command name, flag, and argument
3. **Web console**: `client-apps/web/console/src/app/resource/service-hub/service/_components/` — verified wizard steps, labels, and form fields from `labels.ts` constants
4. **OpenMCF**: `org.openmcf.shared.catalogkind` for service-deployable kinds, `org.openmcf.shared.foreignkey.v1` for `ValueFromRef`

### Key Discovery: Two Deployment Configuration Models

The `DeploymentConfigSource` enum (`git` vs `inline`) is a fundamental architectural distinction that was undocumented. The git path uses kustomize overlays; the inline path uses the web console wizard. Both produce catalog object manifests. The web console wizard labels these as "Git-Based (GitOps)" and "UI-Based (Configure Here)".

### Files Changed

- **Modified**: `public/docs/service-hub/what-is-a-service.md`, `public/docs/service-hub/pipelines.md`, `public/docs/service-hub/index.md`
- **Created**: `public/docs/service-hub/build-methods.md`, `public/docs/service-hub/self-managed-pipelines.md`, `public/docs/service-hub/deployment-targets.md`, `public/docs/service-hub/secrets-and-variables.md`, `public/docs/service-hub/ingress.md`, `public/docs/service-hub/monorepo-support.md`

## Benefits

- **Accuracy**: Every CLI command, field name, and feature claim is traceable to source code
- **Coverage**: 6 previously undocumented features now have dedicated pages
- **Trust**: Developers reading these docs will find what they expect when they use the platform
- **Maintainability**: Each topic has its own focused page rather than being mixed into a monolithic Services page
- **Consistency**: All pages follow the same structure (What/Why/How/Details) and reference conventions

## Impact

- Service Hub documentation grows from 7 to 13 pages
- 9 files changed: +1,225 lines added, -533 lines removed
- Net improvement: +692 lines of source-verified documentation replacing marketing-toned content
- Sidebar now presents a logical reading order from concepts through advanced topics

## Related Work

- Phase 1 (Session 1): Quality fixes across 25 existing pages — `e1a8186`
- Phase 2 (Session 2): 6 new Infra Hub pages — `7f90f6b`
- Phase 3 (Session 3): This work — `7ed5a36`
- New guideline: `clean-slate-principle.md` — existing docs are not a constraint

---

**Status**: Live
**Timeline**: Phase 3 of the Planton Docs Overhaul project (Week 1)
