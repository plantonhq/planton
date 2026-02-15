---
title: "Preset: Dev Basic"
description: "**Tier**: BASIC_SSD (single-zone SSD) **Use case**: Development, testing, CI/CD pipelines"
type: "preset"
rank: "01"
presetSlug: "01-dev-basic"
componentSlug: "filestore-instance"
componentTitle: "Filestore Instance"
provider: "gcp"
icon: "package"
order: 1
---

# Preset: Dev Basic

**Tier**: BASIC_SSD (single-zone SSD)
**Use case**: Development, testing, CI/CD pipelines

## What This Preset Provides

A minimal Filestore instance suitable for development and testing workloads:

- **BASIC_SSD tier**: SSD-backed storage with good read/write performance
- **2.5 TiB capacity**: minimum for BASIC_SSD tier
- **Default VPC**: connects to the project's default network
- **DIRECT_PEERING**: simplest network setup (default)
- **No encryption**: uses Google-managed keys
- **No deletion protection**: easy cleanup in dev environments

## When to Use

- Local development with NFS-dependent applications
- CI/CD pipelines requiring shared file storage
- Testing NFS-based workflows before moving to production
- Small teams needing shared file access

## When NOT to Use

- Production workloads requiring HA (use ENTERPRISE or REGIONAL tier)
- Workloads requiring CMEK encryption
- Applications needing >2.5 TiB of storage (increase capacity_gb)
