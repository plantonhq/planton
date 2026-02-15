---
title: "Presets"
description: "Ready-to-deploy configuration presets for SNS Topic"
type: "preset-list"
componentSlug: "sns-topic"
componentTitle: "SNS Topic"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-standard-topic"
    rank: "01"
    title: "Standard Topic"
    excerpt: "This preset creates a Standard SNS topic with SHA256 message signatures. It is the simplest starting point for a notification or event distribution topic."
  - slug: "02-fifo-with-deduplication"
    rank: "02"
    title: "FIFO Topic with Deduplication"
    excerpt: "This preset creates a FIFO SNS topic with content-based deduplication and high-throughput mode (per message group). Designed for workflows that require exactly-once, ordered message delivery."
  - slug: "03-fanout-to-sqs"
    rank: "03"
    title: "Fan-Out to SQS with Filtering"
    excerpt: "This preset creates a Standard SNS topic with two SQS queue subscriptions, each with a message attribute filter policy. This demonstrates the core SNS fan-out pattern — a single topic distributes..."
---

# SNS Topic Presets

Ready-to-deploy configuration presets for SNS Topic. Each preset is a complete manifest you can copy, customize, and deploy.
