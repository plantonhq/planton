# OCI Compartment Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Provider Framework, Pulumi Module, Terraform Module

## Summary

Added OciCompartment (R04) as the fourth OCI resource kind in OpenMCF, wrapping `oci_identity_compartment`. Compartments are OCI's fundamental resource isolation primitive -- every other OCI resource lives within a compartment. This component is referenced by virtually every other OCI kind via `StringValueOrRef`, making it a critical foundation for the entire OCI catalog.

## Problem Statement / Motivation

After completing the networking foundation (VCN, Subnet, NSG), the next gap was identity infrastructure. In OCI, compartments are the root of the resource hierarchy -- they are mandatory for every resource creation call. Without an OciCompartment component, users had to hard-code compartment OCIDs, losing the infra-chart composability that `StringValueOrRef` provides.

### Pain Points

- Every OCI component's `compartmentId` field already had `default_kind = OciCompartment`, but the component did not yet exist
- No way to manage compartment lifecycles (creation, tagging, deletion policy) through OpenMCF
- Nested compartment hierarchies could not be expressed declaratively

## Solution / What's New

Implemented `OciCompartment` as a lean, single-resource component with four spec fields and one output.

### Spec Design

```
OciCompartmentSpec
├── compartment_id  (StringValueOrRef, required) -- parent compartment
├── name            (string, optional)           -- falls back to metadata.name
├── description     (string, required, min_len=1)
└── enable_delete   (bool, optional)             -- safety net for production
```

**Field naming**: `compartment_id` (not `parentCompartmentId` from the original plan) was chosen to maintain consistency with every other OCI component where field 1 is always `compartment_id` meaning "the compartment this resource lives in." For compartments, that IS the parent. The proto comment clarifies the semantics.

**`enable_delete`**: Exposes OCI's compartment deletion safety mechanism. When false (default), `pulumi destroy` / `terraform destroy` retains the compartment -- preventing accidental deletion of compartments containing active resources.

### Component Structure

```
apis/org/openmcf/provider/oci/ocicompartment/v1/
├── spec.proto              # 4 fields
├── api.proto               # KRM wiring (OciCompartment, OciCompartmentStatus)
├── stack_input.proto       # OciCompartmentStackInput
├── stack_outputs.proto     # compartment_id output
├── spec_test.go            # 11 validation tests
├── iac/
│   ├── pulumi/
│   │   ├── main.go         # Entrypoint
│   │   ├── Pulumi.yaml
│   │   ├── Makefile
│   │   └── module/
│   │       ├── main.go         # Resources() orchestrator
│   │       ├── locals.go       # Locals struct + freeform tags
│   │       ├── outputs.go      # Output constants
│   │       └── compartment.go  # identity.NewCompartment
│   ├── tf/
│   │   ├── provider.tf     # oracle/oci >= 5.0
│   │   ├── main.tf         # oci_identity_compartment resource
│   │   ├── locals.tf       # name fallback + freeform tags
│   │   ├── variables.tf    # metadata + spec variables
│   │   └── outputs.tf      # compartment_id output
│   └── hack/
│       └── manifest.yaml   # Test manifest
```

## Implementation Details

### Proto Layer

- `compartment_id` uses `StringValueOrRef` with `default_kind = OciCompartment` enabling self-referential hierarchies (child compartment referencing parent compartment by name)
- `description` has `min_len = 1` validation since the OCI API requires it
- `name` is optional -- falls back to `metadata.name` in both IaC modules

### Pulumi Module

Uses `identity.NewCompartment` from `github.com/pulumi/pulumi-oci/sdk/v4/go/oci/identity`. The module follows the established flat structure with a dedicated `compartment.go` file. Freeform tags are built from metadata following the standard pattern (resource, resource_kind, resource_id, org, env, labels).

### Terraform Module

Single resource in `main.tf` (`oci_identity_compartment`). The `locals.tf` uses `coalesce(var.spec.name, var.metadata.name)` for the name fallback, matching the Pulumi module's behavior.

### Validation

- 11 spec tests passing (Ginkgo/Gomega): minimal valid, custom name, enable_delete, value_from ref, fully-specified, wrong api_version, wrong kind, missing metadata, missing spec, missing compartment_id, empty description
- `go build` and `go vet` clean
- `terraform validate` passes

## Benefits

- Completes the `StringValueOrRef` chain: users can now reference compartments by name instead of hard-coding OCIDs
- Enables declarative compartment hierarchies via chained OciCompartment resources
- Every downstream OCI component (R05-R37) can now resolve `compartmentId` via infra-chart composability
- `enable_delete` provides a safe default for production while allowing cleanup in dev/test

## Impact

- **OCI catalog**: 4 of 37 resource kinds now implemented (Phase 1: 4/6 complete)
- **Downstream components**: All future OCI components can use `valueFrom` to reference compartments
- **Kind map**: `OciCompartment` registered in `kind_map_gen.go` with the `ProviderOciMap`

## Related Work

- Depends on: OCI provider integration (`2026-02-18-203551-oci-provider-integration.md`)
- Predecessor: OciNetworkSecurityGroup R03 (`2026-02-19-053424-oci-network-security-group-deployment-component.md`)
- Next: R05 OciIdentityPolicy

---

**Status**: Production Ready
