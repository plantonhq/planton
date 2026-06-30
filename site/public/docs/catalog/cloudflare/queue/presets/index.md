---
title: "Presets"
description: "Ready-to-deploy configuration presets for Queue"
type: "preset-list"
componentSlug: "queue"
componentTitle: "Queue"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-worker-consumer"
    rank: "01"
    title: "Preset: Queue with a Worker (push) consumer"
    excerpt: "A queue consumed automatically by a Worker — the most common Queues setup. Cloudflare invokes the Worker with batches of messages as they arrive and autoscales the number of concurrent invocations."
  - slug: "02-http-pull"
    rank: "02"
    title: "Preset: Queue with an HTTP (pull) consumer"
    excerpt: "A queue consumed by external clients that pull and acknowledge messages over the REST API, rather than by a Worker. Use this when the consumer lives outside Cloudflare Workers or needs fine-grained..."
---

# Queue Presets

Ready-to-deploy configuration presets for Queue. Each preset is a complete manifest you can copy, customize, and deploy.
