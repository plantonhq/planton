# Azure Event Hub Namespace

Deploys an Azure Event Hubs namespace with optional event hubs and consumer groups for real-time event streaming. The component bundles the namespace with its streaming entities because a namespace without at least one event hub is incomplete. Supports Basic, Standard, and Premium tiers with auto-inflate throughput scaling, zone redundancy, and Kafka protocol compatibility.

## What Gets Created

When you deploy an AzureEventHubNamespace resource, Planton provisions:

- **Event Hubs Namespace** -- an `eventhub.EventHubNamespace` resource in the specified region and resource group, configured with the chosen SKU tier, throughput capacity, TLS version, and optional auto-inflate scaling
- **Event Hubs** -- an `eventhub.EventHub` for each entry in `eventHubs`, each with configurable partition count and message retention
- **Consumer Groups** -- an `eventhub.ConsumerGroup` for each consumer group defined within an event hub, providing independent read positions for separate consumer applications (Azure auto-creates a `$Default` group per event hub)
- **Azure Tags** -- resource metadata tags applied to the namespace for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or Planton provider config
- **An Azure Resource Group** where the namespace will be created (can reference an AzureResourceGroup resource)
- **A globally unique namespace name** -- the name becomes the endpoint `{name}.servicebus.windows.net`
- **Partition planning** -- partition count cannot be decreased after creation; plan for peak throughput upfront

## Quick Start

Create a file `eventhub.yaml`:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureEventHubNamespace
metadata:
  name: my-events
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureEventHubNamespace.my-events
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-events-ns
  eventHubs:
    - name: telemetry
      partitionCount: 4
```

Deploy:

```shell
planton apply -f eventhub.yaml
```

This creates a Standard-tier Event Hubs namespace with a single `telemetry` event hub having 4 partitions, 1-day message retention, and the auto-created `$Default` consumer group.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the namespace (e.g., `eastus`, `westeurope`). | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Globally unique namespace name. Becomes the endpoint `{name}.servicebus.windows.net`. **ForceNew**. | Required, 6-50 characters, pattern `^[a-zA-Z][-a-zA-Z0-9]{4,48}[a-zA-Z0-9]$` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `sku` | `string` | `"Standard"` | SKU tier. Values: `Basic` (1 consumer group, 1-day retention), `Standard` (20 consumer groups, 7-day retention, Kafka support), `Premium` (dedicated compute, zone redundancy). |
| `capacity` | `int` | `1` | Standard: throughput units (1 TU = 1 MB/s ingress, 2 MB/s egress). Premium: processing units (1, 2, 4, 8, 16). |
| `autoInflateEnabled` | `bool` | `false` | Enable automatic throughput unit scaling (Standard only). |
| `maximumThroughputUnits` | `int` | -- | Maximum throughput units when auto-inflate is enabled. Range: 0-40. |
| `zoneRedundant` | `bool` | `false` | Enable zone redundancy (Standard/Premium). Replicates across availability zones. |
| `minimumTlsVersion` | `string` | `"1.2"` | Minimum TLS version. Values: `1.0`, `1.1`, `1.2`. |
| `publicNetworkAccessEnabled` | `bool` | `true` | Allow public internet access. Set to `false` for private-only access via Private Endpoint. |
| `eventHubs` | `list` | `[]` | Event hubs for streaming data. See event hub fields below. |

**Event hub fields** (each entry in `eventHubs`):

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | -- | Event hub name (required, 1-256 characters) |
| `partitionCount` | `int` | -- | Number of partitions (required, 1-32). Cannot be decreased after creation. |
| `messageRetention` | `int` | `1` | Message retention in days (1-7). |
| `consumerGroups` | `list` | `[]` | Additional consumer groups. Each has `name` (required, 1-50 characters) and optional `userMetadata` (up to 1024 characters). |

## Examples

### Standard Streaming Namespace

A Standard-tier namespace with multiple event hubs for a typical data pipeline:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureEventHubNamespace
metadata:
  name: data-pipeline
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureEventHubNamespace.data-pipeline
spec:
  region: eastus
  resourceGroup: prod-rg
  name: data-pipeline-ns
  capacity: 2
  autoInflateEnabled: true
  maximumThroughputUnits: 10
  eventHubs:
    - name: user-events
      partitionCount: 8
      messageRetention: 7
      consumerGroups:
        - name: analytics-processor
          userMetadata: "Real-time analytics pipeline"
        - name: audit-logger
          userMetadata: "Compliance audit trail"
    - name: system-metrics
      partitionCount: 4
      messageRetention: 3
      consumerGroups:
        - name: monitoring-dashboard
```

### Premium Enterprise Namespace

A Premium-tier namespace with dedicated processing units and private-only access:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureEventHubNamespace
metadata:
  name: enterprise-events
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureEventHubNamespace.enterprise-events
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: enterprise-events-ns
  sku: Premium
  capacity: 4
  zoneRedundant: true
  publicNetworkAccessEnabled: false
  eventHubs:
    - name: transactions
      partitionCount: 32
      messageRetention: 7
      consumerGroups:
        - name: fraud-detection
        - name: reporting
        - name: archiver
```

### IoT Event Ingestion

A namespace configured for IoT device telemetry with high partition counts for parallel processing:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureEventHubNamespace
metadata:
  name: iot-ingestion
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureEventHubNamespace.iot-ingestion
spec:
  region: eastus
  resourceGroup: prod-rg
  name: iot-ingestion-ns
  capacity: 4
  autoInflateEnabled: true
  maximumThroughputUnits: 20
  eventHubs:
    - name: device-telemetry
      partitionCount: 32
      messageRetention: 3
      consumerGroups:
        - name: hot-path
          userMetadata: "Real-time alerting"
        - name: cold-path
          userMetadata: "Batch analytics"
    - name: device-lifecycle
      partitionCount: 4
      messageRetention: 7
```

### Using Foreign Key References

Reference an Planton-managed resource group:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureEventHubNamespace
metadata:
  name: ref-events
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureEventHubNamespace.ref-events
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-events-ns
  eventHubs:
    - name: events
      partitionCount: 8
      messageRetention: 7
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace_id` | `string` | Azure Resource Manager ID of the namespace. Referenced by AzurePrivateEndpoint for private connectivity. |
| `namespace_name` | `string` | Name of the namespace |
| `primary_connection_string` | `string` | Connection string from the default RootManageSharedAccessKey (sensitive). Compatible with Event Hubs SDKs, Kafka clients, and Azure Functions triggers. |
| `primary_key` | `string` | Primary SAS key for authentication (sensitive) |
| `event_hub_ids` | `map<string, string>` | Map of event hub names to their Azure Resource Manager IDs |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) -- provides the resource group for namespace placement
- [AzurePrivateEndpoint](/docs/catalog/azure/azureprivateendpoint) -- establishes private connectivity to the namespace
- [AzureFunctionApp](/docs/catalog/azure/azurefunctionapp) -- serverless functions triggered by Event Hub events
