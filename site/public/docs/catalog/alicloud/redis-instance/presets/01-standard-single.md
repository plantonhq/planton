---
title: "Standard Single-Zone Redis"
description: "This preset creates a minimal Redis 7.0 instance for development and testing, suitable for single-zone deployments with basic configuration."
type: "preset"
rank: "01"
presetSlug: "01-standard-single"
componentSlug: "redis-instance"
componentTitle: "Redis Instance"
provider: "alicloud"
icon: "package"
order: 1
---

# Standard Single-Zone Redis

This preset creates a minimal Redis 7.0 instance for development and testing, suitable for single-zone deployments with basic configuration.

## When to Use

- Development and testing environments
- Proof-of-concept deployments
- Learning and experimentation with Alibaba Cloud Redis
- Environments where high availability is not required

## Key Configuration Choices

- **redis.master.small.default** -- smallest standard instance class; upgrade for production use
- **Redis 7.0** -- latest stable version with full feature support
- **Single zone** -- no secondary_zone_id for cost savings
- **PostPaid billing** (default) -- pay-as-you-go, no commitment
- **VPC password authentication** -- requires password for all connections

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`) | Your deployment region |
| `<your-vswitch-id>` | VSwitch ID to place the instance in | `AliCloudVswitch` stack outputs |
| `<your-instance-name>` | Instance name (2-256 chars) | Choose a descriptive name |
| `<your-password>` | Instance password (8-32 chars, mixed complexity) | Use a secrets manager |
| `<your-application-cidr>` | Application CIDR for IP whitelist (e.g., `10.0.0.0/8`) | Your VPC CIDR range |

## Related Presets

- **02-ha-cluster** -- Use for production workloads with HA and cluster sharding
- **03-production-encrypted** -- Use for security-sensitive environments with SSL and TDE
