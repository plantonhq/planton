---
title: "Application Insights"
description: "Application Insights deployment documentation"
icon: "package"
order: 100
componentName: "azureapplicationinsights"
---

# Azure Application Insights

Deploys an Azure Application Insights resource backed by a Log Analytics Workspace, with configurable application type, data retention, daily ingestion cap, and sampling percentage. The component exports the connection string, instrumentation key, and application ID consumed by downstream compute resources for APM telemetry.

## What Gets Created

When you deploy an AzureApplicationInsights resource, OpenMCF provisions:

- **Application Insights** — an `appinsights.Insights` resource in the specified region and resource group, configured with the chosen application type, retention period, daily data cap, and sampling percentage, linked to a Log Analytics Workspace
- **Azure Tags** — resource metadata tags applied to the Application Insights resource for tracking and governance (resource name, kind, organization, environment)

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the Application Insights resource will be created (can reference an AzureResourceGroup resource)
- **A Log Analytics Workspace** for storing telemetry data (can reference an AzureLogAnalyticsWorkspace resource). Classic (non-workspace) Application Insights is deprecated by Microsoft and is not supported.

## Quick Start

Create a file `app-insights.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationInsights
metadata:
  name: my-app-insights
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureApplicationInsights.my-app-insights
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-app-insights
  workspaceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.OperationalInsights/workspaces/my-workspace
```

Deploy:

```shell
openmcf apply -f app-insights.yaml
```

This creates a workspace-based Application Insights resource with the default application type (`web`), 90-day retention, 100 GB daily data cap, and 100% sampling (full fidelity).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the Application Insights resource (e.g., `eastus`, `westeurope`). Should match the region of the applications being monitored. | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Name of the Application Insights resource. Must be unique within the resource group. | Required, 1-260 characters |
| `workspaceId` | `StringValueOrRef` | Log Analytics Workspace resource ID. Can reference an AzureLogAnalyticsWorkspace resource via `valueFrom`. Workspace-based mode is the only supported mode. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `applicationType` | `string` | `web` | The type of application being monitored. Values: `web` (web applications), `java` (standalone Java), `Node.JS` (Node.js applications), `other` (all other types). This field is ForceNew in Azure -- changing it requires resource recreation. |
| `retentionInDays` | `int32` | `90` | Number of days to retain telemetry data. Allowed values: 30, 60, 90, 120, 180, 270, 365, 550, 730. Free tier includes 90 days; beyond that, retention is billed per GB per month. |
| `dailyDataCapInGb` | `double` | `100` | Daily telemetry ingestion cap in GB. When the cap is reached, ingestion stops until the next UTC day. Useful for controlling costs in development or staging environments. Minimum: 0. |
| `samplingPercentage` | `double` | `100` | Percentage of telemetry data to sample (0-100). Reducing sampling lowers data volume and cost while still providing statistically representative telemetry. Common production values are 25-50%. Set to 100 for full fidelity. |

## Examples

### Basic Web Application Monitoring

A minimal Application Insights resource for monitoring a web application in development:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationInsights
metadata:
  name: dev-web-insights
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureApplicationInsights.dev-web-insights
spec:
  region: eastus
  resourceGroup: dev-rg
  name: dev-web-insights
  workspaceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/dev-rg/providers/Microsoft.OperationalInsights/workspaces/dev-workspace
```

### Cost-Controlled Staging Environment

An Application Insights resource with a low daily data cap and reduced retention for a staging environment:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationInsights
metadata:
  name: staging-insights
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AzureApplicationInsights.staging-insights
spec:
  region: westeurope
  resourceGroup: staging-rg
  name: staging-insights
  workspaceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/staging-rg/providers/Microsoft.OperationalInsights/workspaces/staging-workspace
  applicationType: web
  retentionInDays: 30
  dailyDataCapInGb: 5
  samplingPercentage: 50
```

### Production with Full Retention

A production Application Insights resource with extended retention and full-fidelity telemetry:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationInsights
metadata:
  name: prod-insights
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureApplicationInsights.prod-insights
spec:
  region: eastus
  resourceGroup: prod-rg
  name: prod-insights
  workspaceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.OperationalInsights/workspaces/prod-workspace
  applicationType: web
  retentionInDays: 365
  dailyDataCapInGb: 100
  samplingPercentage: 100
```

### Node.js Application with Sampled Telemetry

An Application Insights resource configured for a Node.js application with 25% sampling to reduce costs at scale:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationInsights
metadata:
  name: nodejs-insights
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureApplicationInsights.nodejs-insights
spec:
  region: southeastasia
  resourceGroup: api-rg
  name: nodejs-insights
  workspaceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/api-rg/providers/Microsoft.OperationalInsights/workspaces/api-workspace
  applicationType: "Node.JS"
  retentionInDays: 90
  dailyDataCapInGb: 25
  samplingPercentage: 25
```

### Using Foreign Key References

Reference OpenMCF-managed resources for the resource group and Log Analytics Workspace instead of hardcoding values:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureApplicationInsights
metadata:
  name: ref-insights
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureApplicationInsights.ref-insights
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-insights
  workspaceId:
    valueFrom:
      kind: AzureLogAnalyticsWorkspace
      name: my-workspace
      field: status.outputs.workspace_id
  retentionInDays: 90
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `appInsightsId` | `string` | Azure Resource Manager ID of the Application Insights resource |
| `instrumentationKey` | `string` | Instrumentation key for classic SDK configuration. Sensitive. Microsoft recommends using `connectionString` for new applications. |
| `connectionString` | `string` | Connection string for SDK configuration. Contains the instrumentation key, ingestion endpoint, and other configuration in a single string. This is the recommended way to configure Application Insights SDKs. Referenced by AzureFunctionApp, AzureLinuxWebApp, and AzureContainerApp. |
| `appId` | `string` | Application ID for programmatic access to Application Insights data via the REST API |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) -- provides the resource group for Application Insights placement
- [AzureLogAnalyticsWorkspace](/docs/catalog/azure/azureloganalyticsworkspace) -- provides the workspace for storing telemetry data
- [AzureFunctionApp](/docs/catalog/azure/azurefunctionapp) -- references the connection string for APM telemetry
- [AzureLinuxWebApp](/docs/catalog/azure/azurelinuxwebapp) -- references the connection string for APM telemetry
- [AzureContainerApp](/docs/catalog/azure/azurecontainerapp) -- references the connection string for APM telemetry
