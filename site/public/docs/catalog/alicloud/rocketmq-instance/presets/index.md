---
title: "Presets"
description: "Ready-to-deploy configuration presets for RocketMQ Instance"
type: "preset-list"
componentSlug: "rocketmq-instance"
componentTitle: "RocketMQ Instance"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-development-single-node"
    rank: "01"
    title: "Development Single-Node RocketMQ"
    excerpt: "A minimal RocketMQ 5.x instance for development and testing. Uses the standard edition with a single-node deployment, which is the cheapest way to get a working message broker for local integration..."
  - slug: "02-production-ha"
    rank: "02"
    title: "Production HA RocketMQ"
    excerpt: "A production-grade RocketMQ 5.x instance using the professional edition with high-availability clustering. Includes example topics (NORMAL and FIFO) with matching consumer groups to demonstrate the..."
  - slug: "03-enterprise-encrypted"
    rank: "03"
    title: "Enterprise Encrypted RocketMQ"
    excerpt: "A mission-critical RocketMQ 5.x instance using the ultimate edition with encryption at rest, public internet access, and subscription billing. Designed for compliance-sensitive environments that..."
---

# RocketMQ Instance Presets

Ready-to-deploy configuration presets for RocketMQ Instance. Each preset is a complete manifest you can copy, customize, and deploy.
