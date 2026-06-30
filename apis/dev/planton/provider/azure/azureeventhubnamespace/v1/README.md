# Azure Event Hub Namespace

## Overview

The **AzureEventHubNamespace** component provisions an Azure Event Hubs namespace with optional event hubs and consumer groups, providing a fully managed, real-time data ingestion service capable of receiving and processing millions of events per second.

Azure Event Hubs is designed for high-throughput event streaming scenarios. Events are published to event hubs by producers and consumed by consumer groups. Each event hub has a fixed number of partitions that determine the degree of downstream parallelism, and consumer groups provide independent views of the event stream so multiple applications can process the same events at their own pace.

## When to Use This Component

Use AzureEventHubNamespace when you need:

- **Telemetry collection** from distributed applications, microservices, or infrastructure at scale
- **Log aggregation** pipelines that ingest millions of log entries per second for centralized processing
- **IoT data ingestion** where devices publish streams of sensor data for real-time and batch analytics
- **Real-time analytics pipelines** feeding Azure Stream Analytics, Apache Spark, or custom consumers
- **Kafka replacement** leveraging Event Hubs' native Kafka protocol support (Standard/Premium) without managing Kafka clusters
- **Event-driven architectures** where services communicate through ordered, partitioned event streams

## SKU Tiers

| Tier | Consumer Groups | Retention | Kafka Support | Auto-Inflate | Max Message Size | SLA |
|------|-----------------|-----------|---------------|--------------|------------------|-------|
| Basic | 1 ($Default only) | 1 day | No | No | 256 KB | 99.95% |
| Standard | Up to 20 | Up to 7 days | Yes | Yes | 1 MB | 99.95% |
| Premium | Up to 100 | Up to 90 days | Yes | N/A (PUs) | 1 MB | 99.95% |

**Recommendation**: Use **Standard** for most production workloads. Use **Premium** only when you need dedicated processing units, zone redundancy, VNet integration, or extended retention.

## Key Configuration

### Namespace-Level Settings

- **`sku`**: Tier selection (Basic, Standard, Premium). Defaults to Standard.
- **`capacity`**: Throughput units for Standard (1-40 TU) or processing units for Premium (1-16 PU).
- **`autoInflateEnabled`**: Standard-only elastic scaling of throughput units based on traffic demand.
- **`maximumThroughputUnits`**: Upper limit (0-40) when auto-inflate is enabled.
- **`zoneRedundant`**: Availability zone redundancy for higher availability.
- **`minimumTlsVersion`**: Defaults to "1.2" for enterprise security.
- **`publicNetworkAccessEnabled`**: Control public internet accessibility.

### Event Hub Settings

- **`partitionCount`**: Determines downstream parallelism (1-32). Cannot be decreased after creation.
- **`messageRetention`**: Days events are retained (1-7 for Standard, up to 90 for Premium). Default: 1.

### Consumer Group Settings

- **`name`**: Unique identifier for the consumer application's view of the stream.
- **`userMetadata`**: Application-specific metadata (team, purpose, application name).

## Event Hub Capture

Event Hub Capture (archiving events to Azure Blob Storage) is **deliberately omitted** from this component. Capture involves storage account dependencies, encoding configuration, and archive naming conventions that add significant complexity. It can be added in a future v2 expansion.

## Outputs

| Output | Description |
|--------|-------------|
| `namespace_id` | Azure Resource Manager ID |
| `namespace_name` | Namespace name |
| `primary_connection_string` | SAS connection string for SDK connectivity |
| `primary_key` | SAS key for manual token generation |
| `event_hub_ids` | Map of event hub name to resource ID |

## Infra Chart Usage

AzureEventHubNamespace is a **leaf resource** in infra chart DAGs. It is consumed by applications via the `primary_connection_string` injected into app settings or environment variables. No downstream Planton resources reference its outputs.

```yaml
# In a container-apps-environment infra chart:
spec:
  appSettings:
    EVENT_HUB_CONNECTION_STRING:
      valueFrom:
        kind: AzureEventHubNamespace
        name: "{{ values.env }}-events"
        fieldPath: status.outputs.primary_connection_string
```
