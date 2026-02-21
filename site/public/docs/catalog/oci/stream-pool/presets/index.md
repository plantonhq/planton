---
title: "Presets"
description: "Ready-to-deploy configuration presets for Stream Pool"
type: "preset-list"
componentSlug: "stream-pool"
componentTitle: "Stream Pool"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-public-kafka-compatible"
    rank: "01"
    title: "Public Kafka-Compatible"
    excerpt: "This preset creates an OCI Stream Pool with Kafka compatibility settings and two pre-defined streams: an `events` stream for application event sourcing and a `commands` stream for command-driven..."
  - slug: "02-private-encrypted"
    rank: "02"
    title: "Private Encrypted"
    excerpt: "This preset creates a production-grade OCI Stream Pool with a private endpoint (VCN-only access), customer-managed KMS encryption, auto-create disabled, maximum 7-day retention, and three streams:..."
---

# Stream Pool Presets

Ready-to-deploy configuration presets for Stream Pool. Each preset is a complete manifest you can copy, customize, and deploy.
