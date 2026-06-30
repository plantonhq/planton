# AwsRedisElasticache Resource Kind

**Date**: February 15, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi Module, Terraform Module, Provider Framework

## Summary

Added AwsRedisElasticache (enum 250) — a managed Redis/Valkey in-memory data store using AWS ElastiCache replication groups. Supports both non-clustered (single primary + read replicas) and clustered (sharded) topologies with encryption, authentication, logging, and full infra-chart composability. This is the seventh new AWS resource kind in the cloud provider expansion project (R07).

## Problem Statement / Motivation

ElastiCache is the most common managed caching layer on AWS, used for session stores, application caches, real-time leaderboards, and message brokers. The Planton catalog had no ElastiCache coverage, forcing users to manage Redis infrastructure outside the declarative framework.

### Pain Points

- No managed caching resource in Planton's AWS provider catalog
- Users had to provision ElastiCache manually, losing dependency wiring and infra-chart composability
- The original planning (T02) designed a single `AwsElasticacheCluster` component, but deep research revealed that Redis/Valkey and Memcached use completely different Terraform resources with ~15 unique fields each

## Solution / What's New

### Component Split Decision

Deep research into the Terraform provider revealed that the original `AwsElasticacheCluster` design needed to be split into three focused components:

- **AwsRedisElasticache** (R07, this session) — Redis/Valkey via `aws_elasticache_replication_group`
- **AwsMemcachedElasticache** (R07a, next session) — Memcached via `aws_elasticache_cluster`
- **AwsServerlessElasticache** (R07b, future session) — Serverless via `aws_elasticache_serverless_cache`

The delta between these three is massive: different TF resources, different topology models, different persistence, different authentication. Three focused components provide better clarity than one overloaded component.

### AwsRedisElasticache Component

**Proto API**: spec.proto with 29 fields, 3 nested messages, 12 CEL validations

- Two topology modes: non-clustered (`num_cache_clusters` 1-6) and clustered (`num_node_groups` + `replicas_per_node_group`)
- StringValueOrRef for `subnet_ids` (→ AwsVpc), `security_group_ids` (→ AwsSecurityGroup), `kms_key_id` (→ AwsKmsKey), `notification_topic_arn` (→ AwsSnsTopic), `auth_token`, and log `destination`
- Encryption: at-rest + in-transit with TLS mode control (preferred/required)
- Authentication: AUTH token vs Redis ACL user groups (mutually exclusive)
- Logging: up to 2 log delivery configs (slow-log, engine-log) to CloudWatch or Kinesis Firehose
- Bundled subnet group and parameter group creation (following AwsRdsCluster pattern)

**Validation tests**: 39 spec tests, all passing

- 14 happy path (minimal, HA, clustered, encryption, auth, parameters, logging, production-ready, data tiering)
- 25 failure scenarios (missing required fields, invalid engine, topology conflicts, range violations, mutual exclusions)

**Pulumi module**: 6 files (main.go, locals.go, outputs.go, subnet_group.go, parameter_group.go, replication_group.go)

**Terraform module**: 5 files (main.tf, locals.tf, outputs.tf, variables.tf, provider.tf) with feature parity

**Documentation**: README.md, examples.md, docs/README.md (architecture deep-dive), catalog page

**Presets**: 3 presets (01-redis-single-node, 02-redis-ha-cluster, 03-redis-clustered-production)

## Implementation Details

### 10 Surprise Findings

During deep research into the Terraform/Pulumi providers, 10 capabilities were discovered that were not in the T02 planning guidance:

1. **ElastiCache Serverless** — separate resource, fundamentally different config → separate component
2. **Cluster Mode Enabled** (sharding) — not in T02, but used by ~40-50% of production deployments
3. **Log delivery configuration** — slow-log and engine-log to CloudWatch/Firehose
4. **user_group_ids** (Redis ACL) — fine-grained auth alternative to auth_token
5. **Data tiering** — r6gd node types with auto SSD offload
6. **Transit encryption mode** — preferred vs required for gradual TLS migration
7. **Notification topic ARN** — SNS alerts for cluster events
8. **Auto minor version upgrade** — maintenance automation
9. **Global replication group** — cross-region (excluded for v1, <5% usage)
10. **Per-shard configuration** — advanced AZ/slot tuning (excluded for v1)

### Key Design Decisions

- **No explicit `cluster_mode` field** — mode auto-inferred from which topology fields are set
- **No convenience security group creation** — simpler than AwsRdsCluster; users create SGs separately
- **`parameter_group_family` is user-specified** — more explicit than auto-derivation from engine/version
- **`auth_token` as StringValueOrRef** — enables referencing secrets from AwsSecretsManager

## Benefits

- Complete managed caching coverage for Redis/Valkey workloads in Planton
- Full infra-chart composability: accepts VPC, SG, KMS, SNS references; exports endpoints, ARN, port
- Production-ready with encryption, HA, multi-AZ, snapshots, logging, and custom parameters
- Clear component boundaries: Redis/Valkey, Memcached, and Serverless are separate, focused components

## Impact

- **End users**: Can now deploy Redis/Valkey caches declaratively with `planton apply`
- **Infra chart authors**: Can compose Redis into microservices, serverless-api, and data-pipeline charts
- **Platform**: 7 of ~32 new AWS resource kinds completed (Phase 1 progress: 47%)

## Related Work

- Previous: R06 AwsStepFunction (2026-02-15)
- Next: R07a AwsMemcachedElasticache, R07b AwsServerlessElasticache
- Parent project: 20260212.01.planton-cloud-provider-expansion

---

**Status**: Production Ready
**Timeline**: Single session (~2 hours)
