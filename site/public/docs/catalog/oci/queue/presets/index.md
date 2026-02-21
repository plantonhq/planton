---
title: "Presets"
description: "Ready-to-deploy configuration presets for Queue"
type: "preset-list"
componentSlug: "queue"
componentTitle: "Queue"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-standard-with-dlq"
    rank: "01"
    title: "Standard Queue with Dead Letter Queue"
    excerpt: "This preset creates an OCI Queue configured for typical microservice-to-microservice asynchronous messaging. Messages are retained for 7 days, consumers have a 30-second visibility window to process..."
  - slug: "02-encrypted-large-messages"
    rank: "02"
    title: "Encrypted Queue with Large Messages and Consumer Groups"
    excerpt: "This preset creates an enterprise-grade OCI Queue with customer-managed KMS encryption, large message support (up to 512 KB), and consumer groups for partitioned consumption. The longer visibility..."
---

# Queue Presets

Ready-to-deploy configuration presets for Queue. Each preset is a complete manifest you can copy, customize, and deploy.
