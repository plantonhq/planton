# Preset: Queue with an HTTP (pull) consumer

A queue consumed by external clients that pull and acknowledge messages over the
REST API, rather than by a Worker. Use this when the consumer lives outside
Cloudflare Workers or needs fine-grained control over consumption rate.

## When to use

- The consumer runs on existing infrastructure (GPU workers, a data pipeline) and
  pulls when it is ready.
- You need to control consumption rate precisely, or autoscale consumers based on
  the queue backlog.

## Key choices

- `consumer.type: http_pull`: no Worker is attached; clients call the pull/ack
  REST API with an API token carrying `Queues` read/write.
- `consumer.settings.visibilityTimeoutMs`: how long a pulled batch is leased
  exclusively before it becomes available again if not acknowledged (max 12h).
- `consumer.settings.batchSize`: messages returned per pull (max 100).

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |

## Pulling messages

Clients pull from `https://api.cloudflare.com/client/v4/accounts/<account>/queues/<queue_id>/messages/pull`
and acknowledge by `lease_id`. The queue id is published as `status.outputs.queue_id`.
