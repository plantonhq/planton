# GcpBigQueryDataset

A GcpBigQueryDataset provisions a BigQuery dataset -- the top-level container
that pins data location and owns the defaults every contained table inherits:
lifecycle (expiration), encryption (CMEK), collation, and the dataset-level
ACL. Tables that infrastructure should own are first-class
[GcpBigQueryTable](../gcpbigquerytable/) resources that reference this
dataset.

## When to Use

Use GcpBigQueryDataset when you need:

- **A managed analytics dataset** for structured data storage and SQL-based analysis
- **Data residency controls** with explicit location selection (multi-regional or regional)
- **Customer-managed encryption** (CMEK) for datasets containing sensitive or regulated data
- **Team-level access control** with an explicit, authoritative ACL for users,
  groups, service accounts, authorized views/routines/datasets, and
  condition-gated grants
- **Lifecycle management** for tables with automatic expiration policies
- **BigQuery Omni federation** (external dataset references to AWS Glue) or
  **open-source catalog interop** (Hive Metastore compatibility for Spark)

## Prerequisites

- A GCP project (the module enables the BigQuery API on it automatically)
- IAM permissions — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs
- For CMEK: an existing KMS key (see [GcpKmsKey](../gcpkmskey/)) with the
  BigQuery service agent granted `roles/cloudkms.cryptoKeyEncrypterDecrypter`

## Quick Start

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryDataset
metadata:
  name: my-analytics-dataset
spec:
  datasetId: analytics_prod
  location: US
```

This creates a BigQuery dataset in the US multi-region in the provider's
default project, with default settings -- Google-managed encryption, 7-day
time travel, and default project-level access.

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `projectId` | StringValueOrRef | No | GCP project ID; provider default project when omitted |
| `datasetId` | string | Yes | Dataset ID (`[A-Za-z0-9_]`, max 1024 chars); immutable |
| `location` | string | Yes | Multi-regional (US, EU) or regional (us-central1); immutable |
| `friendlyName` | string | No | Display name |
| `description` | string | No | Dataset description |
| `labels` | map | No | User labels, merged beneath Planton's platform labels |
| `resourceTags` | map | No | Resource Manager tags (`tagKey namespaced name` -> `value short name`) |
| `defaultTableExpirationMs` | int64 | No | Auto-delete new tables after N ms (min: 3600000) |
| `defaultPartitionExpirationMs` | int64 | No | Auto-delete partitions after N ms |
| `maxTimeTravelHours` | int32 | No | Time travel window, 48-168 hours (default: 168) |
| `isCaseInsensitive` | bool | No | Case-insensitive names (immutable, default: false) |
| `defaultCollation` | string | No | Default collation ("und:ci" for case-insensitive) |
| `storageBillingModel` | string | No | LOGICAL (default) or PHYSICAL |
| `deleteContentsOnDestroy` | bool | No | Delete tables on destroy (default: false) |
| `kmsKeyName` | StringValueOrRef | No | CMEK default key for new tables |
| `access` | list | No | Authoritative access control entries |
| `externalDatasetReference` | object | No | BigQuery Omni federated source (immutable) |
| `externalCatalogOptions` | object | No | Hive Metastore compatibility metadata |
| `deletionPolicy` | string | No | `DELETE` (default: destroy removes the dataset — non-empty datasets still need `deleteContentsOnDestroy`), `PREVENT` (destroy fails), `ABANDON` (leaves the dataset serving unmanaged) |

### Access Entry Fields

Each entry is either a **principal grant** (role + exactly one of the five
principal identities, optionally condition-gated) or a **resource
authorization** (exactly one of view/routine/dataset, no role). Both shapes
are enforced pre-deploy.

| Field | Type | Description |
|-------|------|-------------|
| `role` | string | Role for principal grants (OWNER, WRITER, READER, or predefined); omit for authorizations |
| `userByEmail` | string | Google Account email |
| `groupByEmail` | string | Google Group email |
| `domain` | string | Domain (e.g., "example.com") |
| `specialGroup` | string | projectOwners, projectReaders, projectWriters, allAuthenticatedUsers |
| `iamMember` | string | IAM member expression |
| `view` | object | Authorized view (projectId, datasetId, tableId) |
| `routine` | object | Authorized routine (projectId, datasetId, routineId) |
| `dataset` | object | Authorized dataset (projectId, datasetId, targetTypes) |
| `condition` | object | IAM condition gating a principal grant (expression, title, description, location) |

## Important Notes

**Access is authoritative.** When you specify `access` entries, they become
the dataset's complete ACL and BigQuery removes anything else. Omitting
`access` entirely preserves BigQuery's default access (project
owners/editors/viewers); to keep those defaults alongside custom grants,
list them explicitly.

**Dataset ID cannot contain hyphens.** Unlike most GCP resource names,
BigQuery dataset IDs only allow letters, numbers, and underscores. Use
`analytics_prod` not `analytics-prod`.

**Location is immutable.** Choose carefully -- changing location requires
destroying and recreating the dataset (and all its tables), and every table a
query joins must live in the same location.

**Storage billing model matters.** PHYSICAL billing can reduce costs 60-80%
for highly compressible data (JSON, repeated strings), but bills time-travel
storage too; switching to PHYSICAL is allowed once every 14 days.

## Recorded Skips (with reasons)

- `google_bigquery_dataset_access` (the additive single-grant resource) is
  deliberately not modeled as a kind: mixing it with the dataset's
  authoritative inline `access` causes permanent diffs, and the inline ACL is
  the shape the dataset legitimately owns.
- BigQuery routine / row-access-policy / job / IAM trios / reservation /
  connection / data-transfer / analytics-hub / data-policy resources are
  separate product families outside this kind's boundary.

## Related Kinds

- [GcpBigQueryTable](../gcpbigquerytable/) -- infrastructure-owned tables, views, and external tables in this dataset
- [GcpKmsKey](../gcpkmskey/) -- CMEK encryption key for dataset encryption
- [GcpProject](../gcpproject/) -- Parent GCP project
- [GcpGcsBucket](../gcpgcsbucket/) -- Often paired for data lake architectures

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
