---
title: "Preset: Spot Batch Worker"
description: "A cost-efficient spot instance for batch processing workloads that can tolerate interruption."
type: "preset"
rank: "03"
presetSlug: "03-spot-batch-worker"
componentSlug: "ecsinstance"
componentTitle: "EcsInstance"
provider: "alicloud"
icon: "package"
order: 3
---

# Preset: Spot Batch Worker

A cost-efficient spot instance for batch processing workloads that can tolerate interruption.

## Use Case

- Batch data processing and ETL jobs
- CI/CD build agents
- Machine learning training workloads
- Any workload that is fault-tolerant and can be restarted

## Configuration

- **Instance Type**: `ecs.c7.2xlarge` (8 vCPU, 16 GiB memory, compute-optimized)
- **Image**: Ubuntu 22.04 LTS
- **System Disk**: cloud_essd, 40 GB
- **Data Disk**: cloud_efficiency, 500 GB scratch space (deleted with instance)
- **Authentication**: SSH key pair
- **Billing**: SpotAsPriceGo (pay current market price, up to 90% discount)
- **Public IP**: None

## What's Not Included

- Disk encryption (scratch data is ephemeral)
- Deletion protection (spot instances may be reclaimed)
- RAM role (add if the worker needs to access Alibaba Cloud services)
- SpotWithPriceLimit (uses SpotAsPriceGo for simplicity; change to SpotWithPriceLimit with spotPriceLimit if you need a price cap)
