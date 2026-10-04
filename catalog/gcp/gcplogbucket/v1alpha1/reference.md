# GcpLogBucket

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpLogBucketSpec defines a Cloud Logging bucket — the container where
log entries are STORED: how long they are retained, whether they are
indexed for analytics, who can see which slice (log views), and how the
data links into BigQuery.

One kind covers all four bucket scopes. The `scope` message selects
where the bucket lives — a project (the default; an empty scope uses the
provider's ambient project), a folder, an organization, or a billing
account — and the module creates exactly one of the four provider
resources (google_logging_{project|folder|organization|billing_account}
_bucket_config).

ADOPTION SEMANTICS (provider truth, all scopes): creating a bucket
config whose bucket_id matches an EXISTING bucket adopts and patches it
rather than failing. That is the only way to manage GCP's built-in
`_Default` bucket (every project has one), and the only way at all on
folder/organization/billing scopes — the Logging API creates NEW custom
buckets only under projects; non-project bucket configs always adopt.
The built-in `_Default` and `_Required` buckets are undeletable: a
destroy removes them from management and leaves them in GCP.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpLogBucket
metadata:
  name: my-sample-log-bucket
spec:
  # Where the bucket lives. Omit entirely for a project bucket in the
  # provider's default project — the common case. (Folder, organization,
  # and billing-account scopes take folderId / organizationId /
  # billingAccount instead — those scopes ADOPT existing buckets; the
  # Logging API creates new custom buckets only under projects.)

  # The bucket ID — or "_Default" to adopt the project's built-in bucket
  # and manage its retention. Renaming REPLACES the bucket.
  bucketId: my-sample-log-bucket

  # Immutable; "global" (the default) unless data residency demands a
  # region.
  location: global

  # 1–3650 days; GCP's default is 30. On a LOCKED bucket this can never
  # change again. Lowering it deletes entries older than the new window.
  retentionDays: 45

  # Why this bucket exists and what lands in it.
  description: Sample bucket — error archive with a status-code index

  # Indexed LogEntry fields for faster targeted queries (at most 20).
  indexConfigs:
    - fieldPath: jsonPayload.request.status
      type: INDEX_TYPE_INTEGER

  # Named, filtered slices grantable independently
  # (roles/logging.viewAccessor on a view shows a reader only these
  # entries). viewId is permanent — renames replace the view. View
  # filters speak a RESTRICTED grammar: only log source, resource type,
  # apphub fields, user-defined labels, or LOG_ID() restrictions —
  # severity is NOT a legal view dimension.
  logViews:
    - viewId: run-stderr
      filter: resource.type="cloud_run_revision" AND LOG_ID("run.googleapis.com/stderr")
      description: The on-call slice — Cloud Run stderr without the noise

  # What a destroy does: DELETE (default — stored entries are deleted,
  # no recycle bin), PREVENT (the compliance posture), or ABANDON. Also
  # covers the bucket's views and linked dataset.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.scope` | `GcpLogBucketScope` |  |  |  |
| `spec.scope.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.scope.folderId` | `string \| valueFrom` |  |  | GcpFolder (`status.outputs.folder_id`) |
| `spec.scope.organizationId` | `string` |  |  |  |
| `spec.scope.billingAccount` | `string` |  |  |  |
| `spec.bucketId` | `string` | yes |  |  |
| `spec.location` | `string` |  | `global` |  |
| `spec.description` | `string` |  |  |  |
| `spec.retentionDays` | `int32` |  | `30` |  |
| `spec.locked` | `bool` |  |  |  |
| `spec.enableAnalytics` | `bool` |  |  |  |
| `spec.cmekKmsKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.indexConfigs` | `[]GcpLogBucketIndexConfig` |  |  |  |
| `spec.indexConfigs[].fieldPath` | `string` | yes |  |  |
| `spec.indexConfigs[].type` | `string` | yes |  |  |
| `spec.logViews` | `[]GcpLogBucketLogView` |  |  |  |
| `spec.logViews[].viewId` | `string` | yes |  |  |
| `spec.logViews[].filter` | `string` |  |  |  |
| `spec.logViews[].description` | `string` |  |  |  |
| `spec.linkedBigqueryDataset` | `GcpLogBucketLinkedDataset` |  |  |  |
| `spec.linkedBigqueryDataset.linkId` | `string` | yes |  |  |
| `spec.linkedBigqueryDataset.description` | `string` |  |  |  |
| `spec.scopeSettings` | `GcpLogBucketScopeSettings` |  |  |  |
| `spec.scopeSettings.disableDefaultSink` | `bool` |  |  |  |
| `spec.scopeSettings.kmsKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.scopeSettings.storageLocation` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.scope

`GcpLogBucketScope`

Where the bucket lives. Omit entirely for a project bucket in the
provider's default project — the common case. (Folder, organization,
and billing-account scopes take folder_id / organization_id /
billing_account instead.)

- rule: set at most one of project_id, folder_id, organization_id, or billing_account (empty means the provider's default project)

### spec.scope.projectId

`string | valueFrom`

Project bucket: the owning project — a literal project ID or a
reference to a GcpProject resource.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.scope.folderId

`string | valueFrom`

Folder bucket: the folder's numeric ID (with or without the
"folders/" prefix) -- a literal, or a reference to a GcpFolder resource
(its folder_id output). ADOPT-only: the Logging API creates new custom
buckets only under projects, so a folder bucket adopts one the folder
already has (its _Default or _Required bucket).

- references: GcpFolder (`status.outputs.folder_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFolder, name: <that resource's name>, fieldPath: status.outputs.folder_id}} -- a bare string does not parse

### spec.scope.organizationId

`string`

Organization bucket: the numeric organization ID. ADOPT-only.

### spec.scope.billingAccount

`string`

Billing-account bucket: the billing account ID
(e.g. 012345-6789AB-CDEF01). ADOPT-only.

### spec.bucketId

`string` · required

The bucket ID (the last segment of the bucket resource name), e.g.
"audit-logs" — or "_Default" to ADOPT the project's built-in default
bucket and manage its retention. Changing it REPLACES the bucket.

- rule: {"required":true}

### spec.location

`string`

The bucket location (e.g. "global", "us-central1", "eu"). Immutable —
changing it REPLACES the bucket (and its stored logs). "global" is the
default and the right choice unless data residency demands a region.
Both IaC engines send the value explicitly.

- default: `global`

### spec.description

`string`

Why this bucket exists and what lands in it.

### spec.retentionDays

`int32`

How many days log entries are kept before automatic deletion. GCP's
default is 30; up to 3650 (10 years). On a LOCKED bucket the retention
period can no longer be changed. Both IaC engines send the value
explicitly so the spec default (30) is what GCP applies rather than a
silently different server-side state.

- default: `30`
- rule: retention_days must be between 1 and 3650

### spec.locked

`bool`

Lock the bucket (project scope only). LOCKING IS ONE-WAY: a locked
bucket's retention policy can never be changed or unlocked again, and
the bucket can only be deleted once every entry in it has aged out of
retention. The compliance posture — enable deliberately.

### spec.enableAnalytics

`bool` · optional (explicit presence)

Enable Log Analytics on the bucket (project scope only): entries
become queryable with SQL from the Log Analytics UI and BigQuery
(via linked_bigquery_dataset). ONE-WAY per the provider: analytics
cannot be disabled once enabled. Left unset, nothing is sent — GCP
keeps its own default (disabled) — because analytics enablement is an
atomic, separate API operation the provider performs only on explicit
configuration.

### spec.cmekKmsKey

`string | valueFrom`

Encrypt the bucket with a customer-managed KMS key (CMEK) instead of
Google-managed encryption. The full crypto key resource name
(projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}) — a literal
or a reference to a GcpKmsKey resource. ONE-WAY per the provider:
CMEK cannot be disabled once set (rotating to a DIFFERENT key is
allowed). PREREQUISITE: grant the Logging service account
roles/cloudkms.cryptoKeyEncrypterDecrypter on the key first (find the
account with `gcloud logging settings describe`), or bucket creation
fails.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.indexConfigs

`[]GcpLogBucketIndexConfig`

Indexed LogEntry fields for faster targeted queries — at most 20 (the
provider's own cap).

- rule: {"repeated":{"maxItems":"20"}}

### spec.indexConfigs[].fieldPath

`string` · required

The LogEntry field path to index, e.g. "jsonPayload.request.status".

- rule: {"required":true}

### spec.indexConfigs[].type

`string` · required

The indexed data type: INDEX_TYPE_STRING or INDEX_TYPE_INTEGER (the
Logging API's documented values — server-validated).

- rule: {"required":true}

### spec.logViews

`[]GcpLogBucketLogView`

Log VIEWS on the bucket: named, filtered slices that can be granted
independently (roles/logging.viewAccessor on a view shows a reader
only the entries matching its filter). Each becomes a
google_logging_log_view resource on this bucket.

### spec.logViews[].viewId

`string` · required

The view ID (the last segment of the view resource name). Changing it
REPLACES the view (views are immutable in name and home; only filter
and description update in place).

- rule: {"required":true}

### spec.logViews[].filter

`string`

The filter selecting which of the bucket's entries this view exposes.
Empty exposes every entry in the bucket.

View filters speak a RESTRICTED grammar — NOT the general log filter
language: the API accepts only restrictions on log source
(source()), resource type (resource.type=...), apphub fields,
user-defined labels, and log ID (LOG_ID(...)). Severity is NOT a
legal view dimension — GCP rejects it at create with "Invalid view
filter" (live-verified), even though the same expression is legal in
sinks and log-based metrics.

### spec.logViews[].description

`string`

What this view is for and who should be granted it.

### spec.linkedBigqueryDataset

`GcpLogBucketLinkedDataset`

Link a read-only BigQuery dataset to the bucket so its entries are
queryable from BigQuery directly (requires enable_analytics: true).

### spec.linkedBigqueryDataset.linkId

`string` · required

The link ID — becomes the BigQuery DATASET ID, so it must be a valid
dataset name (letters, numbers, underscores). IMMUTABLE: every field
of a linked dataset (including the description) is create-time-only;
any change REPLACES the link.

- rule: {"required":true}

### spec.linkedBigqueryDataset.description

`string`

What the linked dataset is for. Immutable like the rest of the link.

### spec.scopeSettings

`GcpLogBucketScopeSettings`

Folder/organization scopes only: the scope's LOGGING SETTINGS —
default-sink disable, default CMEK, and default storage location for
buckets created under the scope. A singleton that always exists in
GCP: creating this surface ADOPTS it, and destroying it is a
state-only no-op (the settings object cannot be deleted).

### spec.scopeSettings.disableDefaultSink

`bool`

Disable the scope's automatic _Default sink so logs of NEW child
projects are not copied into their _Default buckets — the
centralized-logging posture (pair it with an aggregating
GcpLoggingSink at the same scope).

### spec.scopeSettings.kmsKey

`string | valueFrom`

The default CMEK key applied to buckets created under the scope. The
full crypto key resource name — a literal or a reference to a
GcpKmsKey resource. The same one-way and grant-first caveats as the
bucket-level cmek_kms_key apply.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.scopeSettings.storageLocation

`string`

The default storage location for buckets created under the scope
(e.g. "us-central1", "eu").

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the bucket (and its stored entries) is deleted; the
               built-in _Default/_Required buckets are undeletable and
               are simply removed from management
  "PREVENT" -- destroy FAILS; protects compliance-mandated log
               storage from accidental teardown
  "ABANDON" -- the bucket is removed from management but keeps
               storing logs in GCP
Also applied to the bucket's log views and linked dataset.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `locked_is_project_scope_only`: locked applies only to project-scoped buckets — the folder/organization/billing bucket resources do not carry the argument
- `analytics_is_project_scope_only`: enable_analytics applies only to project-scoped buckets — the folder/organization/billing bucket resources do not carry the argument
- `linked_dataset_requires_analytics`: linked_bigquery_dataset requires enable_analytics: true (the Logging API links datasets only to analytics-enabled buckets)
- `scope_settings_need_folder_or_org_scope`: scope_settings applies only to folder or organization scoped buckets (google_logging_{folder|organization}_settings)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpLogBucket, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.bucket_name` | `string` | The full resource name of the bucket, e.g. projects/{project}/locations/{location}/buckets/{bucket_id} (or the folders/organizations/billingAccounts form for those scopes). THE composition handle: a GcpLoggingSink routes into this bucket with destination raw_uri "logging.googleapis.com/{bucket_name}", and a bucket-scoped GcpLogMetric references it directly. |
| `status.outputs.linked_dataset_id` | `string` | The BigQuery dataset ID of the linked dataset, when linked_bigquery_dataset is configured (empty otherwise). Query the bucket's entries from BigQuery through this dataset. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.scope.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.scope.folderId` | GcpFolder | `status.outputs.folder_id` |
| `spec.cmekKmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.scopeSettings.kmsKey` | GcpKmsKey | `status.outputs.key_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpLogMetric | `spec.bucketName` | `status.outputs.bucket_name` |

## See Also

- [Overview](../README.md)
