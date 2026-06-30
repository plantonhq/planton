# AzureSubnet Deployment Component

**Date**: February 13, 2026
**Type**: Feature
**Components**: API Definitions, Azure Provider, Pulumi CLI Integration, Resource Management

## Summary

Added `AzureSubnet` (enum 411, id_prefix `azsub`) as a standalone deployment component for Azure Virtual Network subnets. This is the most widely referenced Azure resource in Planton -- 11 downstream resource types consume its `subnet_id` output, making it a critical building block for Azure infra charts including database-stack, enterprise-network-foundation, and container-apps-environment.

## Problem Statement / Motivation

The existing `AzureVpc` resource creates a VNet with a single built-in `nodes_subnet` for AKS. Enterprise Azure architectures require multiple subnets with different configurations: delegated subnets for PostgreSQL and Container Apps, private endpoint subnets, Application Gateway subnets, and management subnets. Without a standalone subnet resource, multi-tier network architectures cannot be composed in infra charts.

### Pain Points

- No way to create additional subnets beyond the AKS nodes_subnet
- Cannot delegate subnets to Azure PaaS services (PostgreSQL, MySQL, Container Apps)
- Cannot configure private endpoint network policies per-subnet
- Downstream resources (11 types) need `subnet_id` references that don't exist yet

## Solution / What's New

A complete deployment component following the forge workflow pattern with both Pulumi and Terraform implementations, full proto validation tests, and production-quality documentation.

### Spec Design Corrections

Deep research against the `azurerm_subnet` Terraform provider schema and Pulumi Azure Native SDK revealed 6 corrections needed from the original T02 plan spec:

1. **Added `resource_group`** -- omitted in original spec but required by all Azure providers and consistent with every other Azure resource pattern
2. **Omitted `region`** -- deliberate deviation; subnets inherit region from VNet, including it would be misleading
3. **Kept `address_prefix` singular** -- Azure supports plural but 99.9% of subnets use one CIDR; IaC modules wrap in list
4. **Replaced `bool private_endpoint_network_policies_enabled`** with `string private_endpoint_network_policies` -- Azure deprecated the boolean in favor of a 4-value enum (Disabled, Enabled, NetworkSecurityGroupEnabled, RouteTableEnabled)
5. **Simplified `AzureSubnetDelegation`** -- flat message with `name`, `service_name`, `actions` instead of nested structure
6. **Fixed enum number** from 410 to 411 (`AzureDnsRecord` already occupies 410)

## Implementation Details

### Proto API (4 files)

- `spec.proto` -- 8 fields with `StringValueOrRef` for `resource_group` and `vnet_id`, CEL validation on `private_endpoint_network_policies`, nested `AzureSubnetDelegation` message
- `stack_outputs.proto` -- 3 outputs: `subnet_id`, `subnet_name`, `address_prefix`
- `api.proto` -- KRM wiring with `azure.planton.dev/v1` api_version
- `stack_input.proto` -- Standard pattern with `AzureProviderConfig`

### Spec Validation Tests (21 tests)

- 9 positive cases: minimal, full metadata, service endpoints, delegation (with/without actions), network policies, all fields
- 12 negative cases: missing required fields, name exceeds max length, invalid enum value, wrong api_version/kind, missing metadata/spec

### Pulumi Module

- `locals.go` -- extracts VNet name from ARM resource ID via string split
- `main.go` -- creates `network.Subnet` with conditional delegation and service endpoints
- Uses `pulumi-azure` v6 (Azure Classic provider), consistent with all other Azure components

### Terraform Module

- `locals.tf` -- extracts VNet name from ARM ID using `split()` + `element()`
- `main.tf` -- `azurerm_subnet` resource with `dynamic` delegation block
- `variables.tf` -- typed object with optional delegation sub-object

### VNet Name Extraction Pattern

Both IaC modules parse the VNet ARM resource ID to extract the VNet name:

```
ARM ID: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{name}
           split by "/"  -->  take last element  -->  VNet name
```

## Benefits

- **Unlocks 11 downstream resources**: Every resource that needs a subnet reference can now use `StringValueOrRef subnet_id`
- **Enables 3 infra charts**: database-stack, enterprise-network-foundation, and container-apps-environment all require AzureSubnet
- **Service delegation support**: PostgreSQL Flexible Server, MySQL Flexible Server, Container App Environment, and App Service VNet integration
- **Granular private endpoint policies**: 4 policy modes for zero-trust architectures
- **Full composability**: `resource_group` and `vnet_id` as `StringValueOrRef` enables proper DAG wiring

## Impact

- **Azure provider**: 17 total Azure resource kinds (11 existing + 6 new including AzureSubnet)
- **Infra charts**: Critical dependency for 3 of 6 planned Azure infra charts
- **Downstream resources**: AKS, Container Apps, PostgreSQL, MySQL, Redis, Private Endpoint, App Gateway, Load Balancer, VM, Function App, Web App

## Related Work

- **DD01**: AzureSubnet as standalone resource (design decision approving separate lifecycle)
- **R04 AzurePublicIp**: Previous resource in the forge queue
- **R06 AzureNetworkSecurityGroup**: Next resource in the forge queue

---

**Status**: Production Ready
**Timeline**: Single session forge
