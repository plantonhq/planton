# Azure Event Grid Event Subscription

Deploys an Azure Event Grid event subscription -- the delivery instruction that routes events from a source (custom topic, domain, domain topic, system topic, resource group, or subscription) to exactly one of seven handler types, with subject and payload filtering, retry tuning, and dead-lettering. The addressing choice is the resource: a scope-addressed subscription attaches to any ARM resource that emits Event Grid events, while a system-topic subscription is created as a child of the topic. A subscription is free at rest -- billing is per operation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions exactly one of:

- **Scope-addressed event subscription** -- attaches to any ARM resource that emits Event Grid events: a custom topic, a domain (receiving every domain topic's events), a single domain topic, a resource group, or a subscription (created when `scope` is set)
- **System-topic event subscription** -- a child of the referenced system topic; Azure models these as children rather than scoped attachments, and the engines create the matching resource type (created when `systemTopicId` is set)

Both shapes carry the same configuration surface: destination, filters, retry policy, dead-letter destination, delivery identity, and delivery properties.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Azure Subscription

- **The event source must exist** -- an Azure Event Grid Topic (the default reference for `scope`), an Azure Event Grid System Topic (for `systemTopicId`), or any other source's ARM ID passed as a literal.
- **The destination must exist** -- Azure validates it at create time. Webhook destinations answer a validation handshake at create: the endpoint must be live, or the create fails, so in charts sequence the handler before the subscription.
- **The dead-letter blob container** (only with `deadLetter`) -- the container must already exist; an Azure Storage Container Infra Component manages one.
- **A managed identity with data-plane access** (only for identity-based delivery) -- the identity must exist ON the source topic and hold the destination's data-plane role (Storage Queue Data Message Sender, Azure Service Bus Data Sender) before the subscription names it.

## Deploy

### Console

Open the deployment store, find **Azure Event Grid Event Subscription**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields: the source, the destination arm, filters, and delivery behavior. Start from the **Queue Fan-Out** or **Filtered Webhook** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventgridEventSubscription
metadata:
  name: orders-to-queue
  org: acme-corp
  env: prod
spec:
  scope:
    valueFrom:
      kind: AzureEventgridTopic
      name: orders-topic
      fieldPath: status.outputs.topic_id
  name: orders-to-queue
  eventDeliverySchema: CloudEventSchemaV1_0
  destination:
    storageQueue:
      storageAccountId:
        valueFrom:
          kind: AzureStorageAccount
          name: app-storage
          fieldPath: status.outputs.storage_account_id
      queueName: order-events
  deadLetter:
    storageAccountId:
      valueFrom:
        kind: AzureStorageAccount
        name: app-storage
        fieldPath: status.outputs.storage_account_id
    storageBlobContainerName: eventgrid-dead-letters
```

```shell
planton apply -f event-subscription.yaml
```

This subscribes a storage queue to every event on the topic, delivering in the CloudEvents envelope, with undeliverable events parked in a blob container instead of dropped. An Infra Job tracks the provisioning in real time.

### InfraChart

When the source and destination deploy in the same InfraChart, wire them with ValueFromRef:

```yaml
spec:
  systemTopicId:
    valueFrom:
      kind: AzureEventgridSystemTopic
      name: storage-events
      fieldPath: status.outputs.system_topic_id
  name: blob-created-to-bus
  destination:
    serviceBusQueueId:
      valueFrom:
        kind: AzureServiceBusQueue
        name: ingest-queue
        fieldPath: status.outputs.queue_id
  includedEventTypes:
    - Microsoft.Storage.BlobCreated
```

The InfraPipeline resolves the dependency graph, deploys the system topic and queue first, then provisions the subscription with the resolved ARM IDs.

## Key Configuration

These are the most important decisions when configuring an event subscription. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Addressing: scope vs systemTopicId** -- Set exactly one. `scope` attaches to any ARM resource that emits Event Grid events (custom topics by default, but also domains, domain topics, resource groups, and subscriptions); `systemTopicId` creates the subscription as a child of a system topic. Both are ForceNew: changing the source destroys and recreates the subscription.

**The destination arm, chosen by failure mode** -- A storage queue is the cheapest at-least-once consumer: workers drain at their own pace and poison messages stay visible. Service Bus adds sessions and its own dead-letter semantics. Webhooks couple your delivery success to someone else's uptime -- pair them with retry tuning and dead-lettering, and remember the endpoint must be live before the create. Note Azure ignores `deliveryProperties` on storage-queue destinations: queue messages carry no custom properties.

**eventDeliverySchema must be one the source can map** -- The default is `EventGridSchema`, which suits system topics and platform events. A topic whose input schema is `CloudEventSchemaV1_0` cannot deliver `EventGridSchema`: ARM rejects the create with `InvalidRequest`, and the check lives only server-side -- no plan catches it. When subscribing to a CloudEvents topic, set `CloudEventSchemaV1_0` explicitly. The schema is ForceNew on both sides, so a mismatch is a redeploy, not an edit.

**Dead-letter first, destination second** -- Event Grid does not queue undeliverable events: after the retry policy gives up, an event without a `deadLetter` destination is gone. Create the blob container, wire dead-lettering, then cut traffic over. Treat a subscription without it as fire-and-forget by explicit choice, never by omission.

**Filters share a 25-value budget** -- `advancedFilter` conditions are ANDed; values within one condition are ORed; Azure caps the total values across all conditions at 25 per subscription. The Terraform engine rejects an overflow at plan time on scope-addressed subscriptions; otherwise Azure rejects at deploy. "Type A from region X OR type B from region Y" is two subscriptions, not one -- and the cheaper, clearer shape anyway, since subscriptions cost nothing at rest and each carries its own metrics. `includedEventTypes` and `subjectFilter` handle the common cases before advanced filters are needed.

**retryPolicy tuned to the destination's latency tolerance** -- Left unset, Azure retries 30 times over 1440 minutes. For latency-sensitive webhook integrations, 10 attempts over a few hours beats a full day of retries; exhausted events dead-letter instead of dropping when `deadLetter` is set.

**deliveryIdentity has a dependency order** -- Delivering as a managed identity (required for destinations with shared keys disabled) needs the identity to exist ON the source topic and hold the destination's data-plane role before the subscription names it. Getting the order wrong fails at create -- or worse, at first delivery. `deadLetterIdentity` follows the same contract for the dead-letter write and requires `deadLetter` to be configured.

**expirationTimeUtc kills silently** -- The subscription stops delivering at the deadline with no error. Ideal for temporary integrations (a partner trial, a migration tap); dangerous when set reflexively on production paths.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureEventgridTopic** (scope-addressed) | `scope` | `status.outputs.topic_id` |
| **AzureEventgridSystemTopic** (system-topic-addressed) | `systemTopicId` | `status.outputs.system_topic_id` |
| **AzureEventHub** (Event Hub destination) | `destination.eventhubId` | `status.outputs.event_hub_id` |
| **AzureServiceBusQueue** (queue destination) | `destination.serviceBusQueueId` | `status.outputs.queue_id` |
| **AzureServiceBusTopic** (topic destination) | `destination.serviceBusTopicId` | `status.outputs.topic_id` |
| **AzureStorageAccount** (storage-queue destination) | `destination.storageQueue.storageAccountId` | `status.outputs.storage_account_id` |
| **AzureStorageAccount** (dead-letter) | `deadLetter.storageAccountId` | `status.outputs.storage_account_id` |
| **AzureUserAssignedIdentity** (identity delivery) | `deliveryIdentity.userAssignedIdentity`, `deadLetterIdentity.userAssignedIdentity` | `status.outputs.identity_id` |

The Azure Function destination's `destination.azureFunction.functionId` addresses one function inside an app (`{function_app_id}/functions/{function_name}`), so it has no single-output default -- pass the ID as a literal or compose it in the manifest. `destination.hybridConnectionId` likewise takes a literal ARM ID (Relay is not a catalog kind).

### What This Kind Provides

`status.outputs` records identifiers only: `event_subscription_id` (whose shape follows the addressing choice -- a scoped path extending the source's ID, or a child path under the system topic) and `event_subscription_name`, which echoes the manifest's `name`. A subscription is the end of the routing chain -- no downstream Infra Component consumes it via ValueFromRef.

## Common Patterns

**Queue fan-out** -- deliver a custom topic's events into a storage queue: pull-based, cheap, tolerant of consumer downtime, with dead-lettering configured up front. Start from the **Queue Fan-Out** preset.

**Filtered webhook** -- deliver only a narrow, high-value slice of a busy topic to a partner's HTTPS endpoint, with retry tuned below the defaults and exhausted events dead-lettered. Start from the **Filtered Webhook** preset.

**Many subscriptions over one mega-filter** -- when a filter approaches the 25-value budget or needs OR-across-conditions logic, split consumers into multiple subscriptions rather than compressing conditions. Subscriptions are free at rest and each carries its own delivery metrics.

**Identity-based delivery** -- for destinations with shared-key access disabled, deliver as the source topic's managed identity: assign the data-plane role first, then create the subscription naming the identity via `deliveryIdentity`.

## Works With

- [**Azure Event Grid Topic**](/infra-catalog/azure-eventgrid-topic) -- the default source for scope-addressed subscriptions
- [**Azure Event Grid System Topic**](/infra-catalog/azure-eventgrid-system-topic) -- the source for platform-event subscriptions
- [**Azure Event Grid Domain**](/infra-catalog/azure-eventgrid-domain) / [**Azure Event Grid Domain Topic**](/infra-catalog/azure-eventgrid-domain-topic) -- alternative scope sources for multi-tenant eventing
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- hosts the destination queue and the dead-letter container
- [**Azure Storage Queue**](/infra-catalog/azure-storage-queue) -- the pull-based destination queue
- [**Azure Storage Container**](/infra-catalog/azure-storage-container) -- the dead-letter blob container
- [**Azure Service Bus Queue**](/infra-catalog/azure-service-bus-queue) / [**Azure Service Bus Topic**](/infra-catalog/azure-service-bus-topic) -- destinations with ordering and DLQ semantics of their own
- [**Azure Event Hub**](/infra-catalog/azure-event-hub) -- the high-throughput streaming destination
- [**Azure Function App**](/infra-catalog/azure-function-app) -- the function destination (`{function_app_id}/functions/{name}`)
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the identity deliveries and dead-letter writes run as
