# OCI Subnet Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Provider Framework, Pulumi Module, Terraform Module

## Summary

Added OciSubnet (R02) as the second OCI resource kind in Planton, providing a complete deployment component for Oracle Cloud Infrastructure subnets with optional inline route table creation. The component covers all practical subnet attributes from the OCI provider and supports both public and private subnet configurations with full infra-chart composability via StringValueOrRef.

## Problem Statement / Motivation

After the OCI VCN (R01) was implemented, the next networking building block -- subnets -- was needed for any useful OCI deployment. Without subnets, no compute instances, OKE clusters, load balancers, or databases can be deployed because all OCI resources that need network connectivity require a subnet OCID.

### Pain Points

- No way to create OCI subnets through Planton
- No custom route table creation capability anywhere in the OCI catalog
- Downstream components (R07-R37) blocked waiting for subnet infrastructure

## Solution / What's New

Implemented `OciSubnet` as a full deployment component following the established OciVcn patterns, with one important addition: optional inline route table creation via `routeRules`.

### Key Design Decisions

**Route table bundling**: Since there is no separate OciRouteTable component in the 37-kind catalog, the subnet component accepts optional `routeRules` that create a dedicated route table owned by the subnet. This is mutually exclusive with `routeTableId` (referencing an external route table). The pattern enables:
- Self-contained subnets with their own routing (provide `routeRules`)
- Subnets sharing an external route table (provide `routeTableId`)
- Subnets using the VCN default route table (provide neither)

**CEL validation** enforces mutual exclusivity between `routeTableId` and `routeRules` at the proto level.

### Component Structure

```
apis/dev/planton/provider/oci/ocisubnet/v1/
├── spec.proto              # 13 fields + RouteRule nested message + CEL validation
├── api.proto               # KRM wiring (OciSubnet, OciSubnetStatus)
├── stack_input.proto        # OciSubnetStackInput
├── stack_outputs.proto      # 5 outputs (subnet_id, domain_name, router_ip/mac, route_table_id)
├── spec_test.go             # 21 Ginkgo specs (valid + invalid cases)
├── iac/pulumi/module/
│   ├── main.go              # Orchestrator: provider setup, optional route table, then subnet
│   ├── locals.go            # Display name fallback, freeform tags from metadata
│   ├── outputs.go           # Output name constants
│   ├── subnet.go            # oci.core.Subnet with conditional routing
│   └── route_table.go       # oci.core.RouteTable with destination type mapping
└── iac/tf/
    ├── main.tf              # oci_core_subnet with conditional route table reference
    ├── variables.tf         # Typed variable definitions matching proto spec
    ├── outputs.tf           # 5 outputs matching stack_outputs.proto
    ├── locals.tf            # Tags, display name, destination type map
    ├── route_table.tf       # oci_core_route_table with dynamic route_rules block
    └── provider.tf          # oracle/oci >= 5.0
```

## Implementation Details

### Spec Fields

| Field | Type | Notes |
|-------|------|-------|
| `compartment_id` | StringValueOrRef | Required, default_kind: OciCompartment |
| `vcn_id` | StringValueOrRef | Required, default_kind: OciVcn |
| `cidr_block` | string | Required |
| `display_name` | string | Falls back to metadata.name |
| `dns_label` | string | Subnet FQDN component |
| `availability_domain` | string | Omit for regional subnet |
| `prohibit_public_ip_on_vnic` | bool | Private subnet control |
| `prohibit_internet_ingress` | bool | Ingress traffic control |
| `dhcp_options_id` | StringValueOrRef | Custom DHCP options |
| `route_table_id` | StringValueOrRef | External route table (mutually exclusive with route_rules) |
| `security_list_ids` | repeated StringValueOrRef | Max 5 per OCI limit |
| `ipv6_cidr_block` | string | Dual-stack subnet support |
| `route_rules` | repeated RouteRule | Creates custom route table (mutually exclusive with route_table_id) |

### Validation Coverage

21 test specs covering:
- Minimal valid subnet, fully-specified subnet, public/private variants
- Route table via routeTableId, via routeRules, via VCN default (neither)
- Multiple route rules (NAT GW + Service GW)
- compartmentId and vcnId via valueFrom references
- Maximum 5 security lists, exceeding 5 (rejected)
- Mutual exclusivity: routeTableId + routeRules (rejected)
- Missing required fields: compartmentId, vcnId, cidrBlock, route rule destination/networkEntityId
- Wrong apiVersion and kind values

## Benefits

- Unblocks all downstream OCI components that require subnet infrastructure
- Provides route table creation without needing a separate component
- Full parity between Pulumi and Terraform implementations
- Composable with OciVcn outputs via StringValueOrRef for infra-chart assembly

## Impact

- **CloudResourceKind enum**: Added `OciSubnet = 3301` in the 3300-3499 OCI range
- **Kind map**: Regenerated with OciSubnet registration
- **Progress**: 2/37 OCI resource kinds complete

## Related Work

- R01 OciVcn: Parent networking component providing VCN, gateways, and default route table
- R03 OciSecurityGroup: Next component in the queue
- DD02 (VCN bundles gateways): Established the gateway bundling pattern that informed route table bundling here

---

**Status**: Production Ready
**Timeline**: Single session implementation
