# OCI DevOps Project Deployment Component

**Date**: February 21, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi Module, Terraform Module, Resource Registration

## Summary

Implemented the OciDevopsProject deployment component -- OCI's organizational container for CI/CD pipelines, code repositories, deployment environments, artifacts, and triggers. This is a single-resource component wrapping `oci_devops_project` with a flattened notification topic configuration, both Pulumi and Terraform modules, and 14 validation tests. Final resource of Phase 11 (Additional Services) for the OCI provider expansion.

## Problem Statement / Motivation

The OpenMCF OCI provider expansion requires coverage of the DevOps Project resource to support teams adopting OCI's managed CI/CD service. A DevOps Project is the top-level organizational unit that all other DevOps resources (build pipelines, deploy pipelines, repositories, connections, triggers) reference by OCID.

### Pain Points

- No declarative way to provision OCI DevOps Projects through OpenMCF
- Teams manually creating DevOps Projects through the OCI Console or ad-hoc Terraform
- No composability with other OpenMCF OCI components for project_id references

## Solution / What's New

Added `OciDevopsProject` (enum 3396, id_prefix `ocidev`) as a deployment component with both Pulumi and Terraform IaC modules.

### Key Design Decisions

- **Flattened notification_config**: The provider nests `topic_id` inside a `notification_config` block, but since the block contains exactly one field, we flatten it to a top-level `notification_topic_id` for cleaner YAML UX
- **No sub-resources bundled**: Repositories, pipelines, connections, triggers, and artifacts are all separate resources with independent lifecycles
- **`devops_project_repository_setting` excluded**: Separate resource with its own lifecycle, not tightly coupled to the project

## Implementation Details

### Proto API (4 files)

- `spec.proto`: 3 fields (compartment_id as StringValueOrRef with OciCompartment default_kind, notification_topic_id as required StringValueOrRef without default_kind since ONS topics are not in the catalog, description as optional string)
- `api.proto`: KRM wiring with OciDevopsProject and OciDevopsProjectStatus messages
- `stack_input.proto`: Standard stack input wrapping target + provider config
- `stack_outputs.proto`: 2 outputs (project_id for downstream composability, namespace for container registry paths)

### Validation Tests (14 tests)

- 6 valid scenarios: minimal fields, with description, compartment via valueFrom, notification topic via valueFrom, both refs via valueFrom, empty description
- 8 invalid scenarios: wrong api_version, wrong kind, missing metadata, missing spec, missing compartment_id, missing notification_topic_id, empty api_version, empty kind

### Pulumi Module (5 Go files)

- `devops.NewProject()` with `ProjectNotificationConfigArgs` reconstructing the provider's nested block from the flattened spec field
- Conditional description assignment (only when non-empty)
- Exports: project_id (ID()), namespace

### Terraform Module (5 files)

- `oci_devops_project.this` with inline `notification_config` block
- Description uses conditional null for empty string
- Standard freeform_tags from locals

### Kind Registration

- `OciDevopsProject = 3396` added under new "Additional Services" section in `cloud_resource_kind.proto`
- `kind_map_gen.go` regenerated successfully

## Benefits

- Declarative DevOps Project provisioning through OpenMCF manifests
- Composability via `project_id` output -- downstream DevOps resources can reference the project using StringValueOrRef valueFrom
- Namespace output enables container registry path construction
- Clean YAML UX with flattened notification topic (no unnecessary nesting)

## Impact

- OCI provider: 34/37 resources now have Start=done
- Phase 11 (Additional Services): 1/2 resources complete (OciDevopsProject done, OciNetworkFirewall pending)
- Downstream composability: build pipelines, deploy pipelines, and repositories can now reference the project_id output

## Validation Results

- `go build`: clean
- `go vet`: clean
- `go test`: 14/14 tests passed
- `terraform validate`: success
- `go build ./pkg/crkreflect/...`: kind map compiles

## Related Work

- Part of the OCI provider expansion sub-project (20260219.01.sp.oracle-cloud-provider)
- R36 OciNetworkFirewall is the other Phase 11 resource (being worked on by another agent)
- All 37 OCI resources will be complete once R34-R36 finish

---

**Status**: Production Ready
