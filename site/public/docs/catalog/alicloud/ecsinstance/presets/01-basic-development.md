---
title: "Preset: Basic Development Instance"
description: "A minimal ECS instance for development and testing workloads."
type: "preset"
rank: "01"
presetSlug: "01-basic-development"
componentSlug: "ecsinstance"
componentTitle: "EcsInstance"
provider: "alicloud"
icon: "package"
order: 1
---

# Preset: Basic Development Instance

A minimal ECS instance for development and testing workloads.

## Use Case

- Local development and integration testing
- Non-production workloads with minimal cost
- Quick prototyping with a general-purpose instance

## Configuration

- **Instance Type**: `ecs.g7.large` (2 vCPU, 8 GiB memory)
- **Image**: Ubuntu 22.04 LTS
- **System Disk**: Default (cloud_essd, 40 GB)
- **Authentication**: SSH key pair
- **Billing**: PostPaid (pay-as-you-go, default)
- **Public IP**: None (no outbound bandwidth)

## What's Not Included

- Data disks
- Public IP / internet bandwidth
- Disk encryption
- Deletion protection
- Spot pricing
- RAM role
