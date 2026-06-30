# OCI Dynamic Group Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi IaC Module, Terraform IaC Module, Provider Framework

## Summary

Added the OciDynamicGroup deployment component (R06, enum 3305) to the OCI provider in Planton. This component manages `oci_identity_dynamic_group` resources -- OCI's mechanism for grouping cloud resources (compute instances, functions, etc.) so they can be granted IAM permissions via policies. Both Pulumi (Go) and Terraform (HCL) modules are implemented with full feature parity.

## Problem Statement / Motivation

OCI uses dynamic groups to enable resource principal authentication -- allowing compute instances, functions, and other resources to authenticate to OCI APIs without stored credentials. This is essential for building secure, credential-free workloads on OCI.

### Pain Points

- No way to manage OCI dynamic groups through Planton
- Instance principal and resource principal authentication patterns require dynamic groups as a prerequisite
- The OCI identity resource family (R04 OciCompartment, R05 OciIdentityPolicy, R06 OciDynamicGroup) must all be available before compute and container phases can properly leverage IAM

## Solution / What's New

Single-resource deployment component wrapping `oci_identity_dynamic_group` with the standard Planton KRM pattern.

### Spec Fields

- `compartmentId` (StringValueOrRef, required) -- must be the tenancy (root compartment) OCID; dynamic groups are tenancy-level IAM resources
- `name` (string, optional) -- falls back to `metadata.name`; unique across all groups in the tenancy, immutable after creation
- `description` (string, required) -- updatable description of the group's purpose
- `matchingRule` (string, required) -- OCI rule syntax defining which resources belong to the group (e.g., `Any {instance.compartment.id = 'ocid1...'}`)

### Outputs

- `dynamicGroupId` -- OCID of the created dynamic group

## Implementation Details

### Files Created

**Proto API** (`apis/dev/planton/provider/oci/ocidynamicgroup/v1/`):
- `spec.proto` -- Spec message with buf-validate rules (required compartment_id, min_len description, min_len matching_rule)
- `api.proto` -- KRM wiring with api_version/kind const validation
- `stack_input.proto` -- IaC module input (target + provider config)
- `stack_outputs.proto` -- Deployment output (dynamic_group_id)
- `spec_test.go` -- 13 Ginkgo/Gomega validation tests (6 valid, 7 invalid scenarios)

**Pulumi Module** (`iac/pulumi/`):
- `module/main.go` -- Entry point with provider setup
- `module/locals.go` -- Name fallback, freeform tag assembly
- `module/dynamic_group.go` -- `identity.NewDynamicGroup()` call with all spec fields
- `module/outputs.go` -- Output constant
- `main.go` -- Pulumi entrypoint with stack input loading

**Terraform Module** (`iac/tf/`):
- `variables.tf` -- metadata + spec variable objects
- `locals.tf` -- Name coalesce, freeform tag merge
- `main.tf` -- `oci_identity_dynamic_group.this` resource
- `outputs.tf` -- dynamic_group_id output
- `provider.tf` -- oracle/oci >= 5.0

**Registration**:
- Added `OciDynamicGroup = 3305` to `cloud_resource_kind.proto` with `id_prefix: "ocidg"`
- Kind map regenerated and verified

### Design Decision: matching_rule as Freeform String

The `matchingRule` field is kept as a plain string rather than a structured proto message. OCI's matching rule language supports `Any {}`, `All {}`, nested conditions, and multiple resource types (`instance`, `resource`, `fnfunc`, etc.). A structured representation would add complexity without value -- platform engineers working with dynamic groups already know the syntax, and the OCI API accepts it as a single string.

### Validation Results

- Go build: clean
- Go vet: clean
- Go test: 13/13 specs passed
- Terraform validate: "Success! The configuration is valid."

## Benefits

- Enables resource principal authentication patterns for OCI workloads through Planton
- Supports infra-chart composability: `compartmentId` accepts `valueFrom` refs to OciCompartment outputs
- Completes the OCI identity resource family (R04 + R05 + R06), unblocking compute and container phases
- Consistent patterns with R04 OciCompartment and R05 OciIdentityPolicy for maintainability

## Impact

- OCI provider: 6 of 37 resource kinds now implemented (16.2%)
- Phase 1 (Foundation) is now 100% complete: OciVcn, OciSubnet, OciSecurityGroup, OciCompartment, OciIdentityPolicy, OciDynamicGroup
- Next phase: Phase 2 (Compute and Containers) starting with R07 OciComputeInstance
- Docs and Presets agents can now pick up OciDynamicGroup for documentation and preset generation

## Related Work

- Predecessor: OciIdentityPolicy (R05) -- same identity service, very similar component shape
- Next: OciComputeInstance (R07) -- begins Phase 2 (Compute and Containers)
- Parent project: 20260212.01.planton-cloud-provider-expansion

---

**Status**: Production Ready
