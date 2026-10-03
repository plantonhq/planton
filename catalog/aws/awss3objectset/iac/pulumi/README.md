# Pulumi Module to Deploy AwsS3ObjectSet

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
name: awss3objectset-pulumi-project
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

# AWS S3 Object Set Pulumi Module

## Introduction

The AWS S3 Object Set Pulumi Module provides a standardized way to upload and manage one or more S3 objects in a target bucket using a Kubernetes-like API resource model. Developers specify their objects in a YAML manifest, and the module creates `s3.BucketObjectv2` resources for each entry through Pulumi.

## Key Features

- **Multi-Object Upload**: Upload multiple objects to a single bucket in one deployment. Each resource is named by its S3 key, so reordering manifest entries never churns unrelated objects.
- **Foreign Key Bucket Reference**: Reference an AwsS3Bucket component or provide a literal bucket name.
- **Content Flexibility**: Support for inline text (`content`) and base64-encoded binary (`content_base64`).
- **Tag Inheritance**: Resource-identity labels, set-level tags, and object-level tags merge in increasing precedence.
- **Full Per-Object Surface**: Presentation headers (content type/disposition/language, cache control, encoding), lowercase-keyed user metadata, website redirects, storage class, per-object encryption overrides (SSE-S3/SSE-KMS with an AwsKmsKey reference), upload checksums, Object Lock retention and legal holds with the governance-bypass force_destroy, and canned ACLs.
- **Status Outputs**: Captures ARNs, ETags, and version IDs for each uploaded object plus the target bucket.

## Architecture

The module iterates over the `objects` list in the spec and creates one `s3.BucketObjectv2` Pulumi resource per entry, named by the object's S3 key. Tags are merged hierarchically: labels, set-level tags, then object-level tags. ARNs, ETags, and version IDs are collected into maps and exported as outputs.

## Usage

Refer to the example section for usage instructions.
