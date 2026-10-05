# AWS S3 Table Bucket

Deploys an S3 table bucket (S3 Tables) with its full contents — namespaces, Apache Iceberg tables, resource policies, and replication — as one declarative analytics storage unit. AWS runs the table maintenance that otherwise consumes a data engineer: compaction, snapshot expiry, and unreferenced-file cleanup happen continuously, tuned by the dials on each table. Athena, EMR, Spark, and any Iceberg engine query the tables in place through the account's catalog integration. A table's schema is create-time input only — evolution happens through query engines, and both IaC modules deliberately ignore post-create schema edits, so a manifest's schema can never trigger a destructive replace.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **S3 Table Bucket** — named after the resource, carrying the bucket-level encryption default, the unreferenced-file-removal dials, and the force-destroy posture
- **Table Bucket Policy** — created only when `resourcePolicy` is set; who can create and query tables in the bucket
- **Table Bucket Replication** — created only when `replication` is set; every table replicates to the destination table buckets through the declared IAM role
- **Namespaces** — one per `namespaces[]` entry, the logical databases query engines address
- **Iceberg Tables** — one per table entry, with its create-time schema and properties, optional per-table encryption, and the compaction and snapshot-management maintenance dials (the table format is module-pinned to ICEBERG — the provider's enum holds exactly that one value)
- **Table Policies** — created only for tables that set their own `resourcePolicy`
- **Table Replication rules** — created only for tables that set their own `replication`, independent of the bucket-level rule
- **AWS Tags** — resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account, including S3 Tables permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A KMS grant for the maintenance service** (only for `aws:kms` encryption) — the S3 Tables maintenance principal (`maintenance.s3tables.amazonaws.com`) needs `kms:GenerateDataKey` and `kms:Decrypt` on the key, or compaction and snapshot expiry silently stop.
- **The analytics catalog integration** (for querying) — one per account, set up in the AWS console or Glue; without it, engines must be pointed manually at each table's warehouse location.
- **A replication role and destination buckets** (only for replication) — an IAM role with the s3tables replication trust and both source and destination buckets in scope; destinations are named by table bucket ARN (1–5 per rule).

## Deploy

### Console

Open the deployment store, find **AWS S3 Table Bucket**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Analytics Lake** preset in the [Presets](#presets) tab for the event-lake starter: one namespace, one schema-bearing events table, tuned maintenance.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsS3TableBucket
metadata:
  name: analytics-lake
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  namespaces:
    - name: analytics
      tables:
        - name: events
          icebergSchema:
            fields:
              - name: event_id
                type: string
                required: true
              - name: user_id
                type: string
              - name: occurred_at
                type: timestamp
              - name: payload
                type: string
          compaction:
            targetFileSizeMb: 256
          snapshotManagement:
            maxSnapshotAgeHours: 168
            minSnapshotsToKeep: 3
```

```shell
planton apply -f table-bucket.yaml
```

This creates a table bucket holding one `analytics` namespace with a four-column `events` table, compacted toward 256 MiB files and keeping a week of time travel. An Infra Job tracks the provisioning in real time.

### InfraChart

When the bucket deploys alongside its encryption key in one chart, wire the key reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  encryption:
    sseAlgorithm: aws:kms
    kmsKeyArn:
      valueFrom:
        kind: AwsKmsKey
        name: lake-key
        fieldPath: status.outputs.key_arn
  namespaces:
    - name: analytics
      tables:
        - name: events
```

The InfraPipeline resolves the dependency graph, provisions the key first, then creates the encrypted table bucket.

## Key Configuration

These are the most important decisions when configuring a table bucket. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The schema is a birth certificate, not a contract** — `icebergSchema` creates the table and is never read back: real schema evolution happens through query engines (ALTER TABLE ADD COLUMN), and both modules ignore post-create changes to the field — editing it after the table exists does nothing, and importing an existing table plans zero-diff against whatever schema the manifest carries. Adding tables is safe; to change a live table's columns, use an engine, then update the manifest to match for the record.

**KMS encryption has a silent failure mode** — with `aws:kms`, miss the maintenance service's key grant and nothing errors: compaction and snapshot expiry just stop, and query performance degrades over weeks as small files pile up. Grant the key BEFORE the first write.

**Maintenance dials are cost-performance trades** — `unreferencedDays` and `nonCurrentDays` decide how long dead files bill before cleanup; `minSnapshotsToKeep` and `maxSnapshotAgeHours` bound time travel. Shrinking retention saves storage and shrinks your undo window in the same edit — decide per table, not globally. Disabling compaction lets small-file buildup degrade queries until an engine compacts manually.

**Namespaces and tables are create-only names** — renaming either replaces it, and a table replacement drops data. Settle the `namespace.table` naming convention before the first deploy; it matters more here than in most kinds.

**forceDestroy is the difference between teardown and a support ticket** — a non-empty table bucket refuses deletion resource-by-resource (tables, then namespaces, then the bucket); `forceDestroy: true` drains it declaratively. Leave it false everywhere data matters. One more teardown wrinkle: a just-deleted bucket's name stays reserved for a short window (AWS answers 409 ConflictException), so automated rebuild flows should wait it out or pick a fresh name.

**Bucket-level or per-table replication** — the bucket-level rule replicates every table; a table's own `replication` entry acts independently of it. Use bucket-level for a uniform DR posture and per-table rules when only specific datasets justify the copy.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** (optional) | `encryption.kmsKeyArn` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional, per table) | `namespaces[].tables[].encryption.kmsKeyArn` | `status.outputs.key_arn` |
| **AwsIamRole** (with replication) | `replication.role` | `status.outputs.role_arn` |
| **AwsIamRole** (per-table replication) | `namespaces[].tables[].replication.role` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `table_bucket_arn` | The table bucket's ARN | Catalog integrations, cross-account policies, and other buckets' replication destinations |
| `table_arns` | Each table's ARN, keyed `namespace//table` | Per-table policy statements and table-level replication references |
| `table_warehouse_locations` | Each table's `s3://...` metadata root, keyed `namespace//table` | Query engines configured manually instead of through the catalog integration |

`owner_account_id` is also exported — the account that owns the bucket, useful for cross-account policy bookkeeping rather than composition.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**The event-lake starter** — one namespace, one schema-bearing events table, a week of time travel, and 256 MiB compaction targets. Query it from Athena the moment the catalog integration is on — no Spark cluster, no table maintenance jobs. The trade against a self-managed Iceberg warehouse is control for zero operations. Start from the **Analytics Lake** preset.

**The compliance lake** — KMS-encrypted at rest (remember the maintenance service's key grant) with bucket-level replication of every table to a replica bucket in another region. The replication role needs the s3tables trust and both buckets in scope; the copy is a second storage bill, taken deliberately. Start from the **Replicated Lake** preset.

## Works With

- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — the customer-managed key for at-rest encryption, wired via `encryption.kmsKeyArn` at the bucket or per table
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the role the replication service assumes, wired via `replication.role`
