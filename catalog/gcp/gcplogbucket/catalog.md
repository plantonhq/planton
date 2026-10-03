# GCP Log Bucket

Creates a Cloud Logging bucket — the container where log entries are STORED: how long they are retained, whether they are indexed for SQL analytics, who can see which slice (log views), and how the data links into BigQuery. One kind covers all four scopes (project, folder, organization, billing account), and adopting a name that already exists — including GCP's built-in `_Default` bucket — patches it into management instead of failing.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Log bucket** -- exactly ONE of `google_logging_{project|folder|organization|billing_account}_bucket_config`, selected by the spec's `scope`
- **Log views** (optional) -- one `google_logging_log_view` per `logViews[]` entry
- **Linked BigQuery dataset** (optional) -- a `google_logging_linked_dataset` (requires analytics)
- **Scope logging settings** (optional, folder/org only) -- the `google_logging_{folder|organization}_settings` singleton
- **Logging API enablement** -- `logging.googleapis.com` enabled on the project scope (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target scope.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **IAM**: `roles/logging.configWriter` on the scope (folder/org scopes need it at that level).
- **CMEK buckets**: grant the Logging service account `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key FIRST, or creation fails.

## Deploy

### Console

Open the deployment store, find **GCP Log Bucket**, and click **Deploy**. The creation wizard walks you through the scope, retention and one-way switches, indexes, log views, and the linked BigQuery dataset. Start from the **Compliance Retention** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpLogBucket
metadata:
  name: audit-logs
  org: acme-corp
  env: prod
spec:
  bucketId: audit-logs
  retentionDays: 400
  indexConfigs:
    - fieldPath: jsonPayload.request.status
      type: INDEX_TYPE_INTEGER
```

```shell
planton apply -f log-bucket.yaml
```

A 400-day project bucket with a status-code index — pair it with a GcpLoggingSink routing audit entries in. An Infra Job tracks the provisioning in real time.

### InfraChart

When wiring CMEK or an explicit project, use ValueFromRef to reference resources deployed in the same InfraPipeline:

```yaml
spec:
  scope:
    projectId:
      valueFrom:
        kind: GcpProject
        name: logging-project
        fieldPath: status.outputs.project_id
  bucketId: audit-logs
  retentionDays: 400
  cmekKmsKey:
    valueFrom:
      kind: GcpKmsKey
      name: logging-cmek
      fieldPath: status.outputs.key_id
```

The InfraPipeline resolves the dependency graph, deploys the project and key first, then provisions the bucket with the resolved values.

## Key Configuration

These are the most important decisions when configuring a log bucket. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**scope** -- omit for a project bucket in the provider's default project (the common case); `folderId` / `organizationId` / `billingAccount` select the other scopes, which are ADOPT-only (the Logging API creates new custom buckets only under projects).

**retentionDays** -- 1 to 3650; GCP's default 30 is sent explicitly. On a locked bucket, retention can never change again.

**One-way doors** -- `locked` (retention frozen, delete only when empty), `enableAnalytics` (SQL over logs; never disables), `cmekKmsKey` (CMEK never turns off; key rotation allowed). Each is deliberate — read the field docs before arming.

**logViews[]** -- named, filtered slices grantable independently (`roles/logging.viewAccessor` on a view shows a reader only its filter's entries) — how one bucket serves audit and app teams with different eyes.

**linkedBigqueryDataset** -- a read-only BigQuery dataset over the bucket (requires analytics); every field is create-time-only.

**deletionPolicy** -- `DELETE` (the default) removes the bucket and its stored entries on destroy (the built-in `_Default`/`_Required` buckets are undeletable and simply drop out of management); `PREVENT` fails the destroy — the posture for compliance-mandated storage; `ABANDON` keeps the bucket storing logs unmanaged. The policy also covers the bucket's log views and linked dataset.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `scope.projectId` | `status.outputs.project_id` |
| **GcpFolder** (optional) | `scope.folderId` | `status.outputs.folder_id` |
| **GcpKmsKey** (optional) | `cmekKmsKey`, `scopeSettings.kmsKey` | `status.outputs.key_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `bucket_name` | Full bucket resource name | GcpLoggingSink destination (`logging.googleapis.com/{bucket_name}`), bucket-scoped GcpLogMetric |
| `linked_dataset_id` | Linked BigQuery dataset ID | Queries from BigQuery tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Compliance retention** -- long retention + status index, PREVENT posture. Start from the **Compliance Retention** preset.

**Analytics + views** -- Log Analytics on, an errors-only view, and a linked BigQuery dataset. Start from the **Analytics And Views** preset.

**Adopt _Default** -- manage the built-in default bucket's retention declaratively. Start from the **Adopt Default Bucket** preset.

## Works With

- [**GCP Logging Sink**](/infra-catalog/gcp-logging-sink) -- routes matching entries INTO this bucket
- [**GCP Log Metric**](/infra-catalog/gcp-log-metric) -- bucket-scoped metrics count entries as they land here
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- supplies the CMEK key
- [**GCP BigQuery Dataset**](/infra-catalog/gcp-big-query-dataset) -- the linked dataset appears beside your other datasets
