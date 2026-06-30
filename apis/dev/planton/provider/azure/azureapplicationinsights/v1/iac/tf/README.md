# AzureApplicationInsights Terraform Module

## Overview

This Terraform module provisions an Azure Application Insights resource using the
`azurerm` provider. It creates a single `azurerm_application_insights` with configurable
application type, workspace integration, retention, daily cap, and sampling percentage.

## Resources Created

- `azurerm_application_insights.main` -- the Application Insights resource

## Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | Planton metadata (name, org, env) |
| `spec` | object | Application Insights specification |

## Outputs

| Output | Description | Sensitive |
|--------|-------------|-----------|
| `app_insights_id` | Azure Resource Manager ID | No |
| `instrumentation_key` | Instrumentation key | Yes |
| `connection_string` | SDK connection string | Yes |
| `app_id` | Application ID for API access | No |

## Usage

```hcl
module "app_insights" {
  source = "./iac/tf"

  metadata = {
    name = "platform-ai"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region              = "eastus"
    resource_group      = "prod-monitoring-rg"
    name                = "prod-platform-ai"
    application_type    = "web"
    workspace_id        = "/subscriptions/.../workspaces/prod-law"
    retention_in_days   = 90
    daily_data_cap_in_gb = 100
    sampling_percentage = 50
  }
}
```
