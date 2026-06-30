# AzureNetworkSecurityGroup Deployment Component

**Date**: February 13, 2026
**Type**: Feature
**Components**: API Definitions, Azure Provider, Pulumi CLI Integration, Resource Management

## Summary

Added AzureNetworkSecurityGroup (R06) as a deployment component in the Azure provider, completing the first networking security resource in the Azure resource expansion project. The NSG bundles an Azure Network Security Group with its security rules, supporting priority-ordered 5-tuple filtering with Allow/Deny decisions. This is the 7th resource forged (including R00 AzureResourceGroup) in the 24-resource Azure expansion effort.

## Problem Statement / Motivation

Enterprise Azure deployments require per-tier network security controls. NSGs are the primary mechanism for implementing network segmentation and the principle of least privilege in Azure. Without an NSG resource kind, the enterprise-network-foundation infra chart (planned as T03) cannot enforce traffic rules between web, application, and data tiers.

### Pain Points

- No way to declare network security policies as code in Planton for Azure
- Enterprise architectures need per-subnet firewall rules that are version-controlled
- The enterprise-network-foundation infra chart requires NSGs as a core building block

## Solution / What's New

A complete AzureNetworkSecurityGroup deployment component with both Pulumi and Terraform implementations, registered at enum 412 in `cloud_resource_kind.proto`.

### Design Corrections from T02 Plan

Seven corrections were applied during implementation based on deep Azure provider research:

1. **Added `resource_group` (StringValueOrRef)** and **`region`** fields -- missing from T02 spec, required by established pattern
2. **Used strings with CEL validation** instead of proto enums for direction/access/protocol -- matches established pattern (R02-R05)
3. **Added `description` field** on rules (max 140 chars) -- not in T02, valuable for operational documentation
4. **Simplified address prefix precedence** -- plural overrides singular in IaC module, no cross-field CEL validation
5. **Dropped plural port ranges** -- singular covers 80% of cases
6. **Renamed message to `AzureSecurityRule`** -- follows naming convention
7. **Confirmed no subnet association** -- NSG-to-subnet association is an infra-chart concern

### Component Architecture

```
AzureNetworkSecurityGroupSpec
├── region (string, required)
├── resource_group (StringValueOrRef → AzureResourceGroup)
├── name (string, 1-80 chars)
└── security_rules (repeated AzureSecurityRule)
    ├── name, description, priority (100-4096)
    ├── direction ("Inbound"/"Outbound"), access ("Allow"/"Deny")
    ├── protocol ("Tcp"/"Udp"/"Icmp"/"*")
    ├── source_port_range (default "*"), destination_port_range (required)
    ├── source_address_prefix (default "*"), destination_address_prefix (default "*")
    └── source_address_prefixes, destination_address_prefixes (plural overrides)
```

## Implementation Details

### Separate Rules Pattern

Security rules are created as separate `network.NetworkSecurityRule` resources (not inline on the NSG) following the AzureUserAssignedIdentity pattern. This provides per-rule error messages, explicit state management, and avoids the Terraform inline-vs-separate conflict.

### Provider-Authentic Values

All enum-like fields use Azure's exact API values as strings with CEL validation:
- Direction: `"Inbound"`, `"Outbound"` (not UPPER_CASE)
- Access: `"Allow"`, `"Deny"`
- Protocol: `"Tcp"`, `"Udp"`, `"Icmp"`, `"*"`

### Files Created

- `apis/dev/planton/provider/azure/azurenetworksecuritygroup/v1/spec.proto` -- Spec with 12 fields per rule
- `apis/dev/planton/provider/azure/azurenetworksecuritygroup/v1/stack_outputs.proto` -- 2 outputs (nsg_id, nsg_name)
- `apis/dev/planton/provider/azure/azurenetworksecuritygroup/v1/api.proto` -- KRM wiring
- `apis/dev/planton/provider/azure/azurenetworksecuritygroup/v1/stack_input.proto` -- Stack input
- `apis/dev/planton/provider/azure/azurenetworksecuritygroup/v1/spec_test.go` -- 30 validation tests
- `apis/dev/planton/provider/azure/azurenetworksecuritygroup/v1/iac/pulumi/module/` -- Pulumi module (main, locals, outputs)
- `apis/dev/planton/provider/azure/azurenetworksecuritygroup/v1/iac/tf/` -- Terraform module (main, variables, outputs, locals, provider)
- Documentation: README.md, examples.md, docs/README.md

### Enum Registration

`AzureNetworkSecurityGroup = 412` in `cloud_resource_kind.proto`, positioned between AzureSubnet (411) and AzurePublicIp (413).

## Benefits

- **Infra chart enablement** -- enterprise-network-foundation can now create per-tier NSGs
- **Complete validation** -- 30 tests covering all valid and invalid input combinations
- **Production-quality docs** -- 6 YAML examples covering minimal, web-tier, app-tier, data-tier, multi-source, and infra-chart patterns
- **Dual IaC** -- Both Pulumi and Terraform with feature parity

## Impact

- **Azure resource coverage**: 7 of 24 resources completed (R00-R06)
- **Downstream consumers**: AzureVirtualMachine (network_security_group_id), infra charts (subnet association)
- **Next resource**: R07 AzurePrivateDnsZone

## Related Work

- **Parent project**: 20260212.05.sp.azure-resource-expansion
- **Previous**: R05 AzureSubnet (2026-02-13)
- **Next**: R07 AzurePrivateDnsZone (enum 415, id_prefix azpdns)
- **Design decisions**: DD03 (Composite Bundling Rules -- NSG bundles rules)

---

**Status**: Production Ready
**Timeline**: Single session
