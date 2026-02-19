# OCI Network Security Group Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Provider Framework, Pulumi Module, Terraform Module

## Summary

Added OciNetworkSecurityGroup (R03) as the third OCI resource kind in OpenMCF, providing a complete deployment component for Oracle Cloud Infrastructure Network Security Groups with inline ingress and egress security rules. The component introduces user-friendly protocol and target type enums, supports all five OCI protocol types (all, TCP, UDP, ICMP, ICMPv6), and enforces OCI's 120-rule limit via CEL validation.

## Problem Statement / Motivation

After OciVcn (R01) and OciSubnet (R02), network security groups were the next critical building block. NSGs are OCI's recommended approach for fine-grained per-VNIC traffic control, preferred over security lists which apply at the subnet level. Multiple downstream components (OKE clusters, compute instances, load balancers) require NSG OCIDs for their security configuration.

### Pain Points

- No way to manage OCI network security groups through OpenMCF
- Downstream components (R07 OciComputeInstance, R08 OciContainerEngineCluster, R11 OciLoadBalancer) blocked waiting for NSG infrastructure
- OCI's raw security rule API uses numeric protocol strings and conditional source/destination fields based on direction, creating a confusing user experience

## Solution / What's New

Implemented `OciNetworkSecurityGroup` as a full deployment component with a thoughtful UX layer over the raw OCI API.

### Key Design Decisions

**Separate ingress/egress rule lists**: Instead of a single `rules` list with an explicit `direction` field (matching the OCI API), the spec uses separate `ingressRules` and `egressRules` repeated fields. This makes direction implicit from the field name, eliminates the conditional "source is required for INGRESS, destination is required for EGRESS" complexity, and produces cleaner YAML.

**Protocol enum**: OCI's API uses protocol number strings (`"1"`, `"6"`, `"17"`, `"58"`, `"all"`). The spec defines a `Protocol` enum with human-readable values (`all`, `tcp`, `udp`, `icmp`, `icmpv6`). The IaC modules handle the mapping internally.

**TargetType enum**: Source and destination types share a `TargetType` enum (`cidr_block`, `service_cidr_block`, `network_security_group`), defaulting to `cidr_block` when unspecified.

**Plain string for source/destination**: These fields are polymorphic (CIDR blocks, service CIDR labels, or NSG OCIDs depending on type). Using `StringValueOrRef` would be semantically misleading for the 95% CIDR case. Users needing NSG-to-NSG references can use OCIDs directly.

**Port range validation**: `PortRange` sub-message validates that ports are in the 1-65535 range and `min <= max` via both field-level and CEL message-level constraints.

### Component Structure

```
apis/org/openmcf/provider/oci/ocinetworksecuritygroup/v1/
├── spec.proto              # 5 top-level fields + 8 nested messages/enums + 2 CEL validations
├── api.proto               # KRM wiring (OciNetworkSecurityGroup, OciNetworkSecurityGroupStatus)
├── stack_input.proto        # OciNetworkSecurityGroupStackInput
├── stack_outputs.proto      # 1 output (network_security_group_id)
├── spec_test.go             # 29 Ginkgo specs (16 valid + 13 invalid cases)
├── iac/pulumi/module/
│   ├── main.go              # Orchestrator: provider setup, create NSG, then rules
│   ├── locals.go            # Display name fallback, freeform tags from metadata
│   ├── outputs.go           # Output name constant
│   ├── nsg.go               # core.NetworkSecurityGroup resource
│   └── security_rules.go    # Iterates ingress + egress rules with protocol/type mapping
└── iac/tf/
    ├── main.tf              # oci_core_network_security_group resource
    ├── security_rules.tf    # oci_core_network_security_group_security_rule with for_each
    ├── variables.tf         # Typed variable definitions with nested rule objects
    ├── outputs.tf           # network_security_group_id output
    ├── locals.tf            # Protocol/type mappings, flattened rules, tags
    └── provider.tf          # OCI provider >= 5.0
```

## Implementation Details

### Proto Spec

The spec defines 8 nested types inside `OciNetworkSecurityGroupSpec`:
- `Protocol` enum (6 values) -- maps to OCI protocol numbers
- `TargetType` enum (4 values) -- maps to OCI source/destination type strings
- `IngressRule` message -- source-based with 8 fields
- `EgressRule` message -- destination-based with 8 fields
- `PortRange` message -- min/max with range validation
- `TcpOptions` message -- destination + source port ranges
- `UdpOptions` message -- destination + source port ranges
- `IcmpOptions` message -- type + optional code (uses proto3 `optional` for code=0 disambiguation)

Two CEL validations:
1. Total rules limit: `ingress_rules.size() + egress_rules.size() <= 120`
2. Port range ordering: `min <= max` (on PortRange message)

### Pulumi Module

The `security_rules.go` file contains three builder functions (`buildTcpOptions`, `buildUdpOptions`, `buildIcmpOptions`) that translate proto messages to Pulumi SDK types. Protocol and target type mapping functions convert enums to the strings OCI expects. Each rule creates a separate `core.NewNetworkSecurityGroupSecurityRule` resource with an implicit dependency on the NSG via `.ID()`.

### Terraform Module

Rules are flattened in `locals.tf` from separate ingress/egress lists into a single `all_rules` map keyed by `"ingress-0"`, `"egress-1"`, etc. Protocol-specific options use `dynamic` blocks to conditionally include `tcp_options`, `udp_options`, or `icmp_options` based on the input.

### Directory Naming

The plan originally proposed `ocinsg/` as the directory name (matching the id_prefix). During implementation, the kind map code generator was found to derive directory names by lowercasing the kind name (`ocinetworksecuritygroup`). The directory was corrected to `ocinetworksecuritygroup/` to match the convention used by OciVcn (`ocivcn/`) and OciSubnet (`ocisubnet/`).

## Benefits

- **User-friendly YAML**: Protocol names (`tcp`, `udp`, `icmp`) instead of numbers; implicit direction from field name
- **Validation at spec level**: 120-rule limit, port range constraints, required source/destination enforced before deployment
- **Full composability**: NSG OCID exported for downstream components via `status.outputs.networkSecurityGroupId`
- **Feature parity**: Both Pulumi and Terraform modules support all five protocol types with port range and ICMP option configuration

## Impact

- **3/37 OCI resource kinds now complete** (OciVcn, OciSubnet, OciNetworkSecurityGroup)
- **Phase 1 Foundation networking complete**: All three networking building blocks are in place
- **Unblocks**: R04-R06 (Identity), R07-R10 (Compute/Containers), R11-R14 (Advanced Networking)

## Related Work

- R01 OciVcn -- Foundation networking, established OCI component patterns
- R02 OciSubnet -- Subnet with inline route tables, established sub-resource bundling pattern
- Next: R04 OciCompartment -- First identity component

---

**Status**: Production Ready
**Timeline**: Single session
