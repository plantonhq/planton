# Standard Resource Group

This preset creates a standard Azure Resource Group, the foundational organizational unit for all Azure resources. Every Azure resource must belong to a resource group. This is the simplest and most common preset -- virtually all Azure deployments begin by creating one or more resource groups.

## When to Use

- Starting any new Azure project or environment
- Grouping related resources for lifecycle management, RBAC, and cost tracking
- Creating environment-specific boundaries (e.g., one resource group per environment)

## Key Configuration Choices

- **Name** (`name`) -- Unique within the Azure subscription. Use a descriptive convention like `{project}-{env}-rg`
- **Region** (`region`) -- Determines where the resource group metadata is stored. Resources inside the group can be in different regions

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group-name>` | Name for the resource group (unique within subscription) | Your naming convention |
| `<azure-region>` | Azure region (e.g., `eastus`, `westeurope`, `southeastasia`) | [Azure regions list](https://learn.microsoft.com/en-us/azure/reliability/availability-zones-overview) |
