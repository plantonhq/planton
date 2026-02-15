---
title: "Preset: Redis Production"
description: "**Use case:** Production workloads requiring encryption, VPC isolation, daily snapshots, and Redis ACL access control."
type: "preset"
rank: "03"
presetSlug: "03-redis-production"
componentSlug: "serverless-elasticache"
componentTitle: "Serverless ElastiCache"
provider: "aws"
icon: "package"
order: 3
---

# Preset: Redis Production

**Use case:** Production workloads requiring encryption, VPC isolation, daily
snapshots, and Redis ACL access control.

**What it creates:**
- A Redis 7.x serverless cache with all production features enabled
- Customer-managed KMS encryption (via `valueFrom` reference to AwsKmsKey)
- VPC placement across 3 private subnets (via `valueFrom` references to AwsVpc)
- Security group attachment (via `valueFrom` reference to AwsSecurityGroup)
- Daily snapshots at 03:00 UTC with 14-day retention
- Redis ACL user group for fine-grained access control
- Scaling: 5–200 GB data, 5,000–500,000 ECPU/sec

**Cost profile:** Higher baseline due to minimum scaling bounds, but appropriate for
production workloads with predictable traffic.

**Prerequisites:**
- An AwsVpc with at least 3 private subnets
- An AwsSecurityGroup allowing inbound traffic on the Redis port (6379)
- An AwsKmsKey for customer-managed encryption
- A Redis ACL user group (created via AWS CLI or console)

**Infra chart integration:** This preset is designed for composition in infra charts.
All cross-resource fields use `valueFrom` references that the platform resolves
into a dependency DAG.
