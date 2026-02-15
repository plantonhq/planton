# Serverless ElastiCache

Provisions an AWS ElastiCache Serverless cache — a fully managed, auto-scaling
in-memory data store with zero node management.

## Overview

ElastiCache Serverless removes all infrastructure decisions. You choose an engine
(Redis, Valkey, or Memcached), set optional scaling bounds, and AWS handles
capacity, replication, patching, and failover automatically.

## Key Features

- **Three engines**: Redis, Valkey (open-source Redis-compatible), and Memcached
- **Auto-scaling**: Compute (ECPU) and storage (GB) scale automatically within configurable bounds
- **Always encrypted**: Data in transit and at rest is always encrypted (customer-managed KMS key optional)
- **VPC support**: Deploy into private subnets with security group controls
- **Snapshots**: Daily automatic snapshots with configurable retention (Redis/Valkey only)
- **Access control**: Redis ACL user groups for fine-grained authentication (Redis/Valkey only)

## When to Use

- Variable or unpredictable traffic patterns
- Zero infrastructure management desired
- Development, staging, or production environments
- Pay-per-use billing preferred over reserved capacity

## Comparison with Provisioned Components

| Feature | Serverless | Provisioned Redis | Provisioned Memcached |
|---|---|---|---|
| Node management | None | Full control | Full control |
| Cost model | Pay-per-use | Per node-hour | Per node-hour |
| Parameter tuning | Not available | Custom parameter groups | Custom parameter groups |
| Global replication | Not available | Available | N/A |
| Data tiering | Not available | Available (r6gd) | N/A |

## Quick Start

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsServerlessElasticache
metadata:
  name: my-cache
  org: my-org
  env: dev
  id: my-cache-dev
spec:
  engine: redis
```

## Resources

- [Spec Reference](../../../../apis/org/openmcf/provider/aws/awsserverlesselasticache/v1/README.md)
- [Examples](../../../../apis/org/openmcf/provider/aws/awsserverlesselasticache/v1/examples.md)
- [Architecture](../../../../apis/org/openmcf/provider/aws/awsserverlesselasticache/v1/docs/README.md)
