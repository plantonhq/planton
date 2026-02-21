# AliCloudRedisInstance

Manages an Alibaba Cloud Redis (KVStore) instance for managed in-memory caching and data storage.

## Overview

Redis is Alibaba Cloud's managed in-memory key-value store, used for caching, session management, real-time analytics, and message brokering. The KVStore service supports both Redis and Memcache engines through a single resource type, with Redis being the overwhelmingly dominant use case.

### What Gets Created

- **KVStore Instance** -- a managed Redis (or Memcache) instance with the selected engine version, instance class, and network configuration

### Architecture Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| Standard (single shard) | Master-replica pair | Development, small workloads |
| Cluster (multi-shard) | Multiple data shards with `shardCount` | High-throughput production |
| Read replicas | Additional read-only nodes with `readOnlyCount` | Read-heavy workloads |

## Build and Test (Localized)

All build and test commands are scoped to this component directory. Never run project-wide `make build`.

```bash
# Go build (Pulumi module)
go build ./apis/org/openmcf/provider/alicloud/alicloudredisinstance/v1/iac/pulumi/...

# Go vet
go vet ./apis/org/openmcf/provider/alicloud/alicloudredisinstance/v1/iac/pulumi/...

# Spec tests
go test ./apis/org/openmcf/provider/alicloud/alicloudredisinstance/v1/...

# Terraform validation
cd apis/org/openmcf/provider/alicloud/alicloudredisinstance/v1/iac/tf
terraform init -backend=false
terraform validate
```

## Configuration Reference

See [catalog-page.md](catalog-page.md) for complete field documentation, examples, and presets.
