# Cloudflare Queue

A managed, guaranteed-delivery message queue for Cloudflare Workers, with an
optional push (Worker) or pull (HTTP) consumer.

## What Gets Created

- A `cloudflare_queue` with optional delivery settings.
- When `consumer` is set, a `cloudflare_queue_consumer` (provisioned separately so
  toggling it never recreates the queue).

## Prerequisites

- A Cloudflare account ID.
- For a worker consumer, a Worker script to consume the queue.

## Configuration Reference

**Required**

- `accountId` — Cloudflare account ID.
- `queueName` — the queue name.

**Optional**

- `settings.deliveryDelay`, `settings.deliveryPaused`, `settings.messageRetentionPeriod`.
- `consumer.type` (`worker` | `http_pull`), `consumer.scriptName`,
  `consumer.deadLetterQueue`, and `consumer.settings.*`.

## Stack Outputs

| Output | Description |
|---|---|
| `queue_id` | The queue ID |
| `queue_name` | The queue name |
| `created_on` | Creation timestamp |
| `modified_on` | Last-modified timestamp |

## Related Components

- CloudflareWorker
- CloudflareR2Bucket
