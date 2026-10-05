# AWS S3 Depth Pass and the Kinesis Family to 90/10

**Date:** 2026-07-06
**Scope:** Kinds #31–#34 — `AwsS3Bucket` (breaking rebuild), `AwsKinesisStream`, `AwsKinesisStreamConsumer`, `AwsKinesisFirehose` (audit-only)

## Summary

`AwsS3Bucket` is rebuilt from an 11-field 80/20 spec to the full provider
surface with every bucket-scoped satellite folded into one declarative
document, closing the Pulumi replication silent-drop parity gap. The Kinesis
family comes to the bar: the stream closes two stale "not in the pinned SDK"
deferrals and gains warm throughput and a folded resource policy, the consumer
gets honest tags and its own resource-policy fold, and firehose ships its
approved audit-only fixes with the destination-breadth deferral recorded. All
four kinds are enrolled in the outputs-conformance and variables.tf drift
guards, the Kinesis family gains first-ever E2E coverage, and all 8 live
dual-engine lanes are green with a zero-orphan sweep.

## AwsS3Bucket (breaking rebuild)

- **Retired (breaking, zero users):** the per-kind `tags` map (identity tags
  derive from metadata platform-wide), the dishonest `is_public` boolean
  (public access is never one flag on AWS), and the proto enums for
  versioning/encryption (provider-authentic strings now).
- **The honest security posture:** `public_access_block` (absence = all four
  guards ON — fully private), `object_ownership` (default
  `BucketOwnerEnforced`, ACLs disabled) with a CEL-gated canned `acl`, and the
  bucket `policy` as a native Struct — the primary access-control surface.
  Both modules always create the public-access-block and ownership satellites
  so the posture is visible in state rather than implied by absence.
- **Full data protection:** `versioning_status` (Enabled/Suspended),
  `encryption` (AES256 / aws:kms / aws:kms:dsse, KMS key reference, bucket
  key), full v6 `lifecycle_rules` (multi-predicate AND filters with tags and
  object-size bounds, date-XOR-days transitions, noncurrent-version handling
  with retained-version counts, multipart-upload cleanup,
  `transition_default_minimum_object_size`), and full `replication` (RTC +
  metrics couplings, replica KMS, cross-account ownership translation,
  delete-marker/existing-object/replica-modification/SSE-KMS arms — all
  couplings CEL-enforced; bucket and role by reference).
- **Full integration surface:** `website` (index-XOR-redirect + routing
  rules), `logging` (bucket reference + Athena-friendly partitioned prefix),
  `cors_rules`, `notification` (EventBridge plus Lambda/SQS/SNS targets by
  reference with event filters), `object_lock_enabled` +
  `object_lock_default_retention` (GOVERNANCE/COMPLIANCE, days XOR years),
  `acceleration_status`, `request_payer`, and per-name
  `intelligent_tiering_configurations` (archive-tier day floors CEL-enforced).
- **Pulumi replication parity gap CLOSED:** the module previously logged a
  warning and silently dropped `replication`; it now implements the full
  surface. The module is refactored per-concern (`bucket.go`, `settings.go`,
  `lifecycle.go`, `replication.go`, `website.go`) matching the sibling shape.
- **Outputs:** `bucket_id`/`bucket_arn` frozen (12 foreign-key consumer
  fields across 11 kinds join on them); added `bucket_domain_name`,
  `website_endpoint`, `website_domain`.
- **Terraform:** rewritten module on a generator-owned typed `variables.tf`;
  pin lifted `>= 5.0` → `>= 6.0.0` for the v6 lifecycle/replication shapes.

## AwsKinesisStream

- **Stale SDK deferrals closed:** `max_record_size_in_kib` (the 10 MiB
  large-record surface) was spec'd but commented as "not available in the
  pinned SDK" in BOTH engines — it has been available since TF provider
  6.20.0 / pulumi-aws v7.x. Now wired in both; the deferral comments deleted.
- **New surface:** `warm_throughput_mib_ps` (pre-provisioned burst capacity
  for ON_DEMAND streams; CEL: conflicts with `shard_count`) and a folded
  `resource_policy` Struct (cross-account grants without role assumption —
  the satellite targets the stream ARN and has no identity of its own).
- **Pin lifted** `= 5.82.0` → `>= 6.48.0` (warm throughput's floor), and the
  TF contract regenerated from a hand-written `type = any` legacy to the
  typed generator-owned shape.

## AwsKinesisStreamConsumer

- **Tags honesty:** the TF module computed identity tags into a local that
  was never applied to the consumer (provider support landed in 6.2.0);
  they now apply. Pin floor `>= 6.2.0`, typed contract regenerated.
- **`resource_policy` fold** for cross-account enhanced fan-out
  (SubscribeToShard/DescribeStreamConsumer grants against the consumer ARN).

## AwsKinesisFirehose (audit-only, as approved)

- Small-defect fixes: the tags local applied under the platform convention,
  and `try(x.value, x)` unwrap shims deleted (the generated contract already
  flattens `StringValueOrRef` to plain strings — the shims masked the real
  contract shape). Typed contract regenerated.
- **Deferral ledger recorded** (README + DD-005): Splunk, Snowflake, Iceberg,
  legacy Elasticsearch, and OpenSearch Serverless destinations, the MSK
  source, and Secrets Manager credentials — each brings its own credential
  surface; the four shipped destinations (extended_s3 / OpenSearch /
  HTTP endpoint / Redshift) carry the real demand today.

## E2E: first-ever Kinesis coverage + the S3 full-surface lane

- **Three new verifiers** (`DescribeStreamSummary`, `DescribeStreamConsumer`,
  `DescribeDeliveryStream`) with the DELETING-status lifecycle semantics the
  DynamoDB/RDS class uses (mid-deletion = absent).
- **Registry prerequisites:** consumer → `[AwsKinesisStream]`; firehose →
  `[AwsS3Bucket, AwsIamRole]` (every destination needs an S3 target and a
  delivery role). The stream gains an install profile
  (`e2e/prerequisite.yaml`, ON_DEMAND + `enforceConsumerDeletion`); the
  shared IAM-role fixture grows its 8th document (a firehose-assumable
  delivery role scoped to the fixture bucket).
- **Scenarios:** stream full-surface (extended retention, AWS-managed-key
  KMS, 2 MiB max record size, shard-level metrics, resource policy whose
  condition value carries a literal `${...}` — the tfvars template-escaping
  probe), consumer chain, firehose Direct PUT → extended_s3 chain (SSE,
  GZIP, `!{...}` prefix expressions), and the S3 full-surface scenario
  (versioning + KMS bucket key + TLS-only policy + multi-predicate lifecycle
  filters + website routing rules + EventBridge notification +
  Intelligent-Tiering).
- **All 8 live dual-engine lanes green:** S3 1m41s+33s (Pulumi) /
  3m15s+1m40s (Terraform, two scenarios each); stream 1m22s / 3m23s;
  consumer chain 1m33s / 2m26s; firehose chain 5m47s / 6m30s. Zero-orphan
  sweep clean across Kinesis, Firehose, S3, and IAM.

## Defects the live lanes caught (all fixed)

- **S3 Terraform null-deref on minimal manifests:** `local.manage_encryption
  && var.spec.encryption.sse_algorithm != ""` — HCL evaluates both `&&`
  operands, so an absent `encryption` block crashed the module. Guarded with
  `try()`.
- **A missing AWS coupling, promoted to CEL:** AWS rejects
  `AbortIncompleteMultipartUpload` on lifecycle rules whose filter uses tags
  or object-size bounds (in-progress uploads have neither). Now a spec CEL
  with three test cases, caught at authoring time instead of apply time.
- **Kinesis resource-policy constraints (recorded in scenarios):** wildcard
  principals are rejected, and the policy's `Resource` must equal the target
  ARN exactly — which is why the consumer scenario cannot carry a static
  policy (a consumer ARN embeds its registration timestamp) and the stream
  scenario spells its deterministic ARN out.
- **Warm throughput is entitlement-gated:** `CreateStream` rejects
  `warm_throughput_mib_ps` unless the account carries the minimum-throughput
  billing commitment. The field stays (the live rejection proved the wiring
  reaches AWS); the scenario documents the exclusion.

## Guards

- All four kinds enrolled in the outputs-conformance suite and the
  `variables.tf` drift guard; kind-registry stubs and kind map regenerated
  for the new prerequisites; full offline gate green (spec tests ×4,
  `tofu init && validate` ×4, Pulumi builds ×4, validate-refs,
  secret-coverage, validate-outputs ×4, e2e discover, `make build-go`).

## Breaking changes

Zero users; no migration. For the record — `AwsS3Bucket`: `tags`,
`is_public`, and the versioning/encryption enums removed; `versioning_status`
and `encryption.sse_algorithm` are provider strings; lifecycle/replication
shapes are the full v6 forms. The six S3-creating charts
(data-analytics, kafka-streaming, ml-workbench, pulumi-backend,
static-website, terraform-backend) were already migrated to the new shape by
the concurrent chart-fleet fix wave — verified: no retired field remains
under `charts/aws`.
