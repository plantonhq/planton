---
title: "Presets"
description: "Ready-to-deploy configuration presets for KMS Key Ring"
type: "preset-list"
componentSlug: "kms-key-ring"
componentTitle: "KMS Key Ring"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-regional-key-ring"
    rank: "01"
    title: "Regional Key Ring"
    excerpt: "This preset creates a KMS key ring in a specific GCP region. It is the most common configuration — co-locating encryption keys with the workloads they protect for lowest latency and data residency..."
  - slug: "02-global-key-ring"
    rank: "02"
    title: "Global Key Ring"
    excerpt: "This preset creates a KMS key ring in the `global` location, making encryption keys accessible from any GCP region without latency penalties associated with cross-region access."
  - slug: "03-multi-region-key-ring"
    rank: "03"
    title: "Multi-Region Key Ring"
    excerpt: "This preset creates a KMS key ring in a multi-region location (e.g., `us`, `europe`, `asia`), providing high availability with automatic replication across all regions within the specified geography..."
---

# KMS Key Ring Presets

Ready-to-deploy configuration presets for KMS Key Ring. Each preset is a complete manifest you can copy, customize, and deploy.
