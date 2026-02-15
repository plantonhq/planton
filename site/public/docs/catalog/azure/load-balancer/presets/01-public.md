---
title: "Public Load Balancer"
description: "This preset creates a public (internet-facing) Azure Load Balancer with Standard SKU, a single backend pool, an HTTP health probe, and a TCP load balancing rule on port 80. This is the standard..."
type: "preset"
rank: "01"
presetSlug: "01-public"
componentSlug: "load-balancer"
componentTitle: "Load Balancer"
provider: "azure"
icon: "package"
order: 1
---

# Public Load Balancer

This preset creates a public (internet-facing) Azure Load Balancer with Standard SKU, a single backend pool, an HTTP health probe, and a TCP load balancing rule on port 80. This is the standard configuration for distributing inbound internet traffic across VMs, VMSS instances, or AKS nodes behind a public IP address.

## When to Use

- Internet-facing web applications that need Layer 4 load balancing across multiple backend instances
- Public APIs or services that require high availability with health-checked traffic distribution
- AKS clusters or VMSS deployments that need a dedicated public entry point separate from Azure-managed LBs
- Replacing or supplementing Azure's default load balancer with explicit control over rules and probes

## Key Configuration Choices

- **Public frontend** (`publicIpId`) -- Associates a Standard SKU public IP for internet-facing traffic. The referenced public IP must be Standard SKU with static allocation
- **Single backend pool** (`backendPools: [default]`) -- One pool for all backends. Pool membership (VMs, VMSS, NICs) is managed outside this component
- **HTTP health probe** (`healthProbes: Http on port 80`) -- Checks `/health` endpoint every 15 seconds; marks backend unhealthy after 2 consecutive failures
- **TCP rule on port 80** (`rules: Tcp 80→80`) -- Routes inbound TCP port 80 to backend port 80 with 4-minute idle timeout

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match backend resources) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-lb-name>` | Name for the load balancer (unique within resource group) | Your naming convention |
| `<public-ip-resource-id>` | Full ARM resource ID of a Standard SKU public IP | Azure portal or `AzurePublicIp` status outputs |

## Related Presets

- **02-internal** -- Use instead for private VNet load balancing without internet exposure
