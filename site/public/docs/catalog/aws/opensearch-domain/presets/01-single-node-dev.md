---
title: "Single-Node Development Domain"
description: "This preset creates a minimal single-node OpenSearch domain suitable for development, prototyping, and learning. The domain is publicly accessible (no VPC) with encryption enabled for security best..."
type: "preset"
rank: "01"
presetSlug: "01-single-node-dev"
componentSlug: "opensearch-domain"
componentTitle: "OpenSearch Domain"
provider: "aws"
icon: "package"
order: 1
---

# Single-Node Development Domain

This preset creates a minimal single-node OpenSearch domain suitable for development, prototyping, and learning. The domain is publicly accessible (no VPC) with encryption enabled for security best practices even in non-production environments.

## When to Use

- Local development and integration testing against a real OpenSearch endpoint
- Prototyping search features, query patterns, or index mappings
- Learning OpenSearch without the cost of a multi-node cluster
- CI/CD environments that need a disposable search backend

## Key Configuration Choices

- **OpenSearch 2.11** (`engineVersion: "OpenSearch_2.11"`) — Latest stable version; update to match your production version
- **t3.small.search** (`instanceType`) — Burstable instance for low-traffic development; provides 2 vCPUs and 2 GiB RAM
- **10 GB gp3 storage** (`volumeSize: 10`) — Minimal storage; increase if indexing significant test data
- **Encryption enabled** (`encryptAtRestEnabled`, `nodeToNodeEncryptionEnabled`) — Security best practices even for dev; matches production configuration to avoid surprises
- **No VPC** — Publicly accessible for easy connectivity from developer machines; secure with access policies or FGAC in real usage
- **No dedicated masters** — Single-node clusters don't benefit from dedicated masters
- **No FGAC** — Add `advancedSecurityOptions` if testing role-based access

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `my-search` | Domain name (3-28 chars, lowercase, hyphens) | Your naming convention |

## Related Presets

- **02-production-vpc** — Use for production workloads with VPC, dedicated masters, and FGAC
- **03-analytics-warm-cold** — Use for analytics workloads with tiered storage
