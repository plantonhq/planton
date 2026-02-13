# AzureLogAnalyticsWorkspace Pulumi Module

## Overview

This Pulumi module provisions an Azure Log Analytics Workspace using the Azure Classic
provider (`pulumi-azure`). It creates a single `operationalinsights.AnalyticsWorkspace`
with configurable SKU, retention, and daily quota.

## Resources Created

- `operationalinsights.AnalyticsWorkspace` -- the Log Analytics Workspace

## Inputs

The module receives an `AzureLogAnalyticsWorkspaceStackInput` containing:

- `target.spec.region` -- Azure region
- `target.spec.resource_group` -- resource group name (resolved from StringValueOrRef)
- `target.spec.name` -- workspace name
- `target.spec.sku` -- pricing tier (default: PerGB2018)
- `target.spec.retention_in_days` -- data retention period (default: 30)
- `target.spec.daily_quota_gb` -- daily ingestion cap (default: -1, unlimited)
- `target.metadata` -- OpenMCF metadata for tagging
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `workspace_id` | Azure Resource Manager ID |
| `workspace_name` | Name of the workspace |
| `primary_shared_key` | Primary authentication key (secret) |
| `secondary_shared_key` | Secondary authentication key (secret) |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest openmcf
```
