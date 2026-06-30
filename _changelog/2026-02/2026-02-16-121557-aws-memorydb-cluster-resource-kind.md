# AWS MemoryDB Cluster Resource Kind

**Date**: February 16, 2026
**Type**: Feature
**Components**: API Definitions, AWS Provider, Pulumi CLI Integration

## Summary

Added AwsMemorydbCluster (R27, enum 342) as a new deployment component for Amazon MemoryDB — a fully managed, Redis-compatible, durable in-memory database. This is the 30th new AWS resource kind added as part of the cloud provider expansion project, completing the first component of Phase 3 (specialized services).

## Problem Statement / Motivation

The AWS resource catalog lacked support for MemoryDB, a distinct service from ElastiCache. While ElastiCache serves as an ephemeral caching layer, MemoryDB provides database-grade durability via a Multi-AZ distributed transaction log — making it suitable as a primary database for applications that need both Redis performance and data persistence.

### Pain Points

- Teams needing durable Redis-compatible storage had no Planton component to manage it
- MemoryDB has a different authentication model (ACL-based vs ElastiCache's auth_token/user_group_ids), different topology (always sharded), and different encryption model (always-on at-rest) — it could not be conflated with the existing ElastiCache component
- No standardized way to manage MemoryDB subnet groups and parameter groups as part of the cluster lifecycle

## Solution / What's New

A complete AwsMemorydbCluster deployment component following the established ElastiCache pattern, adapted for MemoryDB's architectural differences.

### Key Design Decisions

- **ACL by reference, not bundled** — MemoryDB ACLs and Users have independent lifecycles and can be shared across clusters. The component accepts `acl_name` as a required string field (default `"open-access"`) rather than attempting to bundle ACL/User creation inline.
- **Always-sharded topology** — Unlike ElastiCache's clustered-vs-non-clustered mode split, MemoryDB always uses a sharded architecture with `num_shards` and `num_replicas_per_shard`. Simpler for users.
- **Always-on encryption at rest** — No `at_rest_encryption_enabled` toggle (MemoryDB always encrypts). The `kms_key_id` field optionally provides a customer-managed key.
- **CEL presence-aware validation** — The `tls_disabled_requires_open_access` CEL rule uses `has()` checks to correctly handle optional fields whose defaults haven't been applied yet during validation.

## Implementation Details

### Proto API (4 files)

- `spec.proto` — 23 fields covering engine, topology, ACL, networking, encryption, maintenance, snapshots, parameter groups, and advanced options. 4 CEL validations for engine values, TLS/ACL coupling, parameter group family requirement, and snapshot restore mutual exclusion.
- `stack_outputs.proto` — 7 outputs: cluster endpoint address/port, ARN, name, engine patch version, subnet group name, parameter group name.
- `api.proto` — Standard KRM envelope wiring.
- `stack_input.proto` — Standard stack input with AWS provider config.

### Pulumi IaC Module (6 files)

- `main.go` — Provider setup + orchestration: subnet group → parameter group → cluster.
- `locals.go` — Pre-computed values with AWS tags using `awstagkeys` package.
- `outputs.go` — 7 output key constants.
- `subnet_group.go` — Creates `memorydb.SubnetGroup` when `subnet_ids` provided. Name sanitization matching AWS constraints.
- `parameter_group.go` — Creates `memorydb.ParameterGroup` when `parameters` + `parameter_group_family` provided.
- `cluster.go` — Creates `memorydb.Cluster` with all spec fields mapped. Exports cluster endpoint, ARN, name, and engine patch version.

### Terraform Module (5 files)

Feature parity with Pulumi: `provider.tf`, `variables.tf`, `locals.tf`, `main.tf`, `outputs.tf`.

### Validation Tests

18 spec_test.go test cases covering all valid and invalid input combinations:
- Valid: minimal, full production, valkey engine, TLS disabled with open-access, snapshot restore, data tiering
- Invalid: missing engine, missing node_type, invalid engine, TLS+ACL mismatch, parameters without family, mutual exclusion violations, range violations, format violations

### Documentation

- `README.md` — Full configuration reference with MemoryDB vs ElastiCache comparison
- `examples.md` — 4 examples: minimal dev, production HA, high-throughput with data tiering, infra chart reference
- `catalog-page.md` — Source-verified catalog page following the ALB exemplar structure
- 3 presets: dev-single-shard, production-ha, high-throughput

## Benefits

- Teams can now manage durable Redis-compatible databases through Planton's declarative workflow
- Subnet group and parameter group management is bundled, reducing boilerplate
- Cross-resource references (`valueFrom`) enable wiring MemoryDB into larger infra charts
- 3 presets provide immediate starting points for common deployment patterns

## Impact

- **API surface**: +1 CloudResourceKind enum (342), +4 proto files
- **IaC modules**: +11 Go files (Pulumi), +5 Terraform files
- **Documentation**: +8 documentation files (README, examples, catalog, 3 preset pairs)
- **Tests**: +18 validation test cases, all passing

## Related Work

- Part of 20260215.02.sp.aws-resource-expansion (R27 of ~32)
- Pattern reference: AwsRedisElasticache component
- Parent project: 20260212.01.planton-cloud-provider-expansion

---

**Status**: Production Ready
