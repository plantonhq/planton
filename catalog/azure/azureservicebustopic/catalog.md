# Azure Service Bus Topic

Deploys a topic inside an Azure Service Bus namespace -- the publish-subscribe primitive. Publishers send to the topic; each AzureServiceBusSubscription under it receives an independent, optionally filtered copy of the stream. Topics carry the publish-side contract only -- consumer semantics (locks, delivery counts, sessions, dead-lettering) live on each subscription, owned by the consuming team.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Service Bus Topic** -- on the referenced namespace, with your chosen size, message-size, partitioning, TTL, ordering, duplicate-detection, express, and batching dials
- **The administrative gate** -- when `status` is set: the topic deploys Active or Disabled (publishes rejected; subscriptions retained)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureServiceBusNamespace** on **STANDARD or PREMIUM** -- BASIC is queue-only, and Azure rejects topics there at apply time. Reference the namespace's `namespace_id` output via ValueFromRef.
- **A topic name** unique within the namespace -- up to 260 characters, starting and ending with a letter or number; forward slashes create a logical hierarchy (`billing/invoices`).

## Deploy

### Console

Open the deployment store, find **Azure Service Bus Topic**, and click **Deploy**. The creation wizard walks you through the namespace attachment, capacity and partitioning (with every tier contract taught live), the publish-side lifecycle and ordering contract, duplicate detection and express, and the publish gate. Start from the **Event Broadcast Topic** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureServiceBusTopic
metadata:
  name: order-events-topic
  org: acme-corp
  env: prod
spec:
  namespaceId:
    valueFrom:
      kind: AzureServiceBusNamespace
      name: order-bus
      fieldPath: status.outputs.namespace_id
  topicName: order-events
  defaultMessageTtl: P14D
```

```shell
planton apply -f topic.yaml
```

This creates the `order-events` topic on the `order-bus` namespace with a 14-day TTL ceiling; unset dials keep Azure's defaults (tier-default size, 256 KB messages, no ordering guarantee, batching on), and subscriptions arrive as their own kind afterward, referencing this topic. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the topic to a namespace deployed in the same InfraPipeline:

```yaml
spec:
  namespaceId:
    valueFrom:
      kind: AzureServiceBusNamespace
      name: order-bus
      fieldPath: status.outputs.namespace_id
```

The InfraPipeline resolves the dependency graph, deploys the namespace first, then provisions the topic with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Service Bus topic. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Two one-way doors** -- `partitioningEnabled` and `requiresDuplicateDetection` are fixed at creation: changing either later replaces the topic AND every subscription under it, so decide them with the publishing and consuming teams up front.

**The shared size budget** -- `maxSizeInMegabytes` (nine fixed sizes; the four large ones are PREMIUM-only) bounds undelivered data across ALL subscriptions together: one stalled consumer can fill the topic for everyone. Alert on per-subscription backlogs, not just the topic total.

**Duplicate detection** -- with `requiresDuplicateDetection: true`, retried publishes carrying the same MessageId inside `duplicateDetectionHistoryTimeWindow` (Azure's default PT10M) are dropped ONCE, before fan-out -- one broker-side dedup replacing an idempotency implementation in every consumer. Fixed at creation, and incompatible with express.

**Ordered pub/sub** -- `supportOrdering: true` preserves publish order on delivery; pair it with SESSION-AWARE subscriptions (publishers stamp SessionId) for strictly-ordered streams. Ordering alone does not serialize a competing consumer fleet.

**Message TTL as a ceiling** -- each subscription may set a SHORTER TTL for its own view of the stream; none may extend past the topic's. Blank means unbounded.

**The publish gate** -- `status: DISABLED` freezes the publish side while every subscription and its backlog is retained. Per-direction gating (drain one consumer, buffer another) lives on each subscription's own gate.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureServiceBusNamespace** | `namespaceId` | `status.outputs.namespace_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `topic_id` | Azure Resource Manager ID of the topic | The parent reference for every AzureServiceBusSubscription, the scope for topic-level data-plane role assignments (Azure Service Bus Data Sender) via AzureRoleAssignment, and the parent for topic-scoped AzureServiceBusAuthorizationRule SAS rules |
| `topic_name` | The topic's name within the namespace | SDK configuration, Functions bindings, and the forward-to target of sibling entities |
| `namespace_name` | The parent namespace's name, parsed from the resolved reference | The namespace/topic pair without a second reference |

There is deliberately no connection-string output: credentials are minted by AzureServiceBusAuthorizationRule (namespace- or topic-scoped) or granted keyless via Entra data-plane roles on `topic_id`. SDKs connect to the namespace endpoint and address the topic by name.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Event broadcast** -- the everyday fan-out stream with a bounded TTL; subscriptions attach independently, each team owning its own filter and consumer semantics. Start from the **Event Broadcast Topic** preset.

**Ordered dedup topic** -- publish order preserved and duplicate publishes dropped before fan-out, for ledger-style streams consumed through session-aware subscriptions. Start from the **Ordered Topic with Duplicate Detection** preset.

## Works With

- [**Azure Service Bus Namespace**](/infra-catalog/azure-service-bus-namespace) -- the parent namespace every topic references
- [**Azure Service Bus Subscription**](/infra-catalog/azure-service-bus-subscription) -- the consumer-side satellite; each one is an independent, filtered view of this topic
- [**Azure Service Bus Queue**](/infra-catalog/azure-service-bus-queue) -- the point-to-point sibling; subscriptions commonly forward matches into one
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- keyless data-plane grants scoped to `topic_id`
