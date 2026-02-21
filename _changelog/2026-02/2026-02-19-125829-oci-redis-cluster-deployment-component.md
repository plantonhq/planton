# OCI Redis Cluster Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Terraform Module, Protobuf Schemas

## Summary

Implemented the OciRedisCluster deployment component (CloudResourceKind 3334) -- OCI's fully managed, Redis-compatible in-memory caching service supporting both sharded and non-sharded cluster topologies. This is the fifth database resource in the OCI provider (following OciAutonomousDatabase, OciDbSystem, OciMysqlDbSystem, and OciPostgresqlDbSystem), completing the Redis/cache portion of Phase 4.

## Problem Statement / Motivation

OCI Cache (formerly OCI Redis) provides a managed Redis-compatible caching layer for low-latency data access. Platform teams need to provision cache clusters with configurable node counts, memory sizing, and cluster topology (sharded for horizontal scaling, non-sharded for simplicity).

### Pain Points

- Teams need managed Redis-compatible caching without the operational overhead of self-hosted Redis
- Sharded clusters are needed for high-throughput workloads that exceed single-node memory capacity
- Non-sharded clusters with replicas are needed for read-heavy workloads with simple key-space requirements
- Network security groups must be configurable for fine-grained access control
- Custom Redis configuration parameters (maxmemory-policy, timeout, etc.) require Config Set references

## Solution / What's New

A complete deployment component with:

1. **Proto API** -- `spec.proto` with 10 fields, 1 embedded enum (ClusterMode), 1 CEL conditional validation rule
2. **Validation Tests** -- 22 Ginkgo/Gomega tests (13 valid, 9 invalid scenarios), all passing
3. **Pulumi Module** -- `redis.NewRedisCluster()` with conditional field assignment for cluster_mode, shard_count, nsg_ids, and config_set_id across 4 Go files. Five endpoint outputs exported directly from the cluster resource.
4. **Terraform Module** -- `oci_redis_redis_cluster.this` with enum map for cluster_mode, conditional null handling for optional fields. Clean single-resource module with no dynamic blocks needed.
5. **Kind Registration** -- OciRedisCluster=3334 under "Databases" section, kind_map_gen.go regenerated

### Design Decision: Single-Resource Component

Unlike load balancers or DRGs that bundle multiple sub-resources, OCI Cache clusters are a single resource with all configuration inline. Config Sets are separate reusable resources with independent lifecycles and are referenced by OCID rather than bundled.

### Design Decision: ClusterMode as Enum

The `cluster_mode` field uses an embedded proto enum (nonsharded/sharded) rather than a plain string. This provides schema-level documentation of valid values and enables the CEL validation rule that enforces `shard_count > 0` when mode is sharded.

### Design Decision: Float for Node Memory

Node memory uses proto `float` (32-bit) consistent with the pattern established in R09 OciContainerEngineNodePool for OCPUs and memory. Memory values are simple round numbers (2, 4, 8, 16, 32, 64 GB) that are exactly representable in float32.

## Implementation Details

### Five Endpoint Outputs

Redis clusters expose multiple endpoints depending on topology:
- `primary_fqdn` / `primary_endpoint_ip_address` -- primary read-write node (main connection for non-sharded)
- `replicas_fqdn` -- read-replica endpoint for read scaling
- `discovery_fqdn` -- shard discovery endpoint (main connection for sharded clusters)

All five are exported as stack outputs. Some may be empty depending on cluster mode, but exposing all avoids the need to re-deploy when switching topologies.

### CEL Validation

One rule enforces that `shard_count` must be greater than zero when `cluster_mode` is `sharded`. Non-sharded clusters ignore `shard_count` (OCI defaults apply).

### Excluded from v1

- `defined_tags`, `system_tags` -- managed by platform via freeform_tags
- `security_attributes` -- specialized Oracle Zero-Trust Packet Routing feature, very low adoption
- `freeform_tags` -- auto-populated from metadata labels (standard pattern)

## Benefits

- **Managed caching** -- Fully managed Redis-compatible cache with automatic patching
- **Horizontal scaling** -- Sharded mode distributes data across multiple shards for high throughput
- **High availability** -- Multiple nodes per shard (or per cluster) provide automatic failover
- **Network isolation** -- NSG support for fine-grained access control
- **Custom configuration** -- Config Set references for tuning Redis parameters
- **Feature parity** -- Identical capabilities in both Pulumi and Terraform modules

## Impact

- Fifth of six Phase 4 (Databases) resources completed
- Next component: R20 OciNosqlTable
- 19 of 37 OCI resource kinds now implemented

## Related Work

- R15 OciAutonomousDatabase (2026-02-19) -- serverless Oracle database
- R16 OciDbSystem (2026-02-19) -- traditional Oracle on VM/BM
- R17 OciMysqlDbSystem (2026-02-19) -- managed MySQL HeatWave
- R18 OciPostgresqlDbSystem (2026-02-19) -- managed PostgreSQL
- R20 OciNosqlTable (next) -- NoSQL table with secondary indexes

---

**Status**: Production Ready
**Validation**: go build clean, go vet clean, 22/22 tests passed, terraform validate success
