---
title: "Presets"
description: "Ready-to-deploy configuration presets for SQS Queue"
type: "preset-list"
componentSlug: "sqs-queue"
componentTitle: "SQS Queue"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-standard-queue"
    rank: "01"
    title: "Standard Queue"
    excerpt: "This preset creates a Standard SQS queue with SQS-managed encryption and long polling enabled. It is the fastest way to get a production-safe message queue running with sensible defaults."
  - slug: "02-fifo-with-deduplication"
    rank: "02"
    title: "FIFO Queue with Deduplication"
    excerpt: "This preset creates a FIFO SQS queue with content-based deduplication, high-throughput mode, and a dead letter queue for failed messages. Designed for workflows that require exactly-once processing..."
---

# SQS Queue Presets

Ready-to-deploy configuration presets for SQS Queue. Each preset is a complete manifest you can copy, customize, and deploy.
