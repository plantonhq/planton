# Pulumi Module to Deploy AliCloudStorageBucket

This module provisions an Alibaba Cloud OSS bucket with configurable access control, storage class, redundancy, versioning, server-side encryption, lifecycle rules, CORS configuration, and access logging. It creates a single `oss.Bucket` resource and exports the bucket name and both public and VPC-internal endpoints.

Generated from the proto schema for `AliCloudStorageBucket`.

## CLI Usage (Planton Pulumi)

```bash
# Preview changes
planton pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Apply changes
planton pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh state from cloud
planton pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Tear down
planton pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

**Note**: Alibaba Cloud credentials are provided via environment variables (`ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`) or through the Planton provider config -- not in the manifest `spec`.

## What This Module Creates

- **OSS Bucket** (`oss.Bucket`) -- an object storage bucket with the specified storage class and redundancy type
- **Versioning** -- optionally enabled to preserve all object versions
- **Server-Side Encryption** -- optionally configured with AES256 or KMS
- **Lifecycle Rules** -- automated object transitions and expiration
- **CORS Rules** -- cross-origin access for browser-based clients
- **Access Logging** -- server access logs written to a target bucket

Tags from `spec.tags` are merged with system tags (resource name, kind, organization, environment). User tags take precedence on key conflict.

## Prerequisites

- [Planton CLI](https://github.com/plantonhq/planton) installed
- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) installed
- Alibaba Cloud credentials configured via environment variables or Planton provider config

## Usage

1. Create or edit a manifest:

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudStorageBucket
metadata:
  name: my-bucket
spec:
  region: cn-hangzhou
  bucketName: my-app-bucket
```

2. Preview:

```bash
planton pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

3. Apply:

```bash
planton pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `bucket_name` | The bucket name (also serves as the bucket ID) |
| `extranet_endpoint` | Public internet endpoint for the bucket |
| `intranet_endpoint` | VPC-internal endpoint for zero-cost intra-region access |

## Further Reading

- [examples.md](./examples.md) -- runnable manifest examples with CLI commands
- [overview.md](./overview.md) -- module architecture and design decisions
- [hack/manifest.yaml](../hack/manifest.yaml) -- minimal test manifest
