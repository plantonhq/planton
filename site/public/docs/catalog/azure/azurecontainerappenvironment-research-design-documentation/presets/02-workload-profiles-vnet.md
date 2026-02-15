---
title: "Workload Profiles with VNet Integration"
description: "This preset creates a production-grade Azure Container App Environment with VNet injection, internal load balancer (no public internet exposure), zone redundancy, and a D4 dedicated workload profile...."
type: "preset"
rank: "02"
presetSlug: "02-workload-profiles-vnet"
componentSlug: "azurecontainerappenvironment-research-design-documentation"
componentTitle: "AzureContainerAppEnvironment: Research & Design Documentation"
provider: "azure"
icon: "package"
order: 2
---

# Workload Profiles with VNet Integration

This preset creates a production-grade Azure Container App Environment with VNet injection, internal load balancer (no public internet exposure), zone redundancy, and a D4 dedicated workload profile. Apps in this environment have private connectivity to databases, storage, and other VNet resources.

## When to Use

- Production environments requiring private networking and no public internet exposure
- Workloads that need VNet connectivity to databases, storage accounts, or on-premises resources
- High-availability requirements with zone redundancy across 3 availability zones
- Applications needing dedicated compute (guaranteed CPU/memory) rather than serverless Consumption plan

## Key Configuration Choices

- **VNet-injected** (`infrastructureSubnetId`) -- Apps run inside your VNet with private connectivity; subnet must be /21 or larger
- **Internal load balancer** (`internalLoadBalancerEnabled: true`) -- Apps are accessible only from within the VNet; no public internet access
- **Zone redundancy** (`zoneRedundancyEnabled: true`) -- Infrastructure is distributed across availability zones for higher resilience
- **D4 workload profile** (`workloadProfileType: D4`) -- 4 vCPUs, 16 GiB RAM dedicated VMs; pre-warmed with 1 instance, scales to 5
- **Log Analytics linked** (`logAnalyticsWorkspaceId`) -- Centralized logging for monitoring and alerting
- **Consumption still available** -- The Consumption profile is always present alongside dedicated profiles for lightweight workloads

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., `eastus`, `westeurope`) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<environment-name>` | Name for the Container App Environment (lowercase, hyphens, 2-60 chars) | Choose a descriptive name (e.g., `prod-env`) |
| `<infrastructure-subnet-id>` | ARM resource ID of the subnet (/21 or larger) | Azure portal or `AzureSubnet` status outputs |
| `<log-analytics-workspace-id>` | ARM resource ID of the Log Analytics workspace | Azure portal or `AzureLogAnalyticsWorkspace` status outputs |

## Related Presets

- **01-consumption** -- Use instead for development/staging environments that don't need VNet injection or dedicated compute
