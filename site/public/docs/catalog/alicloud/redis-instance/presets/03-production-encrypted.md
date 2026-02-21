---
title: "Production Encrypted Redis"
description: "This preset creates a security-hardened Redis instance with TDE encryption at rest, SSL encryption in transit, subscription billing, and daily backups -- designed for compliance-sensitive..."
type: "preset"
rank: "03"
presetSlug: "03-production-encrypted"
componentSlug: "redis-instance"
componentTitle: "Redis Instance"
provider: "alicloud"
icon: "package"
order: 3
---

# Production Encrypted Redis

This preset creates a security-hardened Redis instance with TDE encryption at rest, SSL encryption in transit, subscription billing, and daily backups -- designed for compliance-sensitive environments.

## When to Use

- Environments subject to compliance requirements (PCI-DSS, HIPAA, SOC 2)
- Applications handling sensitive or confidential data
- Organizations requiring encryption at rest and in transit
- Long-running production deployments benefiting from subscription pricing

## Key Configuration Choices

- **redis.master.large.default** -- production-grade instance class; adjust for your memory needs
- **TDE (Transparent Data Encryption)** -- encrypts data at rest at the storage level; once enabled, cannot be disabled
- **SSL encryption** -- encrypts all client-server traffic in transit
- **Customer-managed KMS key** -- uses your own encryption key for TDE
- **PrePaid (12 months)** -- subscription billing with auto-renewal for cost optimization
- **Daily backups** -- every day at 02:00-03:00 UTC for maximum data protection
- **Cross-AZ deployment** -- primary and standby in different availability zones
- **Deletion protection** -- prevents accidental deletion

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region code (e.g., `cn-hangzhou`) | Your deployment region |
| `<your-vswitch-resource-name>` | VSwitch resource name for `valueFrom` reference | Your AliCloudVswitch resource |
| `<your-instance-name>` | Instance name | Choose a descriptive name |
| `<your-organization>` | Organization identifier | Your org slug |
| `<primary-zone-id>` | Primary AZ (e.g., `cn-hangzhou-a`) | Available AZs in your region |
| `<standby-zone-id>` | Standby AZ (e.g., `cn-hangzhou-b`) | A different AZ from primary |
| `<your-password>` | Instance password (8-32 chars) | Use a secrets manager |
| `<your-kms-key-id>` | KMS key ID for TDE encryption | `AliCloudKmsKey` stack outputs |
| `<your-application-cidr>` | Application CIDR (e.g., `172.16.0.0/12`) | Your VPC CIDR range |
| `<your-compliance-standard>` | Compliance standard tag (e.g., `pci-dss`) | Your compliance requirement |

## Related Presets

- **01-standard-single** -- Use for development and testing (no encryption)
- **02-ha-cluster** -- Use for high-throughput production without encryption requirements
