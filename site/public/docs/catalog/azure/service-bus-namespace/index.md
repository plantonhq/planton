---
title: "Service Bus Namespace"
description: "Service Bus Namespace deployment documentation"
icon: "package"
order: 100
componentName: "azureservicebusnamespace"
---

# Azure Service Bus Namespace

Deploys an Azure Service Bus namespace with optional queues and topics for enterprise message brokering. The component bundles the namespace with its messaging entities because a namespace without at least one queue or topic is incomplete. Supports Basic, Standard, and Premium tiers with duplicate detection, sessions, dead-lettering, message forwarding, and zone redundancy.

## What Gets Created

When you deploy an AzureServiceBusNamespace resource, Planton provisions:

- **Service Bus Namespace** -- a `servicebus.Namespace` resource in the specified region and resource group, configured with the chosen SKU tier, TLS version, and optional Premium capacity settings
- **Queues** -- a `servicebus.Queue` for each entry in `queues`, supporting point-to-point messaging with configurable lock duration, sessions, duplicate detection, dead-lettering, and message forwarding
- **Topics** -- a `servicebus.Topic` for each entry in `topics`, supporting publish-subscribe messaging with configurable partitioning, duplicate detection, and message ordering (Standard and Premium only)
- **Azure Tags** -- resource metadata tags applied to the namespace for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or Planton provider config
- **An Azure Resource Group** where the namespace will be created (can reference an AzureResourceGroup resource)
- **A globally unique namespace name** -- the name becomes the endpoint `{name}.servicebus.windows.net`
- **SKU selection** -- Basic for simple queues only, Standard for queues + topics, Premium for dedicated capacity, VNet integration, and zone redundancy

## Quick Start

Create a file `servicebus.yaml`:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureServiceBusNamespace
metadata:
  name: my-sb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureServiceBusNamespace.my-sb
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-servicebus-ns
  queues:
    - name: orders
```

Deploy:

```shell
planton apply -f servicebus.yaml
```

This creates a Standard-tier Service Bus namespace with a single `orders` queue, TLS 1.2 enforcement, and public network access enabled.

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
| `sku` | `string` | `"Standard"` | SKU tier. Values: `Basic` (queues only), `Standard` (queues + topics, 99.95% SLA), `Premium` (dedicated capacity, VNet, zones). |
| `capacity` | `int` | -- | Messaging units for Premium SKU (1, 2, 4, 8, 16). Each unit provides ~1 MB/s send throughput. Ignored for Basic/Standard. |
| `premiumMessagingPartitions` | `int` | -- | Partitions for Premium SKU (1, 2, 4). **ForceNew**. Ignored for Basic/Standard. |
| `zoneRedundant` | `bool` | `false` | Enable zone redundancy (Premium only). Replicates across availability zones. |
| `minimumTlsVersion` | `string` | `"1.2"` | Minimum TLS version. Values: `1.0`, `1.1`, `1.2`. |
| `publicNetworkAccessEnabled` | `bool` | `true` | Allow public internet access. Set to `false` for private-only access via Private Endpoint. |
| `queues` | `list` | `[]` | Queues for point-to-point messaging. See queue fields below. |
| `topics` | `list` | `[]` | Topics for publish-subscribe messaging (Standard/Premium only). See topic fields below. |

**Queue fields** (each entry in `queues`):

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | -- | Queue name (required, 1-260 characters) |
| `maxSizeInMegabytes` | `int` | _(SKU default)_ | Maximum queue size. Standard: 1024-5120. Premium: up to 81920. |
| `partitioningEnabled` | `bool` | `false` | Enable queue partitioning for higher throughput. **ForceNew**. |
| `defaultMessageTtl` | `string` | _(unbounded)_ | Message time-to-live as ISO 8601 duration (e.g., `P14D`, `PT1H`). |
| `lockDuration` | `string` | `"PT1M"` | Message lock duration. Range: `PT5S` to `PT5M`. |
| `maxDeliveryCount` | `int` | `10` | Max delivery attempts before dead-lettering. Minimum: 1. |
| `requiresDuplicateDetection` | `bool` | `false` | Deduplicate messages by MessageId. **ForceNew**. |
| `requiresSession` | `bool` | `false` | Enable ordered, stateful session processing. **ForceNew**. |
| `deadLetteringOnMessageExpiration` | `bool` | `false` | Move expired messages to dead-letter queue instead of discarding. |
| `forwardTo` | `string` | -- | Auto-forward messages to another queue or topic in the namespace. |
| `forwardDeadLetteredMessagesTo` | `string` | -- | Auto-forward dead-lettered messages to another entity. |

**Topic fields** (each entry in `topics`):

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | -- | Topic name (required, 1-260 characters) |
| `maxSizeInMegabytes` | `int` | _(SKU default)_ | Maximum topic size. |
| `partitioningEnabled` | `bool` | `false` | Enable topic partitioning. **ForceNew**. |
| `defaultMessageTtl` | `string` | _(unbounded)_ | Message time-to-live as ISO 8601 duration. |
| `requiresDuplicateDetection` | `bool` | `false` | Deduplicate messages by MessageId. **ForceNew**. |
| `supportOrdering` | `bool` | `false` | Enable message ordering for session-enabled subscriptions. |

## Examples

### Standard Messaging with Queues and Topics

A Standard-tier namespace with a work queue and an events topic for a typical microservices architecture:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureServiceBusNamespace
metadata:
  name: app-messaging
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureServiceBusNamespace.app-messaging
spec:
  region: eastus
  resourceGroup: prod-rg
  name: app-messaging-ns
  queues:
    - name: order-processing
      maxDeliveryCount: 5
      deadLetteringOnMessageExpiration: true
      defaultMessageTtl: P7D
    - name: notification-delivery
      lockDuration: PT2M
  topics:
    - name: order-events
      defaultMessageTtl: P14D
    - name: inventory-updates
```

### Premium Enterprise Namespace

A Premium-tier namespace with dedicated capacity, zone redundancy, and private-only access:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureServiceBusNamespace
metadata:
  name: enterprise-sb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureServiceBusNamespace.enterprise-sb
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: enterprise-messaging
  sku: Premium
  capacity: 4
  premiumMessagingPartitions: 2
  zoneRedundant: true
  publicNetworkAccessEnabled: false
  queues:
    - name: payment-processing
      requiresSession: true
      requiresDuplicateDetection: true
      maxDeliveryCount: 3
      deadLetteringOnMessageExpiration: true
  topics:
    - name: audit-events
      supportOrdering: true
```

### Event-Driven Microservices with Forwarding

A namespace with message forwarding chains for routing patterns:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureServiceBusNamespace
metadata:
  name: routing-sb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureServiceBusNamespace.routing-sb
spec:
  region: eastus
  resourceGroup: prod-rg
  name: routing-messaging-ns
  queues:
    - name: inbound
      forwardTo: processing
    - name: processing
      maxDeliveryCount: 3
      forwardDeadLetteredMessagesTo: failed-messages
    - name: failed-messages
      defaultMessageTtl: P30D
```

### Using Foreign Key References

Reference an Planton-managed resource group:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureServiceBusNamespace
metadata:
  name: ref-sb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureServiceBusNamespace.ref-sb
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-messaging-ns
  queues:
    - name: tasks
  topics:
    - name: events
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace_id` | `string` | Azure Resource Manager ID of the namespace. Referenced by AzurePrivateEndpoint for private connectivity. |
| `namespace_name` | `string` | Name of the namespace |
| `endpoint` | `string` | Namespace endpoint URL (e.g., `https://{name}.servicebus.windows.net:443/`) |
| `primary_connection_string` | `string` | Connection string from the default RootManageSharedAccessKey (sensitive) |
| `primary_key` | `string` | Primary SAS key for authentication (sensitive) |
| `queue_ids` | `map<string, string>` | Map of queue names to their Azure Resource Manager IDs |
| `topic_ids` | `map<string, string>` | Map of topic names to their Azure Resource Manager IDs |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/resource-group) -- provides the resource group for namespace placement
- [AzurePrivateEndpoint](/docs/catalog/azure/private-endpoint) -- establishes private connectivity to the namespace
- [AzureFunctionApp](/docs/catalog/azure/function-app) -- serverless functions triggered by Service Bus messages
