# ScalewayVpc Resource Kind Implementation and Codegen Phantom Import Fix

**Date**: February 13, 2026
**Type**: Feature
**Components**: API Definitions, Provider Framework, Pulumi Module, Terraform Module, Build System

## Summary

Implemented the first Scaleway resource kind (ScalewayVpc) end-to-end -- proto schemas, Pulumi Go module, Terraform HCL module, and documentation. Also fixed a codegen bug where `kind_map_gen.go` generated imports for all 19 Scaleway resource kinds before their packages existed, causing Gazelle warnings and potential build failures.

## Problem Statement / Motivation

Two issues were addressed in this session:

### 1. Codegen Phantom Imports

After Session 3 registered 19 Scaleway enum values in `cloud_resource_kind.proto`, the `crkreflect` codegen (`pkg/crkreflect/codegen/main.go`) unconditionally generated Go imports for all 19 Scaleway packages in `kind_map_gen.go`. Since no resource kind packages existed yet, this produced 19 Gazelle warnings during `make protos` and would cause `go build` failures because the generated code referenced types in non-existent packages.

### 2. No Scaleway Resource Kinds Existed

The Scaleway provider had scaffolding (enum values, provider helper, label keys) but zero actual resource kinds. ScalewayVpc is the foundation resource (Layer 0) that all other Scaleway resources depend on -- it needed to be implemented first.

### Pain Points

- 19 Gazelle warnings on every `make protos` run, polluting build output
- `kind_map_gen.go` contained unresolvable imports that would break compilation
- No Scaleway resource kinds could be deployed -- the entire provider was non-functional

## Solution / What's New

### Fix: Codegen Directory Existence Guard

Added a directory existence check in `pkg/crkreflect/codegen/main.go` that skips resource kinds whose API packages don't exist on disk yet. This is a 7-line guard clause between the import path computation and the `uniqueAlias()` call:

```go
pkgDir := filepath.Join("apis", "org", "openmcf", "provider", provSlug, lowerKind, "v1")
if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
    fmt.Fprintf(os.Stderr, "skipping %s: package dir %s not found\n", kindName, pkgDir)
    continue
}
```

Skipped kinds are logged to stderr for visibility. As each resource kind is implemented, the codegen automatically picks it up -- no further codegen changes needed.

### Feature: ScalewayVpc Resource Kind

Implemented the complete ScalewayVpc resource kind following the DigitalOcean VPC reference pattern:

```
apis/org/openmcf/provider/scaleway/scalewayvpc/v1/
├── spec.proto              # ScalewayVpcSpec: region, enable_routing, enable_custom_routes_propagation
├── api.proto               # ScalewayVpc resource + ScalewayVpcStatus
├── stack_input.proto        # ScalewayVpcStackInput: target + provider_config
├── stack_outputs.proto      # ScalewayVpcStackOutputs: vpc_id
├── README.md               # Component documentation
├── examples.md             # 4 YAML examples with deployment commands
├── iac/
│   ├── pulumi/
│   │   ├── main.go          # Pulumi entrypoint
│   │   ├── Pulumi.yaml      # Project config
│   │   ├── Makefile          # Build targets
│   │   └── module/
│   │       ├── main.go      # Module orchestration
│   │       ├── vpc.go       # VPC resource creation
│   │       ├── locals.go    # Labels/tags from metadata
│   │       └── outputs.go   # Output constants
│   └── tf/
│       ├── provider.tf      # Scaleway provider config
│       ├── variables.tf     # Input variables
│       ├── locals.tf        # Local values
│       ├── main.tf          # scaleway_vpc resource
│       └── outputs.tf       # Output values
```

#### Key Design Decisions

- **Spec is minimal**: Scaleway VPCs have no CIDR blocks or IP ranges (unlike DigitalOcean/AWS). The spec has only `region`, `enable_routing`, and `enable_custom_routes_propagation`. IP planning happens at the Private Network level (R02).
- **One-way routing flags**: Both `enable_routing` and `enable_custom_routes_propagation` are irreversible toggles -- once enabled, they cannot be disabled. This constraint is documented in proto comments, README, and enforced via Terraform lifecycle `ignore_changes`.
- **Tags from metadata**: Standard OpenMCF labels are automatically applied as Scaleway tags (formatted as `"key=value"` strings). No user-facing `tags` field in the spec -- follows the DigitalOcean VPC pattern.
- **Single output**: `vpc_id` is the only stack output, matching what downstream resources (ScalewayPrivateNetwork) need as a `StringValueOrRef` reference.

## Implementation Details

### Codegen Fix

| File | Change |
|------|--------|
| `pkg/crkreflect/codegen/main.go` | +7 lines: `os.Stat` guard before `uniqueAlias()` call |
| `pkg/crkreflect/kind_map_gen.go` | Regenerated: now includes ScalewayVpc, skips 18 unimplemented kinds |

### ScalewayVpc Resource Kind

| File | Description |
|------|-------------|
| `apis/.../scalewayvpc/v1/spec.proto` | 3 fields: region (required), enable_routing, enable_custom_routes_propagation |
| `apis/.../scalewayvpc/v1/api.proto` | ScalewayVpc message + ScalewayVpcStatus |
| `apis/.../scalewayvpc/v1/stack_input.proto` | Stack input with target + provider config |
| `apis/.../scalewayvpc/v1/stack_outputs.proto` | vpc_id output |
| `apis/.../scalewayvpc/v1/iac/pulumi/main.go` | Loads stack input, delegates to module |
| `apis/.../scalewayvpc/v1/iac/pulumi/module/main.go` | Orchestrates: locals -> provider -> vpc |
| `apis/.../scalewayvpc/v1/iac/pulumi/module/vpc.go` | Creates `network.NewVpc()` via pulumiverse SDK |
| `apis/.../scalewayvpc/v1/iac/pulumi/module/locals.go` | Builds tags from metadata using scalewaylabelkeys |
| `apis/.../scalewayvpc/v1/iac/pulumi/module/outputs.go` | `OpVpcId = "vpc_id"` constant |
| `apis/.../scalewayvpc/v1/iac/tf/main.tf` | `scaleway_vpc` resource with lifecycle management |
| `apis/.../scalewayvpc/v1/iac/tf/variables.tf` | metadata, spec, credential variables |
| `apis/.../scalewayvpc/v1/iac/tf/locals.tf` | Extracts values, builds standard tags |
| `apis/.../scalewayvpc/v1/iac/tf/outputs.tf` | vpc_id, is_default, organization_id, region, created_at |
| `apis/.../scalewayvpc/v1/iac/tf/provider.tf` | scaleway/scaleway ~> 2.0 |
| `apis/.../scalewayvpc/v1/README.md` | Component overview, constraints, use cases |
| `apis/.../scalewayvpc/v1/examples.md` | 4 YAML examples with deployment commands |

Auto-generated files updated by `make protos` and Gazelle:
- `*.pb.go`, `*_pb.ts` proto stubs
- BUILD.bazel files for all new packages
- `pkg/crkreflect/kind_map_gen.go` (ScalewayVpc now registered)

## Benefits

- **Codegen is now incremental-safe**: Enum values can be registered before packages exist without breaking builds
- **First Scaleway resource is deployable**: ScalewayVpc can be provisioned via both Pulumi and Terraform
- **Foundation for R02-R19**: All subsequent Scaleway resource kinds can reference ScalewayVpc's `vpc_id` output
- **Pattern established**: The ScalewayVpc implementation sets the template for all 18 remaining Scaleway resource kinds
- **Zero Gazelle warnings**: `make protos` runs cleanly

## Impact

- **Resource kind authors**: ScalewayVpc is the reference implementation for all subsequent Scaleway kinds
- **Infra chart designers**: The `kapsule-environment` and `serverless-environment` charts can now reference ScalewayVpc as their Layer 0 foundation
- **Build system**: Codegen fix prevents recurrence of phantom import warnings for any future provider that registers enums before implementing resources

## Related Work

- Previous session: `_changelog/2026-02/2026-02-12-232409-scaleway-resource-kinds-scaffolding.md` -- P0 scaffolding (enums, provider helper, label keys)
- Previous session: `_changelog/2026-02/2026-02-12-181851-scaleway-provider-integration.md` -- Provider config and credential management
- Parent project: `20260212.01.openmcf-cloud-provider-expansion` in plantonhq/planton
- Sub-project: `20260212.04.sp.scaleway-resource-kinds` -- R01 was the first resource kind task
- Next: R02 (ScalewayPrivateNetwork) -- first resource with `StringValueOrRef` dependency on ScalewayVpc

---

**Status**: Production Ready
**Timeline**: Single session (R01 of 19-resource implementation queue)
