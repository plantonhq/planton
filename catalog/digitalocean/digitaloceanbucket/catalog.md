# DigitalOcean Spaces Bucket

Deploys a DigitalOcean Spaces object-storage bucket with configurable region and canned ACL, object versioning, lifecycle rules, CORS, a JSON bucket policy, access logging to another bucket, and the force-destroy safety flag. The decisions that stick: bucket names are globally unique per region across all DigitalOcean customers, changing the region replaces the bucket, and versioning can never be removed once enabled -- only suspended.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Spaces Bucket** -- an S3-compatible object-storage bucket; placed in the specified Spaces-capable region, or the provider's default (`nyc3`) when omitted
- **Versioning** -- created only when `versioningEnabled` is true; once enabled it can never be removed, only suspended
- **Lifecycle Rules** -- created only when `lifecycleRules` are provided; expire current or noncurrent object versions and abort stale multipart uploads
- **CORS Configuration** -- created only when `corsRules` are provided; managed through the standalone CORS-configuration resource so rules round-trip with real drift detection
- **Bucket Policy** -- created only when `policy` is provided; a JSON document in S3 policy grammar
- **Access Logging** -- created only when `logging` is provided; delivers S3-style access logs to another bucket

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token AND a Spaces key pair (`spacesAccessId` / `spacesSecretKey`). Spaces is a second credential plane the API token cannot reach. Map the connection as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline authentication.

### DigitalOcean Account

- **Spaces enabled on the account** -- activate Spaces in the control panel before the first bucket.
- **A Spaces key pair** -- mint under API -> Spaces Keys. Required for both provisioners and for verification.
- **A Spaces-capable region** (optional) -- `ams3`, `atl1`, `blr1`, `fra1`, `lon1`, `nyc3`, `sfo2`, `sfo3`, `sgp1`, `syd1`, `tor1`. Omit to use the provider default (`nyc3`). Required the moment you set CORS, policy, or logging.
- **A globally unique bucket name** -- unique per region across ALL DigitalOcean customers, DNS-compatible (lowercase, hyphens, 3–63 characters).

## Deploy

### Console

Open the deployment store, find **DigitalOcean Spaces Bucket**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Private Versioned Bucket** preset in the [Presets](#presets) tab to pre-populate a versioned private bucket with a lifecycle rule.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanBucket
metadata:
  name: app-assets
  org: acme-corp
  env: prod
spec:
  bucketName: acme-app-assets
  region: nyc3
  accessControl: PRIVATE
  versioningEnabled: true
```

```shell
planton apply -f do-bucket.yaml
```

This creates a private, versioned Spaces bucket in NYC3. An Infra Job tracks the provisioning in real time.

### InfraChart

When delivering access logs to another bucket in the same InfraPipeline, use ValueFromRef:

```yaml
spec:
  region: nyc3
  logging:
    targetBucket:
      valueFrom:
        kind: DigitalOceanBucket
        name: access-logs
        fieldPath: status.outputs.bucket_id
    targetPrefix: access-logs/
```

The InfraPipeline resolves the dependency graph, deploys the log-sink bucket first, then enables logging on the source.

## Key Configuration

These are the most important decisions when configuring a Spaces bucket. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Region** -- optional for a bare bucket (the provider defaults to `nyc3`), but required the moment `corsRules`, `policy`, or `logging` is set: those satellites are separate provider resources that cannot inherit the default. Changing the region replaces the bucket, so pin it explicitly on any bucket that might grow satellites later.

**Access control** -- `accessControl` sets the canned ACL: `PRIVATE` (default) restricts access to authenticated calls and signed URLs; `PUBLIC_READ` makes objects reachable by anyone with the URL. Finer grants (per-prefix, per-IP) belong on `policy`.

**Versioning** -- `versioningEnabled: true` keeps every overwrite and delete. It can never be removed, only suspended. Pair it with a lifecycle rule that expires noncurrent versions, or the storage bill grows silently.

**Lifecycle rules** -- expire current versions (`expiration.days` or `expiration.date`), expire noncurrent versions, and abort incomplete multipart uploads. A rule needs at least one action; an expiration sets exactly one trigger.

**CORS** -- `corsRules` let browser applications on other origins read the bucket. Managed through DigitalOcean's standalone CORS resource (the inline block is deprecated and does no drift detection).

**Force destroy** -- `forceDestroy: true` empties the bucket (every object and every version) before delete. Irreversible for the data; leave false for anything you cannot lose.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanBucket** (optional) | `logging.targetBucket` | `status.outputs.bucket_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `bucket_id` | The bucket's name (Spaces buckets have no UUID) | Logging targets (`logging.targetBucket` references), application configuration |
| `region` | The region slug the bucket landed in | Addressing, import (`<region>,<name>`) |
| `endpoint` | Region-level host (`<region>.digitaloceanspaces.com`) | S3 client configuration |
| `bucket_domain_name` | Virtual-host FQDN (`<bucket>.<region>.digitaloceanspaces.com`) | CDN origin, public URLs |
| `urn` | `do:space:<name>` | Project membership |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private versioned bucket** -- private ACL, versioning on, a lifecycle rule that expires noncurrent versions and aborts stale multipart uploads. Start from the **Private Versioned Bucket** preset.

**Public static-site assets** -- public-read ACL with a CORS rule letting the site's origin fetch assets. Start from the **Public Static Site Assets** preset.

## Works With

- [**DigitalOcean Spaces Bucket**](/infra-catalog/digital-ocean-bucket) -- another bucket as the access-log sink
- [**DigitalOcean CDN**](/infra-catalog/digital-ocean-cdn) -- fronts a public bucket at the edge, consuming the `bucket_domain_name` output as its origin
- [**DigitalOcean Spaces Access Key**](/infra-catalog/digital-ocean-spaces-key) -- mints per-bucket read or read-write credentials scoped to this bucket by name
