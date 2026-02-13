# Azure Log Analytics Workspace

Deploys an Azure Log Analytics Workspace with configurable pricing tier, data retention period, and daily ingestion quota. Log Analytics Workspaces are the central data platform for Azure Monitor, collecting and analyzing log and performance data from Azure resources, on-premises servers, and third-party services.

## What Gets Created

When you deploy an AzureLogAnalyticsWorkspace resource, OpenMCF provisions:

- **Log Analytics Workspace** — an `operationalinsights.AnalyticsWorkspace` resource in the specified region and resource group, configured with the chosen SKU pricing tier, retention period, and daily ingestion quota
- **Azure Tags** — resource metadata tags applied to the workspace for tracking and governance, including resource name, kind, organization, and environment

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the workspace will be created (can reference an AzureResourceGroup resource)
- **Workspace naming plan** — names must be 4-63 characters, alphanumeric and hyphens only, starting with a letter, unique within the resource group

## Quick Start

Create a file `log-analytics.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: my-workspace
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureLogAnalyticsWorkspace.my-workspace
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-workspace
```

Deploy:

```shell
openmcf apply -f log-analytics.yaml
```

This creates a Log Analytics Workspace with pay-as-you-go (PerGB2018) pricing, 30-day data retention, and unlimited daily ingestion.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the workspace (e.g., `eastus`, `westeurope`). Choose a region close to the resources that will send logs to minimize egress costs and latency. | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Name of the Log Analytics Workspace. Must be unique within the resource group. Alphanumeric and hyphens only, must start with a letter. | Required, 4-63 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `sku` | `string` | `PerGB2018` | Pricing tier of the workspace. Values: `PerGB2018` (pay-as-you-go, recommended), `CapacityReservation` (commitment tier with discount), `Standalone` (legacy per-node), `PerNode` (legacy OMS per-node). |
| `retentionInDays` | `int32` | `30` | Number of days to retain data. PerGB2018 includes 31 days free; beyond that, retention is billed per GB per month. Range: 30-730. For compliance workloads, 90-365 days is typical. |
| `dailyQuotaGb` | `double` | `-1` | Daily ingestion quota in GB. Set to `-1` for unlimited ingestion. Set to a positive value to cap daily ingestion and prevent cost overruns. When the cap is reached, ingestion stops until the next UTC day. |

## Examples

### Development Workspace with Minimal Retention

A workspace for development with short retention and unlimited ingestion:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: dev-logs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureLogAnalyticsWorkspace.dev-logs
spec:
  region: eastus
  resourceGroup: dev-rg
  name: dev-logs
  sku: PerGB2018
  retentionInDays: 30
```

### Production Workspace with Extended Retention and Ingestion Cap

A production workspace with 180-day retention for audit compliance and a daily ingestion cap to control costs:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: prod-logs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureLogAnalyticsWorkspace.prod-logs
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: prod-logs
  sku: PerGB2018
  retentionInDays: 180
  dailyQuotaGb: 50
```

### Compliance Workspace with Maximum Retention

A workspace for regulatory compliance requiring two-year log retention:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: compliance-logs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureLogAnalyticsWorkspace.compliance-logs
spec:
  region: westeurope
  resourceGroup: compliance-rg
  name: compliance-logs
  sku: PerGB2018
  retentionInDays: 730
  dailyQuotaGb: -1
```

### Using Foreign Key References

Reference an OpenMCF-managed resource group instead of hardcoding the name:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: ref-workspace
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureLogAnalyticsWorkspace.ref-workspace
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-workspace
  retentionInDays: 90
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `workspace_id` | `string` | Azure Resource Manager ID of the Log Analytics Workspace (format: `/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.OperationalInsights/workspaces/{name}`) |
| `workspace_name` | `string` | Name of the Log Analytics Workspace |
| `primary_shared_key` | `string` | Primary shared key for agent authentication. Used by the Log Analytics agent and direct ingestion APIs. Treat as a secret. |
| `secondary_shared_key` | `string` | Secondary shared key for agent authentication. Enables key rotation without downtime. |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) — provides the resource group for workspace placement
- [AzureAksCluster](/docs/catalog/azure/azureakscluster) — AKS clusters send container logs to the workspace via Container Insights
- [AzureContainerAppEnvironment](/docs/catalog/azure/azurecontainerappenvironment) — Container App environments reference the workspace for log collection
