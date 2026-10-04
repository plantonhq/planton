# GcpLoggingSink

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpLoggingSinkSpec defines a Cloud Logging sink — the routing rule that
exports log entries matching a filter to a destination (a GCS bucket, a
BigQuery dataset, a Pub/Sub topic, or any other Logging-supported
destination URI).

One kind covers all four sink scopes. The `scope` message selects where
the sink lives — a project (the default; an empty scope uses the
provider's ambient project), a folder, an organization, or a billing
account — and the module creates the matching GCP resource
(google_logging_{project|folder|organization|billing_account}_sink).

THE post-create step every sink needs: GCP mints a `writer_identity`
service account for the sink, and that identity must be GRANTED write
access on the destination or the sink silently exports nothing. The
writer_identity output (already in IAM member form,
serviceAccount:{email}) exists exactly for that wiring. Grant it with a
standalone grant that depends on both the sink and the destination,
its member referencing this sink's writer_identity output:
  GCS bucket       -- a GcpGcsBucketIamMember, role
                      roles/storage.objectCreator
  Pub/Sub topic    -- a GcpPubSubTopicIamMember, role
                      roles/pubsub.publisher
  BigQuery dataset -- roles/bigquery.dataEditor, granted outside the
                      catalog or as a literal entry in the
                      GcpBigQueryDataset's access list once the identity
                      is known (the dataset's access list is
                      authoritative, so there is no separate grant kind)
Never grant it through the destination kind's own IAM fields: the sink
references its destination, so a destination that references the
sink's identity back is a dependency cycle.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpLoggingSink
metadata:
  name: my-sample-logging-sink
  # GcpGcsBucket is this kind's REGISTRY prerequisite — the harness deploys
  # the bucket fixture first and resolves the reference below; no
  # annotation needed.
spec:
  # Where the sink lives. Omit entirely for a project sink in the
  # provider's default project — the common case. (Folder, organization,
  # and billing-account scopes take folderId / organizationId /
  # billingAccount instead.)

  # The sink name; omit to default to metadata.name. Renaming recreates
  # the sink and mints a NEW writer identity — re-grant the destination.
  sinkName: my-sample-logging-sink

  # Where matching entries export — exactly one arm. The module renders
  # the service-scheme URI (storage.googleapis.com/...) so the manifest
  # references the resource naturally.
  destination:
    gcsBucket:
      valueFrom:
        kind: GcpGcsBucket
        # The bucket kind's PUBLISHED prerequisite fixture name.
        name: planton-oss-e2e-gcsb-prereq
        fieldPath: status.outputs.bucket_id

  # Which entries export. Empty exports EVERYTHING in scope — deliberate
  # for audit archives, expensive otherwise.
  filter: severity>=ERROR

  # Why this sink exists and where its data lands.
  description: Archives error-level logs for compliance review

  # Carve noisy sub-streams out of the export; sample() drops a
  # percentage of matches.
  exclusions:
    - name: drop-health-checks
      filter: httpRequest.requestUrl:"/healthz"
      description: Health-check noise is never evidence

  # Project scope only (default true; other scopes always mint a unique
  # writer). Required true for BigQuery destinations and cross-project
  # export.
  uniqueWriterIdentity: true

  # What a destroy does: DELETE (default), PREVENT, or ABANDON.
  # PREVENT is the honest posture for compliance-mandated exports.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.scope` | `GcpLoggingSinkScope` |  |  |  |
| `spec.scope.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.scope.folderId` | `string \| valueFrom` |  |  | GcpFolder (`status.outputs.folder_id`) |
| `spec.scope.organizationId` | `string` |  |  |  |
| `spec.scope.billingAccount` | `string` |  |  |  |
| `spec.sinkName` | `string` |  |  |  |
| `spec.destination` | `GcpLoggingSinkDestination` | yes |  |  |
| `spec.destination.gcsBucket` | `string \| valueFrom` |  |  | GcpGcsBucket (`status.outputs.bucket_id`) |
| `spec.destination.bigqueryDataset` | `string \| valueFrom` |  |  | GcpBigQueryDataset (`status.outputs.self_link`) |
| `spec.destination.usePartitionedTables` | `bool` |  |  |  |
| `spec.destination.pubsubTopic` | `string \| valueFrom` |  |  | GcpPubSubTopic (`status.outputs.topic_id`) |
| `spec.destination.rawUri` | `string` |  |  |  |
| `spec.filter` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.disabled` | `bool` |  |  |  |
| `spec.exclusions` | `[]GcpLoggingSinkExclusion` |  |  |  |
| `spec.exclusions[].name` | `string` | yes |  |  |
| `spec.exclusions[].filter` | `string` | yes |  |  |
| `spec.exclusions[].description` | `string` |  |  |  |
| `spec.exclusions[].disabled` | `bool` |  |  |  |
| `spec.includeChildren` | `bool` |  |  |  |
| `spec.interceptChildren` | `bool` |  |  |  |
| `spec.uniqueWriterIdentity` | `bool` |  | `true` |  |
| `spec.customWriterIdentity` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.scope

`GcpLoggingSinkScope`

Where the sink lives. Omit entirely for a project sink in the
provider's default project — the common case.

- rule: set at most one of project_id, folder_id, organization_id, or billing_account (empty means the provider's default project)

### spec.scope.projectId

`string | valueFrom`

Project sink: the owning project — a literal project ID or a reference
to a GcpProject resource.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.scope.folderId

`string | valueFrom`

Folder sink: the folder's numeric ID (with or without the "folders/"
prefix) -- a literal, or a reference to a GcpFolder resource (its
folder_id output). The sink lives in the folder and, with
include_children, sees every project and folder beneath it.

- references: GcpFolder (`status.outputs.folder_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFolder, name: <that resource's name>, fieldPath: status.outputs.folder_id}} -- a bare string does not parse

### spec.scope.organizationId

`string`

Organization sink: the numeric organization ID.

### spec.scope.billingAccount

`string`

Billing-account sink: the billing account ID
(e.g. 012345-6789AB-CDEF01).

### spec.sinkName

`string`

The sink name in GCP. Defaults to metadata.name when left empty.
Immutable: changing it destroys and recreates the sink (a new
writer_identity is minted — re-grant destination access).

### spec.destination

`GcpLoggingSinkDestination` · required

Where matching log entries are exported. Exactly one destination arm.

- rule: {"required":true}
- rule: set exactly one destination: gcs_bucket, bigquery_dataset, pubsub_topic, or raw_uri
- rule: use_partitioned_tables applies only to a bigquery_dataset destination

### spec.destination.gcsBucket

`string | valueFrom`

Export to a Cloud Storage bucket (hourly batches of JSON files). The
bucket NAME — a literal or a reference to a GcpGcsBucket resource.
Rendered as storage.googleapis.com/{bucket}. Grant the sink's
writer_identity roles/storage.objectCreator on the bucket with a
GcpGcsBucketIamMember whose member references the writer_identity
output.

- references: GcpGcsBucket (`status.outputs.bucket_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.destination.bigqueryDataset

`string | valueFrom`

Export to a BigQuery dataset (near-real-time, queryable). Accepts the
dataset SELF LINK (https://bigquery.googleapis.com/bigquery/v2/projects/
{p}/datasets/{d} — the GcpBigQueryDataset self_link output) or a bare
projects/{p}/datasets/{d} path; the module normalizes either into the
bigquery.googleapis.com/... destination URI. Grant the writer_identity
roles/bigquery.dataEditor on the dataset -- outside the catalog, or as
a literal entry in the GcpBigQueryDataset's access list once the
identity is known (a reference there would make the dataset depend on
the sink that depends on it). The reference is
containment-exempt: the sink EXPORTS INTO the dataset, it does not live
inside it (the sink's home is its scope).

- references: GcpBigQueryDataset (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBigQueryDataset, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.destination.usePartitionedTables

`bool`

BigQuery destinations only: write into date-partitioned tables
(recommended — enables partition pruning and expiration) instead of
date-sharded table names. Requires unique_writer_identity true (the
provider's own constraint).

### spec.destination.pubsubTopic

`string | valueFrom`

Export to a Pub/Sub topic (streaming; the front door to third-party
log pipelines). The full topic path projects/{p}/topics/{t} — a literal
or a reference to a GcpPubSubTopic resource (its topic_id output).
Rendered as pubsub.googleapis.com/projects/{p}/topics/{t}. Grant the
writer_identity roles/pubsub.publisher on the topic with a
GcpPubSubTopicIamMember whose member references the writer_identity
output.

- references: GcpPubSubTopic (`status.outputs.topic_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.destination.rawUri

`string`

Escape hatch: a complete destination URI passed through verbatim, for
destinations without a first-class arm — chiefly Cloud Logging buckets
(logging.googleapis.com/projects/{p}/locations/{l}/buckets/{b}) until a
dedicated log-bucket kind exists, and cross-project Logging API
destinations.

### spec.filter

`string`

The logs filter selecting entries to export
(https://cloud.google.com/logging/docs/view/logging-query-language),
e.g. severity>=ERROR or resource.type="cloud_run_revision".
Empty exports EVERYTHING in scope — deliberate for audit archives,
expensive for anything else.

### spec.description

`string`

Why this sink exists and where its data lands — write it for the
operator auditing log routing later. At most 8000 characters.

- rule: {"string":{"maxLen":"8000"}}

### spec.disabled

`bool`

If true, the sink keeps its configuration but exports nothing — the
safe way to pause an export without losing the writer identity and
destination grants.

### spec.exclusions

`[]GcpLoggingSinkExclusion`

Log entries matching ANY of these exclusion filters are NOT exported,
even when they match `filter` — carve noisy sub-streams (health checks,
debug logs) out of a broad export.

- rule: {"repeated":{"maxItems":"50"}}

### spec.exclusions[].name

`string` · required

Identifier for this exclusion (letters, digits, underscores, hyphens,
periods; must start alphanumeric; at most 100 characters).

- rule: exclusion name must start with a letter or digit, use only letters, digits, underscores, hyphens, and periods, and be at most 100 characters
- rule: {"required":true}

### spec.exclusions[].filter

`string` · required

The logs filter selecting entries to EXCLUDE. Use the sample()
function to exclude only a percentage (e.g.
sample(insertId, 0.9) drops 90% of matches).

- rule: {"required":true}

### spec.exclusions[].description

`string`

What this exclusion carves out and why.

### spec.exclusions[].disabled

`bool`

If true, this exclusion is ignored (its entries export normally) —
stage or pause a carve-out without deleting it.

### spec.includeChildren

`bool`

Folder/organization scopes only: also export logs from all CHILD
resources (subfolders and projects). Off by default — the sink then
sees only logs at the scope itself.

### spec.interceptChildren

`bool`

Folder/organization scopes only (requires include_children semantics):
INTERCEPT matching logs — they are exported by this sink and NOT routed
onward to the children's own sinks. For centralized compliance capture;
use deliberately, it changes what child projects see.

### spec.uniqueWriterIdentity

`bool` · optional (explicit presence)

Project scope only. If true (default), GCP mints a dedicated service
account as the sink's writer identity — required for exporting across
projects and for BigQuery destinations. Setting false uses the legacy
shared cloud-logs@ account (single-project GCS/Pub/Sub only; not
recommended). Folder/organization/billing sinks always get a unique
writer. Both IaC engines send the value explicitly on project sinks so
behavior is identical regardless of engine.

- default: `true`

### spec.customWriterIdentity

`string`

Project scope only: use a caller-provided service account as the
writer identity instead of a GCP-minted one (format:
serviceAccount:{email} is NOT included — pass the bare email). Requires
unique_writer_identity true semantics; GCP rejects the combination
with false.

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the sink is deleted; export stops immediately (already
               exported data stays in the destination)
  "PREVENT" -- destroy FAILS; protects a compliance-mandated export
               pipeline from accidental teardown
  "ABANDON" -- the sink is removed from management but keeps exporting
               in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `children_flags_need_folder_or_org_scope`: include_children and intercept_children apply only to folder or organization scoped sinks
- `writer_identity_controls_are_project_scope_only`: unique_writer_identity=false and custom_writer_identity apply only to project-scoped sinks — other scopes always create a unique writer identity
- `bigquery_destination_requires_unique_writer`: a BigQuery destination requires unique_writer_identity to remain true

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpLoggingSink, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.sink_name` | `string` | The sink name as it exists in GCP. |
| `status.outputs.writer_identity` | `string` | The service-account identity GCP minted (or adopted) for this sink — format "serviceAccount:{email}", exactly the value an IAM member field takes. THE chart output: grant this identity write access on the destination, or the sink silently exports nothing -- through a standalone grant whose member references this output: a GcpGcsBucketIamMember with roles/storage.objectCreator on a bucket, a GcpPubSubTopicIamMember with roles/pubsub.publisher on a topic. A BigQuery dataset takes roles/bigquery.dataEditor outside the catalog or as a literal entry in its access list once the identity is known. Never through the destination kind's own IAM fields: the sink already depends on the destination, so that grant would be a cycle. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.scope.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.scope.folderId` | GcpFolder | `status.outputs.folder_id` |
| `spec.destination.gcsBucket` | GcpGcsBucket | `status.outputs.bucket_id` |
| `spec.destination.bigqueryDataset` | GcpBigQueryDataset | `status.outputs.self_link` |
| `spec.destination.pubsubTopic` | GcpPubSubTopic | `status.outputs.topic_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpGcsBucketIamMember | `spec.member` | `status.outputs.writer_identity` |
| GcpPubSubTopicIamMember | `spec.member` | `status.outputs.writer_identity` |

## See Also

- [Overview](../README.md)
