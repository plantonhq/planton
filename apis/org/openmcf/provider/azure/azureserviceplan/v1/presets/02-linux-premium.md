# Linux Premium Plan with Zone Redundancy

This preset creates an Azure App Service Plan on the Premium v3 P1v3 tier with 3 Linux workers distributed across availability zones. Premium v3 provides faster processors, SSD storage, double the memory-to-core ratio compared to Standard, and auto-scaling up to 30 instances. Zone balancing ensures high availability across 3 zones.

## When to Use

- Production web apps and APIs requiring zone redundancy and high availability
- Performance-sensitive workloads benefiting from Premium v3 hardware (faster CPUs, SSD storage)
- Applications needing more than 10 instances (Premium scales to 30)
- Teams requiring VNet integration, private endpoints, or larger deployment slots

## Key Configuration Choices

- **P1v3 SKU** (`skuName: P1v3`) -- 2 vCPUs, 8 GiB RAM, SSD storage; auto-scale to 30 instances, 20 staging slots
- **Linux** (`osType: Linux`) -- Runs Linux-based web apps; sets `reserved = true` in the Azure API
- **3 workers** (`workerCount: 3`) -- One per availability zone for even distribution; minimum recommended for zone redundancy
- **Zone balancing** (`zoneBalancingEnabled: true`) -- Distributes instances across availability zones for higher resilience

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., `eastus`, `westeurope`) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<plan-name>` | Name for the App Service Plan (alphanumeric, hyphens, underscores, 1-60 chars) | Choose a descriptive name (e.g., `my-app-plan-p1v3`) |

## Related Presets

- **01-linux-standard** -- Use instead for cost-sensitive workloads that don't need zone redundancy or Premium v3 performance
