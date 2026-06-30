# AWS Serverless ElastiCache Resource Kind

**Date**: February 15, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi Module, Terraform Module, Documentation

## Summary

Added AwsServerlessElasticache (R07b) -- a deployment component for AWS ElastiCache Serverless caches supporting all three engines (Redis, Valkey, Memcached). This is the third and final ElastiCache variant, completing the family alongside AwsRedisElasticache (provisioned Redis/Valkey) and AwsMemcachedElasticache (provisioned Memcached).

## Problem Statement / Motivation

AWS ElastiCache Serverless is a fundamentally different deployment model from provisioned clusters. It uses a completely different Terraform resource (`aws_elasticache_serverless_cache`), a different billing model (pay-per-ECPU and per-GB), and removes all node management. The original AwsElasticacheCluster planning item was split into three focused components during R07 implementation when the deltas between provisioned Redis, provisioned Memcached, and serverless were found to be too large for a single component.

### Pain Points

- Users who want zero-management caching had no serverless option in Planton
- The ElastiCache Serverless resource uses completely different AWS APIs than provisioned clusters
- The ECPU-based scaling model requires different spec fields than node-based provisioned clusters

## Solution / What's New

### Multi-Engine Serverless Cache Component

A single deployment component supporting all three ElastiCache engines with engine-specific field guards via CEL validations. Unlike the provisioned siblings (which were split because they use different Terraform resources), serverless uses a single `aws_elasticache_serverless_cache` resource for all engines.

### Flattened Scaling Limits

The AWS resource nests scaling limits three levels deep (`cache_usage_limits > data_storage > {min, max, unit}` and `cache_usage_limits > ecpu_per_second > {min, max}`). The Planton spec flattens these to four top-level fields (`data_storage_min_gb`, `data_storage_max_gb`, `ecpu_min`, `ecpu_max`) for clean YAML authoring. The `unit` field (always "GB") is hardcoded in IaC modules.

## Implementation Details

### Proto API (13 fields, 11 CEL validations)

```
spec.proto
├── engine (required, CEL: redis/valkey/memcached)
├── major_engine_version
├── description
├── data_storage_max_gb (1-5000)
├── data_storage_min_gb (1-5000, CEL: min <= max)
├── ecpu_max (1000-15000000)
├── ecpu_min (1000-15000000, CEL: min <= max)
├── subnet_ids (StringValueOrRef -> AwsVpc)
├── security_group_ids (StringValueOrRef -> AwsSecurityGroup)
├── kms_key_id (StringValueOrRef -> AwsKmsKey)
├── daily_snapshot_time (CEL: redis/valkey only)
├── snapshot_retention_limit (CEL: redis/valkey only)
└── user_group_id (CEL: redis/valkey only)
```

### Pulumi Module (4 files)

The leanest ElastiCache module -- no subnet groups, no parameter groups:

- `main.go` -- Provider setup, orchestration
- `locals.go` -- Tags from metadata
- `outputs.go` -- Output constants
- `serverless_cache.go` -- Single `elasticache.ServerlessCache` resource with conditional `CacheUsageLimits` block construction

### Terraform Module (5 files)

Uses dynamic blocks for `cache_usage_limits`, `data_storage`, and `ecpu_per_second` to conditionally include scaling limits only when specified.

### Validation Tests (36 tests)

- 20 happy path: minimal Redis/Valkey/Memcached, with limits, VPC, KMS, snapshots, user group, production-ready
- 16 failure: missing engine, invalid engine, min > max ordering, Memcached with Redis-only fields, range violations, ECPU floor checks

## Benefits

- **Zero node management**: Users specify engine + optional limits, AWS handles everything else
- **Pay-per-use**: Ideal for variable workloads, dev/staging, and prototyping
- **Clean spec**: 13 fields vs Redis's 29 and Memcached's 15 -- embraces serverless simplicity
- **Infra chart ready**: StringValueOrRef on all cross-resource fields enables `valueFrom` composition

## Impact

### New Capabilities

- Serverless Redis, Valkey, and Memcached deployment via Planton
- ECPU-based auto-scaling within configurable bounds
- Completes the ElastiCache family (3 of 3 components)

### Files Changed

- 41 files, ~3,097 lines added
- New component: `apis/dev/planton/provider/aws/awsserverlesselasticache/v1/`
- Enum registration: `AwsServerlessElasticache = 253` in `cloud_resource_kind.proto`
- Catalog page: `site/public/docs/catalog/aws/serverless-elasticache.md`
- 3 presets: redis-minimal, memcached-with-limits, redis-production

## Related Work

- AwsRedisElasticache (R07) -- provisioned Redis/Valkey sibling
- AwsMemcachedElasticache (R07a) -- provisioned Memcached sibling
- Part of the AWS resource expansion sub-project (20260215.02.sp.aws-resource-expansion)

---

**Status**: Production Ready
