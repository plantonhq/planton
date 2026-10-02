# GCP Logging Sink

Creates a Cloud Logging sink — the routing rule that exports log entries matching a filter to a Cloud Storage bucket, BigQuery dataset, or Pub/Sub topic. One kind covers all four scopes (project, folder, organization, billing account), the module renders the destination URI from a natural resource reference, and the sink's writer identity is surfaced as THE output to grant on the destination — the mis-wiring this kind exists to remove.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Logging Sink** -- exactly one of the four scope-specific sink resources, selected by `spec.scope`
- **Logging API enablement** -- `logging.googleapis.com` enabled (project-scope sinks only)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target scope.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **The scope**: a project (default), folder, organization, or billing account — folder/org/billing sinks need `roles/logging.configWriter` at that scope.
- **The destination**: a GCS bucket, BigQuery dataset, or Pub/Sub topic, ideally referenced as Planton resources so the grant flow stays in one chart.

## Deploy

### Console

Open the deployment store, find **GCP Logging Sink**, and click **Deploy**. The creation wizard walks you through the scope, the destination reference, the filter and exclusions, and the writer-identity posture. Start from the **Error Archive to GCS** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpLoggingSink
metadata:
  name: error-archive
  org: acme-corp
  env: prod
spec:
  destination:
    gcsBucket:
      value: acme-log-archive
  filter: severity>=ERROR
```

```shell
planton apply -f logging-sink.yaml
```

This exports every ERROR-and-above entry in the project to hourly JSON batches in the bucket. The deploy's second half: grant the `writer_identity` output `roles/storage.objectCreator` on the bucket with a `GcpGcsBucketIamMember`. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying alongside its destination, the sink references the bucket via ValueFromRef:

```yaml
spec:
  destination:
    gcsBucket:
      valueFrom:
        kind: GcpGcsBucket
        name: log-archive
        fieldPath: status.outputs.bucket_id
  filter: severity>=ERROR
```

The InfraPipeline deploys the bucket first, then the sink with the resolved bucket name — and the same chart closes the loop with a `GcpGcsBucketIamMember` (role `roles/storage.objectCreator`, `member` referencing this sink's `status.outputs.writer_identity`), which deploys after both. Never grant through the bucket's own `iamMembers`: the sink references its destination, so a destination that references the sink's identity back would be a dependency cycle.

## Key Configuration

These are the most important decisions when configuring a logging sink. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope** -- omit for the ambient project; set `folderId`/`organizationId` for centralized capture across children (with `includeChildren`), or `billingAccount` for billing-account logs.

**Destination** -- reference the resource, not the URI: `gcsBucket` (hourly batches, cheapest archive), `bigqueryDataset` (near-real-time, queryable; enable `usePartitionedTables`), `pubsubTopic` (streaming to third-party pipelines), or `rawUri` for Cloud Logging buckets.

**Filter and exclusions** -- the filter selects what exports; exclusions carve noisy sub-streams (health checks, debug logs) out of a broad export, with `sample()` support for percentage drops.

**Writer identity** -- GCP mints a service account per sink; the `writer_identity` output is what you grant on the destination. Project sinks can opt into the legacy shared account or bring their own — other scopes always mint one.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpGcsBucket** (optional) | `destination.gcsBucket` | `status.outputs.bucket_id` |
| **GcpBigQueryDataset** (optional) | `destination.bigqueryDataset` | `status.outputs.self_link` |
| **GcpPubSubTopic** (optional) | `destination.pubsubTopic` | `status.outputs.topic_id` |
| **GcpProject** (optional) | `scope.projectId` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `writer_identity` | `serviceAccount:{email}` | The `member` of a `GcpGcsBucketIamMember` / `GcpPubSubTopicIamMember` — the grant that makes the export actually flow |
| `sink_name` | The sink name in GCP | Audit, Logging API cross-references |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Error archive to GCS** -- ERROR-and-above to a bucket: the cheapest compliance archive. Start from the **Error Archive to GCS** preset.

**Audit logs to BigQuery** -- audit entries to partitioned BigQuery tables for SQL-queryable forensics. Start from the **Audit Logs to BigQuery** preset.

**Log stream to Pub/Sub** -- everything (minus carve-outs) streamed to a topic — the front door to Datadog/Splunk-class pipelines. Start from the **Log Stream to Pub/Sub** preset.

## Works With

- [**GCP GCS Bucket**](/cloud-catalog/gcp-gcs-bucket) -- the archival destination
- [**GCP BigQuery Dataset**](/cloud-catalog/gcp-big-query-dataset) -- the queryable destination
- [**GCP Pub/Sub Topic**](/cloud-catalog/gcp-pub-sub-topic) -- the streaming destination
