# Pulumi Module to Deploy AwsS3ObjectSet

## CLI usage (OpenMCF pulumi)

```bash
# Preview
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Update (apply)
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh
openmcf pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Destroy
openmcf pulumi destroy \
  --manifest ../hack/manifest.yaml \
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
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

# AWS S3 Object Set Pulumi Module

## Introduction

The AWS S3 Object Set Pulumi Module provides a standardized way to upload and manage one or more S3 objects in a target bucket using a Kubernetes-like API resource model. Developers specify their objects in a YAML manifest, and the module creates `s3.BucketObjectv2` resources for each entry through Pulumi.

## Key Features

- **Multi-Object Upload**: Upload multiple objects to a single bucket in one deployment.
- **Foreign Key Bucket Reference**: Reference an AwsS3Bucket component or provide a literal bucket name.
- **Content Flexibility**: Support for inline text (`content`) and base64-encoded binary (`content_base64`).
- **Tag Inheritance**: Set-level tags are merged with object-level tags, with object tags taking precedence.
- **Per-Object Metadata**: Configure content type, cache control, content encoding, and ACL per object.
- **Status Outputs**: Captures ETags and version IDs for each uploaded object.

## Architecture

The module iterates over the `objects` list in the spec and creates one `s3.BucketObjectv2` Pulumi resource per entry. Tags are merged hierarchically: labels, set-level tags, then object-level tags. ETags and version IDs are collected into maps and exported as stack outputs.

## Usage

Refer to the example section for usage instructions.
