# AzureApplicationInsights

## Overview

`AzureApplicationInsights` provisions an Azure Application Insights resource --
Azure's Application Performance Management (APM) service. It tracks request
rates, response times, failure rates, dependency calls, exceptions, and custom
telemetry for any instrumented application.

This kind models workspace-based Application Insights only: telemetry is
stored in a referenced `AzureLogAnalyticsWorkspace`. Classic (non-workspace)
mode was retired by Azure in February 2024, so the workspace binding is
required here -- and once set on a resource it can be repointed but never
removed.

## Key Features

- **Full application-type vocabulary** -- all eight Azure kinds (web, java,
  Node.JS, other, ios, phone, store, MobileCenter) as a closed enum with
  case-exact wire mapping (an unmatched raw string would be silently treated
  as ASP.NET by Azure)
- **Cost controls** -- daily data cap with notification toggle, sampling
  percentage, fixed-set retention (30-730 days)
- **Security posture** -- Entra-only ingestion (`local_authentication_enabled:
  false`), private-link-only ingestion and query paths
- **Privacy** -- client IP masking (Azure's default, GDPR-friendly), BYO
  storage for profiler artifacts
- **Governance tags** -- user tags merged over the Planton-derived metadata tags

## When to Use

- APM for anything deployed on `AzureFunctionApp`, `AzureLinuxWebApp`, or
  `AzureContainerApp` -- those kinds reference this resource's
  `connection_string` output
- OpenTelemetry-instrumented workloads exporting to Azure Monitor
- Availability (web-test) monitoring feeding `AzureMonitorMetricAlert`

## Spec Highlights

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | - | Azure region (match the monitored app) |
| `resource_group` | StringValueOrRef | Yes | - | Resource group |
| `application_insights_name` | string | Yes | - | Resource name (1-260 chars; ForceNew) |
| `application_type` | enum | No | WEB | The 8-value Azure vocabulary (ForceNew) |
| `workspace_id` | StringValueOrRef | Yes | - | The Log Analytics Workspace storing telemetry |
| `retention_in_days` | int32 | No | 90 | One of Azure's fixed values (30...730) |
| `daily_data_cap_in_gb` | double | No | 100 | Ingestion stops at the cap until next UTC day |
| `daily_data_cap_notifications_enabled` | bool | No | true | Email when the cap is hit |
| `sampling_percentage` | double | No | 100 | Telemetry sampling (0-100) |
| `ip_masking_enabled` | bool | No | true | Mask client IPs to 0.0.0.0 |
| `local_authentication_enabled` | bool | No | true | Instrumentation-key ingestion allowed |
| `internet_ingestion_enabled` / `internet_query_enabled` | bool | No | true | Public paths (false = AMPLS-only) |
| `force_customer_storage_for_profiler` | bool | No | false | BYO storage for profiler artifacts |
| `tags` | map | No | - | User tags |

## Outputs

| Output | Description |
|--------|-------------|
| `application_insights_id` | ARM resource ID (referenced by web-test metric alerts and diagnostic settings) |
| `application_insights_name` | Resource name |
| `instrumentation_key` | Classic SDK key (secret-bearing; prefer the connection string) |
| `connection_string` | The SDK configuration string -- the seam app kinds reference (secret-bearing) |
| `app_id` | The REST-API query identifier |

## Composition

```yaml
applicationInsightsConnectionString:
  valueFrom:
    kind: AzureApplicationInsights
    name: my-web-app-insights
    fieldPath: status.outputs.connection_string
```

See `presets/` for the standard and production-sampled starting points.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
