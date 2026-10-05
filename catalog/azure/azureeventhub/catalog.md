# Azure Event Hub

Deploys an event hub inside an Azure Event Hubs namespace -- one partitioned, replayable event stream. Producers append events to partitions; consumers read them through consumer groups, each keeping its own offset, so the same stream feeds real-time processing, batch analytics, and archival independently. Hubs are many-per-namespace with independent lifecycles, which is why the hub is a first-class Infra Component referencing the namespace rather than a list folded into it. Kafka clients see the hub as a topic, unchanged. Capture — continuous Avro archival to Blob Storage — folds into the same spec, because Azure models it as a property of the hub.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Event Hub** -- on the referenced namespace, with your chosen partition count and exactly one retention model: a simple day count, or the hour-granular block with Kafka-style log compaction
- **Capture** -- when `captureDescription` is set: continuous archival of every event to Azure Blob Storage in Avro format on a size-or-interval cadence, with SAS or managed-identity authentication
- **The administrative gate** -- when `status` is set: the hub deploys Active, Disabled, or Send-Disabled (drain mode)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureEventHubNamespace** the hub will live in. Reference its `namespace_id` output via ValueFromRef -- the namespace's tier decides what the hub may use (partition counts above 32, retention beyond 7 days, and log compaction are PREMIUM/dedicated or STANDARD-plus concerns).
- **For capture**: an AzureStorageAccount and an AzureStorageContainer for the archives -- referenced by `storage_account_id` and `container_name` outputs.

## Deploy

### Console

Open the deployment store, find **Azure Event Hub**, and click **Deploy**. The creation wizard walks you through the namespace attachment, partition sizing (with every tier contract taught live), the exactly-one retention model, capture to Blob Storage, and the administrative gate. Start from the **Telemetry Stream** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventHub
metadata:
  name: telemetry-stream
  org: acme-corp
  env: prod
spec:
  namespaceId:
    valueFrom:
      kind: AzureEventHubNamespace
      name: telemetry-hubs
      fieldPath: status.outputs.namespace_id
  eventHubName: telemetry
  partitionCount: 8
  messageRetention: 3
```

```shell
planton apply -f event-hub.yaml
```

This creates an eight-partition hub named `telemetry` on the `telemetry-hubs` namespace with a three-day replay window. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the hub to a namespace deployed in the same InfraPipeline:

```yaml
spec:
  namespaceId:
    valueFrom:
      kind: AzureEventHubNamespace
      name: telemetry-hubs
      fieldPath: status.outputs.namespace_id
```

The InfraPipeline resolves the dependency graph, deploys the namespace first, then provisions the hub with the resolved values.

## Key Configuration

These are the most important decisions when configuring an event hub. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Partitions** -- `partitionCount` (1-1024) is the unit of parallelism AND ordering: each partition is an independently consumed, ordered sequence, and downstream parallelism cannot exceed the count. 2-4 suits low throughput, 8-16 typical production, 32+ high-throughput ingestion. Azure caps the count at 32 on shared BASIC/STANDARD namespaces (1024 on PREMIUM/dedicated), never accepts a decrease, and only PREMIUM/dedicated namespaces can increase it later.

**The retention model** -- `messageRetention` is the everyday shape: N days of replayable history (tier ceilings: 1 BASIC, 7 STANDARD, 90 PREMIUM/dedicated; unset lets Azure default to 1 day). `retentionDescription` unlocks hour granularity and Kafka-style log COMPACTION: the latest event per partition key is kept forever -- changelog/table semantics for materialized views and cache warming. The cleanup policy is fixed at creation; tombstones (null-value events marking deletions) stay readable for `tombstoneRetentionTimeInHours`.

**Capture** -- `captureDescription` archives every event to Blob Storage in Avro on a size-or-interval cadence (defaults: 300 seconds / 300 MB): the built-in streaming-to-batch bridge with no consumer application to run. The `archiveNameFormat` must carry all nine placeholders ({Namespace}, {EventHub}, {PartitionId}, {Year}, {Month}, {Day}, {Hour}, {Minute}, {Second}) -- the placeholder path IS your batch layout. Authentication is service-managed SAS by default, or keyless via the namespace's system-assigned identity or a user-assigned identity (grant it Storage Blob Data Contributor on the account and attach it via the namespace's identity block).

**The administrative gate** -- `status` is the designed day-two edit: `SEND_DISABLED` drains the stream before a decommission or migration (producers rejected, consumers keep reading); `DISABLED` freezes both directions while retained events stay stored. Unspecified deploys ACTIVE.

**One-way doors** -- exactly one retention model must be set: `messageRetention` (a day count) or `retentionDescription` (hours / compaction). `eventHubName` and `retentionDescription.cleanupPolicy` are fixed at creation -- changing either replaces the hub and destroys its retained events.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureEventHubNamespace** | `namespaceId` | `status.outputs.namespace_id` |
| **AzureStorageContainer** | `captureDescription.destination.blobContainerName` | `status.outputs.container_name` |
| **AzureStorageAccount** | `captureDescription.destination.storageAccountId` | `status.outputs.storage_account_id` |
| **AzureUserAssignedIdentity** | `captureDescription.destination.storageAuthenticationId` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `event_hub_id` | Azure Resource Manager ID of the hub | The parent for AzureEventHubConsumerGroup and hub-scoped AzureEventHubAuthorizationRule SAS rules, and the scope for hub-level data-plane role assignments (Azure Event Hubs Data Receiver/Sender) via AzureRoleAssignment |
| `event_hub_name` | The hub's name within the namespace | SDK configuration, Functions bindings, and the Kafka topic name |
| `partition_ids` | The hub's partition identifiers | Checkpointing consumers enumerate these -- one active reader per partition per consumer group is the scaling model |

There is deliberately no connection-string output: credentials are minted by AzureEventHubAuthorizationRule (namespace- or hub-scoped) or granted keyless via Entra data-plane roles on `event_hub_id`. SDKs connect to the namespace endpoint and address the hub by name.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Telemetry stream** -- the everyday shape: sized partitions and a multi-day replay window for real-time processing with reprocessing headroom. Start from the **Telemetry Stream** preset.

**Captured archive stream** -- capture enabled so every event lands in Blob Storage as Avro: cold storage and audit trails that outlive the retention window. Start from the **Captured Archive Stream** preset.

**Compacted changelog** -- log compaction keeps the latest event per key forever: materialized views and cache warming read the whole current state by replaying from the start. Start from the **Compacted Changelog Stream** preset.

## Works With

- [**Azure Event Hub Namespace**](/infra-catalog/azure-event-hub-namespace) -- the parent namespace every hub references
- [**Azure Event Hub Consumer Group**](/infra-catalog/azure-event-hub-consumer-group) -- one per consuming application; offsets never collide
- [**Azure Event Hub Authorization Rule**](/infra-catalog/azure-event-hub-authorization-rule) -- hub-scoped SAS credentials referencing `event_hub_id`
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the capture destination's account
- [**Azure Storage Container**](/infra-catalog/azure-storage-container) -- where captured archives land
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- keyless data-plane grants scoped to `event_hub_id`
