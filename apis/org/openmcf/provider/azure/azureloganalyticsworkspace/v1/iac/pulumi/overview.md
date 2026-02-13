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
  │  module/locals.go│  Resolves StringValueOrRef, builds tags
  │  module/outputs.go│ Defines output constant names
  └─────────────────┘
         │
         ▼
  Stack Outputs: workspace_id, workspace_name, primary_shared_key, secondary_shared_key
```

## Key Implementation Details

### StringValueOrRef Resolution

The `resource_group` field uses `StringValueOrRef`. OpenMCF middleware resolves
`valueFrom` references before the IaC module runs. The `resolveStringValueOrRef`
helper in `locals.go` extracts the resolved string value.

### Daily Quota Handling

Azure's `daily_quota_gb` defaults to `-1` (unlimited). The module only sets
`DailyQuotaGb` on the workspace when the value is `>= 0`, avoiding unnecessary
API calls for the default case.
