---
title: "Enterprise Encrypted"
description: "Enterprise-grade AlloyDB cluster with CMEK on cluster data, automated backups, and continuous backups; query insights enabled; require_connectors true; ENCRYPTED_ONLY SSL; 30-day automated backup..."
type: "preset"
rank: "03"
presetSlug: "03-enterprise-encrypted"
componentSlug: "alloydb-cluster"
componentTitle: "AlloyDB Cluster"
provider: "gcp"
icon: "package"
order: 3
---

# Enterprise Encrypted

Enterprise-grade AlloyDB cluster with CMEK on cluster data, automated backups, and continuous backups; query insights enabled; require_connectors true; ENCRYPTED_ONLY SSL; 30-day automated backup retention; 21-day PITR window; and a maintenance window. Uses 8 CPUs and REGIONAL deployment.

## Description

This preset provisions an enterprise-grade AlloyDB cluster for mission-critical workloads. It applies customer-managed encryption (CMEK) to cluster data, automated backups, and continuous backups. Query insights are enabled for performance monitoring, and connector-only access enforces IAM-based authentication via AlloyDB Auth Proxy or Language Connectors.

## Use Case

- Mission-critical applications requiring the highest durability and security
- Workloads subject to compliance or governance requirements (HIPAA, PCI-DSS, FedRAMP)
- Environments that mandate customer-managed encryption keys (CMEK)
- Organizations requiring IAM-based authentication via AlloyDB Auth Proxy or Language Connectors
- Production databases needing query performance monitoring and extended backup retention

## What This Preset Configures

- **8 CPUs** — Higher production capacity for demanding workloads
- **REGIONAL availability** — Multi-zone deployment with automatic failover
- **deletion_protection: true** — Prevents accidental cluster destruction
- **kms_key_name** — CMEK for cluster data at rest
- **initial_user** — Creates a `postgres` superuser for database access
- **automated_backup_policy** — Enabled with 30-day time-based retention (2592000 seconds), CMEK for backup encryption
- **continuous_backup_config** — Enabled with 21-day PITR window, CMEK for continuous backup encryption
- **maintenance_window** — Sunday 3:00 UTC for system updates
- **ssl_mode: ENCRYPTED_ONLY** — All client connections must use TLS
- **require_connectors: true** — Direct IP connections rejected; AlloyDB Auth Proxy or Language Connectors required (IAM-based auth)
- **query_insights_config** — 5 query plans per minute, 1024-char query string length, application tags and client address recorded

## When to Use It

Use this preset when you need enterprise-grade security, compliance controls, and extended backup retention. Ensure KMS keys exist in the same region as the cluster and that the AlloyDB service account has `cloudkms.cryptoKeyEncrypterDecrypter` on each key. With `require_connectors: true`, applications must use AlloyDB Auth Proxy or AlloyDB Language Connectors; direct IP connections are rejected.
