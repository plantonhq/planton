# GCP BigQuery Dataset

Deploys a BigQuery dataset with configurable data location, access control, default table lifecycle policies, storage billing model, and optional CMEK encryption. The dataset is the container tables, views, and routines live in — its location decides where every table's data resides and where queries process, and that choice is permanent.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **BigQuery API enablement** (`bigquery.googleapis.com`) on the target project (never disabled on destroy)
- **BigQuery Dataset** -- a managed dataset in the specified GCP project and location, configured with the chosen storage billing model, time travel window, and default table expiration policies
- **Access Control Entries** -- when `access` entries are specified, authoritative IAM bindings granting roles to users, groups, domains, special groups, IAM members, or authorized views (if omitted, BigQuery applies default project-level access)
- **CMEK Encryption Configuration** -- created only when `kmsKeyName` is set; configures a default Cloud KMS encryption key for all new tables in the dataset
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the dataset will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **Cloud KMS key** (if using CMEK) -- the key must be in the same location as the dataset. The BigQuery service account must have `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key.

## Deploy

### Console

Open the deployment store, find **GCP BigQuery Dataset**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic Analytics Dataset** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryDataset
metadata:
  name: analytics-events
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  datasetId: analytics_events_prod
  location: US
```

```shell
planton apply -f bigquery-dataset.yaml
```

This creates a dataset in the US multi-region with Google-managed encryption, default project-level access, logical storage billing, and 7-day time travel. No table expiration policies are configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the dataset to a GCP project and KMS key deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: bigquery-cmek-key
      fieldPath: status.outputs.key_id
```

The InfraPipeline resolves the dependency graph, deploys the project and KMS key first, then provisions the dataset with CMEK encryption.

## Key Configuration

These are the most important decisions when configuring a BigQuery dataset. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Data location** -- Set `location` to a multi-region (`US`, `EU`) for maximum availability or a specific region (`us-central1`, `europe-west1`) for data residency compliance. The location is immutable after creation and determines where queries are processed.

**Storage billing model** -- `storageBillingModel` defaults to `LOGICAL` (charges per uncompressed bytes). Switch to `PHYSICAL` to charge per compressed bytes on disk, which can reduce storage costs 60-80% for highly compressible data. Evaluate by checking the `INFORMATION_SCHEMA.TABLE_STORAGE` view.

**Access control** -- The `access` field is authoritative: BigQuery removes any entries not listed in the spec. Omit it entirely to inherit default project-level access (owners/editors/viewers). When specifying explicit entries, always include `projectOwners` as OWNER to avoid locking out project administrators.

**Table lifecycle** -- `defaultTableExpirationMs` auto-deletes tables after a duration (minimum 1 hour). Useful for staging datasets where data is transient. `defaultPartitionExpirationMs` does the same for partitions in partitioned tables. Leave both unset for persistent data.

**Time travel** -- `maxTimeTravelHours` controls how far back point-in-time snapshots are available (48-168 hours, default 168). Lower values reduce storage costs but shorten the recovery window.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpKmsKey** (optional) | `kmsKeyName` | `status.outputs.key_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `dataset_id` | Short dataset ID used in SQL queries and API calls | GcpBigQueryTable `datasetId` references, application query configuration, dbt project settings |
| `self_link` | Fully qualified dataset URI | IAM bindings, audit log filters |
| `project` | GCP project containing the dataset | Cross-project dataset references |
| `creation_time` | Dataset creation time in milliseconds since epoch | Auditing and lifecycle tracking |
| `location` | Geographic location of the dataset | Co-locating downstream datasets and tables (every table a query joins must share it) |
| `etag` | Entity tag that changes on every metadata modification | Optimistic-concurrency API callers |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic analytics** -- US multi-region dataset with Google-managed encryption, default project-level access, and logical storage billing. The simplest starting point for analytics workloads. Start from the **Basic Analytics Dataset** preset.

**CMEK encrypted** -- Regional dataset with customer-managed encryption via Cloud KMS and physical storage billing. Designed for compliance scenarios requiring data residency and encryption key control. Start from the **CMEK-Encrypted Dataset** preset.

**Team shared** -- Dataset with explicit access control separating data engineers (WRITER) from data analysts (READER), plus project owners retaining OWNER access. Standard pattern for shared analytics datasets with role-based access. Start from the **Team-Shared Dataset** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the dataset is created
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- provides the Cloud KMS key for default table encryption (CMEK)
- [**GCP BigQuery Table**](/infra-catalog/gcp-big-query-table) -- infrastructure-owned tables, views, and external tables that reference this dataset's `dataset_id` output