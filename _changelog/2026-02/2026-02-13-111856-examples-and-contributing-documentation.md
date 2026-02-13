# Examples and Contributing Documentation Sections

**Date**: February 13, 2026
**Type**: Feature
**Components**: Documentation, Examples, Contributing

## Summary

Created the Examples and Contributing documentation sections for the OpenMCF docs site — 4 new pages across 2 entirely new sections. The Examples section provides a manifest gallery with 10 proto-verified manifests across 7 providers. The Contributing section covers development setup, build workflows, and a comprehensive guide for adding new deployment components.

## Problem Statement / Motivation

The OpenMCF docs site had comprehensive Concepts, CLI, Guides, and Tutorials sections after Phases 1-4, but lacked two critical sections for a production open-source project:

### Pain Points

- No curated, copy-paste-ready manifest examples for users who want to start deploying quickly without following a full tutorial
- No contributor documentation explaining how to set up a development environment, build from source, or add new deployment components
- The `openmcf/examples/` directory only contained 3 YAML files (all Postgres variants), providing minimal coverage of the 198 components across 14 providers
- The component creation workflow (19-step forge process) was documented only as internal Cursor AI rules, not as public contributor-facing documentation

## Solution / What's New

### Examples Section (2 pages)

**`examples/index.md`** — Section navigation hub explaining how to use examples (copy, customize, deploy) with links to the manifest gallery, tutorials, and catalog.

**`examples/manifest-gallery.md`** — 10 curated manifests organized by provider, each verified field-by-field against the component's `spec.proto`:

| Provider | Components | Resource Types |
|----------|-----------|---------------|
| AWS | AwsS3Bucket, AwsRdsInstance, AwsVpc | Storage, Database, Networking |
| GCP | GcpCloudSql, GcpGkeCluster | Database, Kubernetes |
| Azure | AzureAksCluster | Kubernetes |
| Kubernetes | KubernetesDeployment, KubernetesPostgres | Workload, Database |
| Cloudflare | CloudflareWorker | Serverless |
| Civo | CivoVpc | Networking |

The gallery also documents the common metadata pattern for both Pulumi and OpenTofu provisioners, showing how to swap backend labels while keeping the spec identical.

### Contributing Section (2 pages)

**`contributing/index.md`** — Development environment setup covering prerequisites (Go 1.25+, Bazel, Buf, Make), building from source (`make build`, `make protos`, `make build-cli`), running tests (full suite and component-scoped), code style, naming conventions, and the PR submission workflow. Sourced from `CONTRIBUTING.md`, `Makefile`, `go.mod`, and `apis/_rules/`.

**`contributing/adding-components.md`** — Comprehensive guide for creating new deployment components, covering:

- Anatomy of a component (4 proto files + dual IaC modules + docs)
- File structure with directory tree using AwsS3Bucket as reference
- Naming conventions (folder, kind, apiVersion, proto package)
- 6-phase creation workflow (Define API, Register Kind, Pulumi Module, Terraform Module, Documentation, Build & Test)
- Design principles (80/20 rule, deployment-agnostic specs, secure defaults, dual IaC parity)
- Code examples at each step showing actual proto definitions, Go code, and HCL

## Implementation Details

### Manifest Gallery Verification

Each of the 10 gallery manifests was constructed by reading the component's `spec.proto` and translating proto field names to camelCase JSON serialization:

- `aws_region` → `awsRegion`
- `subnet_ids` → `subnetIds` (repeated `StringValueOrRef` with `value` wrapper)
- `database_engine` → `databaseEngine` (enum value like `POSTGRESQL`)
- `system_node_pool` → `systemNodePool` (nested message)

Hack manifests from inside components were used as reference but not copied verbatim — several had issues (wrong kind names, snake_case fields, incomplete specs).

### Contributing Page Sources

The contributing documentation was synthesized from multiple source locations:

- `CONTRIBUTING.md` — fork/clone/branch/PR workflow
- `Makefile` — build targets and their purposes
- `go.mod` — Go version (1.25.0)
- `apis/Makefile` — proto generation pipeline
- `buf/lint/optional-linter/` — custom Buf linting plugin
- `apis/_rules/` — development rules (localized builds, reserved words, default semantics)
- `_rules/deployment-component/forge/` — 19-step component creation workflow (simplified to 6 phases for docs)
- `architecture/deployment-component.md` — component anatomy and ideal state
- AwsS3Bucket component structure — concrete reference example

## Benefits

- **Copy-paste utility**: Users can grab a working manifest and deploy immediately, then customize
- **Contributor onboarding**: New contributors can set up a development environment and understand the component creation workflow without reverse-engineering the codebase
- **Cross-provider coverage**: Gallery spans 7 providers showing the consistent KRM pattern across AWS, GCP, Azure, Kubernetes, Cloudflare, and Civo
- **Source-grounded accuracy**: Every manifest verified against proto definitions, every build command verified against Makefile

## Impact

- OpenMCF docs site now has 7 complete sections (Concepts, CLI, Guides, Tutorials, Examples, Contributing, Catalog)
- Phase 5 of the documentation feature parity project is complete
- Only Phase 6 (final audit and polish) remains

## Related Work

- Concepts section rewrite (2026-02-13-090244)
- CLI documentation rewrite (2026-02-13-093830)
- Tutorials section (2026-02-13-105241)

---

**Status**: Production Ready
**Timeline**: Phase 5 of 6 in the OpenMCF docs feature parity project
