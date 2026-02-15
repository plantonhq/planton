---
title: "Consumption Plan Environment"
description: "This preset creates a minimal Azure Container App Environment using the default Consumption (serverless) plan with no VNet injection. Apps deployed to this environment share Azure-managed networking..."
type: "preset"
rank: "01"
presetSlug: "01-consumption"
componentSlug: "azurecontainerappenvironment-research-design-documentation"
componentTitle: "AzureContainerAppEnvironment: Research & Design Documentation"
provider: "azure"
icon: "package"
order: 1
---

# Consumption Plan Environment

This preset creates a minimal Azure Container App Environment using the default Consumption (serverless) plan with no VNet injection. Apps deployed to this environment share Azure-managed networking and scale to zero when idle. Log Analytics is configured for centralized log collection.

## When to Use

- Development, staging, or small production workloads that don't need VNet connectivity
- Teams getting started with Azure Container Apps and wanting the simplest setup
- Cost-sensitive workloads that benefit from pay-per-use pricing and scale-to-zero
- Applications that can use public endpoints and don't require private networking

## Key Configuration Choices

- **Consumption plan** (no `workloadProfiles`) -- Serverless, pay-per-use; Azure manages the infrastructure and auto-scales
- **No VNet injection** (no `infrastructureSubnetId`) -- Azure manages networking; apps get public endpoints by default
- **External mode** (`internalLoadBalancerEnabled` defaults to `false`) -- Apps can receive traffic from the public internet
- **Log Analytics linked** (`logAnalyticsWorkspaceId`) -- Container app logs are persisted for KQL querying and alerting

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., `eastus`, `westeurope`) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<environment-name>` | Name for the Container App Environment (lowercase, hyphens, 2-60 chars) | Choose a descriptive name (e.g., `dev-env`, `staging-env`) |
| `<log-analytics-workspace-id>` | ARM resource ID of the Log Analytics workspace | Azure portal or `AzureLogAnalyticsWorkspace` status outputs |

## Related Presets

- **02-workload-profiles-vnet** -- Use instead for production environments needing VNet integration, dedicated compute, zone redundancy, and internal-only access
