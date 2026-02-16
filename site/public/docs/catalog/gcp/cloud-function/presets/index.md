---
title: "Presets"
description: "Ready-to-deploy configuration presets for Cloud Function"
type: "preset-list"
componentSlug: "cloud-function"
componentTitle: "Cloud Function"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-http-trigger"
    rank: "01"
    title: "HTTP-Triggered Cloud Function"
    excerpt: "This preset creates a Gen 2 Cloud Function invoked via HTTP requests. It uses all recommended defaults: 256 MiB memory, 60-second timeout, scale-to-zero, and authenticated access. Source code is..."
  - slug: "02-pubsub-event"
    rank: "02"
    title: "Pub/Sub Event-Driven Cloud Function"
    excerpt: "This preset creates a Gen 2 Cloud Function triggered by Pub/Sub messages. It processes one event at a time per instance, uses internal-only ingress (no public HTTP endpoint), and does not retry..."
---

# Cloud Function Presets

Ready-to-deploy configuration presets for Cloud Function. Each preset is a complete manifest you can copy, customize, and deploy.
