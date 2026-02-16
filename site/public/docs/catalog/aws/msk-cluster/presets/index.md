---
title: "Presets"
description: "Ready-to-deploy configuration presets for MSK Cluster"
type: "preset-list"
componentSlug: "msk-cluster"
componentTitle: "MSK Cluster"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-basic-kafka"
    rank: "01"
    title: "Preset: Basic Kafka Cluster"
    excerpt: "A minimal 3-broker MSK cluster suitable for development and testing workloads."
  - slug: "02-production-encrypted"
    rank: "02"
    title: "Preset: Production Encrypted Kafka Cluster"
    excerpt: "A production-grade MSK cluster with customer-managed KMS encryption, tiered storage, comprehensive monitoring, and hardened Kafka server properties."
  - slug: "03-multi-auth-logging"
    rank: "03"
    title: "Preset: Multi-Authentication with Full Logging"
    excerpt: "An MSK cluster demonstrating all three authentication methods and all three log destinations enabled simultaneously. Useful for organizations with diverse client populations and comprehensive audit..."
---

# MSK Cluster Presets

Ready-to-deploy configuration presets for MSK Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
