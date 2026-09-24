# GcpBigQueryDataset

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpBigQueryDatasetSpec defines the configuration for a GCP BigQuery
dataset.

A dataset is the top-level container for tables, views, and routines in
BigQuery. It pins the data's geographic location and owns the defaults
every contained table inherits: lifecycle (expiration), encryption
(CMEK), collation, and the dataset-level ACL. Tables that infrastructure
should own (partitioned fact tables, authorized views, external tables)
are first-class GcpBigQueryTable resources that reference this dataset;
application-owned schemas (dbt models, migrations) can keep managing
their own tables inside the same dataset.

Important behavioral notes:

  - dataset_id, location, and is_case_insensitive are IMMUTABLE —
    changing any of them recreates the dataset (and everything in it).

  - The access field is AUTHORITATIVE: the entries listed here become the
    dataset's complete ACL, and BigQuery removes anything else. If access
    is omitted entirely, BigQuery applies its defaults (project owners =
    OWNER, editors = WRITER, viewers = READER). To keep those defaults
    alongside custom grants, list them explicitly.

  - If delete_contents_on_destroy is false (default) and the dataset
    contains tables, destroy fails — the safety net against deleting
    data with the container. Set true only when contained data is
    intentionally disposable.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryDataset
metadata:
  name: test-bq-dataset
spec:
  projectId:
    value: "test-project"
  datasetId: test_analytics
  location: US
  friendlyName: Test Analytics
  description: Hack manifest exercising the dataset surface offline
  labels:
    team: platform
  defaultTableExpirationMs: 86400000
  maxTimeTravelHours: 96
  storageBillingModel: PHYSICAL
  access:
    - role: OWNER
      specialGroup: projectOwners
    - role: READER
      userByEmail: analyst@example.com
    - view:
        projectId: test-project
        datasetId: shared_views
        tableId: revenue_summary
  # Destroy really destroys in E2E: the live lanes prove the full lifecycle.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.datasetId` | `string` | yes |  |  |
| `spec.location` | `string` | yes |  |  |
| `spec.friendlyName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.resourceTags` | `map<string, string>` |  |  |  |
| `spec.defaultTableExpirationMs` | `int64` |  |  |  |
| `spec.defaultPartitionExpirationMs` | `int64` |  |  |  |
| `spec.maxTimeTravelHours` | `int32` |  |  |  |
| `spec.isCaseInsensitive` | `bool` |  |  |  |
| `spec.defaultCollation` | `string` |  |  |  |
| `spec.storageBillingModel` | `string` |  |  |  |
| `spec.deleteContentsOnDestroy` | `bool` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.access` | `[]GcpBigQueryDatasetAccessEntry` |  |  |  |
| `spec.access[].role` | `string` |  |  |  |
| `spec.access[].userByEmail` | `string` |  |  |  |
| `spec.access[].groupByEmail` | `string` |  |  |  |
| `spec.access[].domain` | `string` |  |  |  |
| `spec.access[].specialGroup` | `string` |  |  |  |
| `spec.access[].iamMember` | `string` |  |  |  |
| `spec.access[].view` | `GcpBigQueryDatasetAccessView` |  |  |  |
| `spec.access[].view.projectId` | `string` | yes |  |  |
| `spec.access[].view.datasetId` | `string` | yes |  |  |
| `spec.access[].view.tableId` | `string` | yes |  |  |
| `spec.access[].routine` | `GcpBigQueryDatasetAccessRoutine` |  |  |  |
| `spec.access[].routine.projectId` | `string` | yes |  |  |
| `spec.access[].routine.datasetId` | `string` | yes |  |  |
| `spec.access[].routine.routineId` | `string` | yes |  |  |
| `spec.access[].dataset` | `GcpBigQueryDatasetAccessDataset` |  |  |  |
| `spec.access[].dataset.projectId` | `string` | yes |  |  |
| `spec.access[].dataset.datasetId` | `string` | yes |  |  |
| `spec.access[].dataset.targetTypes` | `[]string` | yes |  |  |
| `spec.access[].condition` | `GcpBigQueryDatasetAccessCondition` |  |  |  |
| `spec.access[].condition.expression` | `string` | yes |  |  |
| `spec.access[].condition.title` | `string` |  |  |  |
| `spec.access[].condition.description` | `string` |  |  |  |
| `spec.access[].condition.location` | `string` |  |  |  |
| `spec.externalDatasetReference` | `GcpBigQueryDatasetExternalDatasetReference` |  |  |  |
| `spec.externalDatasetReference.externalSource` | `string` | yes |  |  |
| `spec.externalDatasetReference.connection` | `string` | yes |  |  |
| `spec.externalCatalogOptions` | `GcpBigQueryDatasetExternalCatalogOptions` |  |  |  |
| `spec.externalCatalogOptions.defaultStorageLocationUri` | `string` |  |  |  |
| `spec.externalCatalogOptions.parameters` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the dataset is created in. Accepts a literal project ID
or a reference to a GcpProject resource. If omitted, the provider's
default project is used.
Immutable: changing the project destroys and recreates the dataset.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.datasetId

`string` · required

Unique identifier for the dataset within the project. Immutable.
Must contain only letters (upper/lower), numbers, and underscores;
maximum 1024 characters. This is the value SQL queries and downstream
tables reference (e.g. SELECT * FROM `project.dataset.table`).
Example: "analytics_prod", "raw_events"

- rule: {"required":true,"string":{"maxLen":"1024","pattern":"^[0-9A-Za-z_]+$"}}

### spec.location

`string` · required

Geographic location where the dataset and its tables reside. Immutable
— moving data requires a new dataset and a copy job. Use multi-regional
locations ("US", "EU") for maximum availability, or regional locations
("us-central1", "europe-west1") for data-residency requirements. Every
table a query joins must live in the same location.

- rule: {"required":true}

### spec.friendlyName

`string`

User-friendly display name for the dataset.

### spec.description

`string`

Description of the dataset's contents or purpose.

### spec.labels

`map<string, string>`

Labels applied to the dataset for cost attribution and organization.
Merged with Planton's platform labels (which win on key conflicts).

### spec.resourceTags

`map<string, string>`

Resource Manager tags bound to the dataset, as
"tagKeys-namespaced-name" -> "tagValue short name" pairs (e.g.
"123456789012/environment" -> "production"). Unlike labels, tags
participate in IAM conditions and organization policy.

### spec.defaultTableExpirationMs

`int64`

Default lifetime for all NEW tables created in the dataset, in
milliseconds (existing tables keep their expiration). Each table is
deleted this long after its creation time. Minimum: 3600000 (1 hour).
If not set (0), tables do not automatically expire.

- rule: default_table_expiration_ms must be 0 (no expiration) or at least 3600000 (1 hour)

### spec.defaultPartitionExpirationMs

`int64`

Default expiration for partitions in NEW partitioned tables, in
milliseconds. If not set (0), partitions do not automatically expire.

### spec.maxTimeTravelHours

`int32`

Maximum hours of time travel for the dataset (point-in-time recovery
window). Range: 48 to 168 hours (2 to 7 days); GCP defaults to 168
when unset. Lower values reduce storage cost on PHYSICAL billing;
higher values give longer accidental-delete recovery.

- rule: max_time_travel_hours must be 0 (use default 168) or between 48 and 168

### spec.isCaseInsensitive

`bool`

Whether dataset and table names are treated as case-insensitive
("MyTable" and "mytable" become the same table). Immutable.
Default: false (case-sensitive).

### spec.defaultCollation

`string`

Default collation specification for string columns in new tables.
"und:ci" makes string comparison case-insensitive; empty keeps the
default case-sensitive collation.

### spec.storageBillingModel

`string`

Storage billing model for the dataset.
"LOGICAL" (GCP default) bills on uncompressed bytes; "PHYSICAL" bills
on compressed bytes on disk — often 60-80% cheaper for compressible
data, but time-travel storage is billed too. Switching to PHYSICAL is
allowed once every 14 days.

- rule: storage_billing_model must be LOGICAL or PHYSICAL

### spec.deleteContentsOnDestroy

`bool`

If true, destroying the dataset also deletes every table in it. If
false (default), destroy fails while the dataset contains tables —
the guard against deleting data with its container.

### spec.kmsKeyName

`string | valueFrom`

Cloud KMS key for default table encryption (CMEK). All NEW tables are
encrypted with this key unless they specify their own key. The
BigQuery service agent must hold roles/cloudkms.cryptoKeyEncrypterDecrypter
on the key before the first table is written. Format:
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
If not set, tables use Google-managed encryption.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.access

`[]GcpBigQueryDatasetAccessEntry`

Access control entries for the dataset.

AUTHORITATIVE: the entries listed here become the dataset's complete
ACL; BigQuery removes anything not listed. If omitted entirely,
BigQuery applies default access (project owners = OWNER, editors =
WRITER, viewers = READER) — list those explicitly to keep them
alongside custom grants.

Each entry is either a principal grant (role + one identity) or a
resource authorization (view / routine / dataset, no role).

- rule: exactly one of user_by_email, group_by_email, domain, special_group, iam_member, view, routine, or dataset must be set
- rule: role is required for principal grants and must be omitted for view, routine, and dataset authorizations

### spec.access[].role

`string`

The role to grant to a principal. Use basic dataset roles (OWNER,
WRITER, READER) or predefined IAM roles (roles/bigquery.dataOwner,
roles/bigquery.dataEditor, roles/bigquery.dataViewer). BigQuery stores
basic and predefined forms interchangeably (OWNER and
roles/bigquery.dataOwner are the same grant).

Required for principal grants; must be omitted for view/routine/dataset
authorizations (those carry implicit read access).

### spec.access[].userByEmail

`string`

An email address of a Google Account to grant access to.

### spec.access[].groupByEmail

`string`

An email address of a Google Group to grant access to.

### spec.access[].domain

`string`

A domain to grant access to. All users signed in with the domain's
account will be granted access (e.g., "example.com").

### spec.access[].specialGroup

`string`

A special group to grant access to. Valid values:
  "projectOwners"  -- all project owners
  "projectReaders" -- all project viewers
  "projectWriters" -- all project editors
  "allAuthenticatedUsers" -- all authenticated Google accounts

### spec.access[].iamMember

`string`

An IAM member expression to grant access to.
Examples: "allUsers", "serviceAccount:sa@project.iam.gserviceaccount.com"

### spec.access[].view

`GcpBigQueryDatasetAccessView`

A view that is authorized to read the dataset's data (no role).

### spec.access[].view.projectId

`string` · required

The GCP project that contains the view's dataset.

- rule: {"required":true}

### spec.access[].view.datasetId

`string` · required

The dataset that contains the view.

- rule: {"required":true}

### spec.access[].view.tableId

`string` · required

The ID of the authorized view (a table resource whose type is VIEW).

- rule: {"required":true}

### spec.access[].routine

`GcpBigQueryDatasetAccessRoutine`

A routine (UDF / stored procedure) that is authorized to read the
dataset's data (no role).

### spec.access[].routine.projectId

`string` · required

The GCP project that contains the routine's dataset.

- rule: {"required":true}

### spec.access[].routine.datasetId

`string` · required

The dataset that contains the routine.

- rule: {"required":true}

### spec.access[].routine.routineId

`string` · required

The ID of the authorized routine.

- rule: {"required":true}

### spec.access[].dataset

`GcpBigQueryDatasetAccessDataset`

Another dataset whose resources are authorized to read this dataset's
data (no role).

### spec.access[].dataset.projectId

`string` · required

The GCP project that contains the grantee dataset.

- rule: {"required":true}

### spec.access[].dataset.datasetId

`string` · required

The grantee dataset whose resources are authorized to read this
dataset's data.

- rule: {"required":true}

### spec.access[].dataset.targetTypes

`[]string` · required

Which resource types in the grantee dataset receive access.
Currently the BigQuery API supports "VIEWS".

- rule: {"repeated":{"minItems":"1","items":{"string":{"in":["VIEWS"]}}}}

### spec.access[].condition

`GcpBigQueryDatasetAccessCondition`

Optional IAM condition gating a principal grant (for example, a
time-bounded grant). Applies to principal grants only.

### spec.access[].condition.expression

`string` · required

The condition expression in IAM Common Expression Language, e.g.
request.time < timestamp("2030-01-01T00:00:00Z").

- rule: {"required":true}

### spec.access[].condition.title

`string`

Optional short title summarizing the condition's purpose.

### spec.access[].condition.description

`string`

Optional longer description of the condition.

### spec.access[].condition.location

`string`

Optional string indicating the location of the expression for error
reporting (for example, a file and position in that file).

### spec.externalDatasetReference

`GcpBigQueryDatasetExternalDatasetReference`

Makes this dataset a read-only projection of an external source
(e.g. an AWS Glue database) through a BigQuery Omni connection,
instead of a container for BigQuery-managed tables. Immutable.

### spec.externalDatasetReference.externalSource

`string` · required

The external source this dataset mirrors, e.g.
aws-glue://arn:aws:glue:us-east1:1234567:database/database1

- rule: {"required":true}

### spec.externalDatasetReference.connection

`string` · required

The connection (with credentials for the external source) used to read
it. Format: projects/{project}/locations/{location}/connections/{id}

- rule: {"required":true}

### spec.externalCatalogOptions

`GcpBigQueryDatasetExternalCatalogOptions`

Open-source catalog (Hive Metastore compatibility) metadata, letting
engines like Spark address this dataset as a Hive database.

### spec.externalCatalogOptions.defaultStorageLocationUri

`string`

Default storage location for tables in this catalog database, e.g.
gs://bucket/path. Maximum 1024 characters.

- rule: {"string":{"maxLen":"1024"}}

### spec.externalCatalogOptions.parameters

`map<string, string>`

Hive-database-style key/value parameters (the whole map is limited to
30 KB by the API).

### spec.deletionPolicy

`string`

What destroying this resource does to the dataset. Works alongside
delete_contents_on_destroy (which decides whether a NON-EMPTY dataset
may be removed); this switch decides whether removal is attempted at
all:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the dataset is deleted (GCP refuses while tables remain
               unless delete_contents_on_destroy is true)
  "PREVENT" -- destroy FAILS; protects a dataset other projects'
               queries and authorized views depend on
  "ABANDON" -- the dataset is removed from management but keeps
               serving queries in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpBigQueryDataset, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.dataset_id` | `string` | The short dataset ID (same as the spec's dataset_id input). This is the identifier used in BigQuery SQL queries, job configurations, and API calls (e.g., SELECT * FROM `project.dataset.table`). |
| `status.outputs.self_link` | `string` | The fully qualified URI of the dataset. Format: https://bigquery.googleapis.com/bigquery/v2/projects/{project}/datasets/{dataset} |
| `status.outputs.project` | `string` | The GCP project that contains this dataset. Useful when referencing the dataset from a different project context or wiring into downstream resources that need explicit project references. |
| `status.outputs.creation_time` | `int64` | The creation time of the dataset in milliseconds since epoch. |
| `status.outputs.location` | `string` | Geographic location of the dataset (e.g. "US", "europe-west1"). Every table a query joins must share this location, so downstream resources read it to co-locate. |
| `status.outputs.etag` | `string` | The dataset's current entity tag, changing on every metadata modification. Useful for optimistic-concurrency API callers. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpBigQueryTable | `spec.datasetId` | `status.outputs.dataset_id` |
| GcpDatastreamStream | `spec.destinationConfig.bigqueryDestinationConfig.singleTargetDataset.datasetId` | `status.outputs.self_link` |
| GcpGkeCluster | `spec.resourceUsageExport.bigqueryDatasetId` | `status.outputs.dataset_id` |
| GcpLoggingSink | `spec.destination.bigqueryDataset` | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)
