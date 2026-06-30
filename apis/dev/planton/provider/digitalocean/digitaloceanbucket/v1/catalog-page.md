# DigitalOcean Bucket

Deploys a DigitalOcean Spaces bucket, providing S3-compatible object storage in a specified datacenter region. The component configures bucket naming, access control, optional versioning, and tagging, exposing the bucket identifier and regional endpoint as stack outputs.

## What Gets Created

When you deploy a DigitalOceanBucket resource, Planton provisions:

- **Spaces Bucket** — a `digitalocean_spaces_bucket` resource with the specified name, region, ACL, and optional versioning configuration

## Prerequisites

- **DigitalOcean credentials** configured via environment variables or Planton provider config
- **Spaces API keys** (access key ID and secret key) if using S3-compatible access for uploading objects after provisioning
- **A globally unique bucket name** that is DNS-compatible (lowercase alphanumeric and hyphens, 3--63 characters)

## Quick Start

Create a file `bucket.yaml`:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanBucket
metadata:
  name: my-assets
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.DigitalOceanBucket.my-assets
spec:
  bucketName: my-assets
  region: nyc3
```

Deploy:

```shell
planton apply -f bucket.yaml
```

This creates a private Spaces bucket named `my-assets` in the NYC3 region with versioning disabled.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `bucketName` | `string` | Name of the Spaces bucket. Must be DNS-compatible. | Required, 3--63 characters, pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` |
| `region` | `enum` | DigitalOcean datacenter region for the bucket. Valid values: `nyc3`, `sfo3`, `fra1`, `sgp1`, `lon1`, `tor1`, `blr1`, `ams3`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `accessControl` | `enum` | `PRIVATE` | Bucket ACL. `PRIVATE` restricts access to the bucket owner. `PUBLIC_READ` allows unauthenticated read access to all objects. |
| `versioningEnabled` | `bool` | `false` | When `true`, enables object versioning on the bucket. Note: versioning cannot be disabled once enabled, only suspended. |
| `tags` | `string[]` | `[]` | Tags to apply to the bucket. Values must be unique. |

## Examples

### Private Bucket with Tags

A private bucket in Frankfurt tagged for a specific team and environment:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanBucket
metadata:
  name: team-logs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.DigitalOceanBucket.team-logs
spec:
  bucketName: team-logs
  region: fra1
  tags:
    - team:backend
    - env:dev
```

### Public-Read Bucket for Static Assets

A publicly readable bucket for hosting static website assets with versioning enabled:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanBucket
metadata:
  name: static-assets
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.DigitalOceanBucket.static-assets
spec:
  bucketName: static-assets
  region: sfo3
  accessControl: PUBLIC_READ
  versioningEnabled: true
  tags:
    - env:prod
    - purpose:static-hosting
```

### Versioned Backup Bucket

A private bucket with versioning enabled for data backups in Singapore:

```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanBucket
metadata:
  name: db-backups
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.DigitalOceanBucket.db-backups
spec:
  bucketName: db-backups
  region: sgp1
  versioningEnabled: true
  tags:
    - env:prod
    - purpose:backups
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `bucketId` | `string` | Unique identifier for the bucket (format: `region:bucket-name`) |
| `endpoint` | `string` | Regional endpoint URL for the bucket (e.g., `https://my-assets.nyc3.digitaloceanspaces.com`) |

## Related Components

- [DigitalOceanCertificate](/docs/catalog/digitalocean/digitaloceancertificate) — provides TLS certificates for CDN custom domains serving bucket content
- [DigitalOceanDnsRecord](/docs/catalog/digitalocean/digitaloceandnsrecord) — creates DNS records pointing to the bucket endpoint
