# DigitalOcean Spaces Bucket

An S3-compatible object-storage bucket on DigitalOcean Spaces, described once in a Planton manifest: region and canned ACL, object versioning, lifecycle rules that expire current or noncurrent versions and abort stale multipart uploads, CORS for browser applications, a JSON bucket policy, access logging to another bucket, and the force-destroy safety flag.

## What this component models

The spec maps onto DigitalOcean's Spaces bucket plus the three per-bucket settings satellites whose lifecycle is identical to the bucket's:

| Spec field | What it controls |
|---|---|
| `bucketName` | The bucket's name (DNS-compatible, 3–63 chars; unique per region across all DigitalOcean customers) |
| `region` | Spaces-capable region slug; omit to let the provider apply its own default (`nyc3`); changing it replaces the bucket |
| `accessControl` | Canned ACL: `PRIVATE` (default) or `PUBLIC_READ` |
| `versioningEnabled` | Object versioning; once enabled it can never be removed, only suspended |
| `forceDestroy` | When true, destroy empties the bucket — every object AND every object version — before deleting it |
| `lifecycleRules` | Expire current or noncurrent versions and abort incomplete multipart uploads |
| `corsRules` | Browser CORS, managed through the standalone CORS-configuration resource (the bucket's inline `cors_rule` is deprecated and does no drift detection) |
| `policy` | Bucket policy as a JSON document (S3 policy grammar) |
| `logging` | Access logs delivered to another bucket (a literal name or a `DigitalOceanBucket` reference) |

Spaces buckets are not taggable. There is no `tags` field.

## Quick start

The smallest real bucket — DigitalOcean applies its own default region (`nyc3`):

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanBucket
metadata:
  name: app-data
spec:
  bucketName: app-data
```

```shell
planton apply -f app-data.yaml
```

## Outputs

Both provisioners export the identical output set:

| Output | Description |
|---|---|
| `bucket_id` | The bucket's name — Spaces buckets have no separate UUID (import id is `<region>,<name>`) |
| `endpoint` | Region-level host (`<region>.digitaloceanspaces.com`, no scheme, no bucket name) |
| `region` | The region slug the bucket landed in (the provider default when `region` was omitted) |
| `bucket_domain_name` | Virtual-host FQDN (`<bucket>.<region>.digitaloceanspaces.com`) |
| `urn` | `do:space:<name>` |

## Behavior worth knowing

- **Bucket names are globally unique per region** across every DigitalOcean customer. A collision is a create-time error, not a rename.
- **CORS, policy, and logging require an explicit `region`.** Those satellites are separate provider resources whose region argument is required; the spec rejects a satellite without a region before any provisioner runs.
- **Versioning is one-way.** Enabling it keeps every overwrite and delete; flipping `versioningEnabled` back to false only suspends it.
- **`forceDestroy` is irreversible for the bucket's data.** Leave it false for anything you cannot lose; the e2e `full` scenario sets it true so teardown cannot stall on leftover objects.
- **Access logging needs a second bucket** to receive the logs. Logging a bucket to itself works but compounds: reads of the logs generate more logs.
- **Spaces is a second credential plane.** Deploys and verification need `SPACES_ACCESS_KEY_ID` / `SPACES_SECRET_ACCESS_KEY` alongside the DigitalOcean API token. See the [GUIDE](GUIDE.md).
- **Always address the bucket through its `region` output.** Spaces does not redirect: a bucket asked for on the wrong regional endpoint answers `404`, exactly like a bucket that no longer exists. See the [GUIDE](GUIDE.md).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
