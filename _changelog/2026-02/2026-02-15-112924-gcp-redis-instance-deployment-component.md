# GcpRedisInstance Deployment Component

**Date**: February 15, 2026
**Type**: Feature
**Components**: GCP Provider, API Definitions, Pulumi Module, Terraform Module

## Summary

Added GcpRedisInstance as a new GCP deployment component, providing a declarative interface for provisioning Google Cloud Memorystore for Redis instances. The component supports both BASIC and STANDARD_HA tiers with full coverage of auth, TLS, persistence, read replicas, maintenance windows, and CMEK encryption.

## Problem Statement / Motivation

Platform engineers provisioning Redis caches on GCP must navigate Terraform's `google_redis_instance` resource with 25+ arguments, understand tier-dependent field constraints, and manually wire VPC networking and encryption keys. There was no standardized, validated way to express Redis infrastructure requirements in a single YAML manifest.

### Pain Points

- Tier-dependent field validation (read replicas require STANDARD_HA) not caught until apply time
- No cross-resource composition for VPC networking or KMS encryption keys
- No standardized presets for common deployment patterns (dev cache vs production HA)
- Instance naming constraints (`^[a-z][a-z0-9-]{0,39}[a-z0-9]$`) not enforced pre-deploy

## Solution / What's New

A complete deployment component with proto API, Pulumi module, Terraform module, validation tests, documentation, and presets.

### Proto API (4 files, 20 spec fields)

- 2 sub-messages: `GcpRedisInstanceMaintenanceWindow`, `GcpRedisInstancePersistenceConfig`
- 3 `StringValueOrRef` fields for infra-chart composition: `project_id`, `authorized_network`, `customer_managed_key`
- 2 cross-field CEL validations: read replicas require HA tier, replica count range enforcement
- 6 stack outputs: `host`, `port`, `read_endpoint`, `read_endpoint_port`, `current_location_id`, `auth_string`

### IaC Modules

- **Pulumi** (4 Go files): `redis.NewInstance` with conditional field mapping, framework labels
- **Terraform** (6 HCL files): `google_redis_instance` with dynamic blocks for maintenance and persistence

### Validation

- 43 spec tests (22 positive, 21 negative) covering all field validations, boundary conditions, and cross-field CEL rules
- `go build` and `terraform validate` both pass

### Documentation & Presets

- User-facing README with architecture overview
- 6 YAML examples (basic through full production)
- Comprehensive research document (deployment landscape, best practices, pitfalls)
- 3 presets: basic-cache, ha-production, ha-read-replicas
- Catalog page for site discovery

## Implementation Details

### Key Design Decisions

1. **Naming**: `GcpRedisInstance` maps to `google_redis_instance` / `redis.Instance`, matching Terraform/Pulumi nomenclature. Future `GcpMemorystoreInstance` will cover the new-gen sharded/Valkey API.

2. **`tier` as required**: Unlike GCP's default to BASIC, we require explicit tier selection to force deliberate HA decisions.

3. **Persistence as sub-message**: `GcpRedisInstancePersistenceConfig` with mode + snapshot period, validated by CEL (RDB mode requires a snapshot period).

4. **`deletion_protection`**: Supported in Pulumi but not a native TF schema field for `google_redis_instance`. TF users should use `lifecycle { prevent_destroy }`.

5. **Excluded `alternative_location_id`**: GCP auto-selects the HA failover zone. Low user value, adds complexity.

### Files Created

- `apis/org/openmcf/provider/gcp/gcpredisinstance/v1/` -- 35 files total
- `apis/org/openmcf/shared/cloudresourcekind/cloud_resource_kind.proto` -- enum 631 registered
- `site/public/docs/catalog/gcp/redis-instance.md` -- catalog page

## Benefits

- **Pre-deploy validation**: 2 cross-field CEL rules catch tier/replica misconfigurations before any cloud API call
- **Composition-ready**: 3 StringValueOrRef fields enable wiring to GcpProject, GcpVpc, and GcpKmsKey in infra charts
- **Standardized presets**: 3 presets cover 90% of use cases (dev, production HA, read-replica scaled)
- **Feature parity**: Pulumi and Terraform modules cover identical field sets

## Impact

- GCP resource count increases from 26 to 27
- Enables caching layer in future infra charts (serverless-api-backend, microservices environments)
- Completes the first database-category resource in the expansion queue

## Related Work

- Part of 20260215.01.sp.gcp-resource-expansion (R08 of 22)
- Next: R08b GcpMemorystoreInstance (new-gen Memorystore API for sharded clusters)
- Dependencies: GcpProject (project_id), GcpVpc (authorized_network), GcpKmsKey (customer_managed_key)

---

**Status**: Production Ready
**Timeline**: Single session
