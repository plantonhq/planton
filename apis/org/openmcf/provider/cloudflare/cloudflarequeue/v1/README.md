# CloudflareQueue

Provision a Cloudflare Queue: a managed, guaranteed-delivery message queue for
Cloudflare Workers. Producers (a Worker's `queues` binding, or an R2 bucket's
event notifications) write messages; a single consumer reads them — either a
Worker invoked automatically (push) or an external HTTP client (pull).

## Why a first-class Queue

A queue decouples producers from consumers so each scales and fails
independently. Modeling it as its own node lets the resource graph wire a
producing Worker, the queue, and a consuming Worker (or an R2 bucket's event
stream) as explicit, independently-owned edges.

The consumer is configured inline (`consumer`) rather than as a separate kind: at
the resource level a queue has exactly one consumer, with no lifecycle of its own.
The module still provisions the consumer as a distinct provider resource, so
toggling it never recreates the queue.

## Quick start

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareQueue
metadata:
  name: orders-queue
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  queueName: orders-queue
  consumer:
    type: worker
    scriptName:
      valueFrom:
        kind: CloudflareWorker
        name: orders-consumer
        fieldPath: status.outputs.script_name
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `accountId` | yes | 32-char Cloudflare account ID |
| `queueName` | yes | The queue name |
| `settings.deliveryDelay` | no | Seconds to delay delivery of all messages (0–86400) |
| `settings.deliveryPaused` | no | Pause delivery to the consumer |
| `settings.messageRetentionPeriod` | no | Retention seconds (60–86400; 0 = default) |
| `consumer.type` | yes* | `worker` (push) or `http_pull` |
| `consumer.scriptName` | worker only | Consuming Worker (literal name or `CloudflareWorker` ref) |
| `consumer.deadLetterQueue` | no | Dead-letter queue name (literal or `CloudflareQueue` ref) |
| `consumer.settings.batchSize` | no | Messages per batch (1–100) |
| `consumer.settings.maxConcurrency` | worker only | Max concurrent invocations (1–250; 0 = autoscale) |
| `consumer.settings.maxRetries` | no | Max retries (0–100) |
| `consumer.settings.maxWaitTimeMs` | worker only | Batch fill wait in ms (0–60000) |
| `consumer.settings.retryDelay` | no | Re-delivery delay in seconds (0–42300) |
| `consumer.settings.visibilityTimeoutMs` | http_pull only | Lease window in ms (0–43200000) |

\* `consumer` itself is optional; when present, `type` is required.

## Outputs

| Output | Description |
|---|---|
| `queue_id` | The queue ID (referenced by a consumer and event-notification producers) |
| `queue_name` | The queue name (referenced by a Worker producer binding and R2 event notifications) |
| `created_on` | Creation timestamp |
| `modified_on` | Last-modified timestamp |

## Related components

- `CloudflareWorker` — produces to a queue via its `queues` binding; a worker
  consumer is referenced by `consumer.scriptName`.
- `CloudflareR2Bucket` — forwards object events to a queue via `eventNotifications`.
