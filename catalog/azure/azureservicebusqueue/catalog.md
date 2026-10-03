# Azure Service Bus Queue

Deploys a queue inside an Azure Service Bus namespace -- reliable point-to-point messaging with FIFO delivery, at-least-once semantics, PeekLock consumption, and a built-in dead-letter sub-queue. Queues are many-per-namespace with independent lifecycles, which is why the queue is a first-class Infra Component referencing the namespace rather than a list folded into it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Service Bus Queue** -- on the referenced namespace, with your chosen size, lock, delivery, TTL, session, duplicate-detection, express, and dead-letter dials
- **Auto-forwarding chains** -- when `forwardTo` or `forwardDeadLetteredMessagesTo` is set: broker-side routing to another queue or topic in the same namespace, by entity name
- **The administrative gate** -- when `status` is set: the queue deploys Active, Disabled, Send-Disabled (drain mode), or Receive-Disabled (buffer mode)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureServiceBusNamespace** the queue will live in. Reference its `namespace_id` output via ValueFromRef -- the namespace's tier decides what the queue may use (large sizes, custom message sizes, and namespace-dictated partitioning are PREMIUM concerns; express is multi-tenant-only).
- **A queue name** unique within the namespace -- up to 260 characters, starting and ending with a letter or number; forward slashes create a logical hierarchy (`orders/priority`).

## Deploy

### Console

Open the deployment store, find **Azure Service Bus Queue**, and click **Deploy**. The creation wizard walks you through the namespace attachment, capacity and partitioning (with every tier contract taught live), the PeekLock delivery contract, duplicate detection and express, sessions and idle lifecycle, auto-forwarding, and the administrative gate. Start from the **Work Queue** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureServiceBusQueue
metadata:
  name: orders-queue
  org: acme-corp
  env: prod
spec:
  namespaceId:
    valueFrom:
      kind: AzureServiceBusNamespace
      name: order-bus
      fieldPath: status.outputs.namespace_id
  queueName: orders
  maxDeliveryCount: 5
  deadLetteringOnMessageExpiration: true
  defaultMessageTtl: P14D
```

```shell
planton apply -f queue.yaml
```

This creates the `orders` queue on the `order-bus` namespace with a tightened 5-attempt poison threshold, expired messages preserved in the dead-letter sub-queue, and a 14-day TTL; unset dials keep Azure's defaults (1 GB size, 1-minute lock, batching on). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the queue to a namespace deployed in the same InfraPipeline:

```yaml
spec:
  namespaceId:
    valueFrom:
      kind: AzureServiceBusNamespace
      name: order-bus
      fieldPath: status.outputs.namespace_id
```

The InfraPipeline resolves the dependency graph, deploys the namespace first, then provisions the queue with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Service Bus queue. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Three one-way doors** -- `partitioningEnabled`, `requiresDuplicateDetection`, and `requiresSession` are fixed at creation: changing any of them later replaces the queue and drops its messages, so decide them with the producer and consumer teams up front.

**The PeekLock contract** -- `lockDuration` (PT5S to PT5M; Azure's default PT1M) hides a received message while a consumer works on it, and `maxDeliveryCount` (Azure's default 10) is the poison-message circuit breaker: after that many failed deliveries the message moves to the dead-letter sub-queue at `{queue}/$deadletterqueue` instead of redelivering forever.

**Duplicate detection** -- with `requiresDuplicateDetection: true`, retried sends carrying the same MessageId inside `duplicateDetectionHistoryTimeWindow` (Azure's default PT10M) are silently dropped -- idempotent producers with zero application bookkeeping. Fixed at creation, and incompatible with express.

**Sessions** -- `requiresSession: true` turns the queue into strictly ordered per-SessionId sub-streams with exclusive, session-aware consumers. Fixed at creation; plain receivers cannot read a session queue, and auto-forwarding TO a session-aware queue is rejected.

**Auto-forwarding** -- `forwardTo` and `forwardDeadLetteredMessagesTo` chain entities broker-side by entity NAME (never ARM ID) within the same namespace: per-partner intake queues funneling into one work queue, and dead letters centralizing into one poison queue. The target must exist before the queue deploys.

**The administrative gate** -- `status` is the designed day-two edit: `SEND_DISABLED` drains a queue before decommissioning, `RECEIVE_DISABLED` buffers while a consumer fleet redeploys. Unspecified deploys ACTIVE.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureServiceBusNamespace** | `namespaceId` | `status.outputs.namespace_id` |
| **AzureServiceBusQueue** | `forwardTo` / `forwardDeadLetteredMessagesTo` | `status.outputs.queue_name` |
| **AzureServiceBusTopic** | `forwardTo` / `forwardDeadLetteredMessagesTo` | `status.outputs.topic_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `queue_id` | Azure Resource Manager ID of the queue | The scope for queue-level data-plane role assignments (Azure Service Bus Data Receiver/Sender) via AzureRoleAssignment, and the parent for queue-scoped AzureServiceBusAuthorizationRule SAS rules |
| `queue_name` | The queue's name within the namespace | SDK configuration, Functions bindings, and the forward-to target of sibling entities |
| `namespace_name` | The parent namespace's name, parsed from the resolved reference | The namespace/queue pair without a second reference |

There is deliberately no connection-string output: credentials are minted by AzureServiceBusAuthorizationRule (namespace- or queue-scoped) or granted keyless via Entra data-plane roles on `queue_id`. SDKs connect to the namespace endpoint and address the queue by name.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Work queue** -- the everyday competing-consumer shape: a tightened 5-attempt poison threshold and expired messages kept inspectable in the dead-letter sub-queue. Start from the **Work Queue** preset.

**Session FIFO queue** -- strict per-SessionId ordering with duplicate detection for exactly-once-shaped pipelines (account transactions, per-entity event streams). Start from the **Session FIFO Queue** preset.

## Works With

- [**Azure Service Bus Namespace**](/infra-catalog/azure-service-bus-namespace) -- the parent namespace every queue references
- [**Azure Service Bus Topic**](/infra-catalog/azure-service-bus-topic) -- the pub/sub sibling; a legal auto-forward target
- [**Azure Service Bus Subscription**](/infra-catalog/azure-service-bus-subscription) -- fans a topic out; commonly forwards matches into a queue like this one
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- keyless data-plane grants scoped to `queue_id`
