# AzureApplicationInsights

## Overview

`AzureApplicationInsights` provisions an Azure Application Insights resource --
Azure's Application Performance Management (APM) service. It provides deep
observability into web applications, APIs, and microservices by tracking request
rates, response times, failure rates, dependency calls, exceptions, and custom
telemetry.

Application Insights is the standard APM layer in Azure, consumed by:

- **Azure Function Apps** -- serverless function monitoring
- **Azure Web Apps** -- web application monitoring
- **Azure Container Apps** -- containerized application monitoring
- **Any application** instrumented with the Application Insights SDK or OpenTelemetry

This component creates workspace-based Application Insights only. Classic
(non-workspace) Application Insights is deprecated by Microsoft and is not supported.

## Key Features

- **StringValueOrRef resource_group** -- references an `AzureResourceGroup` output,
  enabling proper dependency wiring in infra charts
- **StringValueOrRef workspace_id** -- references an `AzureLogAnalyticsWorkspace`
  output, enforcing the modern workspace-based architecture
- **Telemetry sampling** -- configurable sampling percentage (0-100%) for cost control
- **Daily data cap** -- prevent cost overruns with configurable daily ingestion limits
- **Flexible retention** -- choose from 9 Azure-supported retention periods (30-730 days)
- **Provider-authentic application types** -- `"web"`, `"java"`, `"Node.JS"`, `"other"`
  passed directly to Azure with no conversion

## When to Use

- Before deploying Function Apps, Web Apps, or Container Apps that need APM
- As the monitoring layer in function-app-environment, web-app-environment, and
  container-apps-environment infra charts
- When you need application-level telemetry beyond infrastructure metrics
- When building end-to-end observability with Log Analytics Workspace + App Insights

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | - | Azure region |
| `resource_group` | StringValueOrRef | Yes | - | Resource group (literal or valueFrom) |
| `name` | string | Yes | - | Resource name (1-260 chars) |
| `application_type` | string | No | web | Application type: web, java, Node.JS, other |
| `workspace_id` | StringValueOrRef | Yes | - | Log Analytics Workspace ID (literal or valueFrom) |
| `retention_in_days` | int32 | No | 90 | Data retention (30/60/90/120/180/270/365/550/730) |
| `daily_data_cap_in_gb` | double | No | 100 | Daily ingestion cap in GB |
| `sampling_percentage` | double | No | 100 | Telemetry sampling rate (0-100%) |

## Outputs

| Output | Description |
|--------|-------------|
| `app_insights_id` | Azure Resource Manager ID |
| `instrumentation_key` | Classic instrumentation key (sensitive) |
| `connection_string` | SDK connection string (sensitive, recommended) |
| `app_id` | Application ID for API access |

## Quick Example

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationInsights
metadata:
  name: platform-ai
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  name: prod-platform-ai
  application_type: web
  workspace_id:
    valueFrom:
      kind: AzureLogAnalyticsWorkspace
      name: platform-law
      fieldPath: status.outputs.workspace_id
  retention_in_days: 90
  sampling_percentage: 50
```

## Downstream Resources

Resources that reference this Application Insights:

- **AzureFunctionApp** -- `application_insights_connection_string` via StringValueOrRef
- **AzureLinuxWebApp** -- `application_insights_connection_string` via StringValueOrRef
- **AzureContainerApp** -- connection string passed as environment variable
