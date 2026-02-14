# Standard Log Analytics Workspace

This preset creates an Azure Log Analytics Workspace with pay-as-you-go pricing, 30-day retention, and unlimited ingestion. Log Analytics Workspaces are the central data platform for Azure Monitor -- they are required by AKS Container Insights, Application Insights, Microsoft Sentinel, and most Azure monitoring solutions. This is the standard configuration for most environments.

## When to Use

- Setting up monitoring for AKS clusters (Container Insights requires a workspace)
- Centralizing log collection from Azure resources, VMs, and applications
- Any Azure environment that needs logging and diagnostics

## Key Configuration Choices

- **SKU** (`sku: PerGB2018`) -- Pay-as-you-go pricing per GB ingested. Best for most workloads; switch to `CapacityReservation` only at 100+ GB/day for cost savings
- **Retention** (`retentionInDays: 30`) -- 30 days included free with PerGB2018. Increase to 90+ days for compliance workloads (billed per GB/month beyond 31 days)
- **Unlimited ingestion** (`dailyQuotaGb: -1`) -- No daily cap. Set a positive value to control costs in development environments

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (choose close to monitored resources) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-workspace-name>` | Name for the workspace (4-63 chars, alphanumeric and hyphens) | Your naming convention |
