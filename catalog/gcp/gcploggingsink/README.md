# GCP Logging Sink

Creates a Cloud Logging sink — the routing rule that exports log entries matching a filter to a Cloud Storage bucket, BigQuery dataset, or Pub/Sub topic. One kind covers all four scopes (project, folder, organization, billing account), the module renders the destination URI from a natural resource reference, and the sink's writer identity is surfaced as THE output to grant on the destination — the mis-wiring this kind exists to remove.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Logging Sink** -- exactly one of `logging.ProjectSink` / `FolderSink` / `OrganizationSink` / `BillingAccountSink`, selected by `spec.scope`
- **Logging API enablement** -- `logging.googleapis.com` enabled (project-scope sinks only; other scopes are not project resources)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target scope.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Scope and Destination

- **The scope**: a project (default), folder, organization, or billing account. See [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying identity needs — folder/org/billing sinks require those permissions AT that scope.
- **The destination**: a GCS bucket, BigQuery dataset, or Pub/Sub topic (reference the Planton resources) — or a raw destination URI for log buckets.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpLoggingSink
metadata:
  name: error-archive
spec:
  destination:
    gcsBucket:
      value: my-log-archive-bucket
  filter: severity>=ERROR
```

```shell
planton apply -f logging-sink.yaml
```

Then grant the `writer_identity` output `roles/storage.objectCreator` on the bucket with a `GcpGcsBucketIamMember` whose `member` references the sink's `status.outputs.writer_identity` — without the grant, the sink silently exports nothing.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `destination` | `message` | Exactly one of `gcsBucket`, `bigqueryDataset`, `pubsubTopic`, `rawUri`. | Exactly-one |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `scope` | `message` | ambient project | At most one of `projectId` (ref), `folderId`, `organizationId`, `billingAccount`. |
| `sinkName` | `string` | `metadata.name` | Sink name (ForceNew — recreation mints a NEW writer identity). |
| `filter` | `string` | `""` | Which entries export. Empty exports EVERYTHING in scope. |
| `exclusions` | `list` | `[]` | Carve-outs (name + filter each; `sample()` supported) — at most 50. |
| `includeChildren` / `interceptChildren` | `bool` | `false` | Folder/org only: also export (or intercept) child resources' logs. |
| `uniqueWriterIdentity` | `bool` | `true` | Project scope only; other scopes always mint a unique writer. |
| `customWriterIdentity` | `string` | — | Project scope only: bring your own writer service account. |
| `destination.usePartitionedTables` | `bool` | `false` | BigQuery only: date-partitioned tables (recommended) instead of date-sharded names. |
| `description` / `disabled` | — | — | Operator documentation; pause without losing grants. |
| `deletionPolicy` | `string` | `DELETE` | What destroy does: `DELETE`, `PREVENT`, `ABANDON`. |

### Validation Rules

- **Exactly one destination arm**; `usePartitionedTables` only with BigQuery, and BigQuery requires the unique writer.
- **Children flags** only on folder/org scopes; **writer-identity controls** only on project scope — the other sink resources do not carry the arguments.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `sink_name` | `string` | The sink name in GCP |
| `writer_identity` | `string` | `serviceAccount:{email}` — GRANT THIS on the destination or nothing exports |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **The grant is the deploy's second half**: a standalone grant block that depends on both the sink and the destination, its `member` referencing `writer_identity` — `GcpGcsBucketIamMember` with `roles/storage.objectCreator` for a bucket, `GcpPubSubTopicIamMember` with `roles/pubsub.publisher` for a topic. Never through the destination kind's own IAM fields: the sink references its destination, so a destination that references the sink's identity back would be a dependency cycle. A BigQuery dataset takes `roles/bigquery.dataEditor` outside the catalog, or as a literal `GcpBigQueryDataset.access` entry once the identity is known (that access list is authoritative).
- **Renaming the sink recreates it and mints a NEW writer identity** — re-grant on the destination.
- **`interceptChildren` changes what child projects see** — intercepted logs stop reaching the children's own sinks. Use deliberately for compliance capture.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- [GcpGcsBucket](/docs/catalog/gcp/gcpgcsbucket) — the archival destination (grant the writer through a `GcpGcsBucketIamMember`)
- [GcpBigQueryDataset](/docs/catalog/gcp/gcpbigquerydataset) — the queryable destination
- [GcpPubSubTopic](/docs/catalog/gcp/gcppubsubtopic) — the streaming destination

## Additional Resources

- [Routing and Storage Overview](https://cloud.google.com/logging/docs/routing/overview)
- [Sinks API Reference](https://cloud.google.com/logging/docs/reference/v2/rest/v2/sinks)
- [Logging Query Language](https://cloud.google.com/logging/docs/view/logging-query-language)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
