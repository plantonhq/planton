---
title: "General Purpose Regional EFS"
description: "Regional, encrypted, bursting throughput, backup enabled, no access points. Simplest production-safe starting point."
type: "preset"
rank: "01"
presetSlug: "01-general-purpose-regional"
componentSlug: "elastic-file-system"
componentTitle: "Elastic File System"
provider: "aws"
icon: "package"
order: 1
---

# General Purpose Regional EFS

Regional, encrypted, bursting throughput, backup enabled, no access points. Simplest production-safe starting point.

## When to Use

- EKS pods needing shared persistent storage via the EFS CSI driver
- EC2 instances or ECS tasks that mount EFS directly
- Workloads with predictable or moderate I/O patterns (bursting scales with storage size)
- First EFS deployment when you want minimal configuration

## What It Configures

- **Regional** — No `availabilityZoneName`; file system spans multiple AZs for high availability
- **Encrypted** — AES-256 encryption at rest using AWS-managed key
- **Bursting throughput** — Throughput scales with storage; 50 MiB/s per TiB with bursts up to 100 MiB/s
- **Backup enabled** — Daily backups via AWS Backup
- **No access points** — Mount the root; add access points later if you need per-application isolation

## What to Customize

- Replace placeholders: `<subnet-id-az-a>`, `<subnet-id-az-b>`, `<security-group-id>`
- Add more subnets (one per AZ) for broader availability
- Switch to `throughputMode: elastic` for unpredictable or spiky workloads
- Add access points if using ECS tasks or Lambda with per-app root directories
