# AwsBedrockInvocationLogging — Kind Guide

Authored operational judgment for the Bedrock invocation-logging
singleton: the design decisions behind the spec's shape, and what to
know before operating region-wide model-call auditing.

## Design decisions

- **A settings singleton, not a model-access field.** The
  configuration is region-wide observability delivery, not any one
  model's access — folding it into AwsBedrockModelAccess would make
  multiple access instances fight over one logging object. The region
  is the identity; deploy at most one instance per region.
- **At least one destination is required by CEL.** AWS accepts a
  destination-less configuration shape upstream, but it delivers
  nothing — dead config the spec refuses.
- **The data-type toggles are presence-typed.** AWS defaults all four
  to on; unset inherits that, explicit false is sendable. Turning
  video/embedding off is the usual cost lever for text-first
  workloads.

## Operating invocation logging in production

- **Two authorization models, one configuration.** CloudWatch
  delivery authenticates through the ROLE (trust
  `bedrock.amazonaws.com`, `logs:CreateLogStream` +
  `logs:PutLogEvents` on the group); S3 delivery authenticates
  through the BUCKET POLICY (`s3:PutObject` for the Bedrock service
  principal — add an `aws:SourceAccount` condition in production).
  A role with S3 permissions does nothing for the S3 arm.
- **Size the CloudWatch arm for 256 KB events.** Model payloads
  routinely exceed the CloudWatch event cap; without
  `largeDataDeliveryS3`, oversized bodies are truncated out of the
  stream. CloudWatch for querying + S3 for full-fidelity retention is
  the canonical pairing.
- **Invocation logs contain your prompts.** They are sensitive by
  nature — scope log-group and bucket access like you scope the
  models themselves.
- **AWS validates the permission chain at apply** ("Failed to
  validate permissions for log group") and both engines retry through
  IAM propagation lag; a persistent failure means the role/policy is
  actually wrong, not slow.
- **Bedrock leaves permission-check canaries in your buckets.** Configuring any S3 destination writes zero-byte `amazon-bedrock-logs-permission-check` objects under every configured prefix — with zero invocations ever made — and they survive deleting the logging configuration. A bucket that has ever been a destination is never empty: give it `force_destroy` (or a lifecycle rule) if you expect to delete it later, or its deletion fails `BucketNotEmpty`.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
