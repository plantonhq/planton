# AzureEventgridEventSubscription

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureEventgridEventSubscriptionSpec** defines an Azure Event Grid
event subscription -- the delivery instruction that routes events
from a source to a handler: "when events arrive HERE, filtered like
THIS, deliver them THERE". The source is named by the addressing
choice (exactly one):
  - `scope`: any Azure resource that emits Event Grid events, by ARM
    ID -- a custom topic (AzureEventgridTopic), a domain
    (AzureEventgridDomain, receiving every domain topic's events), a
    single domain topic (AzureEventgridDomainTopic, the per-tenant
    stream), a resource group, or a subscription.
  - `system_topic_id`: an Event Grid SYSTEM topic
    (AzureEventgridSystemTopic) -- Azure models these subscriptions
    as children of the system topic rather than scoped attachments,
    and the engines create the matching resource type.
The destination is exactly one of seven handler arms; filters narrow
which events are delivered. A subscription is free at rest; billing
is per operation.

## Example

```yaml
# Deep-shape example for docs and offline validation: a filtered
# webhook subscription on a custom topic, with delivery properties,
# tuned retry, and dead-lettering. References are literal values so
# the manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventgridEventSubscription
metadata:
  name: test-eventgrid-event-subscription
  id: test-eventgrid-event-subscription
  org: test-org
  env: test
spec:
  scope:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.EventGrid/topics/orders-events
  name: big-orders-webhook
  destination:
    webhook:
      url: https://handler.example.com/events
      maxEventsPerBatch: 10
      preferredBatchSizeInKilobytes: 64
  deliveryProperties:
    - headerName: x-api-key
      type: Static
      value:
        value: test-api-key
      secret: true
    - headerName: x-event-subject
      type: Dynamic
      sourceField: subject
  deadLetter:
    storageAccountId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/deadletters
    storageBlobContainerName: eventgrid-dead-letters
  eventDeliverySchema: CloudEventSchemaV1_0
  includedEventTypes:
    - order.placed
    - order.cancelled
  subjectFilter:
    subjectBeginsWith: /orders/
    caseSensitive: false
  advancedFilter:
    numberGreaterThan:
      - key: data.amount
        value: 1000
    stringIn:
      - key: data.currency
        values:
          - USD
          - EUR
    numberInRange:
      - key: data.riskScore
        ranges:
          - from: 0
            to: 70
  advancedFilteringOnArraysEnabled: true
  labels:
    - orders
    - critical
  expirationTimeUtc: "2027-01-01T00:00:00Z"
  retryPolicy:
    maxDeliveryAttempts: 10
    eventTimeToLive: 240
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.scope` | `string \| valueFrom` |  |  | AzureEventgridTopic (`status.outputs.topic_id`) |
| `spec.systemTopicId` | `string \| valueFrom` |  |  | AzureEventgridSystemTopic (`status.outputs.system_topic_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.destination` | `AzureEventgridEventSubscriptionDestination` | yes |  |  |
| `spec.destination.azureFunction` | `AzureEventgridEventSubscriptionAzureFunctionDestination` |  |  |  |
| `spec.destination.azureFunction.functionId` | `string \| valueFrom` | yes |  |  |
| `spec.destination.azureFunction.maxEventsPerBatch` | `int32` |  |  |  |
| `spec.destination.azureFunction.preferredBatchSizeInKilobytes` | `int32` |  |  |  |
| `spec.destination.eventhubId` | `string \| valueFrom` |  |  | AzureEventHub (`status.outputs.event_hub_id`) |
| `spec.destination.hybridConnectionId` | `string \| valueFrom` |  |  |  |
| `spec.destination.serviceBusQueueId` | `string \| valueFrom` |  |  | AzureServiceBusQueue (`status.outputs.queue_id`) |
| `spec.destination.serviceBusTopicId` | `string \| valueFrom` |  |  | AzureServiceBusTopic (`status.outputs.topic_id`) |
| `spec.destination.storageQueue` | `AzureEventgridEventSubscriptionStorageQueueDestination` |  |  |  |
| `spec.destination.storageQueue.storageAccountId` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.destination.storageQueue.queueName` | `string` | yes |  |  |
| `spec.destination.storageQueue.queueMessageTimeToLiveInSeconds` | `int32` |  |  |  |
| `spec.destination.webhook` | `AzureEventgridEventSubscriptionWebhookDestination` |  |  |  |
| `spec.destination.webhook.url` | `string` | yes |  |  |
| `spec.destination.webhook.maxEventsPerBatch` | `int32` |  |  |  |
| `spec.destination.webhook.preferredBatchSizeInKilobytes` | `int32` |  |  |  |
| `spec.destination.webhook.activeDirectoryTenantId` | `string` |  |  |  |
| `spec.destination.webhook.activeDirectoryAppIdOrUri` | `string` |  |  |  |
| `spec.deliveryIdentity` | `AzureEventgridEventSubscriptionIdentity` |  |  |  |
| `spec.deliveryIdentity.type` | `enum` | yes |  |  |
| `spec.deliveryIdentity.userAssignedIdentity` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.deliveryProperties` | `[]AzureEventgridEventSubscriptionDeliveryProperty` |  |  |  |
| `spec.deliveryProperties[].headerName` | `string` | yes |  |  |
| `spec.deliveryProperties[].type` | `string` | yes |  |  |
| `spec.deliveryProperties[].value` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.deliveryProperties[].sourceField` | `string` |  |  |  |
| `spec.deliveryProperties[].secret` | `bool` |  |  |  |
| `spec.deadLetter` | `AzureEventgridEventSubscriptionDeadLetter` |  |  |  |
| `spec.deadLetter.storageAccountId` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.deadLetter.storageBlobContainerName` | `string` | yes |  |  |
| `spec.deadLetterIdentity` | `AzureEventgridEventSubscriptionIdentity` |  |  |  |
| `spec.deadLetterIdentity.type` | `enum` | yes |  |  |
| `spec.deadLetterIdentity.userAssignedIdentity` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.eventDeliverySchema` | `string` |  | `EventGridSchema` |  |
| `spec.includedEventTypes` | `[]string` |  |  |  |
| `spec.subjectFilter` | `AzureEventgridEventSubscriptionSubjectFilter` |  |  |  |
| `spec.subjectFilter.subjectBeginsWith` | `string` |  |  |  |
| `spec.subjectFilter.subjectEndsWith` | `string` |  |  |  |
| `spec.subjectFilter.caseSensitive` | `bool` |  |  |  |
| `spec.advancedFilter` | `AzureEventgridEventSubscriptionAdvancedFilter` |  |  |  |
| `spec.advancedFilter.boolEquals` | `[]AzureEventgridEventSubscriptionBoolEqualsFilter` |  |  |  |
| `spec.advancedFilter.boolEquals[].key` | `string` | yes |  |  |
| `spec.advancedFilter.boolEquals[].value` | `bool` |  |  |  |
| `spec.advancedFilter.numberGreaterThan` | `[]AzureEventgridEventSubscriptionNumberFilter` |  |  |  |
| `spec.advancedFilter.numberGreaterThan[].key` | `string` | yes |  |  |
| `spec.advancedFilter.numberGreaterThan[].value` | `double` |  |  |  |
| `spec.advancedFilter.numberGreaterThanOrEquals` | `[]AzureEventgridEventSubscriptionNumberFilter` |  |  |  |
| `spec.advancedFilter.numberGreaterThanOrEquals[].key` | `string` | yes |  |  |
| `spec.advancedFilter.numberGreaterThanOrEquals[].value` | `double` |  |  |  |
| `spec.advancedFilter.numberLessThan` | `[]AzureEventgridEventSubscriptionNumberFilter` |  |  |  |
| `spec.advancedFilter.numberLessThan[].key` | `string` | yes |  |  |
| `spec.advancedFilter.numberLessThan[].value` | `double` |  |  |  |
| `spec.advancedFilter.numberLessThanOrEquals` | `[]AzureEventgridEventSubscriptionNumberFilter` |  |  |  |
| `spec.advancedFilter.numberLessThanOrEquals[].key` | `string` | yes |  |  |
| `spec.advancedFilter.numberLessThanOrEquals[].value` | `double` |  |  |  |
| `spec.advancedFilter.numberIn` | `[]AzureEventgridEventSubscriptionNumberListFilter` |  |  |  |
| `spec.advancedFilter.numberIn[].key` | `string` | yes |  |  |
| `spec.advancedFilter.numberIn[].values` | `[]double` | yes |  |  |
| `spec.advancedFilter.numberNotIn` | `[]AzureEventgridEventSubscriptionNumberListFilter` |  |  |  |
| `spec.advancedFilter.numberNotIn[].key` | `string` | yes |  |  |
| `spec.advancedFilter.numberNotIn[].values` | `[]double` | yes |  |  |
| `spec.advancedFilter.numberInRange` | `[]AzureEventgridEventSubscriptionNumberRangeFilter` |  |  |  |
| `spec.advancedFilter.numberInRange[].key` | `string` | yes |  |  |
| `spec.advancedFilter.numberInRange[].ranges` | `[]AzureEventgridEventSubscriptionNumberRange` | yes |  |  |
| `spec.advancedFilter.numberInRange[].ranges[].from` | `double` |  |  |  |
| `spec.advancedFilter.numberInRange[].ranges[].to` | `double` |  |  |  |
| `spec.advancedFilter.numberNotInRange` | `[]AzureEventgridEventSubscriptionNumberRangeFilter` |  |  |  |
| `spec.advancedFilter.numberNotInRange[].key` | `string` | yes |  |  |
| `spec.advancedFilter.numberNotInRange[].ranges` | `[]AzureEventgridEventSubscriptionNumberRange` | yes |  |  |
| `spec.advancedFilter.numberNotInRange[].ranges[].from` | `double` |  |  |  |
| `spec.advancedFilter.numberNotInRange[].ranges[].to` | `double` |  |  |  |
| `spec.advancedFilter.stringBeginsWith` | `[]AzureEventgridEventSubscriptionStringListFilter` |  |  |  |
| `spec.advancedFilter.stringBeginsWith[].key` | `string` | yes |  |  |
| `spec.advancedFilter.stringBeginsWith[].values` | `[]string` | yes |  |  |
| `spec.advancedFilter.stringNotBeginsWith` | `[]AzureEventgridEventSubscriptionStringListFilter` |  |  |  |
| `spec.advancedFilter.stringNotBeginsWith[].key` | `string` | yes |  |  |
| `spec.advancedFilter.stringNotBeginsWith[].values` | `[]string` | yes |  |  |
| `spec.advancedFilter.stringEndsWith` | `[]AzureEventgridEventSubscriptionStringListFilter` |  |  |  |
| `spec.advancedFilter.stringEndsWith[].key` | `string` | yes |  |  |
| `spec.advancedFilter.stringEndsWith[].values` | `[]string` | yes |  |  |
| `spec.advancedFilter.stringNotEndsWith` | `[]AzureEventgridEventSubscriptionStringListFilter` |  |  |  |
| `spec.advancedFilter.stringNotEndsWith[].key` | `string` | yes |  |  |
| `spec.advancedFilter.stringNotEndsWith[].values` | `[]string` | yes |  |  |
| `spec.advancedFilter.stringContains` | `[]AzureEventgridEventSubscriptionStringListFilter` |  |  |  |
| `spec.advancedFilter.stringContains[].key` | `string` | yes |  |  |
| `spec.advancedFilter.stringContains[].values` | `[]string` | yes |  |  |
| `spec.advancedFilter.stringNotContains` | `[]AzureEventgridEventSubscriptionStringListFilter` |  |  |  |
| `spec.advancedFilter.stringNotContains[].key` | `string` | yes |  |  |
| `spec.advancedFilter.stringNotContains[].values` | `[]string` | yes |  |  |
| `spec.advancedFilter.stringIn` | `[]AzureEventgridEventSubscriptionStringListFilter` |  |  |  |
| `spec.advancedFilter.stringIn[].key` | `string` | yes |  |  |
| `spec.advancedFilter.stringIn[].values` | `[]string` | yes |  |  |
| `spec.advancedFilter.stringNotIn` | `[]AzureEventgridEventSubscriptionStringListFilter` |  |  |  |
| `spec.advancedFilter.stringNotIn[].key` | `string` | yes |  |  |
| `spec.advancedFilter.stringNotIn[].values` | `[]string` | yes |  |  |
| `spec.advancedFilter.isNotNull` | `[]AzureEventgridEventSubscriptionKeyFilter` |  |  |  |
| `spec.advancedFilter.isNotNull[].key` | `string` | yes |  |  |
| `spec.advancedFilter.isNullOrUndefined` | `[]AzureEventgridEventSubscriptionKeyFilter` |  |  |  |
| `spec.advancedFilter.isNullOrUndefined[].key` | `string` | yes |  |  |
| `spec.advancedFilteringOnArraysEnabled` | `bool` |  | `false` |  |
| `spec.labels` | `[]string` |  |  |  |
| `spec.expirationTimeUtc` | `string` |  |  |  |
| `spec.retryPolicy` | `AzureEventgridEventSubscriptionRetryPolicy` |  |  |  |
| `spec.retryPolicy.maxDeliveryAttempts` | `int32` |  |  |  |
| `spec.retryPolicy.eventTimeToLive` | `int32` |  |  |  |

## Field Details

### spec.scope

`string | valueFrom`

The ARM ID of the event source, for every source EXCEPT system
topics -- defaults to referencing an AzureEventgridTopic's
topic_id output; pass a domain's, domain topic's, resource
group's, or subscription's ID (literal or explicit valueFrom) for
the other sources. Set exactly one of scope / system_topic_id.

**ForceNew**: changing this destroys and recreates the
subscription.

- references: AzureEventgridTopic (`status.outputs.topic_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureEventgridTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.systemTopicId

`string | valueFrom`

The ARM ID of the Event Grid system topic to subscribe to --
defaults to referencing an AzureEventgridSystemTopic's
system_topic_id output. Set exactly one of scope /
system_topic_id.

**ForceNew**: changing this destroys and recreates the
subscription.

- references: AzureEventgridSystemTopic (`status.outputs.system_topic_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureEventgridSystemTopic, name: <that resource's name>, fieldPath: status.outputs.system_topic_id}} -- a bare string does not parse

### spec.name

`string` · required

The subscription's name -- 3-64 characters; letters, numbers, and
hyphens; unique within its source.

**ForceNew**: changing this destroys and recreates the
subscription.

- rule: Event subscription names must be 3-64 characters of letters, numbers, and hyphens
- rule: {"required":true}

### spec.destination

`AzureEventgridEventSubscriptionDestination` · required

Where matched events are delivered -- exactly one handler arm.

- rule: {"required":true}
- rule: Set exactly one destination arm -- azure_function, eventhub_id, hybrid_connection_id, service_bus_queue_id, service_bus_topic_id, storage_queue, or webhook

### spec.destination.azureFunction

`AzureEventgridEventSubscriptionAzureFunctionDestination`

Deliver to an Azure Function. Set exactly one arm on this block.

### spec.destination.azureFunction.functionId

`string | valueFrom` · required

The FUNCTION's ARM ID -- the function app's ID plus the function
segment:
{function_app_id}/functions/{function_name}. This addresses one
function INSIDE an app (an AzureFunctionApp's function_app_id
output is the prefix), so there is no single-output default --
pass the ID as a literal or compose it in the manifest.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.destination.azureFunction.maxEventsPerBatch

`int32` · optional (explicit presence)

How many events one invocation may receive at most. Leave unset
for Azure's default -- the modules send the value only when set.

- rule: {"int32":{"gte":1}}

### spec.destination.azureFunction.preferredBatchSizeInKilobytes

`int32` · optional (explicit presence)

The preferred batch payload size in kilobytes. Leave unset for
Azure's default -- the modules send the value only when set.

- rule: {"int32":{"gte":1}}

### spec.destination.eventhubId

`string | valueFrom`

Deliver to an Event Hub, by ARM ID -- defaults to referencing an
AzureEventHub's event_hub_id output. Set exactly one arm on this
block.

Events are DELIVERED into the hub; the subscription belongs to the
topic it subscribes to, so on a diagram the reference is access, not
placement.

- references: AzureEventHub (`status.outputs.event_hub_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureEventHub, name: <that resource's name>, fieldPath: status.outputs.event_hub_id}} -- a bare string does not parse

### spec.destination.hybridConnectionId

`string | valueFrom`

Deliver to an Azure Relay hybrid connection, by ARM ID (Relay is
not yet a catalog kind -- pass the ID as a literal or an explicit
valueFrom). Set exactly one arm on this block.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.destination.serviceBusQueueId

`string | valueFrom`

Deliver to a Service Bus queue, by ARM ID -- defaults to
referencing an AzureServiceBusQueue's queue_id output. Set
exactly one arm on this block.

- references: AzureServiceBusQueue (`status.outputs.queue_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureServiceBusQueue, name: <that resource's name>, fieldPath: status.outputs.queue_id}} -- a bare string does not parse

### spec.destination.serviceBusTopicId

`string | valueFrom`

Deliver to a Service Bus topic, by ARM ID -- defaults to
referencing an AzureServiceBusTopic's topic_id output. Set
exactly one arm on this block.

Events are DELIVERED into the topic; the subscription belongs to the
Event Grid topic it subscribes to, so on a diagram the reference is
access, not placement.

- references: AzureServiceBusTopic (`status.outputs.topic_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureServiceBusTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.destination.storageQueue

`AzureEventgridEventSubscriptionStorageQueueDestination`

Deliver to an Azure Storage queue. Set exactly one arm on this
block.

### spec.destination.storageQueue.storageAccountId

`string | valueFrom` · required

The storage account holding the queue, by ARM ID -- defaults to
referencing an AzureStorageAccount's storage_account_id output.

Events are DELIVERED into a queue in this account; the subscription
belongs to the topic it subscribes to, so on a diagram the reference is
access, not placement.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.destination.storageQueue.queueName

`string` · required

The queue's name within that account (the queue must exist --
AzureStorageQueue manages one).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.destination.storageQueue.queueMessageTimeToLiveInSeconds

`int32` · optional (explicit presence)

How long a delivered queue message lives before the queue expires
it, in seconds (-1 = never expires). Leave unset for the
service's default (7 days) -- the modules send the value only
when set.

### spec.destination.webhook

`AzureEventgridEventSubscriptionWebhookDestination`

Deliver to an HTTPS webhook. Set exactly one arm on this block.

### spec.destination.webhook.url

`string` · required

The HTTPS endpoint events are POSTed to. HTTP is rejected --
Azure only delivers over TLS.

- rule: The webhook url must be an https:// URL
- rule: {"required":true}

### spec.destination.webhook.maxEventsPerBatch

`int32` · optional (explicit presence)

How many events one POST may carry at most (1-5000). Leave unset
for Azure's default -- the modules send the value only when set.

- rule: {"int32":{"lte":5000,"gte":1}}

### spec.destination.webhook.preferredBatchSizeInKilobytes

`int32` · optional (explicit presence)

The preferred batch payload size in kilobytes (1-1024). Leave
unset for Azure's default -- the modules send the value only when
set.

- rule: {"int32":{"lte":1024,"gte":1}}

### spec.destination.webhook.activeDirectoryTenantId

`string`

For Entra-ID-protected webhooks: the tenant the destination app
lives in. Sent only when set.

### spec.destination.webhook.activeDirectoryAppIdOrUri

`string`

For Entra-ID-protected webhooks: the application ID or URI Event
Grid acquires a token for. Sent only when set.

### spec.deliveryIdentity

`AzureEventgridEventSubscriptionIdentity`

Deliver AS the source topic's managed identity instead of the
Event Grid service principal -- required for identity-based
access on the destination (e.g. a storage queue with shared keys
disabled). The identity named here must exist ON the source topic
(its identity block), and needs data-plane write access on the
destination.

- rule: user_assigned_identity is required for USER_ASSIGNED and must be unset for SYSTEM_ASSIGNED

### spec.deliveryIdentity.type

`enum` · required

Identity flavor: the source topic's system-assigned identity, or
one of its user-assigned identities.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_eventgrid_event_subscription_identity_type_unspecified` -- Not specified: rejected -- an identity block requires a flavor.
- `SYSTEM_ASSIGNED` -- The source topic's system-assigned identity. Wire value: "SystemAssigned".
- `USER_ASSIGNED` -- One of the source topic's user-assigned identities. Wire value: "UserAssigned".

### spec.deliveryIdentity.userAssignedIdentity

`string | valueFrom`

For USER_ASSIGNED: which of the source topic's user-assigned
identities to deliver as, by ARM ID.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.deliveryProperties

`[]AzureEventgridEventSubscriptionDeliveryProperty`

Extra headers/properties stamped onto every delivered event --
static values (e.g. an API key the handler expects) or values
lifted from the event by JSON path. Azure applies these on every
destination type EXCEPT storage queues (queue messages carry no
custom properties -- entries are ignored there).

- rule: Static entries require value (and may set secret); Dynamic entries require source_field and must not set value or secret

### spec.deliveryProperties[].headerName

`string` · required

The header/property name the destination sees (e.g.
"x-api-key" on a webhook, an application property on Service
Bus/Event Hub messages).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.deliveryProperties[].type

`string` · required

"Static" stamps a fixed value; "Dynamic" lifts the value from
each event by JSON path.

- rule: {"required":true,"string":{"in":["Static","Dynamic"]}}

### spec.deliveryProperties[].value

`string | valueFrom` · sensitive

For "Static": the value to stamp. Often a credential the handler
expects -- reference a managed secret or another resource's
output rather than embedding plaintext in manifests.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.deliveryProperties[].sourceField

`string`

For "Dynamic": the JSON path into the event the value is lifted
from (e.g. "data.system", "subject").

### spec.deliveryProperties[].secret

`bool`

For "Static": hide the value from Azure's own read APIs (they
return "Hidden" instead). Set true for credentials.

### spec.deadLetter

`AzureEventgridEventSubscriptionDeadLetter`

Where events that exhaust their retries are parked (a storage
blob container) instead of being dropped. Strongly recommended
for at-least-once processing: without it, undeliverable events
vanish after the retry policy gives up.

### spec.deadLetter.storageAccountId

`string | valueFrom` · required

The storage account holding the dead-letter container, by ARM ID
-- defaults to referencing an AzureStorageAccount's
storage_account_id output.

Dead-lettered events are WRITTEN into this account; the subscription
belongs to the topic it subscribes to, so on a diagram the reference is
access, not placement.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.deadLetter.storageBlobContainerName

`string` · required

The blob container dead-lettered events are written into (the
container must exist -- AzureStorageContainer manages one).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.deadLetterIdentity

`AzureEventgridEventSubscriptionIdentity`

Dead-letter AS the source topic's managed identity (the
delivery_identity contract, applied to the dead-letter write).
Requires dead_letter to be configured.

- rule: user_assigned_identity is required for USER_ASSIGNED and must be unset for SYSTEM_ASSIGNED

### spec.deadLetterIdentity.type

`enum` · required

Identity flavor: the source topic's system-assigned identity, or
one of its user-assigned identities.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_eventgrid_event_subscription_identity_type_unspecified` -- Not specified: rejected -- an identity block requires a flavor.
- `SYSTEM_ASSIGNED` -- The source topic's system-assigned identity. Wire value: "SystemAssigned".
- `USER_ASSIGNED` -- One of the source topic's user-assigned identities. Wire value: "UserAssigned".

### spec.deadLetterIdentity.userAssignedIdentity

`string | valueFrom`

For USER_ASSIGNED: which of the source topic's user-assigned
identities to deliver as, by ARM ID.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.eventDeliverySchema

`string` · optional (explicit presence)

The envelope events are DELIVERED in. "EventGridSchema" (Azure's
native envelope) is the default; "CloudEventSchemaV1_0" is the
CNCF standard; "CustomInputSchema" passes a custom-schema topic's
events through unmapped. Defaults to "EventGridSchema" -- the
platform sends the default explicitly.

The delivery schema must be one the SOURCE can map: a topic whose
input_schema is CloudEventSchemaV1_0 cannot deliver
EventGridSchema -- ARM rejects the create with InvalidRequest
("cannot be used in combination with the topic's input schema"),
a contract enforced only server-side. When subscribing to a
CloudEvents topic, set this field to CloudEventSchemaV1_0
explicitly; the default only suits EventGridSchema-input sources
(system topics and platform events always emit EventGridSchema).

**ForceNew**: changing the schema destroys and recreates the
subscription.

- default: `EventGridSchema`
- rule: {"string":{"in":["EventGridSchema","CloudEventSchemaV1_0","CustomInputSchema"]}}

### spec.includedEventTypes

`[]string`

Only deliver events of these types (e.g.
"Microsoft.Storage.BlobCreated"). An empty list delivers ALL
event types the source emits.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.subjectFilter

`AzureEventgridEventSubscriptionSubjectFilter`

Only deliver events whose subject matches a prefix/suffix. Leave
unset to deliver regardless of subject.

- rule: Set at least one of subject_begins_with, subject_ends_with, or case_sensitive

### spec.subjectFilter.subjectBeginsWith

`string`

Deliver events whose subject starts with this prefix (e.g.
"/blobServices/default/containers/invoices/").

### spec.subjectFilter.subjectEndsWith

`string`

Deliver events whose subject ends with this suffix (e.g. ".pdf").

### spec.subjectFilter.caseSensitive

`bool` · optional (explicit presence)

Whether prefix/suffix matching is case sensitive. Default: false
(Azure's default).

### spec.advancedFilter

`AzureEventgridEventSubscriptionAdvancedFilter`

Field-level filters over the event payload -- every configured
condition must match for an event to be delivered (conditions are
ANDed; values within one condition are ORed). Azure caps the
TOTAL number of values across all conditions of one subscription
at 25 -- the Terraform engine rejects an overflow at plan time on
scope-addressed subscriptions; system-topic-addressed
subscriptions and the Pulumi engine surface Azure's own deploy-
time rejection.

- rule: An advanced_filter block must carry at least one condition

### spec.advancedFilter.boolEquals

`[]AzureEventgridEventSubscriptionBoolEqualsFilter`

The key's value equals the given boolean.

### spec.advancedFilter.boolEquals[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.boolEquals[].value

`bool`

The boolean the key's value must equal.

### spec.advancedFilter.numberGreaterThan

`[]AzureEventgridEventSubscriptionNumberFilter`

The key's numeric value is strictly greater than the given value.

### spec.advancedFilter.numberGreaterThan[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.numberGreaterThan[].value

`double`

The number the key's value is compared against.

### spec.advancedFilter.numberGreaterThanOrEquals

`[]AzureEventgridEventSubscriptionNumberFilter`

The key's numeric value is greater than or equal to the given
value.

### spec.advancedFilter.numberGreaterThanOrEquals[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.numberGreaterThanOrEquals[].value

`double`

The number the key's value is compared against.

### spec.advancedFilter.numberLessThan

`[]AzureEventgridEventSubscriptionNumberFilter`

The key's numeric value is strictly less than the given value.

### spec.advancedFilter.numberLessThan[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.numberLessThan[].value

`double`

The number the key's value is compared against.

### spec.advancedFilter.numberLessThanOrEquals

`[]AzureEventgridEventSubscriptionNumberFilter`

The key's numeric value is less than or equal to the given value.

### spec.advancedFilter.numberLessThanOrEquals[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.numberLessThanOrEquals[].value

`double`

The number the key's value is compared against.

### spec.advancedFilter.numberIn

`[]AzureEventgridEventSubscriptionNumberListFilter`

The key's numeric value is one of the listed values.

### spec.advancedFilter.numberIn[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.numberIn[].values

`[]double` · required

The alternative values (OR semantics).

- rule: {"repeated":{"minItems":"1","maxItems":"25"}}

### spec.advancedFilter.numberNotIn

`[]AzureEventgridEventSubscriptionNumberListFilter`

The key's numeric value is none of the listed values.

### spec.advancedFilter.numberNotIn[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.numberNotIn[].values

`[]double` · required

The alternative values (OR semantics).

- rule: {"repeated":{"minItems":"1","maxItems":"25"}}

### spec.advancedFilter.numberInRange

`[]AzureEventgridEventSubscriptionNumberRangeFilter`

The key's numeric value falls inside one of the listed ranges.

### spec.advancedFilter.numberInRange[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.numberInRange[].ranges

`[]AzureEventgridEventSubscriptionNumberRange` · required

The alternative ranges (OR semantics), each an inclusive
[from, to] pair.

- rule: {"repeated":{"minItems":"1","maxItems":"25"}}

### spec.advancedFilter.numberInRange[].ranges[].from

`double`

The range's inclusive lower bound.

### spec.advancedFilter.numberInRange[].ranges[].to

`double`

The range's inclusive upper bound.

### spec.advancedFilter.numberNotInRange

`[]AzureEventgridEventSubscriptionNumberRangeFilter`

The key's numeric value falls outside every listed range.

### spec.advancedFilter.numberNotInRange[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.numberNotInRange[].ranges

`[]AzureEventgridEventSubscriptionNumberRange` · required

The alternative ranges (OR semantics), each an inclusive
[from, to] pair.

- rule: {"repeated":{"minItems":"1","maxItems":"25"}}

### spec.advancedFilter.numberNotInRange[].ranges[].from

`double`

The range's inclusive lower bound.

### spec.advancedFilter.numberNotInRange[].ranges[].to

`double`

The range's inclusive upper bound.

### spec.advancedFilter.stringBeginsWith

`[]AzureEventgridEventSubscriptionStringListFilter`

The key's string value starts with one of the listed prefixes.

### spec.advancedFilter.stringBeginsWith[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.stringBeginsWith[].values

`[]string` · required

The alternative values (OR semantics).

- rule: {"repeated":{"minItems":"1","maxItems":"25","items":{"string":{"minLen":"1"}}}}

### spec.advancedFilter.stringNotBeginsWith

`[]AzureEventgridEventSubscriptionStringListFilter`

The key's string value starts with none of the listed prefixes.

### spec.advancedFilter.stringNotBeginsWith[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.stringNotBeginsWith[].values

`[]string` · required

The alternative values (OR semantics).

- rule: {"repeated":{"minItems":"1","maxItems":"25","items":{"string":{"minLen":"1"}}}}

### spec.advancedFilter.stringEndsWith

`[]AzureEventgridEventSubscriptionStringListFilter`

The key's string value ends with one of the listed suffixes.

### spec.advancedFilter.stringEndsWith[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.stringEndsWith[].values

`[]string` · required

The alternative values (OR semantics).

- rule: {"repeated":{"minItems":"1","maxItems":"25","items":{"string":{"minLen":"1"}}}}

### spec.advancedFilter.stringNotEndsWith

`[]AzureEventgridEventSubscriptionStringListFilter`

The key's string value ends with none of the listed suffixes.

### spec.advancedFilter.stringNotEndsWith[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.stringNotEndsWith[].values

`[]string` · required

The alternative values (OR semantics).

- rule: {"repeated":{"minItems":"1","maxItems":"25","items":{"string":{"minLen":"1"}}}}

### spec.advancedFilter.stringContains

`[]AzureEventgridEventSubscriptionStringListFilter`

The key's string value contains one of the listed substrings.

### spec.advancedFilter.stringContains[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.stringContains[].values

`[]string` · required

The alternative values (OR semantics).

- rule: {"repeated":{"minItems":"1","maxItems":"25","items":{"string":{"minLen":"1"}}}}

### spec.advancedFilter.stringNotContains

`[]AzureEventgridEventSubscriptionStringListFilter`

The key's string value contains none of the listed substrings.

### spec.advancedFilter.stringNotContains[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.stringNotContains[].values

`[]string` · required

The alternative values (OR semantics).

- rule: {"repeated":{"minItems":"1","maxItems":"25","items":{"string":{"minLen":"1"}}}}

### spec.advancedFilter.stringIn

`[]AzureEventgridEventSubscriptionStringListFilter`

The key's string value is one of the listed values.

### spec.advancedFilter.stringIn[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.stringIn[].values

`[]string` · required

The alternative values (OR semantics).

- rule: {"repeated":{"minItems":"1","maxItems":"25","items":{"string":{"minLen":"1"}}}}

### spec.advancedFilter.stringNotIn

`[]AzureEventgridEventSubscriptionStringListFilter`

The key's string value is none of the listed values.

### spec.advancedFilter.stringNotIn[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.stringNotIn[].values

`[]string` · required

The alternative values (OR semantics).

- rule: {"repeated":{"minItems":"1","maxItems":"25","items":{"string":{"minLen":"1"}}}}

### spec.advancedFilter.isNotNull

`[]AzureEventgridEventSubscriptionKeyFilter`

The key is present with a non-null value.

### spec.advancedFilter.isNotNull[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilter.isNullOrUndefined

`[]AzureEventgridEventSubscriptionKeyFilter`

The key is absent, null, or undefined.

### spec.advancedFilter.isNullOrUndefined[].key

`string` · required

The JSON path into the event the condition reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.advancedFilteringOnArraysEnabled

`bool` · optional (explicit presence)

Whether a filter key may resolve to an ARRAY in the payload
(matching if any element matches) instead of a single value.
Default: false (Azure's default).

- default: `false`

### spec.labels

`[]string`

Free-form labels stored on the subscription (Event Grid metadata,
not ARM tags).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.expirationTimeUtc

`string` · optional (explicit presence)

When the subscription expires and stops delivering, as an RFC
3339 UTC timestamp (e.g. "2027-01-01T00:00:00Z"). Leave unset for
a subscription that never expires. Useful for temporary
integrations and trials.

- rule: expiration_time_utc must be an RFC 3339 UTC timestamp like 2027-01-01T00:00:00Z

### spec.retryPolicy

`AzureEventgridEventSubscriptionRetryPolicy`

How Event Grid retries failed deliveries before dead-lettering or
dropping. Leave unset to accept Azure's defaults (30 attempts,
1440-minute time-to-live) -- the service owns those defaults, so
the modules send the block only when set.

### spec.retryPolicy.maxDeliveryAttempts

`int32`

How many delivery attempts Event Grid makes before giving up
(1-30).

- rule: {"int32":{"lte":30,"gte":1}}

### spec.retryPolicy.eventTimeToLive

`int32`

How long an undelivered event stays eligible for retry, in
minutes (1-1440).

- rule: {"int32":{"lte":1440,"gte":1}}

## Validation Rules

- `event_subscription_addressing_exactly_one`: Set exactly one of scope or system_topic_id -- the addressing choice determines the source
- `event_subscription_dead_letter_identity_requires_dead_letter`: dead_letter_identity requires dead_letter to be configured

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureEventgridEventSubscription, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.event_subscription_id` | `string` | The subscription's Azure Resource Manager ID. The shape follows the addressing choice: scope-addressed subscriptions extend the source's ID ({scope}/providers/Microsoft.EventGrid/eventSubscriptions/{name}); system-topic subscriptions live under the topic ({system_topic_id}/eventSubscriptions/{name}). |
| `status.outputs.event_subscription_name` | `string` | The subscription's name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.scope` | AzureEventgridTopic | `status.outputs.topic_id` |
| `spec.systemTopicId` | AzureEventgridSystemTopic | `status.outputs.system_topic_id` |
| `spec.destination.eventhubId` | AzureEventHub | `status.outputs.event_hub_id` |
| `spec.destination.serviceBusQueueId` | AzureServiceBusQueue | `status.outputs.queue_id` |
| `spec.destination.serviceBusTopicId` | AzureServiceBusTopic | `status.outputs.topic_id` |
| `spec.destination.storageQueue.storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.deliveryIdentity.userAssignedIdentity` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.deadLetter.storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.deadLetterIdentity.userAssignedIdentity` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## See Also

- [Overview](../README.md)
