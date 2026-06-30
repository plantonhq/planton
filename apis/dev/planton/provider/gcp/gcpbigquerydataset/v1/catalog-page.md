# GCP BigQuery Dataset

Deploys a GCP BigQuery dataset with configurable data location, table lifecycle defaults, access control, and optional CMEK encryption. The dataset serves as the top-level container for tables, views, and routines in BigQuery.

## What Gets Created

When you deploy a GcpBigQueryDataset resource, Planton provisions:

- **BigQuery Dataset** — a `google_bigquery_dataset` resource in the specified project and location, tagged with organization, environment, and resource labels
- **Access Control Entries** — if the `access` field is provided, an authoritative set of IAM bindings granting roles to users, groups, domains, special groups, IAM members, or authorized views
- **CMEK Encryption Configuration** — if `kmsKeyName` is provided, all new tables in the dataset default to encryption with the specified Cloud KMS key

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A GCP project** where the dataset will be created
- **A Cloud KMS key** if enabling customer-managed encryption (optional)
- **BigQuery API** enabled in the target project

## Quick Start

Create a file `bigquery-dataset.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpBigQueryDataset
metadata:
  name: my-dataset
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpBigQueryDataset.my-dataset
spec:
  projectId:
    value: my-gcp-project
  datasetId: analytics_events
  location: US
```

Deploy:

```shell
planton apply -f bigquery-dataset.yaml
```

This creates a BigQuery dataset named `analytics_events` in the `US` multi-region with default access (project owners = OWNER, project editors = WRITER, project viewers = READER).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project where the dataset will be created. Can reference a GcpProject resource via `valueFrom`. | Required |
| `datasetId` | `string` | Unique identifier for the dataset within the project. Only letters, numbers, and underscores. Immutable after creation. | Required; pattern `^[0-9A-Za-z_]+$`; max 1024 chars |
| `location` | `string` | Geographic location where the dataset resides (e.g., `US`, `EU`, `us-central1`). Immutable after creation. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `friendlyName` | `string` | — | User-friendly display name for the dataset. |
| `description` | `string` | — | Description of the dataset's contents or purpose. |
| `defaultTableExpirationMs` | `int64` | `0` (no expiration) | Default lifetime for tables in the dataset, in milliseconds. Minimum 3600000 (1 hour). |
| `defaultPartitionExpirationMs` | `int64` | `0` (no expiration) | Default expiration for partitions in partitioned tables, in milliseconds. |
| `maxTimeTravelHours` | `int32` | `168` (7 days) | Hours of time travel for point-in-time snapshots. Range: 48–168. Lower values reduce storage costs. |
| `isCaseInsensitive` | `bool` | `false` | When `true`, dataset and table names are case-insensitive. Immutable after creation. |
| `defaultCollation` | `string` | — | Default collation for string columns in new tables. Use `und:ci` for case-insensitive collation. |
| `storageBillingModel` | `string` | `LOGICAL` | Billing model: `LOGICAL` (uncompressed bytes) or `PHYSICAL` (compressed bytes, can reduce costs 60–80%). |
| `deleteContentsOnDestroy` | `bool` | `false` | When `true`, all tables are deleted when the dataset is destroyed. When `false`, destroy fails if the dataset contains tables. |
| `kmsKeyName` | `StringValueOrRef` | — | Cloud KMS key for default table encryption (CMEK). Format: `projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{key}`. Can reference a GcpKmsKey resource via `valueFrom`. |
| `access` | `GcpBigQueryDatasetAccessEntry[]` | Default project access | Authoritative access control entries. Entries not listed here are removed. See access entry fields below. |
| `access[].role` | `string` | — | IAM role to grant (e.g., `OWNER`, `WRITER`, `READER`, `roles/bigquery.dataViewer`). Required unless `view` is set. |
| `access[].userByEmail` | `string` | — | Email address of a Google Account. |
| `access[].groupByEmail` | `string` | — | Email address of a Google Group. |
| `access[].domain` | `string` | — | Domain to grant access to (e.g., `example.com`). |
| `access[].specialGroup` | `string` | — | Special group: `projectOwners`, `projectReaders`, `projectWriters`, or `allAuthenticatedUsers`. |
| `access[].iamMember` | `string` | — | IAM member expression (e.g., `serviceAccount:sa@project.iam.gserviceaccount.com`). |
| `access[].view.projectId` | `string` | — | GCP project containing the authorized view. |
| `access[].view.datasetId` | `string` | — | Dataset containing the authorized view. |
| `access[].view.tableId` | `string` | — | Table ID of the authorized view. |

## Examples

### Dataset with Table Expiration

Automatically delete tables after 90 days, useful for staging or transient data:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpBigQueryDataset
metadata:
  name: staging-events
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.GcpBigQueryDataset.staging-events
spec:
  projectId:
    value: my-gcp-project
  datasetId: staging_events
  location: us-central1
  friendlyName: Staging Events
  description: Transient event data with 90-day auto-expiration
  defaultTableExpirationMs: 7776000000
  maxTimeTravelHours: 48
  deleteContentsOnDestroy: true
```

### Dataset with CMEK Encryption and Physical Billing

Production dataset using customer-managed encryption and physical storage billing for cost optimization:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpBigQueryDataset
metadata:
  name: prod-analytics
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpBigQueryDataset.prod-analytics
spec:
  projectId:
    value: my-gcp-project
  datasetId: prod_analytics
  location: US
  friendlyName: Production Analytics
  description: Core analytics dataset with CMEK and physical billing
  storageBillingModel: PHYSICAL
  kmsKeyName:
    value: projects/my-gcp-project/locations/us/keyRings/analytics-ring/cryptoKeys/analytics-key
```

### Dataset with Explicit Access Control

Grant access to specific users, groups, and an authorized view:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpBigQueryDataset
metadata:
  name: finance-data
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpBigQueryDataset.finance-data
spec:
  projectId:
    value: my-gcp-project
  datasetId: finance_data
  location: EU
  friendlyName: Finance Data
  description: Restricted financial data with explicit access grants
  isCaseInsensitive: true
  defaultCollation: "und:ci"
  access:
    - role: OWNER
      userByEmail: data-owner@example.com
    - role: WRITER
      groupByEmail: data-engineers@example.com
    - role: READER
      groupByEmail: analysts@example.com
    - role: READER
      iamMember: "serviceAccount:etl-pipeline@my-gcp-project.iam.gserviceaccount.com"
    - view:
        projectId: my-gcp-project
        datasetId: reporting_views
        tableId: finance_summary
```

### Using Foreign Key References

Reference other Planton-managed resources instead of hardcoding values:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpBigQueryDataset
metadata:
  name: ref-dataset
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpBigQueryDataset.ref-dataset
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  datasetId: warehouse
  location: us-central1
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: warehouse-key
      field: status.outputs.key_id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `dataset_id` | `string` | The short dataset ID (same as the spec's `datasetId` input), used in BigQuery SQL queries and API calls |
| `self_link` | `string` | Fully qualified URI of the dataset (e.g., `https://bigquery.googleapis.com/bigquery/v2/projects/{project}/datasets/{dataset}`) |
| `project` | `string` | The GCP project that contains this dataset |
| `creation_time` | `int64` | Creation time of the dataset in milliseconds since epoch |

## Related Components

- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project where the dataset is created
- [GcpKmsKeyRing](/docs/catalog/gcp/gcpkmskeyring) — provides the key ring containing KMS keys for CMEK encryption
- [GcpKmsKey](/docs/catalog/gcp/gcpkmskey) — provides the Cloud KMS encryption key referenced by `kmsKeyName`
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — creates service accounts that can be granted dataset access
