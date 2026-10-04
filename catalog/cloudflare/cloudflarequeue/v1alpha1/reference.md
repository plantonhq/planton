# CloudflareQueue

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareQueueSpec configures a Cloudflare Queue: a managed, guaranteed-delivery
message queue for Cloudflare Workers. Producers (a Worker's `queues` binding or an R2
bucket's event notifications) write messages; a single consumer reads them.

A queue has at most one consumer at the resource level, so the consumer is modeled as
a nested `consumer` block rather than a separate kind: it has no lifecycle independent
of the queue and is meaningless without it. The deployment still provisions the
underlying consumer as its own provider resource, so editing the consumer never
recreates the queue.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareQueue
metadata:
  name: test-queue
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  queueName: orders-queue
  settings:
    deliveryDelay: 0
    deliveryPaused: false
    messageRetentionPeriod: 86400
  consumer:
    type: worker
    scriptName:
      value: orders-consumer
    settings:
      batchSize: 25
      maxConcurrency: 10
      maxRetries: 3
      maxWaitTimeMs: 1000
      retryDelay: 30
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.queueName` | `string` | yes |  |  |
| `spec.settings` | `CloudflareQueueSettings` |  |  |  |
| `spec.settings.deliveryDelay` | `int64` |  |  |  |
| `spec.settings.deliveryPaused` | `bool` |  |  |  |
| `spec.settings.messageRetentionPeriod` | `int64` |  |  |  |
| `spec.consumer` | `CloudflareQueueConsumer` |  |  |  |
| `spec.consumer.type` | `enum` | yes |  |  |
| `spec.consumer.scriptName` | `string \| valueFrom` |  |  | CloudflareWorker (`status.outputs.script_name`) |
| `spec.consumer.deadLetterQueue` | `string \| valueFrom` |  |  | CloudflareQueue (`status.outputs.queue_name`) |
| `spec.consumer.settings` | `CloudflareQueueConsumerSettings` |  |  |  |
| `spec.consumer.settings.batchSize` | `int64` |  |  |  |
| `spec.consumer.settings.maxConcurrency` | `int64` |  |  |  |
| `spec.consumer.settings.maxRetries` | `int64` |  |  |  |
| `spec.consumer.settings.maxWaitTimeMs` | `int64` |  |  |  |
| `spec.consumer.settings.retryDelay` | `int64` |  |  |  |
| `spec.consumer.settings.visibilityTimeoutMs` | `int64` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns this queue.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.queueName

`string` · required

The queue name (shown in the dashboard; referenced by a Worker's `queues`
producer binding and by R2 event notifications).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"63","pattern":"^[a-zA-Z0-9][a-zA-Z0-9-]*$"}}

### spec.settings

`CloudflareQueueSettings`

Queue-level delivery settings. Omit to accept Cloudflare's defaults.

### spec.settings.deliveryDelay

`int64`

Seconds to delay delivery of every message to the consumer. Leave 0 for no delay.

- rule: delivery_delay must be between 0 and 86400 seconds (24h)

### spec.settings.deliveryPaused

`bool`

Pause delivery to the consumer. Producers can still enqueue while paused.

### spec.settings.messageRetentionPeriod

`int64`

Seconds an unconsumed message is retained before it is dropped. Leave 0 to use
Cloudflare's default (345600 = 4 days); otherwise it must be between 60 seconds
and 1209600 (14 days). Workers Free caps retention at 86400 (24h) -- a plan
limit the API enforces at write time, not a shape rule.

- rule: message_retention_period must be 0 (default) or between 60 and 1209600 seconds (14 days)

### spec.consumer

`CloudflareQueueConsumer`

The queue's single consumer. Omit when nothing consumes the queue yet (producers
can still write to it). A push (worker) consumer runs a Worker automatically; a
pull (http_pull) consumer is read over HTTP by external clients. When adopting
an existing queue that already has a consumer, delete that consumer first (or
omit this field): Cloudflare enforces one consumer per queue and refuses a
duplicate create with error 11004 (measured 2026-08-26), and the consumer has
no import path at provider v5.23.0.

- rule: script_name is required when consumer type is worker
- rule: script_name may only be set when consumer type is worker

### spec.consumer.type

`enum` · required

Whether the queue is consumed by a Worker (push) or pulled over HTTP.

- rule: consumer type must be one of worker, http_pull
- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `consumer_type_unspecified`
- `worker` -- Push consumer: a Worker is invoked automatically with batches of messages.
- `http_pull` -- Pull consumer: external HTTP clients pull batches over the REST API.

### spec.consumer.scriptName

`string | valueFrom`

The consuming Worker's script name, or a reference to a CloudflareWorker resource.
Required for worker (push) consumers; leave unset for http_pull consumers.

- references: CloudflareWorker (`status.outputs.script_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareWorker, name: <that resource's name>, fieldPath: status.outputs.script_name}} -- a bare string does not parse

### spec.consumer.deadLetterQueue

`string | valueFrom`

Optional dead-letter queue name, or a reference to another CloudflareQueue.
Messages that exhaust their retries are delivered here instead of being dropped.

- references: CloudflareQueue (`status.outputs.queue_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareQueue, name: <that resource's name>, fieldPath: status.outputs.queue_name}} -- a bare string does not parse

### spec.consumer.settings

`CloudflareQueueConsumerSettings`

Consumer batching and retry settings.

### spec.consumer.settings.batchSize

`int64`

Maximum messages per batch. Leave 0 for the default (10 for worker, 5 for pull);
the maximum is 100.

- rule: batch_size must be 0 (default) or between 1 and 100

### spec.consumer.settings.maxConcurrency

`int64`

Maximum concurrent consumer invocations (worker consumers only). Leave 0 to let
Cloudflare autoscale (recommended); the maximum is 250.

- rule: max_concurrency must be 0 (autoscale) or between 1 and 250

### spec.consumer.settings.maxRetries

`int64`

Maximum number of retries for a message before it is dropped or sent to the
dead-letter queue. Leave 0 for the default; the maximum is 100.

- rule: max_retries must be between 0 and 100

### spec.consumer.settings.maxWaitTimeMs

`int64`

Milliseconds to wait for a batch to fill before delivering it (worker consumers
only). Leave 0 for the default; the maximum is 60000 (60s).

- rule: max_wait_time_ms must be between 0 and 60000

### spec.consumer.settings.retryDelay

`int64`

Seconds to delay re-delivery of a message after a failed attempt. Leave 0 for no
additional delay; the maximum is 86400 (24h), the same ceiling Cloudflare applies
to every send-or-retry delay.

- rule: retry_delay must be between 0 and 86400 seconds (24h)

### spec.consumer.settings.visibilityTimeoutMs

`int64`

Milliseconds a pulled message is leased exclusively before it becomes available
again (http_pull consumers only). Leave 0 for the default (30s); the maximum is
43200000 (12h).

- rule: visibility_timeout_ms must be between 0 and 43200000 (12h)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareQueue, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.queue_id` | `string` | The Cloudflare-assigned identifier of the queue. A consumer references this value. |
| `status.outputs.queue_name` | `string` | The queue name (echoed; a Worker's producer binding and R2 event notifications reference this value). |
| `status.outputs.created_on` | `string` | RFC3339 timestamp of when the queue was created. |
| `status.outputs.modified_on` | `string` | RFC3339 timestamp of when the queue was last modified. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.consumer.scriptName` | CloudflareWorker | `status.outputs.script_name` |
| `spec.consumer.deadLetterQueue` | CloudflareQueue | `status.outputs.queue_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflarePagesProject | `spec.deploymentConfigs.preview.queueProducers[].queueName` | `status.outputs.queue_name` |
| CloudflarePagesProject | `spec.deploymentConfigs.production.queueProducers[].queueName` | `status.outputs.queue_name` |
| CloudflareQueue | `spec.consumer.deadLetterQueue` | `status.outputs.queue_name` |
| CloudflareR2Bucket | `spec.eventNotifications[].queue` | `status.outputs.queue_id` |
| CloudflareWorker | `spec.queues[].queueName` | `status.outputs.queue_name` |

## See Also

- [Overview](../README.md)
