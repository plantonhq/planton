# GCP Backend Bucket

Deploys a Compute Engine backend bucket (`google_compute_backend_bucket`) — the node that serves a Cloud Storage bucket's objects through an external HTTP(S) load balancer, optionally cached at Google's edge by Cloud CDN. It is the static-content counterpart of a backend service: URL maps route paths like `/assets/*` here while dynamic paths go to backend services.

## What Gets Created

When you deploy a GcpBackendBucket resource, Planton provisions:

- **Backend Bucket** — a `google_compute_backend_bucket` pointing at the origin GCS bucket, with CDN policy, compression, custom headers, and edge security policy
- **Signed-URL Keys** (optional) — one `google_compute_backend_bucket_signed_url_key` per entry in `signedUrlKeys`, for serving private content through the CDN with expiring, tamper-proof links

The origin bucket itself is deliberately NOT created here — reference an existing GcpGcsBucket. One bucket can sit behind several backend buckets with different CDN policies, and swapping the origin is an in-place update that leaves the URL map untouched.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCS bucket** — referenced via `bucketName` (a GcpGcsBucket resource or a literal name); objects must be publicly readable or served via signed URLs
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs

## Quick Start

Create a file `backend-bucket.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBackendBucket
metadata:
  name: static-assets
spec:
  projectId:
    value: my-gcp-project-123
  bucketName:
    valueFrom:
      kind: GcpGcsBucket
      name: my-static-assets
      fieldPath: status.outputs.bucket_id
  enableCdn: true
  cdnPolicy:
    cacheMode: CACHE_ALL_STATIC
    defaultTtl: 3600
```

Deploy:

```shell
planton apply -f backend-bucket.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `bucketName` | `StringValueOrRef` | The Cloud Storage bucket whose objects are served. Can reference a GcpGcsBucket. Mutable — origin swaps are in-place. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | Project that owns the backend bucket (may differ from the bucket's project). Immutable. |
| `backendBucketName` | `string` | `metadata.name` | Cloud-side name (RFC1035). Immutable. |
| `description` | `string` | `""` | What this backend bucket serves and which URL maps use it. |
| `enableCdn` | `bool` | `false` | Cache responses at Google's edge with Cloud CDN. |
| `cdnPolicy` | object | GCP defaults | How responses are cached — see below. |
| `compressionMode` | `string` | disabled | `AUTOMATIC` (gzip/brotli for compressible types) or `DISABLED`. |
| `customResponseHeaders` | `list(string)` | `[]` | Headers the LB adds, `"Header-Name: value"` form; values may use `{cdn_cache_status}`. |
| `edgeSecurityPolicy` | `StringValueOrRef` | none | Cloud Armor EDGE policy (type `CLOUD_ARMOR_EDGE`) filtering requests before cache and origin. Can reference a GcpCloudArmorPolicy. |
| `loadBalancingScheme` | `string` | external | `INTERNAL_MANAGED` for cross-region internal ALBs. Incompatible with CDN. Immutable. |
| `signedUrlKeys` | `list(object)` | `[]` | Up to 3 named signing keys for Cloud CDN signed URLs/cookies. `keyValue` is secret material. |
| `resourceManagerTags` | `map<string,string>` | `{}` | Create-time Resource Manager tags (`tagKeys/{id}` → `tagValues/{id}`) for org policy and IAM conditions. Immutable. |
| `deletionPolicy` | `string` | `DELETE` | What destroy does to the backend bucket AND its signed-URL keys: `DELETE`, `PREVENT` (refuse), or `ABANDON` (keep serving, drop from management). |

### CDN Policy

| Field | Default (GCP) | Description |
|-------|---------------|-------------|
| `cacheMode` | `CACHE_ALL_STATIC` | `CACHE_ALL_STATIC` (static types + honor origin headers), `USE_ORIGIN_HEADERS` (origin controls everything; TTLs must be unset), `FORCE_CACHE_ALL` (cache everything; `maxTtl` must be unset — never use with private content) |
| `clientTtl` / `defaultTtl` / `maxTtl` | 3600 / 3600 / 86400 | Cache lifetimes in seconds; 0 = let GCP default |
| `negativeCaching` + `negativeCachingPolicy` | off | Cache error/redirect responses per status code (codes 300, 301, 308, 404, 405, 410, 421, 451, 501; TTL ≤ 1800) |
| `serveWhileStale` | 0 | Seconds to serve stale content while revalidating in the background |
| `requestCoalescing` | on (API) | Collapse concurrent cache-miss fetches of one object into one origin request |
| `signedUrlCacheMaxAgeSec` | 0 | Freshness window for responses to signed requests |
| `cacheKeyPolicy` | URL only | Which query params (`queryStringWhitelist`) and headers (`includeHttpHeaders`) join the cache key |
| `bypassCacheOnRequestHeaders` | `[]` | Skip the cache for requests carrying these headers (max 5) |

## Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `self_link` | `string` | Self-link URI — the value URL maps reference as a path-rule target |
| `backend_bucket_name` | `string` | Name of the backend bucket in GCP |
| `bucket_name` | `string` | The origin GCS bucket currently being served |

## Deployment Methods

Planton supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Immutability**: `backendBucketName`, `projectId`, and `loadBalancingScheme` are ForceNew. Everything else — including the origin `bucketName` — updates in place.
- **Public objects**: the load balancer does not authenticate to GCS. Objects must be publicly readable (`allUsers` → `roles/storage.objectViewer` on the bucket) unless served exclusively through signed URLs/cookies.
- **Signed-URL keys are secrets**: anyone holding a `keyValue` can mint valid signed URLs. Keys are immutable in GCP — rotate by adding a new key, re-signing, then removing the old one (at most 3 keys exist so rotation always has room). Key values never appear in outputs.
- **Cache invalidation is operational**, not declarative: changing TTLs affects future fills, not existing cache entries. Use `gcloud compute url-maps invalidate-cdn-cache` for immediate eviction, or version asset paths (fingerprinting) to sidestep invalidation entirely.
- **Edge vs backend security policies**: only `CLOUD_ARMOR_EDGE`-type policies attach here; standard Cloud Armor backend policies belong on backend services.

## Related Kinds

- [GcpGcsBucket](/docs/catalog/gcp/gcpgcsbucket) — the origin bucket whose objects are served
- [GcpCloudArmorPolicy](/docs/catalog/gcp/gcpcloudarmorpolicy) — the edge security policy filtering requests
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project that owns the backend bucket

## Additional Resources

- [Backend buckets overview](https://cloud.google.com/load-balancing/docs/backend-bucket)
- [Cloud CDN caching overview](https://cloud.google.com/cdn/docs/caching)
- [Signed URLs and signed cookies](https://cloud.google.com/cdn/docs/private-content)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
