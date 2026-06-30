---
title: "Preset: Queue with a Worker (push) consumer"
description: "A queue consumed automatically by a Worker — the most common Queues setup. Cloudflare invokes the Worker with batches of messages as they arrive and autoscales the number of concurrent invocations."
type: "preset"
rank: "01"
presetSlug: "01-worker-consumer"
componentSlug: "queue"
componentTitle: "Queue"
provider: "cloudflare"
icon: "package"
order: 1
---

# Preset: Queue with a Worker (push) consumer

A queue consumed automatically by a Worker — the most common Queues setup. Cloudflare
invokes the Worker with batches of messages as they arrive and autoscales the
number of concurrent invocations.

## When to use

- You have a Worker that should process messages off a queue (emails, webhooks,
  uploads, fan-out work) without blocking the request path.

## Key choices

- `consumer.scriptName`: reference the consuming `CloudflareWorker` so the graph
  deploys it before wiring the consumer.
- `consumer.deadLetterQueue`: messages that exhaust `maxRetries` are sent here
  instead of being dropped — point it at another `CloudflareQueue`.
- `consumer.settings.maxConcurrency`: leave unset to autoscale (recommended); set
  a number to cap concurrent invocations (e.g. to protect a rate-limited
  downstream).

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
| `<consumer-worker-name>` | Name of the Worker that consumes this queue |

## Producing to the queue

Reference this queue from a `CloudflareWorker` `queues` producer binding:

```yaml
queues:
  - name: ORDERS
    queueName:
      valueFrom:
        kind: CloudflareQueue
        name: orders-queue
        fieldPath: status.outputs.queue_name
```
