# AzureServiceBusQueue

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureServiceBusQueueSpec** defines the configuration for creating a
queue inside an Azure Service Bus namespace: reliable point-to-point
messaging with FIFO delivery, at-least-once semantics, and a built-in
dead-letter sub-queue.

Queues are many-per-namespace with independent lifecycles, which is why
they are a first-class kind referencing the namespace rather than a list
folded into the namespace's spec. The queue's ARM ID is also a
data-plane RBAC scope (grant "Azure Service Bus Data Receiver/Sender" on
exactly this queue) and the target for queue-scoped SAS credentials
(AzureServiceBusAuthorizationRule with queue_id).

**Contracts enforced by Azure at apply time** (they depend on the parent
namespace's tier, which a reference cannot see at validation time):
- PREMIUM namespaces reject express_enabled.
- In a PARTITIONED premium namespace (premium_messaging_partitions of
  2 or 4) every queue must set partitioning_enabled true; in a
  non-partitioned namespace it must stay false. On BASIC/STANDARD,
  per-queue partitioning is a free choice.
- max_message_size_in_kilobytes is PREMIUM-only (multi-tenant tiers are
  fixed at 256 KB).

**ForceNew fields** (changing these replaces the queue and drops any
messages in it): `queue_name`, `partitioning_enabled`,
`requires_duplicate_detection`, `requires_session`.

## Example

```yaml
# Offline-plan manifest: a fully-dialed queue exercising the size
# ladder, duplicate detection with its window, sessions, dead-letter
# routing, the gate-state enum mapping, and forwarding.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureServiceBusQueue
metadata:
  name: test-sb-queue
spec:
  namespaceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ServiceBus/namespaces/hack-servicebus-ns
  queueName: orders/priority
  maxSizeInMegabytes: 5120
  requiresDuplicateDetection: true
  duplicateDetectionHistoryTimeWindow: PT30S
  defaultMessageTtl: P14D
  lockDuration: PT2M
  maxDeliveryCount: 5
  deadLetteringOnMessageExpiration: true
  requiresSession: true
  autoDeleteOnIdle: PT10M
  batchedOperationsEnabled: true
  forwardDeadLetteredMessagesTo:
    value: poison-sink
  status: SEND_DISABLED
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespaceId` | `string \| valueFrom` | yes |  | AzureServiceBusNamespace (`status.outputs.namespace_id`) |
| `spec.queueName` | `string` | yes |  |  |
| `spec.maxSizeInMegabytes` | `int32` |  |  |  |
| `spec.maxMessageSizeInKilobytes` | `int32` |  |  |  |
| `spec.partitioningEnabled` | `bool` |  |  |  |
| `spec.requiresDuplicateDetection` | `bool` |  |  |  |
| `spec.defaultMessageTtl` | `string` |  |  |  |
| `spec.duplicateDetectionHistoryTimeWindow` | `string` |  |  |  |
| `spec.lockDuration` | `string` |  |  |  |
| `spec.maxDeliveryCount` | `int32` |  |  |  |
| `spec.deadLetteringOnMessageExpiration` | `bool` |  |  |  |
| `spec.requiresSession` | `bool` |  |  |  |
| `spec.autoDeleteOnIdle` | `string` |  |  |  |
| `spec.batchedOperationsEnabled` | `bool` |  | `true` |  |
| `spec.expressEnabled` | `bool` |  |  |  |
| `spec.forwardTo` | `string \| valueFrom` |  |  |  |
| `spec.forwardDeadLetteredMessagesTo` | `string \| valueFrom` |  |  |  |
| `spec.status` | `enum` |  |  |  |

## Field Details

### spec.namespaceId

`string | valueFrom` · required

The Service Bus namespace the queue lives in, by ARM ID. References an
AzureServiceBusNamespace's namespace_id output so the namespace and its
queues compose in one manifest set. Fixed at creation.

- references: AzureServiceBusNamespace (`status.outputs.namespace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureServiceBusNamespace, name: <that resource's name>, fieldPath: status.outputs.namespace_id}} -- a bare string does not parse

### spec.queueName

`string` · required

The queue's name -- unique within the namespace, up to 260 characters.
Starts and ends with a letter or number; letters, numbers, periods,
hyphens, underscores, tildes, and forward slashes in between (slashes
create a logical hierarchy, e.g. "orders/priority").

**ForceNew**: changing the name replaces the queue.

- rule: queue_name must start and end with a letter or number and may contain letters, numbers, periods, hyphens, underscores, tildes, and forward slashes (max 260 characters)
- rule: {"required":true,"string":{"minLen":"1","maxLen":"260"}}

### spec.maxSizeInMegabytes

`int32` · optional (explicit presence)

Maximum queue size in megabytes -- how much message data the queue
holds before senders get rejected. Azure only sells fixed sizes:
1024, 2048, 3072, 4096, or 5120 MB on any tier; 10240, 20480, 40960,
or 81920 MB on PREMIUM (large sizes come from the dedicated tier's
storage). Unset lets Azure default the size for the namespace's tier
(1024 MB multi-tenant, 81920 MB premium). For partitioned
BASIC/STANDARD queues Azure multiplies the chosen size by the 16
internal partitions -- the effective capacity is 16x.

- rule: {"int32":{"in":[1024,2048,3072,4096,5120,10240,20480,40960,81920]}}

### spec.maxMessageSizeInKilobytes

`int32` · optional (explicit presence)

The largest message the queue accepts, in kilobytes. PREMIUM only --
multi-tenant tiers are fixed at 256 KB. Range: 1024 (1 MB) to 102400
(100 MB). Unset keeps Azure's default for the tier. Large messages
trade throughput for convenience; keep payloads small and pass blob
references where possible.

- rule: {"int32":{"lte":102400,"gte":1024}}

### spec.partitioningEnabled

`bool` · optional (explicit presence)

Whether the queue is spread across multiple message stores for higher
throughput. On BASIC/STANDARD this is a free per-queue choice. On
PREMIUM it is dictated by the namespace: a partitioned namespace
(premium_messaging_partitions 2 or 4) requires true, a non-partitioned
one requires false.

**ForceNew**: fixed at creation.
Default: false

### spec.requiresDuplicateDetection

`bool` · optional (explicit presence)

Whether Service Bus tracks MessageId values and silently drops
duplicates arriving within duplicate_detection_history_time_window.
Enables idempotent producers that retry sends safely.

**ForceNew**: fixed at creation.
Default: false

### spec.defaultMessageTtl

`string` · optional (explicit presence)

Default time-to-live for messages, as an ISO 8601 duration (e.g.
"P14D", "PT1H"). Messages older than this are removed -- or moved to
the dead-letter sub-queue when
dead_lettering_on_message_expiration is true. Unset means unbounded
(messages never expire). Senders can only shorten this per message,
never extend it.

### spec.duplicateDetectionHistoryTimeWindow

`string` · optional (explicit presence)

How long the duplicate-detection history remembers MessageIds, as an
ISO 8601 duration. Only meaningful with requires_duplicate_detection.
Azure's default: PT10M (10 minutes). Longer windows catch slower
duplicate producers at the cost of broker-side tracking state.

### spec.lockDuration

`string` · optional (explicit presence)

How long a received message stays locked in PeekLock mode before it
returns to the queue for redelivery, as an ISO 8601 duration. Range
PT5S to PT5M. Azure's default: PT1M. Size it to the consumer's
processing time -- too short causes duplicate processing, too long
delays recovery from crashed consumers (SDKs can renew locks).

### spec.maxDeliveryCount

`int32` · optional (explicit presence)

How many delivery attempts a message gets before it is moved to the
dead-letter sub-queue. Minimum 1. Azure's default: 10. Lower values
quarantine poison messages faster; higher values ride out transient
consumer failures.

- rule: {"int32":{"gte":1}}

### spec.deadLetteringOnMessageExpiration

`bool` · optional (explicit presence)

Whether messages that expire (exceed their TTL) are moved to the
dead-letter sub-queue instead of being silently deleted -- keep
expired messages inspectable and re-processable.
Default: false

### spec.requiresSession

`bool` · optional (explicit presence)

Whether sessions are enabled. Sessions group related messages by
SessionId for strict FIFO ordering, exclusive consumption, and
broker-stored session state. Session-aware queues require
session-aware consumers.

**ForceNew**: fixed at creation.
Default: false

### spec.autoDeleteOnIdle

`string` · optional (explicit presence)

Whether the queue is automatically deleted after sitting idle (no
sends or receives) for this ISO 8601 duration. Minimum PT5M. Unset
means never auto-delete -- the right posture for durable
infrastructure; auto-delete suits ephemeral per-instance queues.

### spec.batchedOperationsEnabled

`bool` · optional (explicit presence)

Whether clients may batch multiple operations into one broker call
for higher throughput. Azure's default is true; disable only for
strict per-operation latency requirements.
Default: true

- default: `true`

### spec.expressEnabled

`bool` · optional (explicit presence)

Whether Express Entities are enabled: messages are held in memory
before being written to storage, trading durability for latency.
BASIC/STANDARD only -- PREMIUM rejects express queues. Incompatible
with duplicate detection.
Default: false

### spec.forwardTo

`string | valueFrom`

Auto-forward every message arriving in this queue to another queue or
topic in the SAME namespace, by entity name (not ARM ID) -- the
classic fan-in/routing chain primitive. Reference the target's
queue_name or topic_name output (no default kind: either entity type
is a legal target), or pass a literal name. The target must exist
before the queue; forwarding to a session-aware target is rejected by
Azure.

Containment-exempt: the queue LIVES in its namespace and forwards
TO the target. A topic is a container (its subscriptions live in
it), so a queue forwarding to a topic by reference would otherwise
be drawn inside the topic it feeds; on a diagram the queue stays a
sibling of the topic with a line to it.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.forwardDeadLetteredMessagesTo

`string | valueFrom`

Auto-forward dead-lettered messages to another queue or topic in the
same namespace, by entity name -- centralize poison-message handling
instead of draining every queue's dead-letter sub-queue separately.
Reference the target's queue_name or topic_name output, or pass a
literal name. Access, not placement, like `forward_to`.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.status

`enum`

The queue's gate state: ACTIVE (normal), DISABLED (sends and receives
rejected; messages retained), SEND_DISABLED (receive-only drain mode),
or RECEIVE_DISABLED (accepts sends, delivers nothing). Unspecified
deploys ACTIVE. The transitional server states (Creating, Deleting,
Renaming, ...) are not knobs and are not modeled.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_service_bus_entity_status_unspecified` -- Not specified -- deploys ACTIVE.
- `ACTIVE` -- Normal operation: sends and receives flow.
- `DISABLED` -- Sends and receives are rejected; stored messages are retained.
- `SEND_DISABLED` -- Receive-only drain mode: new sends are rejected, consumers keep draining.
- `RECEIVE_DISABLED` -- Accepts sends but delivers nothing -- buffer while consumers are offline.

## Validation Rules

- `service_bus_queue_express_conflicts_with_duplicate_detection`: express_enabled holds messages in memory before storage, which is incompatible with duplicate detection -- disable one of the two
- `service_bus_queue_dedup_window_requires_dedup`: duplicate_detection_history_time_window only applies when requires_duplicate_detection is true

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureServiceBusQueue, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.queue_id` | `string` | The Azure Resource Manager ID of the queue. The scope for queue-level data-plane role assignments (Azure Service Bus Data Receiver/Sender) and the parent reference for queue-scoped SAS rules. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.ServiceBus/namespaces/{ns}/queues/{name} |
| `status.outputs.queue_name` | `string` | The queue's name -- what SDKs, connection strings, and function bindings reference within the namespace. |
| `status.outputs.namespace_name` | `string` | The name of the Service Bus namespace the queue lives in, parsed from the resolved namespace ID -- saves consumers a second reference when they need the namespace/queue pair. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespaceId` | AzureServiceBusNamespace | `status.outputs.namespace_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureEventgridEventSubscription | `spec.destination.serviceBusQueueId` | `status.outputs.queue_id` |
| AzureServiceBusAuthorizationRule | `spec.queueId` | `status.outputs.queue_id` |

## See Also

- [Overview](../README.md)
