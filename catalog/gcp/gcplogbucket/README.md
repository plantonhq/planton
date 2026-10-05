# GCP Log Bucket

Creates a Cloud Logging bucket — the container where log entries are STORED: how long they are retained, whether they are indexed for SQL analytics, who can see which slice (log views), and how the data links into BigQuery. One kind covers all four scopes (project, folder, organization, billing account), and adopting a name that already exists — including GCP's built-in `_Default` bucket — patches it into management instead of failing.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Log bucket** -- exactly ONE of `google_logging_{project|folder|organization|billing_account}_bucket_config`, selected by the spec's `scope`
- **Log views** (optional) -- one `google_logging_log_view` per `logViews[]` entry — named, independently grantable slices of the bucket
- **Linked BigQuery dataset** (optional) -- a `google_logging_linked_dataset` making the bucket queryable from BigQuery (requires analytics)
- **Scope logging settings** (optional, folder/org only) -- the `google_logging_{folder|organization}_settings` singleton (default-sink disable, default CMEK, default storage location)
- **Logging API enablement** -- `logging.googleapis.com` enabled on the project scope (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target scope.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP

- **Deployer permissions**: the least-privilege permission set the IaC runner's principal needs lives in [`iac/permissions.yaml`](iac/permissions.yaml).
- **CMEK buckets**: grant the Logging service account `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key FIRST (find the account with `gcloud logging settings describe`), or creation fails.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpLogBucket
metadata:
  name: audit-logs
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

## The adoption story

Every bucket variant ADOPTS an existing bucket when `bucketId` matches — that is how GCP's undeletable `_Default` bucket becomes manageable (set `bucketId: _Default` and manage its retention declaratively), and the only mode at all on folder/organization/billing scopes: the Logging API creates NEW custom buckets only under projects. `_Default` and `_Required` never delete — a destroy removes them from management and leaves them in GCP.

## One-way doors (deliberate levers)

- **`locked`** -- a locked bucket's retention can never change and it deletes only when empty. The compliance posture.
- **`enableAnalytics`** -- SQL over your logs; cannot be disabled once enabled.
- **`cmekKmsKey`** -- CMEK cannot be turned off once set (rotating to a different key is allowed).

## Outputs

| Output | Description |
|--------|-------------|
| `bucket_name` | Full resource name — a GcpLoggingSink routes here with destination `raw_uri: logging.googleapis.com/{bucket_name}`; a bucket-scoped GcpLogMetric references it directly |
| `linked_dataset_id` | The linked BigQuery dataset ID (empty when not armed) |

## Works With

- **GcpLoggingSink** -- routes matching entries INTO this bucket (storage and routing are deliberately separate kinds, like GCP's own model)
- **GcpLogMetric** -- bucket-scoped metrics count entries as they land here
- **GcpKmsKey** -- supplies the CMEK key
- **GcpBigQueryDataset** -- the linked dataset appears beside your other datasets, read-only

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
