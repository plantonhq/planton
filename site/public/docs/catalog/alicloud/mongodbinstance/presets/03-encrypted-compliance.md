---
title: "Preset: Encrypted Compliance MongoDB Instance"
description: "A security-hardened MongoDB instance with TDE encryption, SSL, subscription billing, and daily backups for compliance-sensitive workloads."
type: "preset"
rank: "03"
presetSlug: "03-encrypted-compliance"
componentSlug: "mongodbinstance"
componentTitle: "MongodbInstance"
provider: "alicloud"
icon: "package"
order: 3
---

# Preset: Encrypted Compliance MongoDB Instance

A security-hardened MongoDB instance with TDE encryption, SSL, subscription billing, and daily backups for compliance-sensitive workloads.

## Use Case

- Financial services or healthcare data requiring encryption at rest
- PCI-DSS, SOC 2, or HIPAA compliance requirements
- Long-term production deployments with predictable costs

## Configuration

- **Engine**: MongoDB 7.0 with WiredTiger
- **Instance Class**: `mongo.x8.xlarge` (high-performance tier)
- **Storage**: 500 GB on cloud ESSD PL3 (highest IOPS tier)
- **Replication**: 3-node replica set across three AZs
- **Encryption**: TDE enabled with customer-managed KMS key
- **SSL**: Enabled for all client connections
- **Billing**: PrePaid (12-month subscription with 3-month auto-renewal)
- **Backup**: Daily at 02:00-03:00 UTC (all 7 days)
- **Protection**: Release protection enabled
- **Network**: IP whitelist restricted to private RFC 1918 range

## Security Features

- Transparent Data Encryption (TDE) encrypts data at rest
- Customer-managed KMS key for encryption key control
- SSL/TLS for all client-to-server connections
- Restricted IP whitelist (172.16.0.0/12 only)
- Release protection prevents accidental deletion
