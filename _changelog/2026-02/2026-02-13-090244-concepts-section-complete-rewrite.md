# Concepts Section: Complete Rewrite from Scratch

**Date**: February 13, 2026
**Type**: Feature
**Components**: Documentation, API Definitions, CLI, IaC Engines, Module System

## Summary

Rewrote the entire concepts section of the OpenMCF documentation site from scratch -- 9 pages totaling 1,604 lines of new content, all verified against source code. This is Phase 1 of the docs feature parity project, transforming the concepts section from 2 thin pages into 9 deep, interconnected pages that build a developer's complete mental model of OpenMCF.

## Problem Statement / Motivation

The existing concepts section consisted of 2 pages -- a minimal index and an architecture overview -- that covered the framework's surface without depth. Developers evaluating OpenMCF had no documentation explaining how deployment components work, what the KRM manifest model is, how validation catches errors, why dual IaC engines exist, how modules are resolved, or where state is stored.

### Pain Points

- Evaluators could not understand the framework's design philosophy without reading source code
- New users had no conceptual foundation before attempting their first deployment
- Power users had no reference material for module versioning, state backend configuration, or validation rules
- The architecture page mixed explanations with diagrams, covering too many topics at surface level
- No documentation of the 198 component taxonomy across 14 providers

## Solution / What's New

9 pages written from scratch, each answering a single question completely:

| Page | Question | Lines |
|------|----------|-------|
| `index.md` | What is OpenMCF and why does it exist? | ~120 |
| `architecture.md` | How do the pieces fit together? | ~175 |
| `deployment-components.md` | What is a deployment component? | ~210 |
| `manifests.md` | How do I declare what I want? | ~230 |
| `cloud-resource-kinds.md` | What can I deploy, and where? | ~115 |
| `validation.md` | How does OpenMCF catch my mistakes? | ~150 |
| `dual-iac-engines.md` | How does deployment actually happen? | ~140 |
| `module-system.md` | How are IaC modules resolved and managed? | ~120 |
| `state-management.md` | Where does deployment state live? | ~140 |

### Design Principles

- **Fresh start**: Every page designed from scratch -- no incremental improvement on existing content
- **Source code verified**: Every claim traced to protobuf definitions, CLI source, IaC modules, or engine packages
- **Real examples**: YAML manifests from `examples/`, protobuf excerpts from actual component definitions
- **Provider-specific**: Documented the actual provider counts (198 kinds, 14 providers) by counting the `CloudResourceKind` enum

## Implementation Details

### Source Code Verification Map

Each page was verified against specific source files:

- **deployment-components.md**: `apis/org/openmcf/provider/kubernetes/kubernetespostgres/v1/` (all 4 proto files + both IaC directories), `apis/org/openmcf/provider/aws/awss3bucket/v1/` for cross-provider pattern verification
- **manifests.md**: `apis/org/openmcf/shared/metadata.proto`, `internal/cli/iacflags/manifest_source_flags.go`, `internal/cli/iacflags/execution_flags.go`, `examples/*.yaml`
- **cloud-resource-kinds.md**: `cloud_resource_kind.proto` (full 1,101-line enum counted per provider), `cloud_resource_provider.proto`
- **validation.md**: `internal/manifest/manifest_validator.go`, `apis/org/openmcf/shared/foreignkey/v1/foreign_key.proto`, spec.proto validation annotations
- **dual-iac-engines.md**: `iac/pulumi/main.go`, `iac/tf/variables.tf` + `provider.tf`, `stack_input.proto`
- **module-system.md**: `pkg/iac/pulumi/pulumimodule/module_directory.go`, `pkg/iac/tofu/tofumodule/module_directory.go`, `internal/cli/staging/staging.go`, `cmd/openmcf/root/checkout.go` + `pull.go` + `modules_version.go`
- **state-management.md**: `pkg/iac/pulumi/backendconfig/backend_config.go`, `pkg/iac/tofu/backendconfig/backend_config.go` + `validate.go`, `pkg/iac/tofu/tfbackend/tf_backend.go`

### Architecture Page Refactored

The architecture page was refactored from a prose-heavy overview into a diagram-focused page with 3 ASCII architecture diagrams:
1. **Deployment flow**: manifest -> validation -> module resolution -> IaC engine -> cloud provider -> deployed resources
2. **Component anatomy**: the directory structure with annotations explaining each file's role
3. **Three-layer architecture**: API layer, execution layer, infrastructure layer

### Key Discovery: Updated Component Count

During source verification, discovered the actual component count is **198** (not 178 as previously documented):
- Scaleway: 19 components (recently added provider)
- Azure: 12 components (AzureResourceGroup and AzureLogAnalyticsWorkspace added)
- All other provider counts confirmed accurate

## Benefits

- Developers evaluating OpenMCF can now build a complete mental model from documentation alone
- The concepts section scales from "what is this?" to "how does state management work?" in a deliberate learning sequence
- Every page links to related pages, creating a navigable knowledge graph
- Real YAML examples and protobuf excerpts reduce the gap between docs and source code
- Updated component counts (198 kinds, 14 providers) reflect the current state of the framework

## Impact

- **Evaluators**: Can understand OpenMCF's design philosophy, provider-specific approach, and scale without reading source code
- **New users**: Have conceptual foundation before their first deployment
- **Power users**: Have reference material for module versioning, state backends, validation rules, and manifest labels
- **Documentation site**: Concepts section grows from 2 pages to 9 pages with full cross-linking

## Related Work

- Part of project `20260212.03.openmcf-docs-feature-parity`
- Builds on T01 Phase 0 audit (2026-02-12) which fixed 10 existing pages
- Phase 2 (CLI docs expansion) and Phase 3 (Guides expansion) are next

---

**Status**: Production Ready
**Timeline**: Single session (Phase 1 of 6-phase docs project)
