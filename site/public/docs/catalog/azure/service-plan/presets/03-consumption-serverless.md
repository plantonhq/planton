---
title: "Consumption (Serverless) Service Plan"
description: "This preset creates an Azure App Service Plan with the Consumption (Y1) SKU — a fully serverless plan that scales to zero and bills per execution. The Consumption plan is the cheapest option for..."
type: "preset"
rank: "03"
presetSlug: "03-consumption-serverless"
componentSlug: "service-plan"
componentTitle: "Service Plan"
provider: "azure"
icon: "package"
order: 3
---

# Consumption (Serverless) Service Plan

This preset creates an Azure App Service Plan with the Consumption (Y1) SKU — a fully serverless plan that scales to zero and bills per execution. The Consumption plan is the cheapest option for Azure Functions, costing $0 when idle and ~$0.20 per million executions plus ~$0.000016 per GB-second of compute. Each month includes 1 million free executions and 400,000 GB-seconds free.

## When to Use

- Event-driven Azure Functions with sporadic or unpredictable traffic
- Lightweight HTTP APIs with low-to-moderate request volumes
- Background processing triggered by queues, timers, or webhooks
- Cost-optimized deployments where you only pay for actual compute used
- Development and staging environments for serverless functions

## Key Configuration Choices

- **Consumption SKU** (`skuName: Y1`) -- True serverless: scales from 0 to 200 instances automatically. No worker count or scaling configuration needed — Azure manages everything
- **Linux** (`osType: Linux`) -- ForceNew. Linux plans support Node.js, Python, Java, .NET, and PowerShell. Change to Windows if your functions require Windows-specific features
- **No worker count** -- The Consumption plan auto-manages instances. `workerCount`, `zoneBalancingEnabled`, and `maximumElasticWorkerCount` do not apply
- **Cold start** -- Functions may experience 1–10 second cold starts after idle periods. Use the Premium (EP) plan if cold starts are unacceptable

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<plan-name>` | Name for the App Service Plan (1-60 chars; ForceNew) | Choose a descriptive name |

## Related Presets

- **01-linux-standard** -- Use instead for always-on web apps with consistent compute (S1)
- **02-linux-premium** -- Use instead for zone-redundant production with scaling (P1v3)
