# AzureLogAnalyticsWorkspace Pulumi Module -- Architecture Overview

## Purpose

This module is the Pulumi implementation for the `AzureLogAnalyticsWorkspace` OpenMCF
component. It translates the protobuf-defined spec into Azure infrastructure using the
Pulumi Azure Classic SDK.

## Architecture

```
AzureLogAnalyticsWorkspaceStackInput
  ├── target (AzureLogAnalyticsWorkspace)
  │     ├── metadata (name, org, env)
  │     └── spec (region, resource_group, name, sku, retention, quota)
  └── provider_config (credentials)
         │
         ▼
  ┌─────────────────┐
  │  module/main.go  │  Creates azure provider + workspace
  │  module/locals.go│  Extracts spec values, builds tags
  │  module/outputs.go│ Defines output constant names
  └─────────────────┘
         │
         ▼
  Stack Outputs: workspace_id, workspace_name, primary_shared_key, secondary_shared_key
```

## Key Implementation Details

### StringValueOrRef Fields

The `resource_group` field uses `StringValueOrRef`, which supports both literal
string values and references to other resource outputs. The platform middleware
resolves all `valueFrom` references before the IaC module runs, so the module
simply calls `.GetValue()` to extract the resolved string.

### Daily Quota Handling

Azure's `daily_quota_gb` defaults to `-1` (unlimited). The module only sets
`DailyQuotaGb` on the workspace when the value is `>= 0`, avoiding unnecessary
API calls for the default case.
