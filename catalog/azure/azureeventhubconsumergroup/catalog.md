# Azure Event Hub Consumer Group

Deploys a consumer group on an Azure Event Hub -- an independent, named view over one stream. Every group keeps its OWN offset per partition, so real-time processing, batch analytics, and archival all read the same events without ever contending: reading never removes anything, and groups are cursors, not queues. The rule of thumb is one group per consuming APPLICATION -- two applications sharing a group steal each other's partitions and corrupt each other's checkpoints. Both identity coordinates -- the hub reference and the group name -- are fixed at creation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Consumer Group** -- on the referenced event hub, with your chosen application-scoped name
- **Ownership metadata** -- when `userMetadata` is set: a free-form note (owner, app, escalation channel) operators see wherever the group is inspected

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureEventHub** the group will read. Reference its `event_hub_id` output via ValueFromRef -- the hub, its groups, and its credentials compose in one manifest set.
- **A group name** unique within the hub -- up to 50 characters, named after the consuming application. `$Default` is the service-created catch-all Azure adds to every hub; it cannot be declared, and production applications should never share it.

## Deploy

### Console

Open the deployment store, find **Azure Event Hub Consumer Group**, and click **Deploy**. The creation wizard walks you through the hub attachment (with the one-group-per-application model taught live) and the ownership note. Start from the **Per-Application Consumer Group** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventHubConsumerGroup
metadata:
  name: analytics-consumer
  org: acme-corp
  env: prod
spec:
  eventHubId:
    valueFrom:
      kind: AzureEventHub
      name: telemetry-stream
      fieldPath: status.outputs.event_hub_id
  consumerGroupName: analytics
  userMetadata: "owner=data-platform; app=stream-analytics"
```

```shell
planton apply -f consumer-group.yaml
```

This creates a consumer group named `analytics` on the `telemetry-stream` hub, carrying an ownership note operators see wherever the group is inspected. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the group to a hub deployed in the same InfraPipeline:

```yaml
spec:
  eventHubId:
    valueFrom:
      kind: AzureEventHub
      name: telemetry-stream
      fieldPath: status.outputs.event_hub_id
```

The InfraPipeline resolves the dependency graph, deploys the namespace and hub first, then provisions the group with the resolved values.

## Key Configuration

These are the most important decisions when configuring a consumer group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One group per application** -- `consumerGroupName` is what consuming SDK clients are constructed with (and the Kafka consumer group id). An app on its own group can be paused, replayed, and monitored independently; the group count a hub supports is a namespace-tier concern (Azure enforces at apply time).

**The ownership note** -- `userMetadata` (max 1024 characters) is how an operator six months from now knows which team owns this cursor and whether it is safe to delete. A key=value convention keeps it parseable by tooling; it edits in place, so keep it current as ownership moves.

**Renaming resets offsets** -- the hub reference and `consumerGroupName` are fixed at creation. Renaming replaces the group and RESETS every stored offset: consumers restart from the start or end of the stream per their configuration, reprocessing or skipping events. Only `userMetadata` edits in place.

**Tier limits** -- the group count a hub supports is a namespace-tier concern, enforced by Azure at apply time: BASIC hubs allow no groups beyond the service-created `$Default`, STANDARD allows 20 per hub, PREMIUM/dedicated more. A group-per-application convention on a busy STANDARD hub can hit the ceiling -- budget groups like any other quota.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureEventHub** | `eventHubId` | `status.outputs.event_hub_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `consumer_group_id` | Azure Resource Manager ID of the group | Governance and automation addressing the group as an ARM resource |
| `consumer_group_name` | The group's name within the hub | What EventHubConsumerClient (or the Kafka consumer group id) is constructed with, alongside the namespace endpoint and hub name |

The group carries no secrets: credentials come separately -- an AzureEventHubAuthorizationRule with listen rights, or a keyless Entra data-plane grant on the hub's `event_hub_id`.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Per-application group** -- the everyday shape: a group named after the consuming application with a traceable ownership note. Start from the **Per-Application Consumer Group** preset.

## Works With

- [**Azure Event Hub**](/infra-catalog/azure-event-hub) -- the stream every group reads
- [**Azure Event Hub Namespace**](/infra-catalog/azure-event-hub-namespace) -- the namespace owning the endpoint the consumer connects to
- [**Azure Event Hub Authorization Rule**](/infra-catalog/azure-event-hub-authorization-rule) -- the listen-rights credential the consuming application holds
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- keyless Azure Event Hubs Data Receiver grants scoped to the hub
