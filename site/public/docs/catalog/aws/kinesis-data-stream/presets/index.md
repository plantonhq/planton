---
title: "Presets"
description: "Ready-to-deploy configuration presets for Kinesis Data Stream"
type: "preset-list"
componentSlug: "kinesis-data-stream"
componentTitle: "Kinesis Data Stream"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-on-demand-minimal"
    rank: "01"
    title: "Preset: On-Demand Minimal"
    excerpt: "The simplest possible Kinesis stream for development, prototyping, or variable-throughput workloads. AWS manages all capacity automatically."
  - slug: "02-provisioned-encrypted"
    rank: "02"
    title: "Preset: Provisioned Encrypted"
    excerpt: "A provisioned stream with predictable capacity, KMS encryption using the Kinesis-owned key, and 48-hour retention for basic reprocessing."
  - slug: "03-production-analytics"
    rank: "03"
    title: "Preset: Production Analytics"
    excerpt: "Full-featured production stream for analytics pipelines, event sourcing, and high-reliability data ingestion. ON_DEMAND for zero capacity planning with comprehensive monitoring."
---

# Kinesis Data Stream Presets

Ready-to-deploy configuration presets for Kinesis Data Stream. Each preset is a complete manifest you can copy, customize, and deploy.
