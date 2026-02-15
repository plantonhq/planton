---
title: "Linux Standard Plan"
description: "This preset creates an Azure App Service Plan on the Standard S1 tier with a single Linux worker. Standard tier provides auto-scaling up to 10 instances, staging slots, daily backups, and a 99.95%..."
type: "preset"
rank: "01"
presetSlug: "01-linux-standard"
componentSlug: "azureserviceplan-research-design-documentation"
componentTitle: "AzureServicePlan: Research & Design Documentation"
provider: "azure"
icon: "package"
order: 1
---

# Linux Standard Plan

This preset creates an Azure App Service Plan on the Standard S1 tier with a single Linux worker. Standard tier provides auto-scaling up to 10 instances, staging slots, daily backups, and a 99.95% SLA. This is the entry-level production configuration for web apps and APIs.

## When to Use

- Production web apps and APIs that need auto-scaling, staging slots, and SLA
- Linux-based workloads (.NET, Node.js, Python, Java, PHP, Ruby on App Service)
- Teams upgrading from Basic tier to get auto-scale and deployment slot capabilities
- Standard compute needs with predictable pricing (1 vCPU, 1.75 GiB RAM per instance)

## Key Configuration Choices

- **S1 SKU** (`skuName: S1`) -- 1 vCPU, 1.75 GiB RAM; auto-scale to 10 instances, 5 staging slots, daily backups
- **Linux** (`osType: Linux`) -- Runs Linux-based web apps; sets `reserved = true` in the Azure API
- **1 worker** (`workerCount: 1`) -- Single instance; increase for higher throughput or enable auto-scale rules on the app
- **No zone balancing** -- Standard SKU does not support zone redundancy; upgrade to Premium for zone support

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., `eastus`, `westeurope`) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<plan-name>` | Name for the App Service Plan (alphanumeric, hyphens, underscores, 1-60 chars) | Choose a descriptive name (e.g., `my-app-plan-s1`) |

## Related Presets

- **02-linux-premium** -- Use instead for production workloads requiring zone redundancy, higher performance, and more scaling capacity
