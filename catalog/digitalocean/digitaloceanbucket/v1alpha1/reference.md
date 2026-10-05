# DigitalOceanBucket

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

DigitalOceanBucketSpec models the full surface of the
digitalocean_spaces_bucket resource plus the per-bucket settings
satellites whose lifecycle is identical to the bucket's -- CORS
configuration, bucket policy, and access logging -- which the
provisioners manage as part of this kind.

## Example

```yaml
# Example DigitalOceanBucket manifests.
#
# Deploy with: planton apply -f manifest.yaml
#
# The first document is the smallest real bucket (name only; the provider
# applies its own default region, nyc3). The second exercises the full
# surface: explicit region, public-read access, versioning, lifecycle
# rules, CORS for a browser application, a bucket policy, access logging
# to a second bucket, and force_destroy.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanBucket
metadata:
  name: example-dobkt-minimal
spec:
  bucketName: example-dobkt-minimal
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanBucket
metadata:
  name: example-dobkt-full
spec:
  bucketName: example-dobkt-full
  region: fra1
  accessControl: PUBLIC_READ
  versioningEnabled: true
  forceDestroy: true
  lifecycleRules:
    - id: expire-uploads
      prefix: uploads/
      enabled: true
      abortIncompleteMultipartUploadDays: 7
      expiration:
        days: 30
      noncurrentVersionExpiration:
        days: 90
  corsRules:
    - allowedMethods:
        - GET
        - HEAD
      allowedOrigins:
        - "https://www.example.com"
      allowedHeaders:
        - "*"
      exposeHeaders:
        - ETag
      id: site-assets
      maxAgeSeconds: 3600
  policy: |
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Sid": "PublicReadAssets",
          "Effect": "Allow",
          "Principal": "*",
          "Action": "s3:GetObject",
          "Resource": "arn:aws:s3:::example-dobkt-full/assets/*"
        }
      ]
    }
  logging:
    targetBucket:
      value: my-log-sink-bucket
    targetPrefix: access-logs/
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.bucketName` | `string` | yes |  |  |
| `spec.region` | `enum` |  |  |  |
| `spec.accessControl` | `enum` |  |  |  |
| `spec.versioningEnabled` | `bool` |  |  |  |
| `spec.forceDestroy` | `bool` |  |  |  |
| `spec.lifecycleRules` | `[]DigitalOceanBucketLifecycleRule` |  |  |  |
| `spec.lifecycleRules[].id` | `string` |  |  |  |
| `spec.lifecycleRules[].prefix` | `string` |  |  |  |
| `spec.lifecycleRules[].enabled` | `bool` | yes |  |  |
| `spec.lifecycleRules[].abortIncompleteMultipartUploadDays` | `uint32` |  |  |  |
| `spec.lifecycleRules[].expiration` | `DigitalOceanBucketLifecycleExpiration` |  |  |  |
| `spec.lifecycleRules[].expiration.date` | `string` |  |  |  |
| `spec.lifecycleRules[].expiration.days` | `uint32` |  |  |  |
| `spec.lifecycleRules[].expiration.expiredObjectDeleteMarker` | `bool` |  |  |  |
| `spec.lifecycleRules[].noncurrentVersionExpiration` | `DigitalOceanBucketLifecycleNoncurrentVersionExpiration` |  |  |  |
| `spec.lifecycleRules[].noncurrentVersionExpiration.days` | `uint32` | yes |  |  |
| `spec.corsRules` | `[]DigitalOceanBucketCorsRule` |  |  |  |
| `spec.corsRules[].allowedMethods` | `[]string` | yes |  |  |
| `spec.corsRules[].allowedOrigins` | `[]string` | yes |  |  |
| `spec.corsRules[].allowedHeaders` | `[]string` |  |  |  |
| `spec.corsRules[].exposeHeaders` | `[]string` |  |  |  |
| `spec.corsRules[].id` | `string` |  |  |  |
| `spec.corsRules[].maxAgeSeconds` | `uint32` |  |  |  |
| `spec.policy` | `string` |  |  |  |
| `spec.logging` | `DigitalOceanBucketLogging` |  |  |  |
| `spec.logging.targetBucket` | `string \| valueFrom` | yes |  | DigitalOceanBucket (`status.outputs.bucket_id`) |
| `spec.logging.targetPrefix` | `string` | yes |  |  |

## Field Details

### spec.bucketName

`string` · required

Bucket name. Must be DNS-compatible (lowercase letters, digits,
hyphens; 3-63 characters) and unique within the region.

- rule: {"required":true,"string":{"minLen":"3","maxLen":"63","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.region

`enum`

(Optional) Region slug for the bucket. When unset, the provider applies
its own default (nyc3). Spaces is available in a subset of DigitalOcean
regions; the CEL below lists the Spaces-capable slugs by their enum
numbers: nyc3=1, sfo3=2, fra1=3, sgp1=4, lon1=5, tor1=6, blr1=7, ams3=8,
sfo2=11, syd1=12, atl1=13 (nyc1/nyc2 have no Spaces). Changing the
region replaces the bucket.

- rule: region must be a Spaces-capable slug: ams3, atl1, blr1, fra1, lon1, nyc3, sfo2, sfo3, sgp1, syd1, tor1
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

Allowed values (use exactly as shown):

- `digital_ocean_region_unspecified` -- 0: default / unspecified region
- `nyc3` -- new york 3
- `sfo3` -- san francisco 3
- `fra1` -- frankfurt 1
- `sgp1` -- singapore 1
- `lon1` -- london 1
- `tor1` -- toronto 1
- `blr1` -- bangalore 1
- `ams3` -- amsterdam 3
- `nyc1` -- new york 1
- `nyc2` -- new york 2
- `sfo2` -- san francisco 2
- `syd1` -- sydney 1
- `atl1` -- atlanta 1

### spec.accessControl

`enum`

Access control (canned ACL) for the bucket itself: private (default) or
public-read. Per-object ACLs and finer grants are the policy's job.

Allowed values (use exactly as shown):

- `PRIVATE`
- `PUBLIC_READ`

### spec.versioningEnabled

`bool`

Enable object versioning. Once enabled, Spaces versioning can never be
removed -- flipping this back to false suspends it, keeping existing
versions.

### spec.forceDestroy

`bool`

(Optional) When true, the provisioner empties the bucket -- deleting
every object AND every object version -- before destroying it. DANGER:
this makes destroy irreversible for the bucket's data; leave false for
buckets holding anything you cannot lose.

### spec.lifecycleRules

`[]DigitalOceanBucketLifecycleRule`

(Optional) Object lifecycle rules: expire current or noncurrent object
versions and abort stale multipart uploads automatically to control
storage cost.

### spec.lifecycleRules[].id

`string`

(Optional) Unique rule identifier, at most 255 characters. When
omitted, the provider generates one.

- rule: {"string":{"maxLen":"255"}}

### spec.lifecycleRules[].prefix

`string`

(Optional) Object key prefix limiting the rule's scope (e.g. "logs/").
Empty applies the rule to every object. The provider warns when the
prefix starts with "/" -- Spaces keys do not begin with a slash.

### spec.lifecycleRules[].enabled

`bool` · required · optional (explicit presence)

Whether the rule is active. Required so a rule is never silently
inert: enabled: false keeps a rule staged but off, and must be said
explicitly.

- rule: {"required":true}

### spec.lifecycleRules[].abortIncompleteMultipartUploadDays

`uint32`

(Optional) Days after initiation when incomplete multipart uploads are
aborted and their parts removed.

### spec.lifecycleRules[].expiration

`DigitalOceanBucketLifecycleExpiration`

(Optional) Expiration of current object versions.

- rule: set exactly one of date, days, or expired_object_delete_marker

### spec.lifecycleRules[].expiration.date

`string`

(Optional) Absolute expiration date in YYYY-MM-DD form (interpreted as
midnight UTC).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^\\d{4}-\\d{2}-\\d{2}$"}}

### spec.lifecycleRules[].expiration.days

`uint32`

(Optional) Days after object creation when the object expires.

### spec.lifecycleRules[].expiration.expiredObjectDeleteMarker

`bool`

(Optional) On versioned buckets: remove expired object delete markers
that have no noncurrent versions left.

### spec.lifecycleRules[].noncurrentVersionExpiration

`DigitalOceanBucketLifecycleNoncurrentVersionExpiration`

(Optional) Expiration of noncurrent object versions (only meaningful on
versioned buckets).

### spec.lifecycleRules[].noncurrentVersionExpiration.days

`uint32` · required

Days after an object version becomes noncurrent when it is deleted.

- rule: {"required":true,"uint32":{"gte":1}}

### spec.corsRules

`[]DigitalOceanBucketCorsRule`

(Optional) CORS rules letting browser applications on other origins
read the bucket. Managed through the provider's standalone CORS
configuration resource (the bucket's inline cors_rule argument is
deprecated at the pinned provider and does no drift detection).

- rule: {"repeated":{"maxItems":"100"}}

### spec.corsRules[].allowedMethods

`[]string` · required

HTTP methods the rule allows (e.g. GET, HEAD, PUT).

- rule: {"repeated":{"minItems":"1"}}

### spec.corsRules[].allowedOrigins

`[]string` · required

Origins allowed to make cross-origin requests (e.g.
"https://example.com"; "*" allows every origin).

- rule: {"repeated":{"minItems":"1"}}

### spec.corsRules[].allowedHeaders

`[]string`

(Optional) Request headers the rule allows in preflight.

### spec.corsRules[].exposeHeaders

`[]string`

(Optional) Response headers browsers are allowed to read.

### spec.corsRules[].id

`string`

(Optional) Rule identifier, at most 255 characters.

- rule: {"string":{"maxLen":"255"}}

### spec.corsRules[].maxAgeSeconds

`uint32`

(Optional) Seconds browsers may cache the preflight response.

### spec.policy

`string`

(Optional) Bucket policy as a JSON document (S3 policy grammar, e.g.
IP-restricted or key-scoped access). The provider validates the JSON at
plan time and normalizes whitespace/key order when reading it back.

### spec.logging

`DigitalOceanBucketLogging`

(Optional) Access logging: deliver S3-style access logs for this bucket
to another bucket.

### spec.logging.targetBucket

`string | valueFrom` · required

The bucket that receives the access logs. Accepts a bucket name
directly or a reference to another DigitalOceanBucket resource
(resolved from its bucket_id output -- a Spaces bucket's id IS its
name). Logging a bucket to itself works but compounds: reads of the
logs generate more logs.

- references: DigitalOceanBucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.logging.targetPrefix

`string` · required

Prefix for the log object keys (e.g. "logs/").

- rule: {"required":true}

## Validation Rules

- `satellites_require_region`: cors_rules, policy, and logging require region to be set explicitly

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanBucket, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.bucket_id` | `string` | The provider's resource id for the bucket, which IS the bucket name -- Spaces buckets have no separate UUID. Import addressing pairs it with the region ("<region>,<name>"). |
| `status.outputs.endpoint` | `string` | The region-level Spaces endpoint host ("<region>.digitaloceanspaces.com", without the bucket name or a scheme). Pair it with the bucket name for path-style access. |
| `status.outputs.region` | `string` | The region slug the bucket lives in, read back from the API -- set even when spec.region was omitted and the provider default applied. |
| `status.outputs.bucket_domain_name` | `string` | The bucket's virtual-host-style FQDN ("<bucket>.<region>.digitaloceanspaces.com"). |
| `status.outputs.urn` | `string` | The uniform resource name of the bucket ("do:space:<name>"). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.logging.targetBucket` | DigitalOceanBucket | `status.outputs.bucket_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| DigitalOceanBucket | `spec.logging.targetBucket` | `status.outputs.bucket_id` |
| DigitalOceanCdn | `spec.origin` | `status.outputs.bucket_domain_name` |
| DigitalOceanProject | `spec.resources` | `status.outputs.urn` |
| DigitalOceanSpacesKey | `spec.grants[].bucket` | `status.outputs.bucket_id` |

## See Also

- [Overview](../README.md)
