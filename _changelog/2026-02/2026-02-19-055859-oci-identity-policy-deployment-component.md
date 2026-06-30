# OCI Identity Policy Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi IaC Module, Terraform IaC Module, Provider Framework

## Summary

Added the OciIdentityPolicy deployment component (R05, enum 3304) to the OCI provider in Planton. This component manages `oci_identity_policy` resources -- OCI's mechanism for granting IAM access through human-readable policy statements. Both Pulumi (Go) and Terraform (HCL) modules are implemented with full feature parity.

## Problem Statement / Motivation

OCI uses a declarative policy language for IAM access control. Policies are attached to compartments and define who can do what within that compartment's scope. Without an OciIdentityPolicy component, users cannot manage IAM access as code through Planton, making it impossible to build complete OCI environments that include proper access controls.

### Pain Points

- No way to manage OCI IAM policies through Planton
- Infra charts like oke-environment and compute-environment need policy components for complete environment provisioning
- The OCI identity resource family (R04 OciCompartment, R05 OciIdentityPolicy, R06 OciDynamicGroup) must all be available before OKE and compute phases can be fully realized

## Solution / What's New

Single-resource deployment component that wraps `oci_identity_policy` with the standard Planton KRM pattern.

### Spec Fields

- `compartmentId` (StringValueOrRef, required) -- compartment where the policy lives; supports infra-chart composability via `valueFrom`
- `name` (string, optional) -- falls back to `metadata.name`; unique across tenancy, immutable
- `description` (string, required) -- updatable description
- `statements` (repeated string, required, min 1) -- OCI policy language statements
- `versionDate` (string, optional) -- pins policy evaluation to a specific date for compliance stability

### Outputs

- `policyId` -- OCID of the created policy

## Implementation Details

### Files Created

**Proto API** (`apis/dev/planton/provider/oci/ociidentitypolicy/v1/`):
- `spec.proto` -- Spec message with buf-validate rules (required compartment_id, min_len description, min_items statements)
- `api.proto` -- KRM wiring with api_version/kind const validation
- `stack_input.proto` -- IaC module input (target + provider config)
- `stack_outputs.proto` -- Deployment output (policy_id)
- `spec_test.go` -- 14 Ginkgo/Gomega validation tests (6 valid, 8 invalid scenarios)

**Pulumi Module** (`iac/pulumi/`):
- `module/main.go` -- Entry point with provider setup
- `module/locals.go` -- Name fallback, freeform tag assembly
- `module/policy.go` -- `identity.NewPolicy()` call with all spec fields
- `module/outputs.go` -- Output constant
- `main.go` -- Pulumi entrypoint with stack input loading

**Terraform Module** (`iac/tf/`):
- `variables.tf` -- metadata + spec variable objects
- `locals.tf` -- Name coalesce, freeform tag merge
- `main.tf` -- `oci_identity_policy.this` resource
- `outputs.tf` -- policy_id output
- `provider.tf` -- oracle/oci >= 5.0

**Registration**:
- Added `OciIdentityPolicy = 3304` to `cloud_resource_kind.proto` with `id_prefix: "ociply"`
- Kind map regenerated and verified

### Validation Results

- Go build: clean
- Go vet: clean
- Go test: 14/14 specs passed
- Terraform validate: "Success! The configuration is valid."

## Benefits

- Enables IAM-as-code for OCI environments through Planton
- Supports infra-chart composability: `compartmentId` accepts `valueFrom` refs to OciCompartment outputs
- Policy `versionDate` field supports compliance-sensitive organizations that need stable policy evaluation behavior
- Consistent patterns with R04 OciCompartment for maintainability

## Impact

- OCI provider: 5 of 37 resource kinds now implemented (13.5%)
- Phase 1 (Foundation) is 83% complete: OciVcn, OciSubnet, OciSecurityGroup, OciCompartment, OciIdentityPolicy done; OciDynamicGroup remaining
- Docs and Presets agents can now pick up OciIdentityPolicy for documentation and preset generation

## Related Work

- Predecessor: OciCompartment (R04) -- same identity service, very similar component shape
- Next: OciDynamicGroup (R06) -- completes the identity resource family
- Parent project: 20260212.01.planton-cloud-provider-expansion

---

**Status**: Production Ready
