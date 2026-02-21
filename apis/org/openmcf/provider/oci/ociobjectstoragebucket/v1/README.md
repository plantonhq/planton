# OciObjectStorageBucket

## Overview

OciObjectStorageBucket is an OpenMCF component that deploys an OCI Object Storage bucket along with optional lifecycle policies, retention rules, and cross-region replication policies. It provides a single declarative manifest to manage all bucket-related resources as one unit.

## Purpose

Object Storage is OCI's primary durable object store for unstructured data — backups, logs, data lake files, artifacts, and static assets. This component wraps the bucket and its associated policies into a single resource so that lifecycle management, compliance retention, and disaster recovery replication are configured at deploy time rather than as separate manual steps.

## Key Features

- **Single-resource deployment** — one manifest creates the bucket, its lifecycle policy, and all replication policies.
- **Retention rules** — up to 100 inline retention rules with configurable duration and optional time-lock for compliance.
- **Lifecycle management** — automatic archival, tiering transitions, deletion of old objects, and abort of stale multipart uploads based on age and name-pattern filters.
- **Cross-region replication** — declarative replication policies that asynchronously copy objects to destination buckets in other OCI regions.
- **Versioning** — enable, disable, or suspend object version history.
- **Auto-tiering** — automatic Standard ↔ InfrequentAccess transitions based on access patterns.
- **Customer-managed encryption** — optional KMS key for server-side encryption.
- **Event emission** — optional integration with OCI Events for object state change notifications.
- **Foreign key references** — `compartmentId` and `kmsKeyId` support `valueFrom` to reference other OpenMCF-managed resources.

## Constraints

- `storageTier` is immutable after bucket creation (Standard or Archive).
- Replication policy fields (`name`, `destinationBucketName`, `destinationRegionName`) are immutable after creation.
- Destination buckets for replication must exist before the replication policy is created.
- Changing the bucket `name` forces recreation.
- Max 100 retention rules per bucket.
- Lifecycle rules with `lifecycle_abort` action must target `"multipart-uploads"`.
- `objectNameFilter` is not valid when target is `"multipart-uploads"`.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Application data store | Minimal bucket with default private access |
| Compliance archive | Versioning + retention rules with time-lock |
| Data lake with cost optimization | Auto-tiering + lifecycle rules for archival and cleanup |
| Disaster recovery | Cross-region replication to a secondary OCI region |
| Audit logging | Event emission + versioning for immutable log records |

## Production Features

- **Freeform tags** — automatically populated from `metadata.labels`, including `resource_kind`, `resource_id`, `organization`, and `environment`.
- **KMS encryption** — customer-managed keys for encryption at rest when regulatory requirements exceed Oracle-managed defaults.
- **Replication** — asynchronous cross-region replication for RPO-based disaster recovery.
- **Retention locks** — time-locked retention rules prevent early deletion, meeting regulatory hold requirements.
