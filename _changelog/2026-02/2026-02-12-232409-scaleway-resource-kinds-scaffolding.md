# Scaleway Resource Kinds Scaffolding and Enum Registration

**Date**: February 12, 2026
**Type**: Feature
**Components**: API Definitions, Provider Framework, Build System

## Summary

Added foundational scaffolding for 19 Scaleway cloud resource kinds in Planton. This includes registering all enum values in `cloud_resource_kind.proto`, creating the Pulumi provider helper and label keys packages, and adding the first community (pulumiverse) Pulumi SDK dependency. All builds pass -- proto compilation, Go compilation, and Gazelle-managed BUILD.bazel generation.

## Problem Statement / Motivation

Scaleway was integrated as cloud provider #24 in a previous session (provider config, credential management, 6-layer CLI integration), but no actual resource kinds existed yet. Before implementing any of the 19 planned resource kinds (VPC, Private Network, Kapsule Cluster, RDB Instance, etc.), every resource needs:

1. An enum value in `cloud_resource_kind.proto` so the platform can identify and route it
2. A Pulumi provider helper so IaC modules can authenticate with Scaleway
3. A label keys package so resources get consistent `planton-ai_*` labels

### Pain Points

- Without enum registration, no resource kind proto schemas could reference their `CloudResourceKind` value
- Without the Pulumi provider helper, no Pulumi Go module could authenticate
- Without label keys, resources would ship without the standard organizational labels
- The Scaleway Pulumi SDK (`pulumiverse/pulumi-scaleway`) had never been used in this codebase -- needed verification

## Solution / What's New

### Enum Registration (19 kinds, range 2800-2880)

All 19 Scaleway resource kinds registered in `cloud_resource_kind.proto` with categorized grouping and room for future expansion:

| Category | Kinds | Enum Range |
|----------|-------|------------|
| Networking | VPC, PrivateNetwork, PublicGateway, LoadBalancer, InstanceSecurityGroup | 2800-2804 |
| Compute | Instance | 2810 |
| Kubernetes | KapsuleCluster, KapsulePool | 2820-2821 |
| Databases | RdbInstance, RedisCluster, MongodbInstance | 2830-2832 |
| Storage | ObjectBucket, BlockVolume | 2840-2841 |
| Container Registry | ContainerRegistry | 2850 |
| DNS | DnsZone, DnsRecord | 2860-2861 |
| Serverless | ServerlessFunction, ServerlessContainer | 2870-2871 |
| Security | SecretManager | 2880 |

Gaps between groups (e.g., 2804 -> 2810) deliberately reserved for future additions within each category.

### Pulumi Provider Helper

Created `pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider/provider.go` following the DigitalOcean/Civo pattern:

- `Get()` -- builds a `scaleway.Provider` from `ScalewayProviderConfig`, mapping all 6 credential fields (access_key, secret_key, project_id, organization_id, region, zone). Empty fields are left nil for env-var fallback.
- `ProviderResourceName()` -- deterministic naming (`"scaleway"` + suffixes)
- `PulumiOutputName()` -- canonical output names with `"scw_"` prefix

### Label Keys Package

Created `pkg/iac/pulumi/pulumimodule/provider/scaleway/scalewaylabelkeys/scaleway_label_keys.go` -- standard `planton-ai_*` labels (resource, organization, environment, kind, id, name) matching the DigitalOcean and Civo patterns.

### Pulumi SDK Dependency

Added `github.com/pulumiverse/pulumi-scaleway/sdk v1.43.0` to `go.mod`. This is the first pulumiverse (community-maintained) Pulumi provider in the repo -- all prior providers use the official `github.com/pulumi/` namespace. The SDK was verified to have the expected `ProviderArgs` struct and `NewProvider` function before writing the helper.

## Implementation Details

### Design Decisions

- **Region/Zone: Strings, not enums** -- The already-shipped `ScalewayProviderConfig` uses `string region` and `string zone`. Rather than creating typed enums (which would be inconsistent with provider.proto or require a breaking change), we keep strings throughout. Validation happens via `buf-validate` patterns in spec.proto files.
- **Pulumiverse accepted** -- No official Pulumi Scaleway SDK exists. The pulumiverse provider (v1.43.0, Feb 6 2026) is actively maintained and verified to cover the resource types we need.

### Files Changed

| File | Change |
|------|--------|
| `apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto` | +124 lines: 19 enum values with `kind_meta` |
| `go.mod` / `go.sum` | Added `pulumiverse/pulumi-scaleway/sdk v1.43.0` |
| `pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider/provider.go` | New: Pulumi provider helper |
| `pkg/iac/pulumi/pulumimodule/provider/scaleway/scalewaylabelkeys/scaleway_label_keys.go` | New: Label keys package |

Auto-generated files updated by `make protos` and `make reset-gazelle`:
- `cloud_resource_kind.pb.go`, `cloud_resource_kind_pb.ts` (proto codegen)
- `pkg/crkreflect/kind_map_gen.go` (reflection map)
- `MODULE.bazel` (Bazel module dependency)
- BUILD.bazel files for new packages (Gazelle)

## Benefits

- **Unblocks all 19 resource kind implementations** -- every subsequent iteration (R01-R19) can now reference its enum value, use the provider helper, and apply standard labels
- **Build-validated** -- `buf build`, `buf lint`, and `go build ./...` all pass
- **Pattern-consistent** -- follows the same scaffolding structure as DigitalOcean and Civo providers
- **Forward-compatible** -- enum gaps and 200-value range (2800-2999) accommodate future Scaleway services

## Impact

- **Resource kind authors**: Can now begin implementing any of the 19 Scaleway kinds
- **Frontend**: TypeScript stubs generated with all new enum values
- **Backend**: Go reflection map updated with all kind metadata

## Related Work

- Previous session: `_changelog/2026-02/2026-02-12-181851-scaleway-provider-integration.md` -- Scaleway provider config and credential management
- Parent project: `20260212.01.planton-cloud-provider-expansion` in plantonhq/planton
- Sub-project: `20260212.04.sp.scaleway-resource-kinds` -- this P0 was the first implementation task
- Next: R01 (ScalewayVpc) -- foundation resource, first actual deployment component

---

**Status**: Production Ready
**Timeline**: Single session (P0 of 22-task implementation queue)
