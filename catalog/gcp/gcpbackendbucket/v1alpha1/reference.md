# GcpBackendBucket

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpBackendBucketSpec defines a Compute Engine backend bucket — the piece
that serves a Cloud Storage bucket's objects through an external HTTP(S)
load balancer, optionally cached at Google's edge by Cloud CDN. It is the
static-content counterpart of a backend service: URL maps route paths like
/assets/* to a backend bucket while dynamic paths go to backend services.

The backend bucket is deliberately a separate node from the bucket itself:
one GCS bucket can sit behind several backend buckets with different CDN
policies, and swapping the origin bucket is an in-place update that leaves
the URL map untouched.

CDN behavior is a policy ON this resource, not a separate GCP object —
enable_cdn turns caching on and cdn_policy tunes how responses are cached,
keyed, and invalidated.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBackendBucket
metadata:
  name: my-sample-backend-bucket
spec:
  # GCP project that owns the backend bucket.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Cloud-side name; omit to default to metadata.name.
  backendBucketName: static-assets

  # The GCS bucket whose objects are served (or reference a GcpGcsBucket).
  bucketName:
    value: my-static-assets-bucket

  description: Serves the app's fingerprinted static assets through Cloud CDN

  # Cache at Google's edge.
  enableCdn: true
  cdnPolicy:
    # Cache static content types; honor origin headers for the rest.
    cacheMode: CACHE_ALL_STATIC
    defaultTtl: 3600
    clientTtl: 1800
    # Cache 404s briefly so missing-asset storms don't hammer the origin.
    negativeCaching: true
    negativeCachingPolicy:
      - code: 404
        ttl: 60
    # Collapse concurrent cache-miss fetches of the same object.
    requestCoalescing: true

  # Compress compressible content types for clients that ask.
  compressionMode: AUTOMATIC

  # Surface the CDN verdict for debugging.
  customResponseHeaders:
    - "X-Cache-Status: {cdn_cache_status}"

  # Ephemeral test fixture: delete the backend bucket (and any signed-URL
  # keys) on destroy.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.backendBucketName` | `string` |  |  |  |
| `spec.bucketName` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.bucket_id`) |
| `spec.description` | `string` |  |  |  |
| `spec.enableCdn` | `bool` |  |  |  |
| `spec.cdnPolicy` | `GcpBackendBucketCdnPolicy` |  |  |  |
| `spec.cdnPolicy.cacheMode` | `string` |  |  |  |
| `spec.cdnPolicy.clientTtl` | `int32` |  |  |  |
| `spec.cdnPolicy.defaultTtl` | `int32` |  |  |  |
| `spec.cdnPolicy.maxTtl` | `int32` |  |  |  |
| `spec.cdnPolicy.negativeCaching` | `bool` |  |  |  |
| `spec.cdnPolicy.negativeCachingPolicy` | `[]GcpBackendBucketNegativeCachingPolicy` |  |  |  |
| `spec.cdnPolicy.negativeCachingPolicy[].code` | `int32` | yes |  |  |
| `spec.cdnPolicy.negativeCachingPolicy[].ttl` | `int32` |  |  |  |
| `spec.cdnPolicy.serveWhileStale` | `int32` |  |  |  |
| `spec.cdnPolicy.requestCoalescing` | `bool` |  |  |  |
| `spec.cdnPolicy.signedUrlCacheMaxAgeSec` | `int32` |  |  |  |
| `spec.cdnPolicy.cacheKeyPolicy` | `GcpBackendBucketCacheKeyPolicy` |  |  |  |
| `spec.cdnPolicy.cacheKeyPolicy.queryStringWhitelist` | `[]string` |  |  |  |
| `spec.cdnPolicy.cacheKeyPolicy.includeHttpHeaders` | `[]string` |  |  |  |
| `spec.cdnPolicy.bypassCacheOnRequestHeaders` | `[]GcpBackendBucketBypassCacheOnRequestHeader` |  |  |  |
| `spec.cdnPolicy.bypassCacheOnRequestHeaders[].headerName` | `string` | yes |  |  |
| `spec.compressionMode` | `string` |  |  |  |
| `spec.customResponseHeaders` | `[]string` |  |  |  |
| `spec.edgeSecurityPolicy` | `string \| valueFrom` |  |  | GcpCloudArmorPolicy (`status.outputs.policy_self_link`) |
| `spec.loadBalancingScheme` | `string` |  |  |  |
| `spec.signedUrlKeys` | `[]GcpBackendBucketSignedUrlKey` |  |  |  |
| `spec.signedUrlKeys[].name` | `string` | yes |  |  |
| `spec.signedUrlKeys[].keyValue` | `string` (sensitive) | yes |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the backend bucket (which may differ from the
project owning the GCS bucket — cross-project origins are valid).
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing it destroys and recreates the backend bucket.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.backendBucketName

`string`

Name of the backend bucket in GCP. Must be 1-63 characters: lowercase
letters, digits, and hyphens; must start with a letter and end with a
letter or digit. If not specified, defaults to metadata.name.
Immutable: changing it destroys and recreates the backend bucket,
briefly breaking every URL map that references the old self_link.

- rule: backend_bucket_name must be RFC1035-compliant: 1-63 lowercase letters, digits, or hyphens; must start with a letter and end with a letter or digit

### spec.bucketName

`string | valueFrom` · required

The Cloud Storage bucket whose objects are served — the origin.
Reference a GcpGcsBucket resource or provide the bucket name directly.
Mutable: pointing at a different bucket is an in-place update, which
makes origin swaps (e.g. blue/green static releases) cheap.
Objects must be publicly readable (or served via signed URLs/cookies) —
the load balancer does not authenticate to the bucket.

- references: GcpGcsBucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.description

`string`

What this backend bucket serves and which URL maps use it — write it for
the operator tracing a route later. Mutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.enableCdn

`bool`

Cache responses at Google's edge with Cloud CDN. Off by default: without
it every request is proxied to the bucket. Turning it on activates
cdn_policy (or sensible CDN defaults when cdn_policy is omitted).
Cannot be enabled together with load_balancing_scheme INTERNAL_MANAGED —
Cloud CDN only fronts external load balancers. Mutable.

### spec.cdnPolicy

`GcpBackendBucketCdnPolicy`

How Cloud CDN caches responses from this origin. Only meaningful with
enable_cdn — GCP ignores the policy while CDN is off.

- rule: with cache_mode USE_ORIGIN_HEADERS the origin's headers control lifetimes — remove client_ttl, default_ttl, and max_ttl (GCP would silently ignore them)
- rule: with cache_mode FORCE_CACHE_ALL every response is cached for default_ttl — remove max_ttl (GCP would silently ignore it)

### spec.cdnPolicy.cacheMode

`string`

What gets cached. CACHE_ALL_STATIC (the GCP default) caches static
content types and honors origin cache headers for the rest;
USE_ORIGIN_HEADERS caches only what the origin explicitly marks
cacheable (TTL fields must be unset — the origin controls lifetimes);
FORCE_CACHE_ALL caches everything, ignoring origin headers (never
combine with private or per-user content; max_ttl must be unset).

- rule: cache_mode must be CACHE_ALL_STATIC, USE_ORIGIN_HEADERS, or FORCE_CACHE_ALL

### spec.cdnPolicy.clientTtl

`int32`

Seconds a response may be cached by browsers and other downstream
caches (sets the max-age clients see; GCP default 3600, max 86400).
Keep it shorter than default_ttl so edge caches revalidate before
clients do.

- rule: client_ttl must be between 0 and 86400 seconds (1 day)

### spec.cdnPolicy.defaultTtl

`int32`

Seconds the edge caches a response when the origin sets no caching
headers (GCP default 3600, max 31622400 = 1 year). The workhorse TTL
for CACHE_ALL_STATIC and FORCE_CACHE_ALL.

- rule: default_ttl must be between 0 and 31622400 seconds (1 year)

### spec.cdnPolicy.maxTtl

`int32`

Upper bound in seconds on any cache lifetime, capping even origin
headers that ask for longer (GCP default 86400, max 31622400). Not
allowed with USE_ORIGIN_HEADERS or FORCE_CACHE_ALL cache modes.

- rule: max_ttl must be between 0 and 31622400 seconds (1 year)

### spec.cdnPolicy.negativeCaching

`bool`

Cache error responses (404s, redirects) at the edge so failing paths do
not hammer the origin. Pair with negative_caching_policy to set
per-status TTLs; without it GCP applies default lifetimes.

### spec.cdnPolicy.negativeCachingPolicy

`[]GcpBackendBucketNegativeCachingPolicy`

Per-status-code TTLs for negative caching. Only effective with
negative_caching enabled. Codes limited by GCP to 300, 301, 308, 404,
405, 410, 421, 451, and 501.

### spec.cdnPolicy.negativeCachingPolicy[].code

`int32` · required

The HTTP status code to cache. GCP supports 300, 301, 308, 404, 405,
410, 421, 451, and 501.

- rule: code must be one of 300, 301, 308, 404, 405, 410, 421, 451, or 501 — the status codes Cloud CDN can negative-cache
- rule: {"required":true}

### spec.cdnPolicy.negativeCachingPolicy[].ttl

`int32`

Seconds responses with this status are cached at the edge
(0 to 1800 = 30 minutes).

- rule: ttl must be between 0 and 1800 seconds (30 minutes)

### spec.cdnPolicy.serveWhileStale

`int32`

Seconds the edge may keep serving a stale response while it revalidates
with the origin in the background (max 86400; 0 disables). Smooths over
brief origin outages for content that tolerates slight staleness.

- rule: serve_while_stale must be between 0 and 86400 seconds (1 day)

### spec.cdnPolicy.requestCoalescing

`bool`

Collapse concurrent cache-miss requests for the same object into one
origin fetch. Protects the origin from thundering herds on cache
expiry of popular objects.

### spec.cdnPolicy.signedUrlCacheMaxAgeSec

`int32`

Seconds a response to a SIGNED request stays fresh in the cache before
revalidation (max 86400). Only meaningful with signed URLs or cookies;
after this window the edge revalidates, though the signature's own
expiry still governs access.

- rule: signed_url_cache_max_age_sec must be between 0 and 86400 seconds (1 day)

### spec.cdnPolicy.cacheKeyPolicy

`GcpBackendBucketCacheKeyPolicy`

What forms the cache key beyond the URL host and path. Leave unset to
ignore query strings and headers entirely — the best hit rate for
immutable, fingerprinted assets.

### spec.cdnPolicy.cacheKeyPolicy.queryStringWhitelist

`[]string`

Query parameters included in the cache key (all others are ignored).
Include only parameters that genuinely change the response — e.g. an
image resizer's "w" and "h" — so equivalent requests share a cache
entry.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.cdnPolicy.cacheKeyPolicy.includeHttpHeaders

`[]string`

Request headers whose values join the cache key — for origins that vary
responses by header (e.g. Accept for image format negotiation). Each
distinct value creates a separate cache entry, so keep this list short.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.cdnPolicy.bypassCacheOnRequestHeaders

`[]GcpBackendBucketBypassCacheOnRequestHeader`

Skip the cache entirely for requests carrying any of these headers
(at most 5) — an escape hatch for debugging or per-request freshness
(e.g. a Pragma: no-cache internal tooling header).

- rule: {"repeated":{"maxItems":"5"}}

### spec.cdnPolicy.bypassCacheOnRequestHeaders[].headerName

`string` · required

The header name to match (case-insensitive); any value triggers the
bypass.

- rule: {"required":true}

### spec.compressionMode

`string`

Whether the load balancer compresses responses (gzip/brotli) for clients
that ask for it. AUTOMATIC compresses compressible content types;
DISABLED (the GCP default when unset) never compresses. Compression is
applied by the load balancer, not the bucket — objects stay uncompressed
at the origin. Mutable.

- rule: compression_mode must be AUTOMATIC or DISABLED

### spec.customResponseHeaders

`[]string`

Response headers the load balancer adds to every response served from
this backend, in "Header-Name: value" form. Values may use variables
like {cdn_cache_status}. Typical uses: security headers
(Strict-Transport-Security) and cache observability
(X-Cache-Status: {cdn_cache_status}). Mutable.

- rule: {"repeated":{"maxItems":"25","items":{"string":{"pattern":"^[^:]+:.*$"}}}}

### spec.edgeSecurityPolicy

`string | valueFrom`

Cloud Armor EDGE security policy filtering requests before they reach
the cache or the origin (rate limiting and geo/IP blocking at the edge).
Reference a GcpCloudArmorPolicy of type CLOUD_ARMOR_EDGE — standard
backend policies are not valid here. Mutable.

- references: GcpCloudArmorPolicy (`status.outputs.policy_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudArmorPolicy, name: <that resource's name>, fieldPath: status.outputs.policy_self_link}} -- a bare string does not parse

### spec.loadBalancingScheme

`string`

Which load balancer family this backend serves. Leave unset for global
EXTERNAL HTTP(S) load balancers — the overwhelmingly common case for
static content. INTERNAL_MANAGED serves cross-region internal
Application Load Balancers instead, and is incompatible with Cloud CDN.
Immutable: changing it destroys and recreates the backend bucket.

- rule: load_balancing_scheme must be INTERNAL_MANAGED, or left unset for external load balancers

### spec.signedUrlKeys

`[]GcpBackendBucketSignedUrlKey`

Keys for signing Cloud CDN signed URLs and signed cookies — the
mechanism for serving private content from the cache with expiring,
tamper-proof links. GCP allows at most 3 keys per backend bucket so one
can be rotated while another stays live. Each key's material is a
secret; rotate by adding a new key, re-signing URLs, then removing the
old one.

- rule: {"repeated":{"maxItems":"3"}}

### spec.signedUrlKeys[].name

`string` · required

Name of the key, referenced by the key_name parameter of signed URLs.
Must be 1-63 characters: lowercase letters, digits, and hyphens; must
start with a letter and end with a letter or digit. Immutable: renaming
replaces the key, invalidating URLs signed with the old name.

- rule: {"required":true,"string":{"pattern":"^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$"}}

### spec.signedUrlKeys[].keyValue

`string` · required · sensitive

The 128-bit signing key, base64url-encoded (RFC 4648 §5) — generate one
with: head -c 16 /dev/urandom | base64 | tr '+/' '-_'. 22 characters of
base64url, with or without the trailing == padding. Anyone holding this
value can mint valid signed URLs, so it is handled as a secret.
Immutable per key name: rotating means adding a new key and removing
the old. The base64url shape is taught here rather than enforced by a
validation rule, because sensitive fields hold a managed-secret
reference on consuming platforms and a content-shape rule would
reject every reference.

- rule: {"required":true}

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the backend bucket for org-policy and
IAM conditions. Keys in the form "tagKeys/{id}", values
"tagValues/{id}". Create-time only: changing them later replaces the
backend bucket.

### spec.deletionPolicy

`string`

Deletion policy for the backend bucket AND its signed-URL keys — one
switch governs both objects this kind manages:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- both are deleted (GCP refuses to delete the backend
               bucket while a URL map still references it); the GCS
               bucket behind it is untouched — it belongs to its own
               kind
  "PREVENT" -- destroy FAILS; protects a CDN origin that URL maps may
               still route to
  "ABANDON" -- both are removed from management but keep serving in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `cdn_requires_external_scheme`: Cloud CDN cannot be enabled on an INTERNAL_MANAGED backend bucket — CDN only fronts external load balancers

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpBackendBucket, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.self_link` | `string` | Self-link URI of the backend bucket. This is the value URL maps reference as a default service or path-rule target — the composition handle for routing static content. Format: https://www.googleapis.com/compute/v1/projects/{project}/global/backendBuckets/{name} |
| `status.outputs.backend_bucket_name` | `string` | Name of the backend bucket as it exists in GCP. |
| `status.outputs.bucket_name` | `string` | Name of the Cloud Storage bucket currently serving as the origin, echoed for tooling that resolves the serving chain. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.bucketName` | GcpGcsBucket | `status.outputs.bucket_id` |
| `spec.edgeSecurityPolicy` | GcpCloudArmorPolicy | `status.outputs.policy_self_link` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpUrlMap | `spec.defaultService` | `status.outputs.self_link` |
| GcpUrlMap | `spec.defaultCustomErrorResponsePolicy.errorService` | `status.outputs.self_link` |
| GcpUrlMap | `spec.pathMatchers[].defaultService` | `status.outputs.self_link` |
| GcpUrlMap | `spec.pathMatchers[].defaultCustomErrorResponsePolicy.errorService` | `status.outputs.self_link` |
| GcpUrlMap | `spec.pathMatchers[].pathRules[].service` | `status.outputs.self_link` |
| GcpUrlMap | `spec.pathMatchers[].pathRules[].customErrorResponsePolicy.errorService` | `status.outputs.self_link` |
| GcpUrlMap | `spec.pathMatchers[].routeRules[].customErrorResponsePolicy.errorService` | `status.outputs.self_link` |
| GcpUrlMap | `spec.tests[].service` | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)
