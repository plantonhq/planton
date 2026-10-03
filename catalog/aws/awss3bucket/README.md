# AwsS3Bucket

The **AwsS3Bucket** resource provisions and manages Amazon S3 buckets through Planton. It models the bucket and every bucket-scoped behavior — versioning, encryption, public-access posture, policies, lifecycle management, replication, static website hosting, access logging, CORS, event notifications, Object Lock, transfer acceleration, requester pays, Intelligent-Tiering archive configurations, attribute-based access control (ABAC), storage-class analytics, inventory reports, request metrics, and S3 Metadata tables — as one declarative document.

## Spec Fields

### Bucket Root

- **region**: AWS region for the bucket. Required.
- **force_destroy**: Delete all objects (including versions) when the bucket is destroyed. Leave `false` for production data.
- **object_lock_enabled**: Enable S3 Object Lock (WORM) at creation. Immutable after creation; requires `versioning_status: Enabled`. Pair with `object_lock_default_retention` to apply a default retention window.
- **bucket_namespace**: Empty keeps the classic `global` namespace; `account-regional` scopes the bucket name to this account and region (account-scoped endpoints). Chosen at creation; changing it replaces the bucket.

### Versioning and Encryption

- **versioning_status**: `Enabled`, `Suspended`, or empty (unversioned — the AWS default). Once enabled, a bucket cannot return to the never-versioned state; use `Suspended` instead.
- **encryption**: Default server-side encryption for new objects. When absent, AWS's own SSE-S3 (AES256) default applies.
  - **sse_algorithm**: `AES256` (SSE-S3, free), `aws:kms` (SSE-KMS), or `aws:kms:dsse` (dual-layer SSE-KMS).
  - **kms_key_id**: Customer-managed KMS key. Accepts a literal ARN or a `valueFrom` reference to an `AwsKmsKey`.
  - **bucket_key_enabled**: Reuse a bucket-level data key to cut KMS request costs by up to 99%. Recommended with SSE-KMS.
  - **blocked_encryption_types**: Encryption types PUT requests may no longer use — `SSE-C` (AWS's own default block on new buckets since March 2026) and/or `NONE`. Set it to state the posture explicitly or to manage it deliberately.

### Public Access, Ownership, and Policy

- **public_access_block**: The four public-access guards (`block_public_acls`, `block_public_policy`, `ignore_public_acls`, `restrict_public_buckets`). **When the block is absent, all four guards are enabled** — new buckets are fully private by default. Set the block (flipping specific guards to `false`) only for deliberately public buckets.
- **object_ownership**: `BucketOwnerEnforced` (default — ACLs disabled, policy-only access control), `BucketOwnerPreferred`, or `ObjectWriter` (both re-enable ACLs for legacy patterns).
- **acl**: Canned ACL (`private`, `public-read`, `log-delivery-write`, ...). Only valid when `object_ownership` re-enables ACLs — enforced by validation.
- **policy**: The bucket resource policy as a native YAML/JSON structure — the primary access-control surface (cross-account grants, TLS-only conditions, public-read statements, service permissions).

### Lifecycle Management

- **lifecycle_rules**: Storage-class transitions and expiration. Each rule carries a filter (prefix, tags, and object-size bounds, AND-combined), current-version `transitions` (days — including `0` for transition-at-upload — or an absolute date, to STANDARD_IA / ONEZONE_IA / INTELLIGENT_TIERING / GLACIER_IR / GLACIER / DEEP_ARCHIVE), `expiration`, noncurrent-version transitions/expiration (with `newer_noncurrent_versions` retention), and incomplete-multipart-upload cleanup.
- **transition_default_minimum_object_size**: `all_storage_classes_128K` (AWS default) or `varies_by_storage_class`.

### Replication

- **replication**: Cross-region or same-region replication. Requires `versioning_status: Enabled` (enforced by validation).
  - **role_arn**: IAM role S3 assumes — literal ARN or a reference to an `AwsIamRole`.
  - **rules**: Per-rule scope filter, destination (bucket ARN or `AwsS3Bucket` reference, storage class, cross-account `account` + ownership translation, replica KMS key, Replication Time Control + metrics), delete-marker replication, existing-object replication, replica-modification sync, and SSE-KMS object replication.

### Website, Logging, CORS

- **website**: Static website hosting — `index_document_suffix`/`error_document_key` (website mode) XOR `redirect_all_requests_to` (redirect bucket), plus conditional `routing_rules`.
- **logging**: Server access logging to a target bucket (literal name or `AwsS3Bucket` reference), with optional Athena-friendly partitioned log-object keys (`partitioned_prefix_date_source`).
- **cors_rules**: Cross-origin rules for browser access (methods, origins, headers, preflight cache).

### Event Notifications

- **notification**: Object-event delivery. `eventbridge: true` routes all events to EventBridge (no permission setup needed). `lambda_functions` / `queues` / `topics` target specific consumers by ARN or reference (`AwsLambda`, `AwsSqsQueue`, `AwsSnsTopic`) with event-type and key prefix/suffix filters — note the target's resource policy must grant S3 delivery **before** the notification is configured.

### Object Lock, Acceleration, Requester Pays, Intelligent-Tiering

- **object_lock_default_retention**: Default WORM retention (`GOVERNANCE` or `COMPLIANCE`, days XOR years) for new objects. COMPLIANCE cannot be shortened or bypassed by anyone until expiry.
- **acceleration_status**: `Enabled`/`Suspended` — Transfer Acceleration through CloudFront edge locations.
- **request_payer**: `BucketOwner` (default) or `Requester`.
- **intelligent_tiering_configurations**: Named archive-tier configurations (ARCHIVE_ACCESS ≥ 90 days, DEEP_ARCHIVE_ACCESS ≥ 180 days) for objects in the INTELLIGENT_TIERING storage class.

### Governance: ABAC, Analytics, Inventory, Metrics, S3 Metadata

- **abac_status**: `Enabled` lets IAM policies authorize by principal tags vs bucket tags (attribute-based access control); `Disabled` states the off posture; empty leaves AWS's default unmanaged.
- **analytics_configurations**: Named storage-class-analysis configurations (prefix/tag scoped) with an optional daily CSV `export` to a bucket — the data source for choosing lifecycle transition ages.
- **inventory_configurations**: Named scheduled inventory reports (`Daily`/`Weekly`, `All`/`Current` versions, CSV/ORC/Parquet) with selectable metadata columns, delivered to a destination bucket (which may be the bucket itself — ARNs are name-derived). The destination needs a bucket policy allowing `s3.amazonaws.com` to `s3:PutObject` for delivery; use `disabled: true` to keep a configuration defined but paused.
- **metrics_configurations**: Named CloudWatch request-metrics configurations scoped by access point, prefix, and/or tags (standard CloudWatch per-metric cost).
- **metadata_configuration**: S3 Metadata — AWS-managed Apache Iceberg tables of this bucket's objects: the change-journal table (with its required record-expiration policy, ≥ 7 days when enabled) and an optional live inventory table, each with optional KMS encryption. Query them with Athena/Spark instead of listing the bucket.

## Outputs

- **bucket_id**: Bucket name — the identifier log destinations, code sources, and object sets reference.
- **bucket_arn**: Bucket ARN (`arn:aws:s3:::name`) — used in IAM/bucket policies and by ARN-consuming services.
- **region**: Region the bucket lives in.
- **bucket_regional_domain_name**: `name.s3.region.amazonaws.com` — the regional endpoint, also the right CloudFront origin domain.
- **bucket_domain_name**: `name.s3.amazonaws.com` — the legacy global-path domain.
- **hosted_zone_id**: Route53 hosted zone ID of the bucket's region, for alias records.
- **website_endpoint** / **website_domain**: Populated only when website hosting is configured; the direct HTTP website address and the Route53 alias target for it.

## How It Works

When you define an AwsS3Bucket resource, Planton:

1. **Creates the bucket** named from `metadata.name` (immutable, globally unique) with identity tags.
2. **Applies the security posture** — public-access block (all guards on unless relaxed), ownership controls (`BucketOwnerEnforced` unless overridden), optional canned ACL, and the bucket policy.
3. **Configures data protection** — versioning, default encryption, Object Lock default retention, and replication (rules, RTC, replica KMS).
4. **Configures data management** — lifecycle rules, Intelligent-Tiering archive configurations, transfer acceleration, requester pays.
5. **Configures integration surfaces** — website hosting, server access logging, CORS, and event notifications.

Both the Terraform/OpenTofu and Pulumi modules implement the same contract with identical outputs.

## Use Cases

### Application Data Store
A private, versioned, KMS-encrypted bucket with noncurrent-version pruning — the default posture for anything holding real data.

### Static Website
A website-mode bucket with a scoped public-read policy and relaxed policy guards — or better, a private bucket used as a CloudFront origin with Origin Access Control.

### Log Archive
A destination bucket with lifecycle tiering (STANDARD_IA → GLACIER) and hard expiration; other buckets point their `logging.target_bucket` here.

### Event-Driven Processing
`notification.eventbridge: true` streams object events into EventBridge rules, or Lambda/SQS/SNS targets deliver directly to consumers.

### Disaster Recovery
Cross-region replication with Replication Time Control to a versioned destination bucket in another region, replicating SSE-KMS objects under a destination-region key.

## Important Notes

- **Bucket names are global and immutable** — `metadata.name` must be unique across ALL AWS accounts and cannot be renamed.
- **Public access is deliberate** — serving public content is best done via CloudFront + Origin Access Control; a directly public bucket needs both a policy statement AND relaxed `public_access_block` guards.
- **Versioning is one-way** — enabled buckets can only be suspended, never unversioned; AWS rejects the transition at apply time.
- **SQS/SNS/Lambda notification targets must grant delivery first** — configure the queue policy / topic policy / Lambda invoke permission before pointing a notification at it, or AWS rejects the configuration.
- **COMPLIANCE-mode Object Lock is absolute** — not even the root account can shorten or bypass retention until it expires.

## References

- [Amazon S3 User Guide](https://docs.aws.amazon.com/AmazonS3/latest/userguide/Welcome.html)
- [Blocking public access](https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-control-block-public-access.html)
- [Lifecycle management](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html)
- [Replication](https://docs.aws.amazon.com/AmazonS3/latest/userguide/replication.html)
- [Object Lock](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lock.html)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
