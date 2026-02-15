---
title: "R2 Bucket"
description: "R2 Bucket deployment documentation"
icon: "package"
order: 100
componentName: "cloudflarer2bucket"
---

# Cloudflare R2 Bucket

Deploys a Cloudflare R2 object storage bucket with a configurable location hint and optional custom domain access via an R2 managed domain. The component supports all six R2 location regions and integrates with CloudflareDnsZone for custom domain configuration.

## What Gets Created

When you deploy a CloudflareR2Bucket resource, OpenMCF provisions:

- **R2 Bucket** — a `cloudflare_r2_bucket` resource in the specified Cloudflare account with the configured location hint
- **R2 Custom Domain** — created only when `customDomain.enabled` is `true`, attaches a custom domain to the bucket via a `cloudflare_r2_custom_domain` resource so the bucket is accessible at the specified hostname

## Prerequisites

- **Cloudflare credentials** configured via environment variables or OpenMCF provider config
- **A Cloudflare account ID** (32-character hex string) with R2 enabled
- **A Cloudflare DNS zone** if enabling custom domain access (the domain must be within that zone)

## Quick Start

Create a file `r2-bucket.yaml`:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: my-bucket
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CloudflareR2Bucket.my-bucket
spec:
  bucketName: my-app-assets
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  location: WNAM
```

Deploy:

```shell
openmcf apply -f r2-bucket.yaml
```

This creates an R2 bucket named `my-app-assets` in Western North America with no public or custom domain access.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `bucketName` | `string` | Name of the R2 bucket. Must be DNS-compatible. | 3–63 characters, pattern `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` |
| `accountId` | `string` | Cloudflare account ID where the bucket is created. | Exactly 32 hex characters, pattern `^[0-9a-fA-F]{32}$` |
| `location` | `enum` | Location hint for the bucket's primary storage region. | One of: `auto`, `WNAM`, `ENAM`, `WEUR`, `EEUR`, `APAC`, `OC` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `publicAccess` | `bool` | `false` | Expose the bucket via Cloudflare's managed `r2.dev` public URL. Note: this currently requires manual enablement via the Cloudflare Dashboard or API. |

| `customDomain.enabled` | `bool` | `false` | Enables custom domain access for the bucket. When `true`, `customDomain.zoneId` and `customDomain.domain` are required. |
| `customDomain.zoneId` | `string` | — | Cloudflare Zone ID where the custom domain is configured. Can reference a CloudflareDnsZone resource via `valueFrom`. Required when `customDomain.enabled` is `true`. |
| `customDomain.domain` | `string` | — | Fully qualified domain name for accessing the bucket (e.g., `media.example.com`). Must be within the zone specified by `customDomain.zoneId`. Maximum 253 characters. Required when `customDomain.enabled` is `true`. |

## Examples

### Basic Bucket in Auto Region

A minimal R2 bucket where Cloudflare selects the optimal storage location:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: logs-bucket
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CloudflareR2Bucket.logs-bucket
spec:
  bucketName: app-logs
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  location: auto
```

### Bucket with Custom Domain

An R2 bucket accessible via a custom domain, useful for serving static assets:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: media-bucket
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareR2Bucket.media-bucket
spec:
  bucketName: prod-media-assets
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  location: ENAM
  customDomain:
    enabled: true
    zoneId: a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6
    domain: media.example.com
```

### Full-Featured Bucket with Foreign Key References

Production configuration referencing an OpenMCF-managed DNS zone for custom domain setup:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareR2Bucket
metadata:
  name: prod-assets
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareR2Bucket.prod-assets
spec:
  bucketName: prod-static-assets
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  location: WEUR
  customDomain:
    enabled: true
    zoneId:
      valueFrom:
        kind: CloudflareDnsZone
        name: my-zone
        field: status.outputs.zone_id
    domain: assets.example.com
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `bucket_name` | `string` | Name of the created R2 bucket |
| `bucket_url` | `string` | The accessible bucket URL (R2 public endpoint or base S3 API URL) |
| `custom_domain_url` | `string` | The custom domain URL (e.g., `https://media.example.com`). Only set when `customDomain.enabled` is `true`. |

## Related Components

- [CloudflareDnsZone](/docs/catalog/cloudflare/dns-zone) — provides the DNS zone referenced by `customDomain.zoneId`
- [CloudflareWorker](/docs/catalog/cloudflare/worker) — commonly deployed alongside R2 to handle object transformations or access control
- [CloudflareKvNamespace](/docs/catalog/cloudflare/kv-namespace) — key-value storage often paired with R2 for metadata indexing
