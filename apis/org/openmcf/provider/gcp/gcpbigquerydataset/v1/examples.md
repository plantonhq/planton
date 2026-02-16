# GcpBigQueryDataset -- Examples

## Example 1: Minimal Analytics Dataset

The simplest configuration: a dataset in the US multi-region with default settings.
Suitable for development, prototyping, or non-sensitive workloads.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigQueryDataset
metadata:
  name: dev-analytics
spec:
  projectId:
    value: "my-gcp-project"
  datasetId: analytics_dev
  location: US
```

**Notes:**
- Google-managed encryption (default)
- 7-day time travel window (default)
- Default project-level access (owners=OWNER, editors=WRITER, viewers=READER)
- Tables do not auto-expire

## Example 2: Production Dataset with CMEK Encryption

For regulated workloads that require customer-managed encryption keys. All tables
created in this dataset are automatically encrypted with the specified KMS key.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigQueryDataset
metadata:
  name: prod-financial-data
spec:
  projectId:
    value: "my-prod-project"
  datasetId: financial_data_prod
  location: us-central1
  friendlyName: "Production Financial Data"
  description: "Financial reporting dataset -- PII and PCI scope"
  storageBillingModel: PHYSICAL
  kmsKeyName:
    value: "projects/my-prod-project/locations/us-central1/keyRings/prod-ring/cryptoKeys/bq-cmek"
```

**Notes:**
- Regional location (`us-central1`) for data residency
- PHYSICAL billing for cost savings on highly structured financial data
- CMEK encryption for compliance (PCI, SOX)
- No `deleteContentsOnDestroy` -- prevents accidental data loss

## Example 3: Team-Shared Dataset with Access Control

A dataset with explicit access control for a data analytics team. Demonstrates
role-based access for users, groups, and authorized views.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigQueryDataset
metadata:
  name: team-analytics
spec:
  projectId:
    value: "my-gcp-project"
  datasetId: analytics_shared
  location: EU
  friendlyName: "Shared Analytics"
  description: "Shared dataset for the data analytics team"
  access:
    - role: OWNER
      specialGroup: projectOwners
    - role: WRITER
      groupByEmail: "data-engineers@example.com"
    - role: READER
      groupByEmail: "data-analysts@example.com"
    - role: READER
      userByEmail: "external-consultant@partner.com"
```

**Notes:**
- EU multi-region for GDPR data residency
- Access is **authoritative** -- only the listed entries will exist
- Project owners retain OWNER access explicitly
- Engineers can write (create/update tables), analysts can only read

## Example 4: Cross-Resource Reference (Infra Chart Pattern)

When composing resources in an infra chart, use `valueFrom` to reference the
project and KMS key from other OpenMCF resources.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigQueryDataset
metadata:
  name: data-warehouse
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: analytics-project
      fieldPath: status.outputs.project_id
  datasetId: data_warehouse
  location: US
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: warehouse-cmek-key
      fieldPath: status.outputs.key_id
  maxTimeTravelHours: 168
  deleteContentsOnDestroy: true
```

**Notes:**
- `valueFrom` creates dependency edges -- project and KMS key are provisioned first
- `deleteContentsOnDestroy: true` allows clean teardown in development environments
- Maximum time travel (7 days) for recovery flexibility

## Example 5: Dataset with Table Auto-Expiration

For temporary or staging datasets where tables should be automatically cleaned up
after a retention period.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigQueryDataset
metadata:
  name: staging-events
spec:
  projectId:
    value: "my-gcp-project"
  datasetId: staging_events
  location: US
  friendlyName: "Staging Events"
  description: "Temporary event data -- tables auto-expire after 7 days"
  defaultTableExpirationMs: 604800000
  defaultPartitionExpirationMs: 604800000
  maxTimeTravelHours: 48
  deleteContentsOnDestroy: true
```

**Notes:**
- `defaultTableExpirationMs: 604800000` = 7 days (tables auto-deleted)
- `defaultPartitionExpirationMs: 604800000` = 7 days (partitions auto-deleted)
- `maxTimeTravelHours: 48` = 2 days (minimum, reduces storage costs)
- `deleteContentsOnDestroy: true` = clean teardown

## Example 6: Case-Insensitive Dataset

For workloads where case sensitivity in table and column names causes friction.
Once enabled, "MyTable" and "mytable" refer to the same table.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpBigQueryDataset
metadata:
  name: case-insensitive-ds
spec:
  projectId:
    value: "my-gcp-project"
  datasetId: case_insensitive_analytics
  location: US
  isCaseInsensitive: true
  defaultCollation: "und:ci"
```

**Notes:**
- `isCaseInsensitive` is **immutable** after creation -- cannot be changed later
- `defaultCollation: "und:ci"` makes string comparisons case-insensitive by default
- Useful for datasets migrated from case-insensitive database systems
