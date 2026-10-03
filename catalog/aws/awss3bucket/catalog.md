# AWS S3 Bucket

Deploys an S3 bucket with its full behavioral surface folded into one document: public-access guards, object ownership, bucket policy, versioning and Object Lock, default encryption (including blocked encryption types), lifecycle rules and Intelligent-Tiering archive, replication, static website hosting, CORS, access logging, event notifications, the transfer/cost knobs, and the governance surface — ABAC, storage-class analytics, inventory reports, request metrics, and S3 Metadata tables. New buckets are private by default — all four public-access guards on, ACLs disabled, every object encrypted.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **S3 Bucket** — named after the resource (bucket names are immutable and globally unique across all AWS accounts)
- **Bucket-scoped configurations** — only the blocks the spec sets: versioning state, default encryption, public-access block, ownership controls, bucket policy, lifecycle rules, replication configuration, website configuration, logging, CORS, notifications, Object Lock default retention, Intelligent-Tiering archive configurations, acceleration, request-payer, ABAC, storage-class-analysis configurations, inventory configurations, request-metrics configurations, and the S3 Metadata tables
- **AWS Tags** — resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A KMS key** (optional) — for SSE-KMS default encryption or replica encryption. Reference an AwsKmsKey Infra Component or pass a key ARN.
- **A replication role and destination buckets** (only for replication) — an IAM role trusting `s3.amazonaws.com` with read access here and replicate permissions on every destination; versioning must be Enabled on the source and every destination.
- **Notification targets with delivery grants** (only for eventing) — SQS/SNS/Lambda targets must permit `s3.amazonaws.com` BEFORE the notification is configured (queue/topic policy or Lambda resource permission — AwsLambda's `invoke_permissions` models the Lambda side). The EventBridge arm needs no grant.
- **A log-delivery bucket** (only for access logging) — same region, never the bucket itself; under BucketOwnerEnforced ownership it needs a policy granting `logging.s3.amazonaws.com`.

## Deploy

### Console

Open the deployment store, find **AWS S3 Bucket**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Private Encrypted Bucket**, **Public Static Website Bucket**, or **Log Archive with Lifecycle Tiering** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsS3Bucket
metadata:
  name: acme-app-artifacts
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  versioningStatus: Enabled
```

```shell
planton apply -f s3-bucket.yaml
```

This creates a fully private, versioned bucket — all four public-access guards on, ACLs disabled, SSE-S3 encryption by AWS default. An Infra Job tracks the provisioning in real time.

### InfraChart

When wiring KMS encryption or replication, use ValueFromRef to reference resources deployed in the same InfraPipeline:

```yaml
spec:
  encryption:
    sseAlgorithm: aws:kms
    kmsKeyId:
      valueFrom:
        kind: AwsKmsKey
        name: data-key
        fieldPath: status.outputs.key_arn
    bucketKeyEnabled: true
  replication:
    roleArn:
      valueFrom:
        kind: AwsIamRole
        name: replication-role
        fieldPath: status.outputs.role_arn
    rules:
      - id: dr-copy
        destination:
          bucketArn:
            valueFrom:
              kind: AwsS3Bucket
              name: dr-bucket
              fieldPath: status.outputs.bucket_arn
```

The InfraPipeline resolves the dependency graph, deploys the key, role, and destination bucket first, then provisions this bucket with the resolved ARNs.

## Key Configuration

These are the most important decisions when configuring a bucket. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Access posture** — Leaving `publicAccessBlock` unset keeps ALL FOUR guards on: nothing can make the bucket public. Set the block (relaxing specific guards) only when the bucket must serve public content directly; for production sites, prefer CloudFront with Origin Access Control and keep the bucket private. `objectOwnership` defaults to BucketOwnerEnforced (ACLs disabled); `acl` is only valid under the two legacy ownership modes. The `policy` document is the primary access-control surface.

**Versioning** — `versioningStatus: Enabled` keeps every version of every object: the foundation for replication, Object Lock, and recovery from accidental overwrite/delete. AWS never allows returning to the unversioned state — only Suspended. Bound the cost with lifecycle `noncurrentVersionExpiration`.

**Object Lock (WORM)** — `objectLockEnabled` is create-time only and requires versioning Enabled. Pair it with `objectLockDefaultRetention`: GOVERNANCE allows privileged bypass; COMPLIANCE is immutable for everyone — including the root account — until retention expires.

**Encryption** — AWS already encrypts every object with SSE-S3; set `encryption` only to move to `aws:kms`/`aws:kms:dsse` with a customer-managed key. Enable `bucketKeyEnabled` with KMS to cut key-request costs by up to 99%.

**Lifecycle** — `lifecycleRules` are the standard cost levers: storage-class transitions (days XOR absolute date per transition), expiration (exactly one of days/date/delete-marker cleanup), noncurrent-version pruning, and multipart-upload abort (prefix-filtered rules only). `intelligentTieringConfigurations` opt INTELLIGENT_TIERING objects into the archive tiers after 90/180+ days without access.

**Replication** — `replication` copies objects asynchronously to destination buckets: cross-region for disaster recovery, same-region for cross-account backup. Not retroactive — rules cover new objects; backfill with S3 Batch Operations. Replication Time Control (15-minute SLA) requires metrics.

**Website hosting** — `website` serves EITHER a normal site (`indexDocumentSuffix` + optional error document and routing rules) OR a pure redirect bucket (`redirectAllRequestsTo`), never both. The website endpoint is HTTP-only and requires public reads.

**Eventing** — `notification` delivers object events to EventBridge (no grant needed) and/or Lambda/SQS/SNS targets (grant delivery BEFORE configuring). `logging` delivers request-level access logs; partitioned prefixes make them Athena-queryable.

**Governance and insight** — `inventoryConfigurations` deliver scheduled object manifests (encryption/replication/lock posture per object) to a bucket — including the bucket itself; `analyticsConfigurations` observe access patterns and export findings that justify lifecycle transition ages; `metricsConfigurations` add CloudWatch request metrics per prefix/tag scope; `metadataConfiguration` maintains queryable Iceberg tables (change journal + live inventory) so you find objects with Athena instead of listing; `abacStatus` enables tag-based access control. Inventory/analytics DELIVERY needs a bucket policy allowing `s3.amazonaws.com` to `s3:PutObject` on the destination.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** (optional) | `encryption.kmsKeyId` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional, per rule) | `replication.rules[].destination.replicaKmsKeyId` | `status.outputs.key_arn` |
| **AwsIamRole** (with replication) | `replication.roleArn` | `status.outputs.role_arn` |
| **AwsS3Bucket** (optional, per rule) | `replication.rules[].destination.bucketArn` | `status.outputs.bucket_arn` |
| **AwsS3Bucket** (with logging) | `logging.targetBucket` | `status.outputs.bucket_id` |
| **AwsLambda** (optional, per target) | `notification.lambdaFunctions[].lambdaFunctionArn` | `status.outputs.function_arn` |
| **AwsSqsQueue** (optional, per target) | `notification.queues[].queueArn` | `status.outputs.queue_arn` |
| **AwsSnsTopic** (optional, per target) | `notification.topics[].topicArn` | `status.outputs.topic_arn` |
| **AwsS3Bucket** (optional, per config) | `analyticsConfigurations[].export.bucketArn` | `status.outputs.bucket_arn` |
| **AwsS3Bucket** (optional, per config) | `inventoryConfigurations[].destination.bucketArn` | `status.outputs.bucket_arn` |
| **AwsKmsKey** (optional, per config) | `inventoryConfigurations[].destination.sseKmsKeyId` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional, per table) | `metadataConfiguration.*TableEncryption.kmsKeyArn` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `bucket_id` | The bucket name | Logging targets, CLI addressing, SDK configuration |
| `bucket_arn` | ARN of the bucket | IAM policies, replication destinations, notification scoping |
| `bucket_regional_domain_name` | Regional endpoint hostname | CloudFront origins (the recommended origin form) |
| `hosted_zone_id` | The S3 Route 53 hosted zone for this region | Route 53 alias records to the bucket |
| `website_endpoint` | Website endpoint (when hosting is configured) | DNS records for direct website hosting |
| `website_domain` | Website domain (when hosting is configured) | Route 53 alias records |

`status.outputs` also echoes `region` back and carries `bucket_domain_name`, the legacy global endpoint — neither is normally wired downstream (use the regional domain name for origins).

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private encrypted bucket** — the production default: fully private, versioned, KMS-encrypted with the bucket key on. Start from the **Private Encrypted Bucket** preset.

**Public static website** — website hosting with relaxed guards and a public-read policy; the pattern for internal or throwaway sites (production sites belong behind CloudFront). Start from the **Public Static Website Bucket** preset.

**Log archive** — a destination for logs with tiering transitions to Glacier and scheduled expiration. Start from the **Log Archive with Lifecycle Tiering** preset.

**Governed data lake** — a private, versioned lake bucket that audits and optimizes itself: day-zero Intelligent-Tiering on the raw zone, weekly inventory delivered to the bucket itself, storage-class analytics on the curated zone to justify future transitions, request metrics, and S3 Metadata tables queried with Athena instead of listing. Start from the **Governed Data Lake** preset.

## Works With

- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed keys for default and replica encryption
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the role S3 assumes to replicate
- [**AWS Lambda**](/infra-catalog/aws-lambda) — object-event processing targets (and buckets hold Lambda code archives)
- [**AWS SQS Queue**](/infra-catalog/aws-sqs-queue) / [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) — event delivery targets
- [**AWS CloudFront**](/infra-catalog/aws-cloud-front) — the TLS/caching front for content served from private buckets
