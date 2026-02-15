# GcpBigQueryDataset

A GcpBigQueryDataset provisions a BigQuery dataset -- the top-level organizational
container for tables, views, and routines in Google BigQuery.

## When to Use

Use GcpBigQueryDataset when you need:

- **A managed analytics dataset** for structured data storage and SQL-based analysis
- **Data residency controls** with explicit location selection (multi-regional or regional)
- **Customer-managed encryption** (CMEK) for datasets containing sensitive or regulated data
- **Team-level access control** with fine-grained IAM roles for users, groups, and service accounts
- **Lifecycle management** for tables with automatic expiration policies

## Prerequisites

- A GCP project with the BigQuery API enabled
- Appropriate IAM permissions (`roles/bigquery.dataOwner` or `roles/bigquery.admin`)
- For CMEK: an existing KMS key (see [GcpKmsKey](../gcpkmskey/v1/))

## Quick Start

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigQueryDataset
metadata:
  name: my-analytics-dataset
spec:
  projectId:
    value: "my-gcp-project"
  datasetId: analytics_prod
  location: US
```

This creates a BigQuery dataset in the US multi-region with default settings --
Google-managed encryption, 7-day time travel, and default project-level access.

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `projectId` | StringValueOrRef | Yes | GCP project ID |
| `datasetId` | string | Yes | Dataset ID (`[A-Za-z0-9_]`, max 1024 chars) |
| `location` | string | Yes | Multi-regional (US, EU) or regional (us-central1) |
| `friendlyName` | string | No | Display name |
| `description` | string | No | Dataset description |
| `defaultTableExpirationMs` | int64 | No | Auto-delete tables after N ms (min: 3600000) |
| `defaultPartitionExpirationMs` | int64 | No | Auto-delete partitions after N ms |
| `maxTimeTravelHours` | int32 | No | Time travel window, 48-168 hours (default: 168) |
| `isCaseInsensitive` | bool | No | Case-insensitive names (immutable, default: false) |
| `defaultCollation` | string | No | Default collation ("und:ci" for case-insensitive) |
| `storageBillingModel` | string | No | LOGICAL (default) or PHYSICAL |
| `deleteContentsOnDestroy` | bool | No | Delete tables on destroy (default: false) |
| `kmsKeyName` | StringValueOrRef | No | CMEK encryption key |
| `access` | list | No | Access control entries (authoritative) |

### Access Entry Fields

| Field | Type | Description |
|-------|------|-------------|
| `role` | string | IAM role (OWNER, WRITER, READER, or predefined) |
| `userByEmail` | string | Google Account email |
| `groupByEmail` | string | Google Group email |
| `domain` | string | Domain (e.g., "example.com") |
| `specialGroup` | string | projectOwners, projectReaders, projectWriters, allAuthenticatedUsers |
| `iamMember` | string | IAM member expression |
| `view` | object | Authorized view reference (project_id, dataset_id, table_id) |

## Important Notes

**Access is authoritative.** When you specify `access` entries, BigQuery removes
any entries not in your spec. Omitting `access` entirely preserves BigQuery's
default access (project owners/editors/viewers).

**Dataset ID cannot contain hyphens.** Unlike most GCP resource names, BigQuery
dataset IDs only allow letters, numbers, and underscores. Use `analytics_prod`
not `analytics-prod`.

**Location is immutable.** Choose carefully -- changing location requires
destroying and recreating the dataset (and all its tables).

**Storage billing model matters.** PHYSICAL billing can reduce costs 60-80% for
highly compressible data (JSON, repeated strings). Consider this for cost-sensitive
workloads.

## Related Components

- [GcpKmsKey](../gcpkmskey/v1/) -- CMEK encryption key for dataset encryption
- [GcpProject](../gcpproject/v1/) -- Parent GCP project
- [GcpGcsBucket](../gcpgcsbucket/v1/) -- Often paired for data lake architectures
