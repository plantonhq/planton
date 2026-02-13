# Scaleway Container Registry Resource Kind (R14)

**Date**: February 13, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Terraform Modules, Provider Framework

## Summary

Implemented the ScalewayContainerRegistry resource kind (R14) -- a standalone, regional resource wrapping `scaleway_registry_namespace`. This is the fourteenth Scaleway resource kind and the first in the container/registry tier. Provides a declarative interface for creating OCI-compliant container image registries on Scaleway with Docker-compatible endpoints.

## Problem Statement / Motivation

The Scaleway provider expansion requires container registry support for teams to store and distribute container images. Without a registry resource kind:

- Kapsule cluster workloads cannot pull from Scaleway-hosted private registries through the infra-chart composition model
- Serverless functions and containers (R17, R18) will have no declarative way to reference their image source registry
- CI/CD pipelines lack a IaC-managed push target

### Pain Points

- No existing OpenMCF resource for Scaleway container image storage
- Future serverless resource kinds (R17, R18) need registry endpoint outputs for `valueFrom` references
- Kapsule cluster deployments need registry endpoints for `imagePullSecrets` configuration

## Solution / What's New

A standalone resource kind following the established Scaleway pattern (closest to ScalewayBlockVolume in simplicity), wrapping a single `scaleway_registry_namespace` resource.

### Key Features

- **Private by default** -- `is_public` defaults to `false`, requiring authentication for image pulls
- **Regional deployment** -- Available in fr-par, nl-ams, pl-waw with region-specific Docker endpoints
- **Docker-native endpoint** -- Outputs the Docker endpoint URL (`rg.<region>.scw.cloud/<namespace-name>`) for direct use in `docker login`/`push`/`pull`
- **Infra-chart ready** -- Exports `endpoint` and `namespace_id` for downstream `valueFrom` references

## Implementation Details

### File Structure

17 new files under `apis/org/openmcf/provider/scaleway/scalewaycontainerregistry/v1/`:

| Category | Files | Description |
|---|---|---|
| Proto schemas | 4 | api.proto, spec.proto, stack_input.proto, stack_outputs.proto |
| Pulumi Go module | 6 | main.go, Pulumi.yaml, module/{main,locals,registry,outputs}.go |
| Terraform HCL | 5 | main.tf, variables.tf, outputs.tf, locals.tf, provider.tf |
| Documentation | 2 | README.md, examples.md |

### Spec Design

Three spec fields covering the 80% use case:

| Field | Type | Required | Description |
|---|---|---|---|
| `region` | string | Yes | Scaleway region (immutable after creation) |
| `description` | string | No | Human-readable namespace description |
| `is_public` | bool | No | Public pull access (default: false) |

### Stack Outputs

| Output | Format | Downstream Use |
|---|---|---|
| `namespace_id` | `{region}/{uuid}` | Terraform state, API references |
| `endpoint` | `rg.<region>.scw.cloud/<name>` | Docker login/push/pull, K8s imagePullSecrets |
| `namespace_name` | string | Observability, CI/CD pipeline variables |
| `region` | string | Co-location verification |

### Discovery: No Tag Support

During build verification, discovered that **Scaleway Container Registry namespaces do not support tags** in either the Pulumi SDK (`registry.NamespaceArgs` has no `Tags` field) or the Terraform provider (`scaleway_registry_namespace` rejects the `tags` argument). This is unlike most other Scaleway resources.

**Resolution:** Removed tag handling from both Pulumi `locals.go` (simplified to not build tag slices) and Terraform `locals.tf`/`main.tf` (removed `standard_tags`). Documented this as a Scaleway API limitation in both IaC modules and README.

### Pulumi SDK

Uses the preferred `registry.NewNamespace()` from the `scaleway/registry` subpackage (not the deprecated top-level `scaleway.NewRegistryNamespace`), consistent with the subpackage migration pattern established in R07 (KapsuleCluster) and R12 (ObjectBucket).

## Benefits

- **Composition ready** -- The `endpoint` output enables downstream serverless resources (R17, R18) and Kubernetes workloads to reference the registry via infra-chart `valueFrom` patterns
- **Secure defaults** -- Private registry by default prevents accidental public exposure of proprietary images
- **Minimal spec** -- Three fields for a complete registry, reducing configuration burden
- **Consistent patterns** -- Follows the established standalone resource patterns from R01-R13

## Impact

- **Scaleway provider**: 14 of 19 resource kinds now complete (74%)
- **Storage/Registry tier**: Complete (ObjectBucket + BlockVolume + ContainerRegistry)
- **Downstream enabler**: Unlocks R17 (ServerlessFunction) and R18 (ServerlessContainer) which need registry endpoints

## Related Work

- **R12: ScalewayObjectBucket** -- Similar standalone resource, but uses map-based tags (S3-compatible). ContainerRegistry uses neither tag format.
- **R13: ScalewayBlockVolume** -- Closest pattern match (standalone, simple, leaf resource with flat tags -- though registry has no tags at all).
- **R07: ScalewayKapsuleCluster** -- Kapsule workloads consume registry endpoints for image pulls.
- **R17/R18: Serverless Function/Container** -- Will use `endpoint` output via `StringValueOrRef` for image source configuration.

---

**Status**: Production Ready
