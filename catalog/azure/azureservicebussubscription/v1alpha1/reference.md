# AzureServiceBusSubscription

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureServiceBusSubscriptionSpec** defines the configuration for
creating a subscription under an Azure Service Bus topic: an independent,
optionally filtered view of the topic's message stream, with its own
consumer semantics (lock duration, delivery counts, sessions,
dead-lettering).

Subscriptions are many-per-topic and are typically owned by the
consuming team rather than the team that provisioned the namespace --
which is why they are a first-class kind referencing the topic.

**Filter rules fold inside the subscription** (the `rules` list): a
rule has no life outside its subscription and nothing references one.
Rules combine with OR semantics -- a message is delivered once if ANY
rule matches. Azure auto-creates a catch-all rule named `$Default`
(match everything) on every new subscription, so rules declared here
are ADDITIVE alongside it: the subscription keeps receiving everything
until the catch-all is removed. The management plane cannot remove or
overwrite a service-created rule declaratively (the providers refuse to
adopt an existing resource on create), so `$Default` is not a legal
declared name -- for restrictive delivery, remove the catch-all once
after creation:
`az servicebus topic subscription rule delete --name '$Default' ...`
(a one-time action; it never comes back unless the subscription is
recreated).

**ForceNew fields** (changing these replaces the subscription and
resets its read position): `subscription_name`, `requires_session`,
and the client_scoped_subscription block's identity fields.

## Example

```yaml
# Offline-plan manifest: a subscription exercising BOTH filter-rule
# families (SQL with an action, correlation with property matchers),
# dead-letter routing, and the client-scoped block.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureServiceBusSubscription
metadata:
  name: test-sb-subscription
spec:
  topicId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ServiceBus/namespaces/hack-servicebus-ns/topics/billing-invoices
  subscriptionName: audit-consumer
  maxDeliveryCount: 5
  lockDuration: PT2M
  defaultMessageTtl: P7D
  deadLetteringOnMessageExpiration: true
  deadLetteringOnFilterEvaluationError: false
  forwardDeadLetteredMessagesTo:
    value: poison-sink
  status: ACTIVE
  rules:
    - ruleName: emea-orders
      filterType: SQL_FILTER
      sqlFilter: "region = 'emea' AND quantity > 10"
      action: "SET sys.Label = 'routed'"
    - ruleName: order-created
      filterType: CORRELATION_FILTER
      correlationFilter:
        label: order-created
        contentType: application/json
        properties:
          tenant: contoso
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.topicId` | `string \| valueFrom` | yes |  | AzureServiceBusTopic (`status.outputs.topic_id`) |
| `spec.subscriptionName` | `string` | yes |  |  |
| `spec.maxDeliveryCount` | `int32` | yes |  |  |
| `spec.lockDuration` | `string` |  |  |  |
| `spec.defaultMessageTtl` | `string` |  |  |  |
| `spec.autoDeleteOnIdle` | `string` |  |  |  |
| `spec.deadLetteringOnMessageExpiration` | `bool` |  |  |  |
| `spec.deadLetteringOnFilterEvaluationError` | `bool` |  | `true` |  |
| `spec.requiresSession` | `bool` |  |  |  |
| `spec.batchedOperationsEnabled` | `bool` |  |  |  |
| `spec.forwardTo` | `string \| valueFrom` |  |  |  |
| `spec.forwardDeadLetteredMessagesTo` | `string \| valueFrom` |  |  |  |
| `spec.status` | `enum` |  |  |  |
| `spec.clientScopedSubscription` | `AzureServiceBusClientScopedSubscription` |  |  |  |
| `spec.clientScopedSubscription.clientId` | `string` |  |  |  |
| `spec.clientScopedSubscription.shareable` | `bool` |  |  |  |
| `spec.rules` | `[]AzureServiceBusSubscriptionRule` |  |  |  |
| `spec.rules[].ruleName` | `string` | yes |  |  |
| `spec.rules[].filterType` | `enum` |  |  |  |
| `spec.rules[].sqlFilter` | `string` | yes |  |  |
| `spec.rules[].correlationFilter` | `AzureServiceBusCorrelationFilter` |  |  |  |
| `spec.rules[].correlationFilter.correlationId` | `string` | yes |  |  |
| `spec.rules[].correlationFilter.messageId` | `string` | yes |  |  |
| `spec.rules[].correlationFilter.to` | `string` | yes |  |  |
| `spec.rules[].correlationFilter.replyTo` | `string` | yes |  |  |
| `spec.rules[].correlationFilter.label` | `string` | yes |  |  |
| `spec.rules[].correlationFilter.sessionId` | `string` | yes |  |  |
| `spec.rules[].correlationFilter.replyToSessionId` | `string` | yes |  |  |
| `spec.rules[].correlationFilter.contentType` | `string` | yes |  |  |
| `spec.rules[].correlationFilter.properties` | `map<string, string>` |  |  |  |
| `spec.rules[].action` | `string` | yes |  |  |

## Field Details

### spec.topicId

`string | valueFrom` · required

The topic this subscription attaches to, by ARM ID. References an
AzureServiceBusTopic's topic_id output so the topic and its
subscriptions compose in one manifest set. Fixed at creation.

- references: AzureServiceBusTopic (`status.outputs.topic_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureServiceBusTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.subscriptionName

`string` · required

The subscription's name -- unique within the topic, up to 50
characters. Starts with a letter, number, or underscore; letters,
numbers, periods, hyphens, and underscores in between.

**ForceNew**: changing the name replaces the subscription and resets
its read position -- undelivered messages in the old subscription are
lost.

- rule: subscription_name must be up to 50 characters of letters, numbers, periods, hyphens, and underscores, starting and ending with a letter, number, or underscore
- rule: {"required":true,"string":{"minLen":"1","maxLen":"50"}}

### spec.maxDeliveryCount

`int32` · required

How many delivery attempts a message gets before it is moved to the
subscription's dead-letter sub-queue. Minimum 1. Lower values
quarantine poison messages faster; higher values ride out transient
consumer failures. Azure has no server default here -- the
subscription's tolerance for redelivery is an explicit choice (10 is
the queue-side convention).

- rule: {"required":true,"int32":{"gte":1}}

### spec.lockDuration

`string` · optional (explicit presence)

How long a received message stays locked in PeekLock mode before it
returns to the subscription for redelivery, as an ISO 8601 duration.
Range PT5S to PT5M. Azure's default: PT1M. Size it to the consumer's
processing time (SDKs can renew locks).

### spec.defaultMessageTtl

`string` · optional (explicit presence)

Default time-to-live for messages in this subscription, as an ISO
8601 duration. May be shorter than the topic's TTL (the shorter one
wins), never effectively longer. Unset inherits the topic's TTL.

### spec.autoDeleteOnIdle

`string` · optional (explicit presence)

Whether the subscription is automatically deleted after sitting idle
for this ISO 8601 duration. Minimum PT5M. Unset means never
auto-delete. Auto-delete suits ephemeral per-instance subscriptions
(each consumer instance materializes its own view and cleans up after
itself).

### spec.deadLetteringOnMessageExpiration

`bool` · optional (explicit presence)

Whether messages that expire (exceed their TTL) are moved to the
subscription's dead-letter sub-queue instead of being silently
deleted.
Default: false

### spec.deadLetteringOnFilterEvaluationError

`bool` · optional (explicit presence)

Whether messages that FAIL filter evaluation (a SQL filter that
throws on a malformed message) are dead-lettered instead of dropped.
Azure's default is true -- keep it unless noisy malformed producers
flood the dead-letter sub-queue.
Default: true

- default: `true`

### spec.requiresSession

`bool` · optional (explicit presence)

Whether sessions are enabled. Session-aware subscriptions deliver
each session's messages in strict order to one consumer at a time,
with broker-stored session state. Pair with the topic's
support_ordering for ordered publish-subscribe.

**ForceNew**: fixed at creation.
Default: false

### spec.batchedOperationsEnabled

`bool` · optional (explicit presence)

Whether clients may batch multiple operations into one broker call
for higher throughput.
Default: false (Azure's subscription-side default)

### spec.forwardTo

`string | valueFrom`

Auto-forward every message arriving in this subscription to another
queue or topic in the SAME namespace, by entity name (not ARM ID) --
the fan-out-then-collect routing primitive: subscriptions filter,
forwarding funnels the matches into a work queue. Reference the
target's queue_name or topic_name output (no default kind: either
entity type is a legal target), or pass a literal name. The target
must exist before the subscription.

Containment-exempt: the subscription LIVES in its own topic and
forwards TO the target. When the target is another topic (a
container), the subscription would otherwise be torn between the
topic it belongs to and the topic it feeds; on a diagram it stays
in its topic with a line out.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.forwardDeadLetteredMessagesTo

`string | valueFrom`

Auto-forward dead-lettered messages to another queue or topic in the
same namespace, by entity name -- centralize poison-message handling.
Reference the target's queue_name or topic_name output, or pass a
literal name. Access, not placement, like `forward_to`.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.status

`enum`

The subscription's gate state: ACTIVE (normal), DISABLED (delivery
and new arrivals stopped), or RECEIVE_DISABLED (arrivals accumulate,
delivery paused). Unspecified deploys ACTIVE.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_service_bus_subscription_status_unspecified` -- Not specified -- deploys ACTIVE.
- `ACTIVE` -- Normal operation: matched messages arrive and are delivered.
- `DISABLED` -- Delivery and new arrivals are stopped; stored messages are retained.
- `RECEIVE_DISABLED` -- Matched messages keep accumulating but delivery is paused -- buffer while consumers are offline.

### spec.clientScopedSubscription

`AzureServiceBusClientScopedSubscription`

Client-scoped (client-affine) subscription -- the JMS 2.0 shared/
unshared durable-subscription model, where the subscription is bound
to a specific client ID. Leave unset for the normal shared
subscription every Service Bus consumer uses.

### spec.clientScopedSubscription.clientId

`string`

The client ID the subscription is scoped to. Azure stores the entity
as `{name}$${client_id}$$D` internally; the modules and outputs keep
the user-facing name. **ForceNew**: fixed at creation.

### spec.clientScopedSubscription.shareable

`bool` · optional (explicit presence)

Whether multiple clients presenting the same client ID may consume
concurrently (JMS "shared" vs "unshared" durable subscription).
Azure's default is true. **ForceNew**: fixed at creation.

### spec.rules

`[]AzureServiceBusSubscriptionRule`

The subscription's filter rules. Each rule admits messages by SQL
expression or by correlation matching; rules combine with OR
semantics ALONGSIDE Azure's auto-created `$Default` catch-all (see
the message comment for the one-time catch-all removal that makes
declared filters restrictive). An empty list keeps just the
catch-all (deliver everything).

- rule: sql_filter is required with filter_type SQL_FILTER and must be absent with CORRELATION_FILTER
- rule: correlation_filter is required with filter_type CORRELATION_FILTER and must be absent with SQL_FILTER

### spec.rules[].ruleName

`string` · required

The rule's name -- unique within the subscription, up to 50
characters. `$Default` is reserved: Azure creates that catch-all
rule itself, and an existing service-created resource cannot be
adopted by a declared one (the providers' import check refuses it).

- rule: $Default is the service-created catch-all rule and cannot be declared -- name the rule for what it admits, and remove the catch-all once out-of-band for restrictive delivery (az servicebus topic subscription rule delete --name '$Default')
- rule: {"required":true,"string":{"minLen":"1","maxLen":"50"}}

### spec.rules[].filterType

`enum`

How this rule matches messages: SQL_FILTER (a SQL-92-like expression
over system and user properties -- flexible, evaluated per message)
or CORRELATION_FILTER (equality matching on correlation properties --
cheaper at high throughput).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_service_bus_filter_type_unspecified` -- Not specified -- invalid; choose SQL_FILTER or CORRELATION_FILTER.
- `SQL_FILTER` -- A SQL-92-like expression over system (sys.*) and user properties. Flexible; evaluated per message.
- `CORRELATION_FILTER` -- Equality matching on correlation properties -- cheaper than SQL at high throughput.

### spec.rules[].sqlFilter

`string` · required · optional (explicit presence)

The SQL filter expression, e.g. "sys.Label = 'important' AND
quantity > 10". Required with (and only valid with) SQL_FILTER. Max
1024 characters.

- rule: {"string":{"minLen":"1","maxLen":"1024"}}

### spec.rules[].correlationFilter

`AzureServiceBusCorrelationFilter`

The correlation matcher. Required with (and only valid with)
CORRELATION_FILTER.

- rule: a correlation filter needs at least one matcher -- set correlation_id, message_id, to, reply_to, label, session_id, reply_to_session_id, content_type, or a properties entry

### spec.rules[].correlationFilter.correlationId

`string` · required · optional (explicit presence)

Match the message's CorrelationId.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].correlationFilter.messageId

`string` · required · optional (explicit presence)

Match the message's MessageId.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].correlationFilter.to

`string` · required · optional (explicit presence)

Match the message's To address.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].correlationFilter.replyTo

`string` · required · optional (explicit presence)

Match the message's ReplyTo address.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].correlationFilter.label

`string` · required · optional (explicit presence)

Match the message's Label (Subject).

- rule: {"string":{"minLen":"1"}}

### spec.rules[].correlationFilter.sessionId

`string` · required · optional (explicit presence)

Match the message's SessionId.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].correlationFilter.replyToSessionId

`string` · required · optional (explicit presence)

Match the message's ReplyToSessionId.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].correlationFilter.contentType

`string` · required · optional (explicit presence)

Match the message's ContentType.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].correlationFilter.properties

`map<string, string>`

Match user-defined message properties by exact key/value equality.

### spec.rules[].action

`string` · required · optional (explicit presence)

An optional SQL action that annotates matched messages (SET
sys.Label = 'processed', SET priority = 'high') before delivery --
routing metadata without touching the producer.

- rule: {"string":{"minLen":"1"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureServiceBusSubscription, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.subscription_id` | `string` | The Azure Resource Manager ID of the subscription. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.ServiceBus/namespaces/{ns}/topics/{topic}/subscriptions/{name} |
| `status.outputs.subscription_name` | `string` | The subscription's name -- what consumers reference (together with the topic name) when receiving. |
| `status.outputs.topic_name` | `string` | The name of the topic the subscription attaches to, parsed from the resolved topic ID. |
| `status.outputs.namespace_name` | `string` | The name of the Service Bus namespace, parsed from the resolved topic ID -- saves consumers a second reference when they need the namespace/topic/subscription triple. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.topicId` | AzureServiceBusTopic | `status.outputs.topic_id` |

## See Also

- [Overview](../README.md)
