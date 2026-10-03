# Pulumi Module to Deploy AwsS3Bucket

## CLI usage (Planton pulumi)

```bash
# Preview
planton pulumi preview \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Update (apply)
planton pulumi update \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh
planton pulumi refresh \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Destroy
planton pulumi destroy \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

## Debugging

This module includes a `debug.sh` helper. To enable debugging, edit `Pulumi.yaml` and uncomment the `runtime.options.binary` line so Pulumi runs the program via the script:

```yaml
name: aws-module-test-pulumi-project
runtime:
  name: go
#  options:
#    binary: ./debug.sh
```

Then make the script executable and run your command (e.g., `preview` or `update`). See `docs/pages/docs/guide/debug-pulumi-modules.mdx` for full instructions.

```bash
chmod +x debug.sh
planton pulumi preview \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

# AWS S3 Bucket Pulumi Module

## Introduction

This Pulumi module (Go) deploys an Amazon S3 bucket from an `AwsS3Bucket` API resource definition. It provisions the bucket root resource plus one satellite resource per configured spec block, mirroring how AWS itself models bucket behavior — and implements the exact same contract as the Terraform module, with identical outputs.

## What It Creates

- **Bucket root** — named from `metadata.name`, with identity tags and `force_destroy` handling
- **Security posture (always created)** — the public-access block (all four guards on unless the spec relaxes them) and ownership controls (`BucketOwnerEnforced` unless overridden), plus the optional canned ACL and bucket policy
- **Data protection** — versioning, default server-side encryption (SSE-S3 / SSE-KMS / DSSE-KMS with optional bucket key), Object Lock default retention, and the full replication configuration (rule filters, delete-marker replication, RTC + metrics, replica KMS keys, cross-account ownership translation)
- **Data management** — lifecycle configuration (filters with the single-vs-`and` document shaping, transitions, expiration, noncurrent-version handling, multipart cleanup), per-name Intelligent-Tiering archive configurations, transfer acceleration, requester pays
- **Integration surfaces** — website or redirect-all hosting with routing rules, server access logging with partitioned key formats, CORS, and event notifications (EventBridge and/or Lambda/SQS/SNS targets)

The module code is organized by concern (`bucket.go`, `lifecycle.go`, `replication.go`, `website.go`, `settings.go`) so each satellite's provider quirks are documented where they are handled.

## Module Structure

- **`main.go` (module root)** — loads the IaC input and delegates to `module.Resources`
- **`module/main.go`** — orchestrates the bucket and its satellites, exports outputs (website outputs are exported as empty strings when hosting is not configured so the output contract stays shape-stable across engines)
- **`module/locals.go`** — identity tags and normalized views of the spec
- **AWS Provider** — built via the shared provider builder from the IaC input's `provider_config`, resolving static keys, keyless web identity, or the ambient credential chain

## Usage

See the CLI usage section above; a working manifest lives at [`../../e2e/manifest.yaml`](../../e2e/manifest.yaml) and remixable starting points in [`../../presets/`](../../presets/).
