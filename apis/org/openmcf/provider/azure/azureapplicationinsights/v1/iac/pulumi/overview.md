# AzureApplicationInsights Pulumi Module -- Architecture Overview

## Purpose

This module is the Pulumi implementation for the `AzureApplicationInsights` OpenMCF
component. It translates the protobuf-defined spec into Azure infrastructure using the
Pulumi Azure Classic SDK.

## Architecture

```
AzureApplicationInsightsStackInput
  ├── target (AzureApplicationInsights)
  │     ├── metadata (name, org, env)
  │     └── spec (region, resource_group, name, application_type,
  │              workspace_id, retention, daily_cap, sampling)
  └── provider_config (credentials)
         │
         ▼
  ┌─────────────────┐
  │  module/main.go  │  Creates azure provider + Application Insights
  │  module/locals.go│  Extracts spec values, resolves StringValueOrRef, builds tags
  │  module/outputs.go│ Defines output constant names
  └─────────────────┘
         │
         ▼
  Stack Outputs: app_insights_id, instrumentation_key, connection_string, app_id
```

## Key Implementation Details

### StringValueOrRef Fields

Two fields use `StringValueOrRef`:
- `resource_group` -- references an AzureResourceGroup output
- `workspace_id` -- references an AzureLogAnalyticsWorkspace output

The platform middleware resolves all `valueFrom` references before the IaC module runs,
so the module simply calls `.GetValue()` to extract the resolved string. No helper
functions or conversion logic needed.

### Application Type Passthrough

The `application_type` field is a plain string that matches Azure's exact API values
(`"web"`, `"java"`, `"Node.JS"`, `"other"`). The Pulumi module passes this value
directly to Azure -- no enum-to-string conversion needed.

### Sensitive Outputs

The `instrumentation_key` and `connection_string` outputs contain authentication
credentials. In the Pulumi stack, these are automatically treated as secrets because
the underlying Azure resource marks them as sensitive.
