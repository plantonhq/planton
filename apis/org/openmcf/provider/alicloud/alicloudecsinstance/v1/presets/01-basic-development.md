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
