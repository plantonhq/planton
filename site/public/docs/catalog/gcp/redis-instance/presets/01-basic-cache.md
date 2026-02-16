---
title: "Basic Cache"
description: "This preset provisions a minimal Memorystore for Redis instance using the BASIC tier with 1 GB memory. It is ideal for development, testing, or lightweight caching workloads where high availability..."
type: "preset"
rank: "01"
presetSlug: "01-basic-cache"
componentSlug: "redis-instance"
componentTitle: "Redis Instance"
provider: "gcp"
icon: "package"
order: 1
---

# Basic Cache

This preset provisions a minimal Memorystore for Redis instance using the BASIC tier with 1 GB memory. It is ideal for development, testing, or lightweight caching workloads where high availability and authentication are not required.

## When to Use

- Local development and integration testing
- CI/CD pipelines that need a temporary Redis instance
- Lightweight application caching with minimal cost
- Proof-of-concept or prototyping environments

## Key Configuration

- **BASIC tier** — single-node instance with no replication or SLA
- **1 GB memory** — smallest available size; increase for larger datasets
- **Minimal fields** — only project, instance name, region, tier, and memory; no auth, TLS, or VPC attachment

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the instance will be created | GCP Console or `GcpProject` outputs |
| `<instance-name>` | Name for this Redis instance (2-40 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `dev-cache`) |
| `<gcp-region>` | GCP region for the instance (e.g., `us-central1`) | [GCP regions](https://cloud.google.com/about/locations) |

## Related Presets

- **02-ha-production** — STANDARD_HA with auth, TLS, persistence, and deletion protection
- **03-ha-read-replicas** — STANDARD_HA with read replicas for read-heavy workloads
