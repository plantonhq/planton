# Cloudflare R2 Bucket

Deploys a Cloudflare R2 object storage bucket together with its bucket-scoped configuration: location and jurisdiction, default storage class, managed (r2.dev) public access, custom domains, CORS, object lifecycle, and object lock. Custom domains integrate with CloudflareDnsZone via foreign-key references.

## What Gets Created

When you deploy a CloudflareR2Bucket resource, Planton provisions:

- **R2 Bucket** — a `cloudflare_r2_bucket` in the specified account, with the configured location hint, jurisdiction, and default storage class.
- **Managed Public Domain** — when `publicAccess` is `true`, a `cloudflare_r2_managed_domain` enabling the bucket's `r2.dev` URL (published as the `public_url` output).
- **Custom Domains** — one `cloudflare_r2_custom_domain` per enabled entry in `customDomains`, serving the bucket over your own hostnames.
- **CORS** — a `cloudflare_r2_bucket_cors` when `cors.rules` are provided.
- **Lifecycle** — a `cloudflare_r2_bucket_lifecycle` when `lifecycle.rules` are provided (storage-class transitions, object expiration, multipart-upload cleanup).
- **Object Lock** — a `cloudflare_r2_bucket_lock` when `lock.rules` are provided (write-once retention).

## Prerequisites

- **Cloudflare credentials** configured via environment variables or Planton provider config
- **A Cloudflare account ID** (32-character hex string) with R2 enabled
- **A Cloudflare DNS zone** for any custom domain (the domain must be within that zone)

## Quick Start

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareR2Bucket
metadata:
  name: my-bucket
spec:
  bucketName: my-app-assets
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  location: wnam
```

```shell
planton apply -f r2-bucket.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `bucketName` | `string` | Name of the R2 bucket. Must be DNS-compatible. | 3–63 chars, pattern `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` |
| `accountId` | `string` | Cloudflare account ID where the bucket is created. | Exactly 32 hex characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `location` | `enum` | `auto` | Location hint: `auto`, `wnam`, `enam`, `weur`, `eeur`, `apac`, `oc`. Best-effort, honored only at creation. |
| `jurisdiction` | `string` | `default` | Data residency: `default`, `eu`, or `fedramp`. Fixed at creation; applies to the bucket and all sub-resources. |
| `storageClass` | `enum` | `Standard` | Default storage class for new objects: `Standard` or `InfrequentAccess`. |
| `publicAccess` | `bool` | `false` | Enable the managed `r2.dev` public URL (development-grade; rate-limited). |
| `customDomains[]` | `list` | `[]` | Custom domains serving the bucket. Each: `enabled`, `zoneId` (literal or `valueFrom` a CloudflareDnsZone), `domain`, optional `minTls` (`1.0`–`1.3`) and `ciphers`. |
| `cors.rules[]` | `list` | `[]` | CORS rules. Each: `allowed.methods` (GET/PUT/POST/DELETE/HEAD), `allowed.origins`, optional `allowed.headers`, `exposeHeaders`, `maxAgeSeconds`. |
| `lifecycle.rules[]` | `list` | `[]` | Lifecycle rules: `id`, `conditions.prefix`, `enabled`, optional `abortMultipartUploadsTransition`, `deleteObjectsTransition`, and `storageClassTransitions` (Age/Date conditions). |
| `lock.rules[]` | `list` | `[]` | Object-lock rules: `id`, `enabled`, optional `prefix`, and a `condition` of type `Age`, `Date`, or `Indefinite`. |

## Example: public bucket with custom domain and CORS

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareR2Bucket
metadata:
  name: media-bucket
spec:
  bucketName: prod-media-assets
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  location: weur
  publicAccess: true
  customDomains:
    - enabled: true
      zoneId:
        valueFrom:
          kind: CloudflareDnsZone
          name: my-zone
          field: status.outputs.zone_id
      domain: media.example.com
      minTls: "1.2"
  cors:
    rules:
      - allowed:
          methods: [GET, HEAD]
          origins: ["https://app.example.com"]
        maxAgeSeconds: 3600
```

## Stack Outputs

After deployment, these outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `bucket_name` | `string` | Name of the created R2 bucket |
| `bucket_url` | `string` | The S3-compatible API URL for the bucket |
| `custom_domain_urls` | `list(string)` | URLs of the configured custom domains (one per enabled custom domain) |
| `public_url` | `string` | The managed `r2.dev` public URL when `publicAccess` is enabled; empty otherwise |

## Related Components

- [CloudflareDnsZone](/docs/catalog/cloudflare/cloudflarednszone) — provides the DNS zone referenced by `customDomains[].zoneId`
- [CloudflareWorker](/docs/catalog/cloudflare/cloudflareworker) — commonly deployed alongside R2 to handle object transformations or access control
- [CloudflareKvNamespace](/docs/catalog/cloudflare/cloudflarekvnamespace) — key-value storage often paired with R2 for metadata indexing
