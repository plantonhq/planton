---
title: "Preset: Production Enterprise"
description: "**Tier**: ENTERPRISE (regional HA) **Use case**: Production workloads requiring high availability and security"
type: "preset"
rank: "02"
presetSlug: "02-production-enterprise"
componentSlug: "filestore-instance"
componentTitle: "Filestore Instance"
provider: "gcp"
icon: "package"
order: 2
---

# Preset: Production Enterprise

**Tier**: ENTERPRISE (regional HA)
**Use case**: Production workloads requiring high availability and security

## What This Preset Provides

A production-grade Filestore instance with enterprise features:

- **ENTERPRISE tier**: regional high availability with automatic failover
- **1 TiB capacity**: configurable starting point, increase as needed
- **PRIVATE_SERVICE_ACCESS**: secure network connectivity via private services
- **ROOT_SQUASH**: root users on clients mapped to anonymous UID for security
- **Deletion protection**: prevents accidental destruction
- **IP range restriction**: only 10.0.0.0/8 RFC1918 addresses can mount

## When to Use

- Production applications requiring shared NFS storage
- Workloads that cannot tolerate zone-level outages
- Environments with strict security requirements
- GKE clusters needing persistent shared volumes

## When NOT to Use

- Development or testing (use dev-basic preset)
- Cost-sensitive workloads (ENTERPRISE is the most expensive tier)
- Workloads requiring < 1 TiB (consider GCS instead)
