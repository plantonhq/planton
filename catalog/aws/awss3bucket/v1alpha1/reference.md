# AwsS3Bucket

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsS3BucketSpec defines the desired configuration for an AWS S3 bucket.

S3 is object storage: buckets hold objects (files) addressed by key, and nearly
every AWS architecture touches one — static websites, log destinations, Lambda
code archives, data-lake storage, artifact stores, backup targets.

AWS models the bucket itself as a small resource and every behavioral setting
(versioning, encryption, lifecycle, replication, website hosting, notifications,
and so on) as a bucket-scoped configuration with the bucket's own lifecycle.
This spec folds those settings into one document so a bucket is fully described
in one place; none of them is independently referenceable, so none deserves to
be its own resource kind.

Notes:
- The bucket name comes from `metadata.name` and cannot be changed after
  creation (AWS bucket names are immutable and globally unique across all
  AWS accounts).
- New buckets are private by default: all four public-access-block guards are
  enabled unless `public_access_block` explicitly relaxes them, and object
  ownership defaults to `BucketOwnerEnforced` (ACLs disabled). Serving public
  content is best done through CloudFront with Origin Access Control; direct
  public buckets are the exception, not the rule.
- Since January 2023 AWS encrypts every object with SSE-S3 (AES256) by
  default, so the `encryption` block only needs to be set to switch to
  SSE-KMS/DSSE-KMS or to enable the S3 Bucket Key cost optimization.
- Credentials, region wiring, and deployment workflow live outside this spec
  in stack inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsS3Bucket
metadata:
  name: awss3bucket-demo
spec:
  region: us-east-1
  versioningStatus: Enabled
  encryption:
    sseAlgorithm: AES256
    # Refuse customer-provided-key uploads (states the March-2026 AWS
    # default posture explicitly).
    blockedEncryptionTypes:
      - SSE-C
  lifecycleRules:
    - id: prune-noncurrent
      noncurrentVersionExpiration:
        noncurrentDays: 30
      abortIncompleteMultipartUploadDays: 7
    # Day-zero transition: objects under ingest/ enter INTELLIGENT_TIERING
    # on the day they are uploaded.
    - id: tier-at-ingest
      filter:
        prefix: ingest/
      transitions:
        - days: 0
          storageClass: INTELLIGENT_TIERING
  # Attribute-based access control: authorize by principal tags vs bucket tags.
  abacStatus: Enabled
  # Storage-class analysis with a daily CSV export delivered to this bucket
  # itself (S3 bucket ARNs are name-derived, so the self-ARN is a literal).
  analyticsConfigurations:
    - name: hot-data-analysis
      filterPrefix: data/
      export:
        bucketArn:
          value: arn:aws:s3:::awss3bucket-demo
        prefix: analytics/
  # Weekly inventory report of every object version, delivered to this
  # bucket itself.
  inventoryConfigurations:
    - name: weekly-audit
      includedObjectVersions: All
      frequency: Weekly
      optionalFields:
        - Size
        - EncryptionStatus
        - ReplicationStatus
      destination:
        bucketArn:
          value: arn:aws:s3:::awss3bucket-demo
        format: Parquet
        prefix: inventory/
  # CloudWatch request metrics for the hot data prefix.
  metricsConfigurations:
    - name: hot-data-requests
      filterPrefix: data/
  # S3 Metadata: queryable Iceberg tables of this bucket's objects — the
  # change journal (expiring records after 90 days) plus the live
  # inventory table.
  metadataConfiguration:
    inventoryTableEnabled: true
    journalRecordExpiration:
      enabled: true
      days: 90
  forceDestroy: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.forceDestroy` | `bool` |  |  |  |
| `spec.objectLockEnabled` | `bool` |  |  |  |
| `spec.bucketNamespace` | `string` |  |  |  |
| `spec.versioningStatus` | `string` |  |  |  |
| `spec.encryption` | `AwsS3BucketEncryption` |  |  |  |
| `spec.encryption.sseAlgorithm` | `string` |  |  |  |
| `spec.encryption.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.encryption.bucketKeyEnabled` | `bool` |  |  |  |
| `spec.encryption.blockedEncryptionTypes` | `[]string` |  |  |  |
| `spec.publicAccessBlock` | `AwsS3BucketPublicAccessBlock` |  |  |  |
| `spec.publicAccessBlock.blockPublicAcls` | `bool` |  | `true` |  |
| `spec.publicAccessBlock.blockPublicPolicy` | `bool` |  | `true` |  |
| `spec.publicAccessBlock.ignorePublicAcls` | `bool` |  | `true` |  |
| `spec.publicAccessBlock.restrictPublicBuckets` | `bool` |  | `true` |  |
| `spec.objectOwnership` | `string` |  |  |  |
| `spec.acl` | `string` |  |  |  |
| `spec.policy` | `object` |  |  |  |
| `spec.transitionDefaultMinimumObjectSize` | `string` |  |  |  |
| `spec.lifecycleRules` | `[]AwsS3BucketLifecycleRule` |  |  |  |
| `spec.lifecycleRules[].id` | `string` | yes |  |  |
| `spec.lifecycleRules[].status` | `string` |  |  |  |
| `spec.lifecycleRules[].filter` | `AwsS3BucketLifecycleFilter` |  |  |  |
| `spec.lifecycleRules[].filter.prefix` | `string` |  |  |  |
| `spec.lifecycleRules[].filter.tags` | `map<string, string>` |  |  |  |
| `spec.lifecycleRules[].filter.objectSizeGreaterThan` | `int64` |  |  |  |
| `spec.lifecycleRules[].filter.objectSizeLessThan` | `int64` |  |  |  |
| `spec.lifecycleRules[].transitions` | `[]AwsS3BucketLifecycleTransition` |  |  |  |
| `spec.lifecycleRules[].transitions[].days` | `int32` |  |  |  |
| `spec.lifecycleRules[].transitions[].date` | `string` |  |  |  |
| `spec.lifecycleRules[].transitions[].storageClass` | `string` | yes |  |  |
| `spec.lifecycleRules[].expiration` | `AwsS3BucketLifecycleExpiration` |  |  |  |
| `spec.lifecycleRules[].expiration.days` | `int32` |  |  |  |
| `spec.lifecycleRules[].expiration.date` | `string` |  |  |  |
| `spec.lifecycleRules[].expiration.expiredObjectDeleteMarker` | `bool` |  |  |  |
| `spec.lifecycleRules[].noncurrentVersionTransitions` | `[]AwsS3BucketNoncurrentVersionTransition` |  |  |  |
| `spec.lifecycleRules[].noncurrentVersionTransitions[].noncurrentDays` | `int32` |  |  |  |
| `spec.lifecycleRules[].noncurrentVersionTransitions[].newerNoncurrentVersions` | `int32` |  |  |  |
| `spec.lifecycleRules[].noncurrentVersionTransitions[].storageClass` | `string` | yes |  |  |
| `spec.lifecycleRules[].noncurrentVersionExpiration` | `AwsS3BucketNoncurrentVersionExpiration` |  |  |  |
| `spec.lifecycleRules[].noncurrentVersionExpiration.noncurrentDays` | `int32` |  |  |  |
| `spec.lifecycleRules[].noncurrentVersionExpiration.newerNoncurrentVersions` | `int32` |  |  |  |
| `spec.lifecycleRules[].abortIncompleteMultipartUploadDays` | `int32` |  |  |  |
| `spec.replication` | `AwsS3BucketReplication` |  |  |  |
| `spec.replication.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.replication.rules` | `[]AwsS3BucketReplicationRule` | yes |  |  |
| `spec.replication.rules[].id` | `string` | yes |  |  |
| `spec.replication.rules[].priority` | `int32` |  |  |  |
| `spec.replication.rules[].status` | `string` |  |  |  |
| `spec.replication.rules[].filter` | `AwsS3BucketReplicationFilter` |  |  |  |
| `spec.replication.rules[].filter.prefix` | `string` |  |  |  |
| `spec.replication.rules[].filter.tags` | `map<string, string>` |  |  |  |
| `spec.replication.rules[].destination` | `AwsS3BucketReplicationDestination` | yes |  |  |
| `spec.replication.rules[].destination.bucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.replication.rules[].destination.account` | `string` |  |  |  |
| `spec.replication.rules[].destination.storageClass` | `string` |  |  |  |
| `spec.replication.rules[].destination.changeReplicaOwnershipToDestination` | `bool` |  |  |  |
| `spec.replication.rules[].destination.replicaKmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.replication.rules[].destination.metricsEnabled` | `bool` |  |  |  |
| `spec.replication.rules[].destination.replicationTimeControlEnabled` | `bool` |  |  |  |
| `spec.replication.rules[].deleteMarkerReplication` | `bool` |  |  |  |
| `spec.replication.rules[].existingObjectReplication` | `bool` |  |  |  |
| `spec.replication.rules[].replicateReplicaModifications` | `bool` |  |  |  |
| `spec.replication.rules[].replicateSseKmsEncryptedObjects` | `bool` |  |  |  |
| `spec.website` | `AwsS3BucketWebsite` |  |  |  |
| `spec.website.indexDocumentSuffix` | `string` |  |  |  |
| `spec.website.errorDocumentKey` | `string` |  |  |  |
| `spec.website.redirectAllRequestsTo` | `AwsS3BucketWebsiteRedirectAll` |  |  |  |
| `spec.website.redirectAllRequestsTo.hostName` | `string` | yes |  |  |
| `spec.website.redirectAllRequestsTo.protocol` | `string` |  |  |  |
| `spec.website.routingRules` | `[]AwsS3BucketWebsiteRoutingRule` |  |  |  |
| `spec.website.routingRules[].condition` | `AwsS3BucketWebsiteRoutingRuleCondition` |  |  |  |
| `spec.website.routingRules[].condition.httpErrorCodeReturnedEquals` | `string` |  |  |  |
| `spec.website.routingRules[].condition.keyPrefixEquals` | `string` |  |  |  |
| `spec.website.routingRules[].redirect` | `AwsS3BucketWebsiteRoutingRuleRedirect` | yes |  |  |
| `spec.website.routingRules[].redirect.hostName` | `string` |  |  |  |
| `spec.website.routingRules[].redirect.httpRedirectCode` | `string` |  |  |  |
| `spec.website.routingRules[].redirect.protocol` | `string` |  |  |  |
| `spec.website.routingRules[].redirect.replaceKeyPrefixWith` | `string` |  |  |  |
| `spec.website.routingRules[].redirect.replaceKeyWith` | `string` |  |  |  |
| `spec.logging` | `AwsS3BucketLogging` |  |  |  |
| `spec.logging.targetBucket` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.logging.targetPrefix` | `string` |  |  |  |
| `spec.logging.partitionedPrefixDateSource` | `string` |  |  |  |
| `spec.corsRules` | `[]AwsS3BucketCorsRule` |  |  |  |
| `spec.corsRules[].id` | `string` |  |  |  |
| `spec.corsRules[].allowedMethods` | `[]string` | yes |  |  |
| `spec.corsRules[].allowedOrigins` | `[]string` | yes |  |  |
| `spec.corsRules[].allowedHeaders` | `[]string` |  |  |  |
| `spec.corsRules[].exposeHeaders` | `[]string` |  |  |  |
| `spec.corsRules[].maxAgeSeconds` | `int32` |  |  |  |
| `spec.notification` | `AwsS3BucketNotification` |  |  |  |
| `spec.notification.eventbridge` | `bool` |  |  |  |
| `spec.notification.lambdaFunctions` | `[]AwsS3BucketLambdaNotification` |  |  |  |
| `spec.notification.lambdaFunctions[].lambdaFunctionArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.notification.lambdaFunctions[].events` | `[]string` | yes |  |  |
| `spec.notification.lambdaFunctions[].filterPrefix` | `string` |  |  |  |
| `spec.notification.lambdaFunctions[].filterSuffix` | `string` |  |  |  |
| `spec.notification.queues` | `[]AwsS3BucketQueueNotification` |  |  |  |
| `spec.notification.queues[].queueArn` | `string \| valueFrom` | yes |  | AwsSqsQueue (`status.outputs.queue_arn`) |
| `spec.notification.queues[].events` | `[]string` | yes |  |  |
| `spec.notification.queues[].filterPrefix` | `string` |  |  |  |
| `spec.notification.queues[].filterSuffix` | `string` |  |  |  |
| `spec.notification.topics` | `[]AwsS3BucketTopicNotification` |  |  |  |
| `spec.notification.topics[].topicArn` | `string \| valueFrom` | yes |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.notification.topics[].events` | `[]string` | yes |  |  |
| `spec.notification.topics[].filterPrefix` | `string` |  |  |  |
| `spec.notification.topics[].filterSuffix` | `string` |  |  |  |
| `spec.objectLockDefaultRetention` | `AwsS3BucketObjectLockDefaultRetention` |  |  |  |
| `spec.objectLockDefaultRetention.mode` | `string` | yes |  |  |
| `spec.objectLockDefaultRetention.days` | `int32` |  |  |  |
| `spec.objectLockDefaultRetention.years` | `int32` |  |  |  |
| `spec.accelerationStatus` | `string` |  |  |  |
| `spec.requestPayer` | `string` |  |  |  |
| `spec.intelligentTieringConfigurations` | `[]AwsS3BucketIntelligentTieringConfiguration` |  |  |  |
| `spec.intelligentTieringConfigurations[].name` | `string` | yes |  |  |
| `spec.intelligentTieringConfigurations[].status` | `string` |  |  |  |
| `spec.intelligentTieringConfigurations[].filterPrefix` | `string` |  |  |  |
| `spec.intelligentTieringConfigurations[].filterTags` | `map<string, string>` |  |  |  |
| `spec.intelligentTieringConfigurations[].tiers` | `[]AwsS3BucketIntelligentTieringTier` | yes |  |  |
| `spec.intelligentTieringConfigurations[].tiers[].accessTier` | `string` | yes |  |  |
| `spec.intelligentTieringConfigurations[].tiers[].days` | `int32` |  |  |  |
| `spec.abacStatus` | `string` |  |  |  |
| `spec.analyticsConfigurations` | `[]AwsS3BucketAnalyticsConfiguration` |  |  |  |
| `spec.analyticsConfigurations[].name` | `string` | yes |  |  |
| `spec.analyticsConfigurations[].filterPrefix` | `string` |  |  |  |
| `spec.analyticsConfigurations[].filterTags` | `map<string, string>` |  |  |  |
| `spec.analyticsConfigurations[].export` | `AwsS3BucketAnalyticsExport` |  |  |  |
| `spec.analyticsConfigurations[].export.bucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.analyticsConfigurations[].export.bucketAccountId` | `string` |  |  |  |
| `spec.analyticsConfigurations[].export.prefix` | `string` |  |  |  |
| `spec.inventoryConfigurations` | `[]AwsS3BucketInventoryConfiguration` |  |  |  |
| `spec.inventoryConfigurations[].name` | `string` | yes |  |  |
| `spec.inventoryConfigurations[].disabled` | `bool` |  |  |  |
| `spec.inventoryConfigurations[].includedObjectVersions` | `string` | yes |  |  |
| `spec.inventoryConfigurations[].frequency` | `string` | yes |  |  |
| `spec.inventoryConfigurations[].optionalFields` | `[]string` |  |  |  |
| `spec.inventoryConfigurations[].filterPrefix` | `string` |  |  |  |
| `spec.inventoryConfigurations[].destination` | `AwsS3BucketInventoryDestination` | yes |  |  |
| `spec.inventoryConfigurations[].destination.bucketArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.inventoryConfigurations[].destination.format` | `string` | yes |  |  |
| `spec.inventoryConfigurations[].destination.accountId` | `string` |  |  |  |
| `spec.inventoryConfigurations[].destination.prefix` | `string` |  |  |  |
| `spec.inventoryConfigurations[].destination.sseKmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.inventoryConfigurations[].destination.sseS3` | `bool` |  |  |  |
| `spec.metricsConfigurations` | `[]AwsS3BucketMetricsConfiguration` |  |  |  |
| `spec.metricsConfigurations[].name` | `string` | yes |  |  |
| `spec.metricsConfigurations[].filterAccessPointArn` | `string` |  |  |  |
| `spec.metricsConfigurations[].filterPrefix` | `string` |  |  |  |
| `spec.metricsConfigurations[].filterTags` | `map<string, string>` |  |  |  |
| `spec.metadataConfiguration` | `AwsS3BucketMetadataConfiguration` |  |  |  |
| `spec.metadataConfiguration.inventoryTableEnabled` | `bool` |  |  |  |
| `spec.metadataConfiguration.inventoryTableEncryption` | `AwsS3BucketMetadataTableEncryption` |  |  |  |
| `spec.metadataConfiguration.inventoryTableEncryption.sseAlgorithm` | `string` | yes |  |  |
| `spec.metadataConfiguration.inventoryTableEncryption.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.metadataConfiguration.journalRecordExpiration` | `AwsS3BucketMetadataJournalRecordExpiration` | yes |  |  |
| `spec.metadataConfiguration.journalRecordExpiration.enabled` | `bool` |  |  |  |
| `spec.metadataConfiguration.journalRecordExpiration.days` | `int32` |  |  |  |
| `spec.metadataConfiguration.journalTableEncryption` | `AwsS3BucketMetadataTableEncryption` |  |  |  |
| `spec.metadataConfiguration.journalTableEncryption.sseAlgorithm` | `string` | yes |  |  |
| `spec.metadataConfiguration.journalTableEncryption.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |

## Field Details

### spec.region

`string` · required

The AWS region where the bucket will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.forceDestroy

`bool`

Delete all objects (including all versions and delete markers) when the
bucket is destroyed, so a non-empty bucket does not block teardown.
Irreversible — leave false for production buckets holding real data.

### spec.objectLockEnabled

`bool`

Enable S3 Object Lock on the bucket (WORM — write once, read many).
Cannot be changed after creation and requires versioning. Enabling the
flag alone only makes the bucket lock-capable; pair it with
`object_lock_default_retention` to apply a default retention window to
every new object.

### spec.bucketNamespace

`string`

Bucket namespace. Empty keeps the classic "global" namespace (bucket
names unique across all AWS accounts worldwide); "account-regional"
scopes the name to this account and region instead — names only need to
be unique within the account+region, and the bucket is addressed through
account-scoped endpoints. Chosen at creation; changing it replaces the
bucket.

- rule: bucket_namespace must be one of: account-regional, global

### spec.versioningStatus

`string`

Bucket versioning state. When empty the bucket is left unversioned (the
AWS default). "Enabled" keeps every version of every object — the
foundation for replication, Object Lock, and protection against
accidental overwrite/delete. "Suspended" stops minting new versions but
keeps existing ones. AWS does not allow returning to the never-versioned
state once versioning has been enabled, so flipping Enabled back to
empty is rejected by AWS at apply time — use Suspended instead.
Versioned buckets accrue storage for every version; pair with lifecycle
`noncurrent_version_expiration` to bound costs.

- rule: versioning_status must be one of: Enabled, Suspended (or empty for unversioned)

### spec.encryption

`AwsS3BucketEncryption`

Default server-side encryption for new objects. When unset, AWS applies
its own SSE-S3 (AES256) default — set this only to move to KMS-based
encryption or to tune the bucket key.

- rule: kms_key_id is only used when sse_algorithm is aws:kms or aws:kms:dsse

### spec.encryption.sseAlgorithm

`string`

Encryption algorithm for new objects. "AES256" is SSE-S3 (S3-managed
keys, free — also the AWS account-wide default when this block is
absent). "aws:kms" is SSE-KMS: CloudTrail-audited key usage and
customer-controlled key policy, billed per KMS request unless the bucket
key is enabled. "aws:kms:dsse" is dual-layer SSE-KMS for workloads that
require two independent layers of encryption.

- rule: sse_algorithm must be one of: AES256, aws:kms, aws:kms:dsse

### spec.encryption.kmsKeyId

`string | valueFrom`

Customer-managed KMS key for SSE-KMS/DSSE-KMS. Accepts a key ARN or a
reference to an AwsKmsKey resource. When omitted with a KMS algorithm,
AWS uses the account's AWS-managed `aws/s3` key — functional, but without
customer control over the key policy or rotation.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.encryption.bucketKeyEnabled

`bool`

Use an S3 Bucket Key to reduce SSE-KMS costs. A bucket-level key is
derived from the KMS key and reused for objects, cutting KMS API calls
by up to 99% on KMS-encrypted buckets. Recommended whenever SSE-KMS is
in use; no effect under AES256.

### spec.encryption.blockedEncryptionTypes

`[]string`

Encryption types PUT requests may no longer use on this bucket.
"SSE-C" blocks customer-provided-key encryption (AWS blocks SSE-C on
new buckets by default since March 2026 — set this to state the posture
explicitly, or to re-permit SSE-C by managing the block without it);
"NONE" blocks unencrypted uploads that override the bucket default.

- rule: {"repeated":{"items":{"cel":[{"id":"blocked_encryption_type_valid","message":"blocked_encryption_types entries must be one of: NONE, SSE-C","expression":"this == 'NONE' || this == 'SSE-C'"}]}}}

### spec.publicAccessBlock

`AwsS3BucketPublicAccessBlock`

Public access guard rails. When unset, ALL FOUR guards are enabled — the
secure default for every new bucket. Set this block (flipping specific
guards to false) only when the bucket must serve public content directly,
e.g. a public static website. Each guard is independent; see the block's
field comments for what each one controls.

### spec.publicAccessBlock.blockPublicAcls

`bool` · optional (explicit presence)

Reject new ACLs that grant public access (PUT requests carrying public
ACLs fail). Unset keeps the guard on.

- default: `true`

### spec.publicAccessBlock.blockPublicPolicy

`bool` · optional (explicit presence)

Reject bucket policies that grant public access (the policy PUT fails).
Unset keeps the guard on.

- default: `true`

### spec.publicAccessBlock.ignorePublicAcls

`bool` · optional (explicit presence)

Ignore all existing public ACLs when evaluating access. Unset keeps the
guard on.

- default: `true`

### spec.publicAccessBlock.restrictPublicBuckets

`bool` · optional (explicit presence)

Restrict access to this bucket to AWS service principals and authorized
users within the bucket owner's account, even if a policy grants public
access. Unset keeps the guard on.

- default: `true`

### spec.objectOwnership

`string`

Object Ownership setting. Empty defaults to "BucketOwnerEnforced" — ACLs
are disabled and the bucket owner owns every object; access is managed
purely through policies (the AWS-recommended model). "BucketOwnerPreferred"
and "ObjectWriter" re-enable ACLs for legacy cross-account upload patterns
that depend on them.

- rule: object_ownership must be one of: BucketOwnerEnforced, BucketOwnerPreferred, ObjectWriter

### spec.acl

`string`

Canned ACL applied to the bucket. Only meaningful when `object_ownership`
re-enables ACLs (BucketOwnerPreferred or ObjectWriter) — enforced by
validation. Prefer bucket policies over ACLs; this exists for legacy
integrations (e.g. "log-delivery-write" for classic log delivery,
"aws-exec-read" for EC2 AMI bundle reads).

- rule: acl must be a canned ACL: private, public-read, public-read-write, authenticated-read, aws-exec-read, log-delivery-write

### spec.policy

`object`

Bucket resource policy as a standard IAM policy document. This is the
primary access-control surface for a bucket: cross-account read/write
grants, TLS-only conditions, public-read statements for website buckets,
and service permissions (CloudFront OAC, log delivery) all live here.
Note: statements granting public access also require the corresponding
`public_access_block` guards to be relaxed, otherwise AWS blocks the
policy at apply time.

### spec.transitionDefaultMinimumObjectSize

`string`

Minimum object size for the default transition behavior across all
lifecycle rules. "all_storage_classes_128K" (the AWS default) skips
transitioning objects smaller than 128 KB, which would otherwise cost
more in transition requests than they save in storage;
"varies_by_storage_class" applies the legacy per-class minimums.

- rule: transition_default_minimum_object_size must be one of: all_storage_classes_128K, varies_by_storage_class

### spec.lifecycleRules

`[]AwsS3BucketLifecycleRule`

Lifecycle rules that transition objects to cheaper storage classes and
expire objects/versions on a schedule. The standard levers for cost
control: tier logs to Glacier, expire temporary data, prune noncurrent
versions on versioned buckets, and abort stale multipart uploads.

- rule: a lifecycle rule must define at least one action (transition, expiration, noncurrent-version handling, or multipart-upload abort)
- rule: abort_incomplete_multipart_upload_days cannot be combined with a filter using tags or object-size bounds (AWS rejects it; prefix-only filters are allowed)

### spec.lifecycleRules[].id

`string` · required

Unique identifier for the rule within the bucket.

- rule: {"string":{"minLen":"1","maxLen":"255"}}

### spec.lifecycleRules[].status

`string`

Rule state. Empty defaults to "Enabled"; set "Disabled" to keep a rule
defined but inactive.

- rule: status must be one of: Enabled, Disabled

### spec.lifecycleRules[].filter

`AwsS3BucketLifecycleFilter`

Which objects the rule applies to. When unset the rule covers every
object in the bucket.

### spec.lifecycleRules[].filter.prefix

`string`

Key prefix, e.g. "logs/".

### spec.lifecycleRules[].filter.tags

`map<string, string>`

Object tags that must all be present.

### spec.lifecycleRules[].filter.objectSizeGreaterThan

`int64`

Minimum object size in bytes (exclusive). 0 means no lower bound.

- rule: {"int64":{"gte":"0"}}

### spec.lifecycleRules[].filter.objectSizeLessThan

`int64`

Maximum object size in bytes (exclusive). 0 means no upper bound.

- rule: {"int64":{"gte":"0"}}

### spec.lifecycleRules[].transitions

`[]AwsS3BucketLifecycleTransition`

Storage-class transitions for current object versions, e.g. to
STANDARD_IA after 30 days and DEEP_ARCHIVE after 365. Each entry names a
target class and when to move.

- rule: exactly one of days or date must be set

### spec.lifecycleRules[].transitions[].days

`int32` · optional (explicit presence)

Days after object creation to transition. Mutually exclusive with `date`.
Presence-typed because AWS allows 0 ("transition on the upload day" —
common for INTELLIGENT_TIERING/GLACIER_IR at ingest): unset means "use
date instead", an explicit 0 is a real day-zero transition.

- rule: {"int32":{"gte":0}}

### spec.lifecycleRules[].transitions[].date

`string`

Absolute transition date in RFC3339 format (e.g. "2027-01-01T00:00:00Z").
Mutually exclusive with `days`.

### spec.lifecycleRules[].transitions[].storageClass

`string` · required

Target storage class.

- rule: storage_class must be one of: STANDARD_IA, ONEZONE_IA, INTELLIGENT_TIERING, GLACIER_IR, GLACIER, DEEP_ARCHIVE
- rule: {"required":true}

### spec.lifecycleRules[].expiration

`AwsS3BucketLifecycleExpiration`

Expiration (deletion) of current object versions.

- rule: exactly one of days, date, or expired_object_delete_marker must be set

### spec.lifecycleRules[].expiration.days

`int32`

Days after object creation to expire.

- rule: {"int32":{"gte":0}}

### spec.lifecycleRules[].expiration.date

`string`

Absolute expiration date in RFC3339 format. Mutually exclusive with `days`.

### spec.lifecycleRules[].expiration.expiredObjectDeleteMarker

`bool`

Remove delete markers that have no remaining noncurrent versions
("expired object delete markers") — housekeeping for versioned buckets
that also prune noncurrent versions.

### spec.lifecycleRules[].noncurrentVersionTransitions

`[]AwsS3BucketNoncurrentVersionTransition`

Storage-class transitions for noncurrent (superseded) versions on
versioned buckets.

### spec.lifecycleRules[].noncurrentVersionTransitions[].noncurrentDays

`int32`

Days after an object version becomes noncurrent to transition it.

- rule: {"int32":{"gte":0}}

### spec.lifecycleRules[].noncurrentVersionTransitions[].newerNoncurrentVersions

`int32`

Keep this many newest noncurrent versions in their current class; the
transition applies only to older ones. 0 transitions all noncurrent
versions on schedule.

- rule: {"int32":{"gte":0}}

### spec.lifecycleRules[].noncurrentVersionTransitions[].storageClass

`string` · required

Target storage class.

- rule: storage_class must be one of: STANDARD_IA, ONEZONE_IA, INTELLIGENT_TIERING, GLACIER_IR, GLACIER, DEEP_ARCHIVE
- rule: {"required":true}

### spec.lifecycleRules[].noncurrentVersionExpiration

`AwsS3BucketNoncurrentVersionExpiration`

Permanent deletion of noncurrent versions — the essential cost control
for versioned buckets.

### spec.lifecycleRules[].noncurrentVersionExpiration.noncurrentDays

`int32`

Days after an object version becomes noncurrent to delete it permanently.

- rule: {"int32":{"gte":1}}

### spec.lifecycleRules[].noncurrentVersionExpiration.newerNoncurrentVersions

`int32`

Keep this many newest noncurrent versions regardless of age; only older
ones are deleted. 0 deletes all noncurrent versions on schedule.

- rule: {"int32":{"gte":0}}

### spec.lifecycleRules[].abortIncompleteMultipartUploadDays

`int32`

Abort incomplete multipart uploads this many days after initiation,
reclaiming storage from failed uploads. 7 days is a common choice.

- rule: {"int32":{"gte":0}}

### spec.replication

`AwsS3BucketReplication`

Replication configuration. Copies objects (asynchronously) to one or more
destination buckets — cross-region for disaster recovery, same-region for
cross-account backup or log aggregation. Requires versioning on both the
source (enforced here) and every destination bucket.

### spec.replication.roleArn

`string | valueFrom` · required

IAM role S3 assumes to replicate. The role must trust
`s3.amazonaws.com`, be able to read from this bucket, and be able to
write (`s3:ReplicateObject`/`s3:ReplicateDelete`) to every destination.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.replication.rules

`[]AwsS3BucketReplicationRule` · required

Replication rules. Rules with overlapping scopes are disambiguated by
`priority`.

- rule: {"repeated":{"minItems":"1","maxItems":"1000"}}
- rule: replicating SSE-KMS encrypted objects requires destination.replica_kms_key_id

### spec.replication.rules[].id

`string` · required

Unique identifier for the rule.

- rule: {"string":{"minLen":"1","maxLen":"255"}}

### spec.replication.rules[].priority

`int32`

Priority among rules with overlapping scopes; higher wins. Must be unique
across rules.

- rule: {"int32":{"gte":0}}

### spec.replication.rules[].status

`string`

Rule state. Empty defaults to "Enabled".

- rule: status must be one of: Enabled, Disabled

### spec.replication.rules[].filter

`AwsS3BucketReplicationFilter`

Which objects the rule replicates. When unset the rule covers the whole
bucket. Predicates combine with AND (same convention as lifecycle
filters; tag-scoped rules require delete-marker replication to stay
disabled per AWS rules — AWS validates this at apply time).

### spec.replication.rules[].filter.prefix

`string`

Key prefix, e.g. "important/".

### spec.replication.rules[].filter.tags

`map<string, string>`

Object tags that must all be present.

### spec.replication.rules[].destination

`AwsS3BucketReplicationDestination` · required

Where and how replicas are stored.

- rule: {"required":true}
- rule: replication_time_control_enabled requires metrics_enabled (AWS requires replication metrics with RTC)
- rule: change_replica_ownership_to_destination requires account (the destination account ID)

### spec.replication.rules[].destination.bucketArn

`string | valueFrom` · required

Destination bucket. Accepts a bucket ARN (arn:aws:s3:::name) or a
reference to an AwsS3Bucket resource. The destination must have
versioning enabled.

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.replication.rules[].destination.account

`string`

Destination AWS account ID for cross-account replication. Empty for
same-account.

### spec.replication.rules[].destination.storageClass

`string`

Storage class for replicas. Empty keeps each source object's class.
The domain mirrors AWS's PUT Bucket replication contract at provider
6.58.0: the full storage-class set including EXPRESS_ONEZONE (directory
buckets), OUTPOSTS, SNOW, and FSX_ONTAP — minus FSX_OPENZFS, which AWS's
own API contract states "is not an accepted value when replicating
objects". OUTPOSTS/SNOW/FSX_ONTAP apply only to their matching
destination bucket types; AWS validates that pairing at apply time.

- rule: storage_class must be one of: STANDARD, STANDARD_IA, ONEZONE_IA, INTELLIGENT_TIERING, GLACIER_IR, GLACIER, DEEP_ARCHIVE, REDUCED_REDUNDANCY, EXPRESS_ONEZONE, OUTPOSTS, SNOW, FSX_ONTAP

### spec.replication.rules[].destination.changeReplicaOwnershipToDestination

`bool`

Transfer replica ownership to the destination account (cross-account
replication where the destination owns its copies). Requires `account`.

### spec.replication.rules[].destination.replicaKmsKeyId

`string | valueFrom`

KMS key (in the destination region/account) used to encrypt replicas of
SSE-KMS objects. Required when the rule sets
`replicate_sse_kms_encrypted_objects`.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.replication.rules[].destination.metricsEnabled

`bool`

Emit replication metrics (bytes pending, latency) to CloudWatch.
Required for — and implied by — Replication Time Control.

### spec.replication.rules[].destination.replicationTimeControlEnabled

`bool`

Enable S3 Replication Time Control: an SLA that 99.99% of objects
replicate within 15 minutes, with metrics and events. Extra per-GB cost;
requires `metrics_enabled`.

### spec.replication.rules[].deleteMarkerReplication

`bool`

Replicate delete markers to the destination, keeping deletions in sync.
AWS default is NOT to replicate them (replicas outlive source deletions).

### spec.replication.rules[].existingObjectReplication

`bool`

Replicate existing objects (those created before the rule). Requires an
AWS Support-activated Batch Replication entitlement on some accounts;
most new setups replicate only new objects and backfill with S3 Batch
Operations.

### spec.replication.rules[].replicateReplicaModifications

`bool`

Replicate changes to replica metadata (replica modification sync) — used
for bi-directional replication topologies.

### spec.replication.rules[].replicateSseKmsEncryptedObjects

`bool`

Replicate objects encrypted with SSE-KMS (skipped by default). Requires
`destination.replica_kms_key_id` (enforced by validation), and the
replication role must be able to decrypt with the source key and encrypt
with the destination key.

### spec.website

`AwsS3BucketWebsite`

Static website hosting configuration. Serves bucket content over the
region's website endpoint (HTTP only, no TLS). For production sites,
front the bucket with CloudFront (which adds TLS, caching, and lets the
bucket stay private via Origin Access Control) and leave this unset;
direct website hosting suits internal or throwaway sites.

- rule: set either index_document_suffix (website mode) or redirect_all_requests_to (redirect mode), not both
- rule: error_document_key requires index_document_suffix (website mode)
- rule: routing_rules require index_document_suffix (website mode)

### spec.website.indexDocumentSuffix

`string`

Index document object suffix served for directory-style requests,
e.g. "index.html".

### spec.website.errorDocumentKey

`string`

Object key served on 4XX errors, e.g. "error.html".

### spec.website.redirectAllRequestsTo

`AwsS3BucketWebsiteRedirectAll`

Redirect every request to another host — makes this a pure redirect
bucket (e.g. apex-domain → www).

### spec.website.redirectAllRequestsTo.hostName

`string` · required

Host to redirect to, e.g. "www.example.com".

- rule: {"string":{"minLen":"1"}}

### spec.website.redirectAllRequestsTo.protocol

`string`

Protocol for the redirect. Empty preserves the request protocol.

- rule: protocol must be one of: http, https

### spec.website.routingRules

`[]AwsS3BucketWebsiteRoutingRule`

Conditional redirect rules evaluated per request (e.g. redirect a prefix
to another host or rewrite key prefixes).

### spec.website.routingRules[].condition

`AwsS3BucketWebsiteRoutingRuleCondition`

When the rule applies. Omit the condition entirely to apply the redirect
to every request; a present-but-empty condition is rejected (AWS requires
at least one condition field when the Condition element is sent).

- rule: a routing-rule condition must set http_error_code_returned_equals and/or key_prefix_equals (omit the condition to match every request)

### spec.website.routingRules[].condition.httpErrorCodeReturnedEquals

`string`

Apply when the response would have this HTTP error code, e.g. "404".

### spec.website.routingRules[].condition.keyPrefixEquals

`string`

Apply to requests whose key starts with this prefix, e.g. "docs/".

### spec.website.routingRules[].redirect

`AwsS3BucketWebsiteRoutingRuleRedirect` · required

What the redirect does.

- rule: {"required":true}
- rule: a routing-rule redirect must set at least one of: host_name, http_redirect_code, protocol, replace_key_prefix_with, replace_key_with
- rule: replace_key_prefix_with and replace_key_with are mutually exclusive

### spec.website.routingRules[].redirect.hostName

`string`

Redirect target host. Empty keeps the original host.

### spec.website.routingRules[].redirect.httpRedirectCode

`string`

HTTP redirect code, e.g. "301". Empty uses the AWS default (301).

### spec.website.routingRules[].redirect.protocol

`string`

Protocol for the redirect. Empty preserves the request protocol.

- rule: protocol must be one of: http, https

### spec.website.routingRules[].redirect.replaceKeyPrefixWith

`string`

Replace the matched key prefix with this value (prefix rewrite).
Mutually exclusive with replace_key_with.

### spec.website.routingRules[].redirect.replaceKeyWith

`string`

Replace the entire key with this value. Mutually exclusive with
replace_key_prefix_with.

### spec.logging

`AwsS3BucketLogging`

Server access logging. Delivers request logs to another bucket (with some
delay). The target bucket must be in the same region and must allow log
delivery — either via its ACL (legacy) or, under BucketOwnerEnforced
ownership, via a bucket policy granting `logging.s3.amazonaws.com`.

### spec.logging.targetBucket

`string | valueFrom` · required

Destination bucket for access logs. Accepts a bucket name or a reference
to an AwsS3Bucket resource. Must be in the same region; must not be the
bucket itself (logging loops amplify storage).

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.logging.targetPrefix

`string`

Key prefix for log objects, e.g. "logs/my-bucket/".

### spec.logging.partitionedPrefixDateSource

`string`

Date source for partitioned log-object keys
("[prefix]/[account]/[region]/[bucket]/[yyyy]/[mm]/[dd]/..."), which makes
logs directly queryable by Athena partitions. "EventTime" partitions by
when the request happened, "DeliveryTime" by when the log was delivered.
Empty uses the flat (non-partitioned) key format.

- rule: partitioned_prefix_date_source must be one of: EventTime, DeliveryTime

### spec.corsRules

`[]AwsS3BucketCorsRule`

CORS rules for browser-based access, required when a web application on
another origin reads from or writes to the bucket directly (presigned
uploads, font/asset serving).

- rule: {"repeated":{"maxItems":"100"}}

### spec.corsRules[].id

`string`

Optional identifier for the rule (shows up in error messages).

- rule: {"string":{"maxLen":"255"}}

### spec.corsRules[].allowedMethods

`[]string` · required

HTTP methods the origin may use, e.g. ["GET", "PUT"].

- rule: {"repeated":{"minItems":"1","items":{"cel":[{"id":"cors_method_valid","message":"allowed_methods entries must be one of: GET, PUT, POST, DELETE, HEAD","expression":"this == 'GET' || this == 'PUT' || this == 'POST' || this == 'DELETE' || this == 'HEAD'"}]}}}

### spec.corsRules[].allowedOrigins

`[]string` · required

Origins allowed to make requests, e.g. ["https://example.com"]. "*"
allows any origin.

- rule: {"repeated":{"minItems":"1"}}

### spec.corsRules[].allowedHeaders

`[]string`

Headers allowed in the actual request (matched against
Access-Control-Request-Headers in preflight).

### spec.corsRules[].exposeHeaders

`[]string`

Response headers the browser is allowed to read.

### spec.corsRules[].maxAgeSeconds

`int32`

Seconds the browser may cache the preflight response.

- rule: {"int32":{"gte":0}}

### spec.notification

`AwsS3BucketNotification`

Event notifications for object-level events (created, removed, restored,
replicated...). Targets are Lambda functions, SQS queues, SNS topics, or
EventBridge. Note: SQS/SNS/Lambda targets must grant S3 permission to
deliver BEFORE the notification is configured (queue/topic policy or
Lambda resource permission), or AWS rejects the configuration at apply
time; the EventBridge arm needs no such grant.

### spec.notification.eventbridge

`bool`

Deliver all bucket events to Amazon EventBridge (the default event bus),
where rules route them anywhere. The most flexible arm and the only one
requiring no delivery permission setup.

### spec.notification.lambdaFunctions

`[]AwsS3BucketLambdaNotification`

Lambda function targets.

### spec.notification.lambdaFunctions[].lambdaFunctionArn

`string | valueFrom` · required

Lambda function to invoke. Accepts a function ARN or a reference to an
AwsLambda resource. The function must carry a resource-based permission
allowing `s3.amazonaws.com` to invoke it (AwsLambda's
`invoke_permissions` models this).

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.notification.lambdaFunctions[].events

`[]string` · required

Event types to deliver, e.g. ["s3:ObjectCreated:*", "s3:ObjectRemoved:*"].

- rule: {"repeated":{"minItems":"1"}}

### spec.notification.lambdaFunctions[].filterPrefix

`string`

Only deliver events for keys starting with this prefix.

### spec.notification.lambdaFunctions[].filterSuffix

`string`

Only deliver events for keys ending with this suffix, e.g. ".jpg".

### spec.notification.queues

`[]AwsS3BucketQueueNotification`

SQS queue targets. The queue policy must allow `s3.amazonaws.com` to
send messages (scoped to this bucket's ARN) before the notification is
configured.

### spec.notification.queues[].queueArn

`string | valueFrom` · required

SQS queue to deliver to. Accepts a queue ARN or a reference to an
AwsSqsQueue resource.

- references: AwsSqsQueue (`status.outputs.queue_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSqsQueue, name: <that resource's name>, fieldPath: status.outputs.queue_arn}} -- a bare string does not parse

### spec.notification.queues[].events

`[]string` · required

Event types to deliver, e.g. ["s3:ObjectCreated:*"].

- rule: {"repeated":{"minItems":"1"}}

### spec.notification.queues[].filterPrefix

`string`

Only deliver events for keys starting with this prefix.

### spec.notification.queues[].filterSuffix

`string`

Only deliver events for keys ending with this suffix.

### spec.notification.topics

`[]AwsS3BucketTopicNotification`

SNS topic targets. The topic policy must allow `s3.amazonaws.com` to
publish (scoped to this bucket's ARN) before the notification is
configured.

### spec.notification.topics[].topicArn

`string | valueFrom` · required

SNS topic to publish to. Accepts a topic ARN or a reference to an
AwsSnsTopic resource.

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.notification.topics[].events

`[]string` · required

Event types to deliver, e.g. ["s3:ObjectCreated:*"].

- rule: {"repeated":{"minItems":"1"}}

### spec.notification.topics[].filterPrefix

`string`

Only deliver events for keys starting with this prefix.

### spec.notification.topics[].filterSuffix

`string`

Only deliver events for keys ending with this suffix.

### spec.objectLockDefaultRetention

`AwsS3BucketObjectLockDefaultRetention`

Default Object Lock retention applied to every new object. Requires
`object_lock_enabled` (enforced by validation). GOVERNANCE mode can be
bypassed by principals with special permission; COMPLIANCE mode cannot be
shortened or bypassed by anyone — including the root account — until the
retention period expires, so treat COMPLIANCE with care.

- rule: exactly one of days or years must be set

### spec.objectLockDefaultRetention.mode

`string` · required

Retention mode. GOVERNANCE allows privileged bypass
(s3:BypassGovernanceRetention); COMPLIANCE is immutable for everyone
until expiry.

- rule: mode must be one of: GOVERNANCE, COMPLIANCE
- rule: {"required":true}

### spec.objectLockDefaultRetention.days

`int32`

Retention period in days. Mutually exclusive with `years`.

- rule: {"int32":{"gte":0}}

### spec.objectLockDefaultRetention.years

`int32`

Retention period in years. Mutually exclusive with `days`.

- rule: {"int32":{"gte":0}}

### spec.accelerationStatus

`string`

Transfer Acceleration state. "Enabled" routes uploads/downloads through
CloudFront edge locations for faster long-distance transfers (extra cost
per GB). Once enabled, use "Suspended" to turn it off — the setting
cannot be removed entirely.

- rule: acceleration_status must be one of: Enabled, Suspended

### spec.requestPayer

`string`

Who pays for requests and data transfer. Empty defaults to "BucketOwner".
"Requester" shifts request/transfer costs to the caller — common for
large public datasets.

- rule: request_payer must be one of: BucketOwner, Requester

### spec.intelligentTieringConfigurations

`[]AwsS3BucketIntelligentTieringConfiguration`

Archive-tier configurations for objects stored in the INTELLIGENT_TIERING
storage class. Each named configuration opts a scope of objects into the
Archive Access and/or Deep Archive Access tiers after a period without
access. Only affects objects already in INTELLIGENT_TIERING (via lifecycle
transition or direct upload).

### spec.intelligentTieringConfigurations[].name

`string` · required

Name of the configuration, unique within the bucket.

- rule: {"string":{"minLen":"1"}}

### spec.intelligentTieringConfigurations[].status

`string`

Configuration state. Empty defaults to "Enabled".

- rule: status must be one of: Enabled, Disabled

### spec.intelligentTieringConfigurations[].filterPrefix

`string`

Key prefix scoping the configuration. Combined (AND) with filter_tags.

### spec.intelligentTieringConfigurations[].filterTags

`map<string, string>`

Object tags scoping the configuration. Combined (AND) with filter_prefix.

### spec.intelligentTieringConfigurations[].tiers

`[]AwsS3BucketIntelligentTieringTier` · required

Archive tiers to enable and when. ARCHIVE_ACCESS requires at least 90
days without access, DEEP_ARCHIVE_ACCESS at least 180 (enforced by
validation, mirroring AWS limits; both max 730).

- rule: {"repeated":{"minItems":"1"}}
- rule: days must be at least 90 for ARCHIVE_ACCESS and at least 180 for DEEP_ARCHIVE_ACCESS (max 730)

### spec.intelligentTieringConfigurations[].tiers[].accessTier

`string` · required

Archive tier to move objects into.

- rule: access_tier must be one of: ARCHIVE_ACCESS, DEEP_ARCHIVE_ACCESS
- rule: {"required":true}

### spec.intelligentTieringConfigurations[].tiers[].days

`int32`

Consecutive days without access before objects move to this tier.

- rule: {"int32":{"gte":1}}

### spec.abacStatus

`string`

ABAC state for the bucket. "Enabled" lets IAM policies authorize requests
by comparing principal tags against the bucket's tags (attribute-based
access control); "Disabled" turns it off explicitly. Empty leaves the
bucket at AWS's default (disabled) without managing the setting.

- rule: abac_status must be one of: Enabled, Disabled

### spec.analyticsConfigurations

`[]AwsS3BucketAnalyticsConfiguration`

Storage-class-analysis configurations. Each named configuration observes
access patterns for a scope of objects and (optionally) exports daily
findings to a bucket as CSV — the data source for choosing lifecycle
transition ages. Analysis observes for at least 30 days before results
stabilize.

### spec.analyticsConfigurations[].name

`string` · required

Name of the configuration, unique within the bucket (up to 64
characters per the AWS analytics-id contract).

- rule: {"string":{"minLen":"1","maxLen":"64"}}

### spec.analyticsConfigurations[].filterPrefix

`string`

Key prefix scoping the analysis. Combined (AND) with filter_tags; leave
both empty to analyze the whole bucket.

### spec.analyticsConfigurations[].filterTags

`map<string, string>`

Object tags scoping the analysis. Combined (AND) with filter_prefix.

### spec.analyticsConfigurations[].export

`AwsS3BucketAnalyticsExport`

Daily CSV export of the analysis findings to a bucket. Without it the
findings are only visible in the S3 console. (AWS's V_1 schema and CSV
format are the only values the API accepts, so only the destination is
configurable.)

### spec.analyticsConfigurations[].export.bucketArn

`string | valueFrom` · required

Destination bucket for the CSV findings. Accepts a bucket ARN
(arn:aws:s3:::name) or a reference to an AwsS3Bucket resource. The
bucket must allow `s3.amazonaws.com` to `s3:PutObject` (same-account
destinations get the grant via bucket policy).

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.analyticsConfigurations[].export.bucketAccountId

`string`

Account ID that owns the destination bucket. Empty assumes the
bucket owner's account.

### spec.analyticsConfigurations[].export.prefix

`string`

Key prefix for exported findings, e.g. "analytics/".

### spec.inventoryConfigurations

`[]AwsS3BucketInventoryConfiguration`

Inventory report configurations. Each named configuration delivers a
scheduled manifest of objects and selected metadata (encryption status,
replication status, object-lock state, ...) to a destination bucket —
the audit backbone for large buckets where ListObjects is impractical.
The destination bucket needs a policy allowing `s3.amazonaws.com` to
`s3:PutObject` (AWS documents the exact statement); delivery starts
within 48 hours of configuration.

### spec.inventoryConfigurations[].name

`string` · required

Name of the configuration, unique within the bucket (up to 64
characters per the AWS inventory-id contract).

- rule: {"string":{"minLen":"1","maxLen":"64"}}

### spec.inventoryConfigurations[].disabled

`bool`

Keep the configuration defined but stop generating reports. Unset means
active (the AWS default) — the zero value never fights the provider's
enabled-by-default contract.

### spec.inventoryConfigurations[].includedObjectVersions

`string` · required

Which object versions the report covers: "All" (every version on
versioned buckets) or "Current" (only the latest of each object).

- rule: included_object_versions must be one of: All, Current
- rule: {"required":true}

### spec.inventoryConfigurations[].frequency

`string` · required

Report generation schedule.

- rule: frequency must be one of: Daily, Weekly
- rule: {"required":true}

### spec.inventoryConfigurations[].optionalFields

`[]string`

Metadata columns to include beyond bucket/key/version. The report is the
cheap way to audit encryption, replication, and object-lock posture at
scale.

- rule: {"repeated":{"items":{"cel":[{"id":"inventory_optional_field_valid","message":"optional_fields entries must be valid inventory fields (Size, LastModifiedDate, StorageClass, ETag, IsMultipartUploaded, ReplicationStatus, EncryptionStatus, ObjectLockRetainUntilDate, ObjectLockMode, ObjectLockLegalHoldStatus, IntelligentTieringAccessTier, BucketKeyStatus, ChecksumAlgorithm, ObjectAccessControlList, ObjectOwner, LifecycleExpirationDate)","expression":"this in ['Size', 'LastModifiedDate', 'StorageClass', 'ETag', 'IsMultipartUploaded', 'ReplicationStatus', 'EncryptionStatus', 'ObjectLockRetainUntilDate', 'ObjectLockMode', 'ObjectLockLegalHoldStatus', 'IntelligentTieringAccessTier', 'BucketKeyStatus', 'ChecksumAlgorithm', 'ObjectAccessControlList', 'ObjectOwner', 'LifecycleExpirationDate']"}]}}}

### spec.inventoryConfigurations[].filterPrefix

`string`

Key prefix scoping which objects the report covers. Empty covers the
whole bucket.

### spec.inventoryConfigurations[].destination

`AwsS3BucketInventoryDestination` · required

Where and how the report files are delivered.

- rule: {"required":true}
- rule: set at most one of sse_s3 or sse_kms_key_id for report encryption

### spec.inventoryConfigurations[].destination.bucketArn

`string | valueFrom` · required

Destination bucket for report files. Accepts a bucket ARN
(arn:aws:s3:::name) or a reference to an AwsS3Bucket resource — a bucket
may deliver its inventory to itself (its own ARN is derived from the
name: arn:aws:s3:::<name>). The destination must allow `s3.amazonaws.com`
to `s3:PutObject` via bucket policy.

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.inventoryConfigurations[].destination.format

`string` · required

Report file format.

- rule: format must be one of: CSV, ORC, Parquet
- rule: {"required":true}

### spec.inventoryConfigurations[].destination.accountId

`string`

Account ID that owns the destination bucket. Empty assumes the bucket
owner's account.

### spec.inventoryConfigurations[].destination.prefix

`string`

Key prefix for report files, e.g. "inventory/".

### spec.inventoryConfigurations[].destination.sseKmsKeyId

`string | valueFrom`

Encrypt report files with SSE-KMS using this key. Accepts a key ARN or
a reference to an AwsKmsKey resource; the key policy must allow
`s3.amazonaws.com` to use it. Mutually exclusive with `sse_s3`.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.inventoryConfigurations[].destination.sseS3

`bool`

Encrypt report files with SSE-S3 (S3-managed keys). Mutually exclusive
with `sse_kms_key_id`; leave both unset to deliver with the destination
bucket's default encryption.

### spec.metricsConfigurations

`[]AwsS3BucketMetricsConfiguration`

Request-metrics configurations. Each named configuration publishes
CloudWatch request metrics (AllRequests, 4xxErrors, FirstByteLatency,
...) for a scope of objects, at CloudWatch's standard per-metric cost.
Whole-bucket storage metrics are free and always on; these add
request-level visibility for a prefix, tag set, or access point.

### spec.metricsConfigurations[].name

`string` · required

Name of the configuration, unique within the bucket (up to 64
characters per the AWS metrics-id contract).

- rule: {"string":{"minLen":"1","maxLen":"64"}}

### spec.metricsConfigurations[].filterAccessPointArn

`string`

Scope metrics to requests through this access point (access point ARN).

### spec.metricsConfigurations[].filterPrefix

`string`

Key prefix scoping the metrics. Combined (AND) with the other predicates.

### spec.metricsConfigurations[].filterTags

`map<string, string>`

Object tags scoping the metrics. Combined (AND) with the other
predicates.

### spec.metadataConfiguration

`AwsS3BucketMetadataConfiguration`

S3 Metadata configuration. Maintains queryable Apache Iceberg tables of
this bucket's object metadata in an AWS-managed table bucket: a journal
table records every change (uploads, deletes, metadata updates) and an
optional live inventory table mirrors the current object set. Query them
with Athena/Spark to find objects without listing the bucket. Table
storage and updates bill per S3 Tables pricing.

### spec.metadataConfiguration.inventoryTableEnabled

`bool`

Maintain the live inventory table (current state of every object) in
addition to the journal. Off keeps only the change journal.

### spec.metadataConfiguration.inventoryTableEncryption

`AwsS3BucketMetadataTableEncryption`

Encryption for the inventory table. Only meaningful while
`inventory_table_enabled` is true.

- rule: kms_key_arn is only used when sse_algorithm is aws:kms

### spec.metadataConfiguration.inventoryTableEncryption.sseAlgorithm

`string` · required

Encryption algorithm for the table: "AES256" (S3-managed keys) or
"aws:kms" (customer-managed KMS key).

- rule: sse_algorithm must be one of: AES256, aws:kms
- rule: {"required":true}

### spec.metadataConfiguration.inventoryTableEncryption.kmsKeyArn

`string | valueFrom`

KMS key for aws:kms encryption. Accepts a key ARN or a reference to an
AwsKmsKey resource; the key policy must allow the S3 Tables maintenance
principal to use it.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.metadataConfiguration.journalRecordExpiration

`AwsS3BucketMetadataJournalRecordExpiration` · required

Expiration policy for journal-table records — the required cost control
for the change log (records accrue per object change).

- rule: {"required":true}
- rule: days is required (7–2147483647) when expiration is enabled, and must be unset when disabled

### spec.metadataConfiguration.journalRecordExpiration.enabled

`bool`

Expire journal records after `days`. Disabled keeps records forever
(stated explicitly — AWS requires the policy either way).

### spec.metadataConfiguration.journalRecordExpiration.days

`int32`

Days to retain journal records (at least 7). Only set when `enabled`.

- rule: {"int32":{"gte":0}}

### spec.metadataConfiguration.journalTableEncryption

`AwsS3BucketMetadataTableEncryption`

Encryption for the journal table.

- rule: kms_key_arn is only used when sse_algorithm is aws:kms

### spec.metadataConfiguration.journalTableEncryption.sseAlgorithm

`string` · required

Encryption algorithm for the table: "AES256" (S3-managed keys) or
"aws:kms" (customer-managed KMS key).

- rule: sse_algorithm must be one of: AES256, aws:kms
- rule: {"required":true}

### spec.metadataConfiguration.journalTableEncryption.kmsKeyArn

`string | valueFrom`

KMS key for aws:kms encryption. Accepts a key ARN or a reference to an
AwsKmsKey resource; the key policy must allow the S3 Tables maintenance
principal to use it.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

## Validation Rules

- `acl_requires_non_enforced_ownership`: acl can only be set when object_ownership is BucketOwnerPreferred or ObjectWriter (ACLs are disabled under BucketOwnerEnforced)
- `object_lock_retention_requires_object_lock`: object_lock_default_retention requires object_lock_enabled to be true
- `object_lock_requires_versioning`: object_lock_enabled requires versioning_status to be Enabled (AWS enables versioning automatically for Object Lock buckets; state it explicitly so the manifest is honest)
- `replication_requires_versioning`: replication requires versioning_status to be Enabled on the source bucket

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsS3Bucket, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.bucket_id` | `string` | id (name) of the S3 bucket created on AWS |
| `status.outputs.bucket_arn` | `string` | ARN (Amazon Resource Name) of the S3 bucket Format: arn:aws:s3:::bucket-name Used for IAM policies, bucket policies, and cross-account access |
| `status.outputs.region` | `string` | AWS region where the bucket is created |
| `status.outputs.bucket_regional_domain_name` | `string` | Regional domain name of the S3 bucket Format: bucket-name.s3.region.amazonaws.com Used for accessing bucket via regional endpoint |
| `status.outputs.hosted_zone_id` | `string` | Hosted zone ID for the S3 bucket's region Used for Route53 alias records pointing to S3 bucket |
| `status.outputs.bucket_domain_name` | `string` | Global (legacy path-style) domain name of the S3 bucket Format: bucket-name.s3.amazonaws.com |
| `status.outputs.website_endpoint` | `string` | Website endpoint for the bucket, populated only when static website hosting is configured Format: bucket-name.s3-website-region.amazonaws.com (or s3-website.region for newer regions) Used as a CloudFront custom origin or direct HTTP website address |
| `status.outputs.website_domain` | `string` | Website domain of the region's S3 website service, populated only when static website hosting is configured Format: s3-website-region.amazonaws.com Used for Route53 alias records pointing to the website endpoint |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.encryption.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.replication.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.replication.rules[].destination.bucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.replication.rules[].destination.replicaKmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.logging.targetBucket` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.notification.lambdaFunctions[].lambdaFunctionArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.notification.queues[].queueArn` | AwsSqsQueue | `status.outputs.queue_arn` |
| `spec.notification.topics[].topicArn` | AwsSnsTopic | `status.outputs.topic_arn` |
| `spec.analyticsConfigurations[].export.bucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.inventoryConfigurations[].destination.bucketArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.inventoryConfigurations[].destination.sseKmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.metadataConfiguration.inventoryTableEncryption.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.metadataConfiguration.journalTableEncryption.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAlb | `spec.accessLogs.bucket` | `status.outputs.bucket_id` |
| AwsAlb | `spec.connectionLogs.bucket` | `status.outputs.bucket_id` |
| AwsAlb | `spec.healthCheckLogs.bucket` | `status.outputs.bucket_id` |
| AwsBackupReportPlan | `spec.deliveryChannel.s3BucketName` | `status.outputs.bucket_id` |
| AwsBedrockAgent | `spec.actionGroups[].apiSchema.s3.bucketName` | `status.outputs.bucket_id` |
| AwsBedrockAgentCoreRuntime | `spec.artifact.code.s3.bucket` | `status.outputs.bucket_id` |
| AwsBedrockAgentCoreTools | `spec.browsers[].recording.s3Location.bucket` | `status.outputs.bucket_id` |
| AwsBedrockAgentCoreTools | `spec.browsers[].enterprisePolicies[].s3.bucket` | `status.outputs.bucket_id` |
| AwsBedrockFlow | `spec.definition.nodes[].retrieval.bucketName` | `status.outputs.bucket_id` |
| AwsBedrockFlow | `spec.definition.nodes[].storage.bucketName` | `status.outputs.bucket_id` |
| AwsBedrockInvocationLogging | `spec.cloudwatch.largeDataDeliveryS3.bucketName` | `status.outputs.bucket_id` |
| AwsBedrockInvocationLogging | `spec.s3.bucketName` | `status.outputs.bucket_id` |
| AwsBedrockKnowledgeBase | `spec.dataSources[].s3.bucketArn` | `status.outputs.bucket_arn` |
| AwsCloudTrail | `spec.s3BucketName` | `status.outputs.bucket_id` |
| AwsCloudwatchSynthetics | `spec.canary.artifactBucket` | `status.outputs.bucket_id` |
| AwsCloudwatchSynthetics | `spec.canary.code.s3Bucket` | `status.outputs.bucket_id` |
| AwsCodeBuildProject | `spec.artifacts.location` | `status.outputs.bucket_id` |
| AwsCodeBuildProject | `spec.secondaryArtifacts[].location` | `status.outputs.bucket_id` |
| AwsCodeBuildProject | `spec.cache.location` | `status.outputs.bucket_id` |
| AwsCodeBuildProject | `spec.logsConfig.s3Logs.bucket` | `status.outputs.bucket_id` |
| AwsCodePipeline | `spec.artifactStores[].location` | `status.outputs.bucket_id` |
| AwsCognitoUserPool | `spec.logConfigurations[].s3BucketArn` | `status.outputs.bucket_arn` |
| AwsConfigConformancePack | `spec.deliveryS3Bucket` | `status.outputs.bucket_id` |
| AwsConfigRecorder | `spec.deliveryChannel.s3BucketName` | `status.outputs.bucket_id` |
| AwsDynamodb | `spec.importTable.s3Bucket` | `status.outputs.bucket_id` |
| AwsEcsTaskDefinition | `spec.volumes[].s3files.fileSystemArn` | `status.outputs.bucket_arn` |
| AwsEventBridgePipe | `spec.logConfiguration.s3.bucketName` | `status.outputs.bucket_id` |
| AwsGlobalAccelerator | `spec.flowLogs.s3Bucket` | `status.outputs.bucket_id` |
| AwsGuardDuty | `spec.publishingDestination.bucketArn` | `status.outputs.bucket_arn` |
| AwsGuardDutyMalwareProtectionPlan | `spec.s3BucketName` | `status.outputs.bucket_id` |
| AwsKinesisFirehose | `spec.extendedS3.bucketArn` | `status.outputs.bucket_arn` |
| AwsKinesisFirehose | `spec.extendedS3.s3Backup.bucketArn` | `status.outputs.bucket_arn` |
| AwsKinesisFirehose | `spec.opensearch.s3Config.bucketArn` | `status.outputs.bucket_arn` |
| AwsKinesisFirehose | `spec.opensearchServerless.s3Config.bucketArn` | `status.outputs.bucket_arn` |
| AwsKinesisFirehose | `spec.httpEndpoint.s3Config.bucketArn` | `status.outputs.bucket_arn` |
| AwsKinesisFirehose | `spec.redshift.s3Config.bucketArn` | `status.outputs.bucket_arn` |
| AwsKinesisFirehose | `spec.redshift.s3Backup.bucketArn` | `status.outputs.bucket_arn` |
| AwsKinesisFirehose | `spec.splunk.s3Config.bucketArn` | `status.outputs.bucket_arn` |
| AwsKinesisFirehose | `spec.snowflake.s3Config.bucketArn` | `status.outputs.bucket_arn` |
| AwsKinesisFirehose | `spec.iceberg.s3Config.bucketArn` | `status.outputs.bucket_arn` |
| AwsLambda | `spec.s3.bucket` | `status.outputs.bucket_id` |
| AwsLambdaLayer | `spec.code.bucket` | `status.outputs.bucket_id` |
| AwsMskCluster | `spec.logging.s3.bucket` | `status.outputs.bucket_id` |
| AwsMwaaEnvironment | `spec.sourceBucketArn` | `status.outputs.bucket_arn` |
| AwsNlb | `spec.accessLogs.bucket` | `status.outputs.bucket_id` |
| AwsPrivateCa | `spec.revocation.crl.s3BucketName` | `status.outputs.bucket_id` |
| AwsS3Bucket | `spec.replication.rules[].destination.bucketArn` | `status.outputs.bucket_arn` |
| AwsS3Bucket | `spec.logging.targetBucket` | `status.outputs.bucket_id` |
| AwsS3Bucket | `spec.analyticsConfigurations[].export.bucketArn` | `status.outputs.bucket_arn` |
| AwsS3Bucket | `spec.inventoryConfigurations[].destination.bucketArn` | `status.outputs.bucket_arn` |
| AwsS3ObjectSet | `spec.bucket` | `status.outputs.bucket_id` |
| AwsS3ObjectSet | `spec.objects[].copyFrom.sourceBucket` | `status.outputs.bucket_id` |
| AwsSagemakerPipeline | `spec.definitionS3Location.bucket` | `status.outputs.bucket_id` |
| AwsSsmAssociation | `spec.outputLocation.s3BucketName` | `status.outputs.bucket_id` |
| AwsSsmMaintenanceWindow | `spec.tasks[].invocation.runCommand.outputS3Bucket` | `status.outputs.bucket_id` |

## See Also

- [Overview](../README.md)
