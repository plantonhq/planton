# AzurePrivateDnsZone Deployment Component

**Date**: February 13, 2026
**Type**: Feature
**Components**: Azure Provider, API Definitions, Pulumi IaC, Terraform IaC

## Summary

Added AzurePrivateDnsZone (R07) as a new deployment component in the Azure resource expansion initiative. This component creates an Azure Private DNS Zone with a bundled Virtual Network link, enabling private name resolution for Private Link scenarios (database servers, Key Vault, Storage) and custom internal DNS. Two spec corrections from deep provider research were applied: adding the missing `resource_group` field and a new `registration_enabled` toggle for VM auto-registration.

## Problem Statement / Motivation

Azure Private Endpoints require properly configured private DNS zones to function correctly. Without a privatelink zone linked to the VNet, service FQDNs resolve to public IPs instead of private endpoint IPs, completely bypassing the private connectivity. This component is a critical dependency for the upcoming database-stack infra chart (R11-R15) and any Private Link-based architecture.

### Pain Points

- No way to provision private DNS zones through OpenMCF before this component
- Database servers (PostgreSQL, MySQL) requiring VNet integration need a corresponding privatelink zone
- Manual DNS zone creation is error-prone (zone names must exactly match Azure's predefined patterns)
- Missing infrastructure-as-code support for private DNS in the Azure resource collection

## Solution / What's New

A complete deployment component following the OpenMCF forge pattern:

- **4 proto files** with buf.validate rules and CEL validation for DNS zone name format
- **Pulumi module** using `privatedns.NewZone` and `privatedns.NewZoneVirtualNetworkLink`
- **Terraform module** using `azurerm_private_dns_zone` and `azurerm_private_dns_zone_virtual_network_link`
- **17 validation tests** covering all valid and invalid input combinations
- **Production-quality documentation** with 7 YAML examples covering privatelink, custom DNS, and infra chart patterns

### Spec Corrections from Provider Research

Two corrections were discovered during deep research into the Azure Terraform provider source code:

1. **Added `resource_group` field** (StringValueOrRef, required) -- missing from the original T02 spec design. This was the same omission pattern found and corrected in R05 (AzureSubnet).

2. **Added `registration_enabled` field** (optional bool, default false) -- enables the auto-registration use case for custom internal DNS zones while defaulting to the correct behavior for privatelink zones.

### Design Decision: No Region Field

Private DNS zones are global Azure resources with no location parameter. This was confirmed in both the Terraform provider source and the Pulumi SDK. The component correctly omits the `region` field.

## Implementation Details

### Proto API

- `spec.proto`: 4 fields (resource_group, name, vnet_id, registration_enabled)
- Zone name validated via CEL regex for DNS domain format
- `resource_group` and `vnet_id` use StringValueOrRef with default_kind annotations
- `registration_enabled` uses optional + default pattern

### IaC Modules

Both Pulumi and Terraform modules create two resources with feature parity:
1. Private DNS zone (global, no region)
2. VNet link with auto-derived name (`{metadata.name}-vnet-link`)

### Bundling Rationale (DD03)

A private DNS zone without a VNet link is unreachable from any VNet. The VNet link is bundled as a structural dependency, following the same pattern as NSG + rules and UserAssignedIdentity + role assignments.

## Benefits

- Enables Private Link connectivity for all upcoming database resources (R11-R15)
- Supports both privatelink and custom internal DNS use cases
- StringValueOrRef on resource_group and vnet_id enables infra chart composition
- 17 comprehensive validation tests ensure correctness
- Consistent with all other Azure resource patterns in the collection

## Impact

- **Downstream consumers**: AzurePrivateEndpoint, AzurePostgresqlFlexibleServer, AzureMysqlFlexibleServer
- **Infra charts**: database-stack (primary), enterprise-network-foundation (optional)
- **Enum registration**: CloudResourceKind 415 (AzurePrivateDnsZone, id_prefix: azpdns)
- **Resource count**: 8 of 24 Azure resources now completed

## Related Work

- Part of 20260212.05.sp.azure-resource-expansion (23 new Azure resources)
- Follows R06 AzureNetworkSecurityGroup (same session patterns)
- Prerequisite for R08 AzurePrivateEndpoint (next in queue)
- Foundation for T03 database-stack infra chart

---

**Status**: Production Ready
**Tests**: 17/17 passing
**Build**: Go build + proto generation successful
