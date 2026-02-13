# AzureApplicationInsights Pulumi Module

## Overview

This Pulumi module provisions an Azure Application Insights resource using the Azure
Classic provider (`pulumi-azure`). It creates a single `appinsights.Insights` resource
with configurable application type, workspace integration, retention, daily cap, and
sampling percentage.

## Resources Created

- `appinsights.Insights` -- the Application Insights resource

## Inputs

The module receives an `AzureApplicationInsightsStackInput` containing:

- `target.spec.region` -- Azure region
- `target.spec.resource_group` -- resource group name (resolved from StringValueOrRef)
- `target.spec.name` -- Application Insights name
- `target.spec.application_type` -- application type (default: web)
- `target.spec.workspace_id` -- Log Analytics Workspace ID (resolved from StringValueOrRef)
- `target.spec.retention_in_days` -- data retention period (default: 90)
- `target.spec.daily_data_cap_in_gb` -- daily ingestion cap (default: 100)
- `target.spec.sampling_percentage` -- telemetry sampling rate (default: 100)
- `target.metadata` -- OpenMCF metadata for tagging
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `app_insights_id` | Azure Resource Manager ID |
| `instrumentation_key` | Instrumentation key (secret) |
| `connection_string` | SDK connection string (secret) |
| `app_id` | Application ID for API access |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest openmcf
```
