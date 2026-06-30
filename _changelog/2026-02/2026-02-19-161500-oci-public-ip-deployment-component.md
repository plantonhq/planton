# OCI Public IP Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Provider Framework

## Summary

Implemented the OciPublicIp deployment component (R14, enum 3323) -- the fourth and final resource in Phase 3 (Advanced Networking). Manages OCI reserved and ephemeral public IPv4 addresses with optional BYOIP pool support. This completes the Advanced Networking phase.

## Problem Statement / Motivation

Planton's OCI provider had load balancers and DRG connectivity but lacked direct public IP management.

### Pain Points

- No way to allocate and manage persistent public IP addresses through Planton
- No support for reserved IPs that survive instance termination (critical for DNS records, firewall allowlists)
- No BYOIP (Bring Your Own IP) pool integration for enterprise migration scenarios
- Ephemeral IPs required manual assignment outside Planton

## Solution / What's New

Designed and implemented a single-resource OciPublicIp component that manages `oci_core_public_ip` with support for both reserved and ephemeral lifetime modes.

### Design Decision: String Field for Lifetime

The `lifetime` field uses a plain string with `in`-list validation (`"RESERVED"`, `"EPHEMERAL"`) rather than a protobuf enum. This was necessary because `reserved` is a protobuf keyword and cannot be used as an enum value name. The string approach matches OCI's API values exactly, so IaC modules pass the value through without conversion.

### Conditional Validation

A CEL expression enforces that ephemeral IPs must have a `private_ip_id` assigned at creation. Reserved IPs may be created unassigned and attached later.

## Implementation Details

### Proto API (spec.proto)

- **5 fields**: compartment_id (StringValueOrRef, required), lifetime (string, required, in-list validated), display_name (optional, falls back to metadata.name), private_ip_id (StringValueOrRef, conditionally required for ephemeral), public_ip_pool_id (StringValueOrRef, optional for BYOIP)
- **CEL validation**: `this.lifetime != 'EPHEMERAL' || has(this.private_ip_id)`
- **No default_kind** on private_ip_id and public_ip_pool_id -- there are no corresponding Planton components (private IPs are created implicitly with VNICs; IP pools are not in the 37-resource catalog)

### Stack Outputs

- `public_ip_id` -- OCID of the created public IP resource
- `ip_address` -- the allocated IPv4 address (primary output for DNS, firewall rules)

### Validation Tests (spec_test.go)

- **16 Ginkgo/Gomega tests** (9 valid scenarios, 7 invalid scenarios)
- Valid: minimal reserved, reserved with display_name, reserved with private_ip_id, reserved with BYOIP pool, compartment_id via valueFrom, private_ip_id via valueFrom, minimal ephemeral with private_ip_id, fully-specified reserved
- Invalid: wrong api_version, wrong kind, missing metadata, missing spec, missing compartment_id, empty lifetime, invalid lifetime string, ephemeral without private_ip_id (CEL)

### Pulumi Module (4 Go files)

| File | Purpose |
|------|---------|
| `main.go` | `Resources()` entry point |
| `locals.go` | Display name fallback, freeform tags |
| `public_ip.go` | `core.NewPublicIp()` with conditional private_ip_id and public_ip_pool_id |
| `outputs.go` | Output key constants |

### Terraform Module (5 HCL files)

- `main.tf` -- `oci_core_public_ip.this` with conditional null handling for optional OCID fields
- `locals.tf` -- freeform tags, display_name coalesce
- `variables.tf` -- metadata + spec type definitions with optional object fields
- `outputs.tf` -- public_ip_id, ip_address
- `provider.tf` -- OCI provider >= 5.0

### Kind Registration

- `OciPublicIp = 3323` added to CloudResourceKind enum under "Advanced Networking" section
- `kind_map_gen.go` regenerated with new entry

## Benefits

- **Stable addressing** -- reserved IPs persist across instance lifecycle for DNS and firewall rules
- **Ephemeral convenience** -- automatic IP assignment tied to compute instance lifecycle
- **BYOIP support** -- enterprise customers can use their own IP ranges
- **Composable** -- outputs feed into DNS records, firewall rules, and documentation via StringValueOrRef

## Impact

- **OCI Provider**: 14th resource kind implemented (14/37 total)
- **Phase 3 Complete**: All 4 Advanced Networking components done (LoadBalancer, NetworkLoadBalancer, DynamicRoutingGateway, PublicIp)
- **Users**: Can now allocate and manage public IP addresses through Planton, enabling stable DNS records and firewall configurations

## Validation Results

- `go build` -- clean
- `go vet` -- clean
- `go test` -- 16/16 passed
- `terraform validate` -- success

## Related Work

- **R11 OciApplicationLoadBalancer** -- first Phase 3 component
- **R12 OciNetworkLoadBalancer** -- second Phase 3 component
- **R13 OciDynamicRoutingGateway** -- third Phase 3 component
- **R15 OciAutonomousDatabase** -- next component (Phase 4: Databases)

---

**Status**: Production Ready
