# GcpBigQueryDataset -- Research & Design Documentation

## BigQuery Dataset in the GCP Ecosystem

BigQuery is Google Cloud's fully managed, serverless data warehouse designed for
large-scale analytics. A **dataset** is the top-level organizational container
within BigQuery -- it holds tables, views, routines (UDFs), and models. Datasets
are scoped to a project and bound to a geographic location.

The key architectural insight: datasets are **infrastructure**, while tables are
**schema**. This distinction drives the design boundary for Planton -- we model
the dataset (infrastructure provisioning, access control, encryption, lifecycle
policies) but deliberately exclude tables, views, and routines (which belong to
application code, dbt, or migration scripts).

## Deployment Landscape

### Method Comparison

| Method | Dataset | Tables | Access | CMEK | Lifecycle |
|--------|---------|--------|--------|------|-----------|
| GCP Console | Yes | Yes | Yes | Yes | Manual |
| `bq` CLI | Yes | Yes | Yes | Yes | Manual |
| Terraform (`google_bigquery_dataset`) | Yes | Separate resource | Yes (authoritative) | Yes | IaC |
| Pulumi (`bigquery.Dataset`) | Yes | Separate resource | Yes | Yes | IaC |
| Planton (this component) | Yes | Excluded | Yes (authoritative) | Yes | IaC |
| dbt | No | Yes (models) | No | No | Schema |
| Dataform | No | Yes (SQL workflows) | No | No | Schema |

Planton fills the IaC gap for dataset provisioning with cross-resource composability
that Terraform and Pulumi lack natively.

## Field Analysis

### Immutable Fields (ForceNew)

These fields cannot be changed after creation. Any change destroys and recreates
the dataset:

- `dataset_id` -- the BigQuery dataset identifier
- `location` -- geographic data residency
- `is_case_insensitive` -- case sensitivity behavior

### Mutable Fields

- `friendly_name`, `description` -- metadata
- `default_table_expiration_ms`, `default_partition_expiration_ms` -- lifecycle
- `max_time_travel_hours` -- time travel configuration
- `default_collation` -- collation settings
- `storage_billing_model` -- billing model
- `default_encryption_configuration` -- CMEK key
- `access` -- access control entries
- `labels` -- managed by Planton framework
- `delete_contents_on_destroy` -- safety flag (Terraform/Pulumi-specific)

### Labels Support

BigQuery datasets support GCP labels. The Planton framework applies standard labels:

- `planton-resource: true`
- `planton-resource-name: <dataset_id>`
- `planton-resource-kind: gcpbigquerydataset`
- `planton-organization: <metadata.org>` (if set)
- `planton-environment: <metadata.env>` (if set)
- `planton-resource-id: <metadata.id>` (if set)

## Access Control Model

BigQuery datasets support two access control models:

### 1. Dataset-Level Access (What Planton Models)

Access entries are embedded in the dataset resource itself. Each entry grants a
role to one identity. This is the **authoritative** model -- BigQuery enforces
that only the specified entries exist.

Identity types supported:
- `user_by_email` -- individual Google Account
- `group_by_email` -- Google Group
- `domain` -- entire domain (e.g., "example.com")
- `special_group` -- projectOwners, projectReaders, projectWriters, allAuthenticatedUsers
- `iam_member` -- generic IAM member expression
- `view` -- authorized view (grants data access without a role)

### 2. IAM Policy Binding (Not Modeled)

Managed separately via `google_bigquery_dataset_iam_*` resources. These are
**additive** -- they don't remove existing access. Planton does not model these
separately because dataset-level access is sufficient for most use cases and
provides a single source of truth.

### Default Access Behavior

When `access` is omitted from the spec, BigQuery applies default access:
- Project owners get `OWNER`
- Project editors get `WRITER`
- Project viewers get `READER`

When `access` is specified, it becomes authoritative -- unlisted entries are removed.
This is a critical behavioral difference that users must understand.

## CMEK (Customer-Managed Encryption Keys)

BigQuery encrypts all data at rest by default using Google-managed keys. CMEK
provides additional control:

- The KMS key must be in the **same region** as the dataset
- For multi-regional datasets (US, EU), the KMS key must also be multi-regional
- CMEK is set at the dataset level (affects all new tables)
- Individual tables can override with their own CMEK
- Changing the CMEK key only affects new table data; existing data remains
  encrypted with the previous key

### CMEK and Planton Composability

The `kms_key_name` field uses `StringValueOrRef` with `default_kind = GcpKmsKey`.
In infra charts, this enables:

```yaml
kmsKeyName:
  valueFrom:
    kind: GcpKmsKey
    name: analytics-cmek
    fieldPath: status.outputs.key_id
```

This creates a dependency edge: the KMS key is provisioned before the dataset.

## Storage Billing Model

BigQuery offers two billing models for storage:

| Model | Charges For | Best For |
|-------|------------|----------|
| LOGICAL (default) | Uncompressed data size | Small datasets, predictable billing |
| PHYSICAL | Compressed on-disk size | Large datasets with compressible data |

PHYSICAL billing typically reduces storage costs 60-80% for:
- JSON columns (highly compressible)
- Repeated string values
- Wide tables with many nullable columns

The tradeoff: PHYSICAL billing includes time travel and fail-safe storage in the
billing calculation, while LOGICAL billing does not.

## Time Travel

BigQuery's time travel feature allows querying data at any point within the
configured window (48-168 hours). This affects:

- **Cost**: Time travel storage is billed under PHYSICAL model
- **Recovery**: Longer windows provide more recovery options
- **Point-in-time queries**: `SELECT * FROM table FOR SYSTEM_TIME AS OF timestamp`

Reducing `max_time_travel_hours` to 48 (minimum) saves storage costs but limits
the recovery window to 2 days.

## Infra-Chart Composability

GcpBigQueryDataset is a **Layer 1** resource in infra chart topology:

```
Layer 0: GcpProject
Layer 0-1: GcpKmsKeyRing -> GcpKmsKey
Layer 1: GcpBigQueryDataset (references Project, optionally KmsKey)
Layer 2+: Application tables, views, dbt models (not IaC)
```

The dataset participates in these infra chart patterns:

- **data-analytics-environment**: BigQuery Dataset + Dataproc + PubSub + GCS + SA
- **ml-notebook-environment**: BigQuery Dataset + Vertex AI Notebook + GCS + SA
- **event-pipeline**: BigQuery Dataset + PubSub + Cloud Function + GCS

### Key Outputs for Composition

| Output | Used By |
|--------|---------|
| `dataset_id` | Application code, dbt, SQL queries |
| `self_link` | API references, audit trails |
| `project` | Cross-project dataset references |

## Deliberate Exclusions

| Feature | Reason |
|---------|--------|
| `external_dataset_reference` | Federated datasets backed by external sources (AWS Glue). Niche feature. |
| `external_catalog_dataset_options` | Open source catalog integration. Niche. |
| `resource_tags` | GCP resource tags (different from labels). Not widely adopted. |
| Access `condition` | CEL-based conditional access. Advanced IAM feature. |
| Access `routine` | Authorized routine access. Rare in IaC. |
| Access `dataset` | Authorized dataset access. Rare in IaC. |
| Tables, views, routines | Schema-level concerns managed by dbt/application code. |

These can be added in future versions if demand materializes.

## Best Practices

1. **Use descriptive dataset IDs** -- e.g., `analytics_prod`, `raw_events_2024`
2. **Choose location carefully** -- immutable after creation, affects latency and compliance
3. **Set CMEK for regulated data** -- PCI, HIPAA, SOX workloads
4. **Consider PHYSICAL billing** -- significant savings for compressible data
5. **Be explicit about access** -- document who has access and why
6. **Use table expiration for staging** -- prevent data sprawl in dev/staging
7. **Keep time travel at 168 hours** -- maximum recovery window unless cost is critical
