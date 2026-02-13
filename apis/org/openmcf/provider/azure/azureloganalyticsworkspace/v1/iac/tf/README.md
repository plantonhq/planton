# AzureLogAnalyticsWorkspace Terraform Module

## Overview

This Terraform module provisions an Azure Log Analytics Workspace using the `azurerm`
provider. It creates a single `azurerm_log_analytics_workspace` with configurable
SKU, retention, and daily ingestion quota.

## Resources Created

- `azurerm_log_analytics_workspace.main` -- the Log Analytics Workspace

## Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | OpenMCF metadata (name, org, env) |
| `spec` | object | Workspace specification (region, resource_group, name, sku, retention, quota) |

## Outputs

| Output | Description | Sensitive |
|--------|-------------|-----------|
| `workspace_id` | Azure Resource Manager ID | No |
| `workspace_name` | Name of the workspace | No |
| `primary_shared_key` | Primary authentication key | Yes |
| `secondary_shared_key` | Secondary authentication key | Yes |

## Usage

```hcl
module "law" {
  source = "./iac/tf"

  metadata = {
    name = "platform-law"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region            = "eastus"
    resource_group    = "prod-monitoring-rg"
    name              = "prod-platform-law"
    sku               = "PerGB2018"
    retention_in_days = 90
    daily_quota_gb    = -1
  }
}
```
