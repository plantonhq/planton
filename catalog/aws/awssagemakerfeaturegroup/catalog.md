# AWS SageMaker Feature Group

Deploys an Amazon SageMaker Feature Store feature group — a declared schema of 1–2500 ML features over an online store for low-latency serving, an offline S3/Glue store for training datasets, or both. Every record is anchored by a record-identifier feature and an event-time feature; the online store can hard-delete records on a TTL clock, and the offline store lands every write under an auto-created Glue Data Catalog table. Almost everything is create-time structure: only the online TTL and the throughput mode update in place, so the schema and store choices deserve deliberate design before data lands.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SageMaker Feature Group** — named from `metadata.name`, carrying the feature schema (scalars, or List / Set / Vector collections with a dimension), the online and/or offline store configuration, and the throughput mode. The offline store's Glue table is auto-created by AWS in the `sagemaker_featurestore` database unless `disableGlueTableCreation` opts out.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with SageMaker Feature Store control-plane permissions (`sagemaker:CreateFeatureGroup` and its siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An IAM role trusting `sagemaker.amazonaws.com`, wired via `roleArn`. With an offline store, CreateFeatureGroup validates this role AGAINST THE BUCKET at create time: it assumes the role and calls `s3:GetBucketAcl`, and writes carry `s3:PutObjectAcl`. A role that can merely read and write objects fails the create with a `ValidationException` reading "Invalid S3Uri" — grant both ACL verbs, or attach AWS's `AmazonSageMakerFeatureStoreAccess` managed policy (it only matches buckets whose name contains "sagemaker").
- For an offline store: the S3 bucket behind `s3Uri` (only for that feature).

## Deploy

### Console

Open the deployment store, find **AWS SageMaker Feature Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the feature schema with its record-identifier and event-time features, and the store choices. Start from the **Realtime Serving Features** preset in the [Presets](#presets) tab for an online-only group, or the **Training and Serving Features** preset for the dual-store shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerFeatureGroup
metadata:
  name: customer-features
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  recordIdentifierFeatureName: customer_id
  eventTimeFeatureName: event_time
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: feature-store-role
      fieldPath: status.outputs.role_arn
  featureDefinitions:
    - name: customer_id
      type: String
    - name: event_time
      type: Fractional
    - name: sessions_last_7d
      type: Integral
    - name: avg_order_value
      type: Fractional
  onlineStore:
    enabled: true
    ttl:
      unit: Days
      value: 30
  offlineStore:
    s3Uri: s3://acme-feature-store/customers/
```

```shell
planton apply -f feature-group.yaml
```

This creates a dual-store group: four features served online with 30-day record expiry, every write also landing in S3 under an auto-created Glue table. An Infra Job tracks the provisioning in real time.

### InfraChart

When the group deploys alongside its execution role in one chart, wire the role reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  recordIdentifierFeatureName: customer_id
  eventTimeFeatureName: event_time
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: feature-store-role
      fieldPath: status.outputs.role_arn
  featureDefinitions:
    - name: customer_id
      type: String
    - name: event_time
      type: Fractional
  onlineStore:
    enabled: true
```

The InfraPipeline resolves the dependency graph, creates the role first, then the feature group that assumes it.

## Key Configuration

These are the most important decisions when configuring a feature group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Treat the schema as forever** — Everything except the online TTL and throughput is create-time only: changing a feature, a store, the storage tier, or even the `description` replaces the group. Design the schema — and decide whether training will ever need an offline store — before data lands, because stores cannot be added later without replacement.

**Online, offline, or both** — An enabled online store serves GetRecord at low latency; an offline store lands every write in S3 for training datasets and point-in-time queries. The dual-store shape writes one record to both, which is how train/serve skew is avoided. At least one store is required, and an online block with `enabled: false` does not count — the spec rejects a group that stores nothing.

**TTL is your one online lever** — Records hard-delete at event time plus `ttl`. Size it to serving freshness and adjust it freely: it is the only online-store setting that updates in place.

**Collections need InMemory storage** — List / Set / Vector features (vectors carry a `vectorDimension`, 1–8192) require the online store's `InMemory` tier, which bills an at-rest floor even when idle. Reach for vector features only when embeddings truly serve online; otherwise keep the Standard tier.

**Start on-demand** — `throughput` defaults to pay-per-request and updates in place, so move to `Provisioned` capacity once traffic is predictable rather than guessing up front. Capacity units are rejected outside Provisioned mode.

**Offline data outlives the group** — Deleting the group leaves its S3 objects in place by AWS design, and the auto-created Glue table survives the delete too. Budget the bucket's lifecycle separately, and clean both yourself when a group's data must go.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |
| **AwsKmsKey** | `onlineStore.kmsKeyArn`, `offlineStore.kmsKeyArn` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `feature_group_name` | The group's AWS identity — what PutRecord ingestion and GetRecord serving key on | Ingestion pipelines and serving-time feature lookup configuration |
| `feature_group_arn` | Amazon Resource Name of the group | IAM policies scoping feature-store access to this group |

## Common Patterns

**Online-only serving group** — an enabled online store, no offline store, a TTL sized to feature freshness: the shape for request-time feature lookup where no training dataset lives in this group. Start from the **Realtime Serving Features** preset.

**Dual-store training-and-serving group** — one write path feeding both stores, with the offline side queryable through the auto-created Glue table. The extra S3 and Glue footprint buys point-in-time training datasets that exactly match what was served. Start from the **Training and Serving Features** preset.

**Iceberg offline store** — set `tableFormat: Iceberg` at creation for faster offline queries and compaction. It is a create-time choice like every other store setting, so decide before the first record — converting later means replacing the group.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the role SageMaker assumes to persist offline-store data, wired via `roleArn`
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption for either store's data at rest
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — the offline store's landing bucket, referenced by URI in `s3Uri`
- [**AWS SageMaker Pipeline**](/infra-catalog/aws-sagemaker-pipeline) — pipelines that ingest into or build training datasets from the group
