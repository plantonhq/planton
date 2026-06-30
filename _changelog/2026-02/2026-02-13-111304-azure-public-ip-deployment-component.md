# AzurePublicIp Deployment Component

**Date**: February 13, 2026
**Type**: Feature
**Components**: API Definitions, Azure Provider, Pulumi CLI Integration, Terraform Module

## Summary

Added the `AzurePublicIp` deployment component (enum 413, id_prefix `azpip`) to Planton, provisioning Standard SKU static Public IP Addresses in Azure. This is the 5th resource (R04) in the Azure resource expansion project, and the first networking resource in the queue. The component deliberately omits the retired Basic SKU and always-Dynamic allocation, hardcoding Standard/Static in IaC modules for a clean, modern spec.

## Problem Statement / Motivation

The Azure resource expansion project requires 20 more networking, database, serverless, messaging, and CDN resources. Public IP addresses are a foundational networking primitive -- they are referenced by load balancers, application gateways, and NAT gateways. Without a standalone AzurePublicIp resource, downstream networking resources cannot compose properly in infra charts.

### Pain Points

- No standalone Public IP resource existed in Planton for Azure
- The existing AzureNatGateway created inline Public IPs, preventing reuse across resources
- Enterprise network architectures need explicit control over Public IP lifecycle, DNS labels, and zone placement

## Solution / What's New

A complete deployment component at `apis/dev/planton/provider/azure/azurepublicip/v1/` with:

- 4 proto files with buf-validate constraints (including CEL for domain_name_label format)
- Pulumi module using `network.NewPublicIp` from `pulumi-azure/sdk/v6`
- Terraform module using `azurerm_public_ip`
- 25 validation tests (all passing)
- Production-quality documentation (README, examples, research docs)

### Key Design Decision: Standard-Only

Azure retired the Basic SKU on September 30, 2025. Standard SKU requires static allocation. Rather than exposing two enum fields (`sku` and `allocation_method`) where only one combination is valid, both values are hardcoded in IaC modules. The spec contains only fields users actually configure.

## Implementation Details

### Spec Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string (required) | Azure region |
| `resource_group` | StringValueOrRef (required) | References AzureResourceGroup |
| `name` | string (required) | Public IP name, 1-80 chars |
| `domain_name_label` | string (optional) | DNS label with CEL regex validation |
| `zones` | repeated string (optional) | Availability zones ["1","2","3"] |
| `idle_timeout_in_minutes` | optional int32 | TCP idle timeout, 4-30, default 4 |

### Outputs

| Output | Description |
|--------|-------------|
| `public_ip_id` | ARM resource ID (primary consumer output) |
| `ip_address` | Allocated static IPv4 address |
| `fqdn` | FQDN if domain_name_label set |
| `public_ip_name` | Resource name |

### Files Created

- `apis/dev/planton/provider/azure/azurepublicip/v1/` -- 4 proto files, spec_test.go, docs
- `apis/dev/planton/provider/azure/azurepublicip/v1/iac/pulumi/` -- Go module (main, locals, outputs)
- `apis/dev/planton/provider/azure/azurepublicip/v1/iac/tf/` -- Terraform module (5 files)
- `apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto` -- enum 413 registered

## Benefits

- Standalone Public IP lifecycle management for enterprise Azure architectures
- Composable via `StringValueOrRef public_ip_id` for downstream AzureLoadBalancer, AzureApplicationGateway, AzureNatGateway
- Clean spec with zero deprecated options (no Basic SKU, no Dynamic allocation)
- `idle_timeout_in_minutes` as a production tunable for long-lived connections
- Zone-redundant deployments for production resilience

## Impact

- **Downstream resources**: AzureLoadBalancer (R09) and AzureApplicationGateway (R10) will reference `public_ip_id` via StringValueOrRef
- **Infra charts**: Enables the `enterprise-network-foundation` chart (Public IP as Layer 1 resource)
- **Existing resources**: AzureNatGateway could be refactored to reference an external AzurePublicIp instead of creating inline IPs

## Related Work

- R00: AzureResourceGroup (2026-02-13) -- Layer 0 foundation
- R01: AzureLogAnalyticsWorkspace (2026-02-13) -- first resource with StringValueOrRef resource_group
- R02: AzureApplicationInsights (2026-02-13) -- monitoring layer
- R03: AzureUserAssignedIdentity (2026-02-13) -- identity layer
- **Next**: R05 AzureSubnet -- networking layer, depends on AzureVpc

---

**Status**: Production Ready
**Test Results**: 25/25 passing
