---
title: "Presets"
description: "Ready-to-deploy configuration presets for EcsInstance"
type: "preset-list"
componentSlug: "ecsinstance"
componentTitle: "EcsInstance"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-basic-development"
    rank: "01"
    title: "Preset: Basic Development Instance"
    excerpt: "A minimal ECS instance for development and testing workloads."
  - slug: "02-production-web-server"
    rank: "02"
    title: "Preset: Production Web Server"
    excerpt: "A production-grade ECS instance with encrypted disks, public IP, deletion protection, and a RAM role."
  - slug: "03-spot-batch-worker"
    rank: "03"
    title: "Preset: Spot Batch Worker"
    excerpt: "A cost-efficient spot instance for batch processing workloads that can tolerate interruption."
---

# EcsInstance Presets

Ready-to-deploy configuration presets for EcsInstance. Each preset is a complete manifest you can copy, customize, and deploy.
