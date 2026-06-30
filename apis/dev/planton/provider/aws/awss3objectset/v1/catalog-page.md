# AWS S3 Object Set

Deploys one or more objects into an existing AWS S3 bucket, supporting inline text content and base64-encoded binary content. The component manages objects declaratively alongside infrastructure, making it suitable for configuration files, static assets, and seed data.

## What Gets Created

When you deploy an AwsS3ObjectSet resource, Planton provisions:

- **S3 Object (one per entry)** — an `aws_s3_bucket_object_v2` resource for each item in the `objects` list, uploaded to the target bucket with the specified key, content, content type, caching headers, and tags
- **Merged Tags** — each object receives tags merged from three sources in increasing precedence: resource labels, set-level `tags`, and per-object `tags`

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **An existing S3 bucket** — either a literal bucket name or a deployed AwsS3Bucket resource to reference via `valueFrom`
- **The bucket's AWS region** — must match the region specified in `region`

## Quick Start

Create a file `s3-objects.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsS3ObjectSet
metadata:
  name: my-objects
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsS3ObjectSet.my-objects
spec:
  region: us-east-1
  bucket: my-app-bucket
  objects:
    - key: config/app.json
      content: '{"env": "dev", "debug": true}'
```

Deploy:

```shell
planton apply -f s3-objects.yaml
```

This uploads a single JSON configuration file to the `config/app.json` key in the target bucket.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `bucket` | `StringValueOrRef` | The target S3 bucket where objects will be uploaded. Can be a literal bucket name or a reference to an AwsS3Bucket resource. | Required. Can reference `AwsS3Bucket` resource via `valueFrom` (resolves `status.outputs.bucket_id`). |
| `region` | `string` | The AWS region where the S3 bucket is located (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |
| `objects` | `AwsS3Object[]` | The list of S3 objects to upload to the target bucket. | Minimum 1 item |
| `objects[].key` | `string` | The S3 object key (path within the bucket). | Minimum length 1 |
| `objects[].content` | `string` | Inline UTF-8 text content for the object. Exactly one of `content` or `contentBase64` must be set. | — |
| `objects[].contentBase64` | `string` | Base64-encoded binary content for the object. Exactly one of `content` or `contentBase64` must be set. | — |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `tags` | `map<string, string>` | `{}` | Tags applied to all objects in the set. Individual object tags are merged with these, with object-level tags taking precedence. |
| `objects[].contentType` | `string` | `application/octet-stream` | The MIME content type of the object (e.g., `application/json`, `text/html`, `image/png`). |
| `objects[].cacheControl` | `string` | — | The caching behavior for the object (e.g., `max-age=86400` for 24-hour caching, `no-cache`). |
| `objects[].contentEncoding` | `string` | — | How the content is encoded (e.g., `gzip`, `br`). Set this if the content has been pre-compressed. |
| `objects[].tags` | `map<string, string>` | `{}` | Tags specific to this object. Merged with set-level tags (object tags take precedence). |
| `objects[].acl` | `string` | — | The canned ACL for this object (e.g., `private`, `public-read`). If not specified, inherits the bucket's default ACL. |

## Examples

### Multiple Configuration Files

Upload several configuration files to a shared bucket:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsS3ObjectSet
metadata:
  name: app-config
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsS3ObjectSet.app-config
spec:
  region: us-east-1
  bucket: my-app-bucket
  objects:
    - key: config/app.json
      content: '{"env": "dev", "logLevel": "debug"}'
      contentType: application/json
    - key: config/feature-flags.json
      content: '{"darkMode": true, "betaSignup": false}'
      contentType: application/json
    - key: config/robots.txt
      content: |
        User-agent: *
        Disallow: /admin/
      contentType: text/plain
```

### Static Website Assets with Caching

Upload pre-compressed static assets with cache headers and public read access:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsS3ObjectSet
metadata:
  name: website-assets
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsS3ObjectSet.website-assets
spec:
  region: us-west-2
  bucket: my-website-bucket
  tags:
    project: website
    managed-by: planton
  objects:
    - key: index.html
      content: |
        <!DOCTYPE html>
        <html><head><title>My Site</title></head>
        <body><h1>Hello</h1></body></html>
      contentType: text/html
      cacheControl: no-cache
      acl: public-read
    - key: assets/style.css
      content: "body { font-family: sans-serif; margin: 0; }"
      contentType: text/css
      cacheControl: max-age=31536000
      acl: public-read
```

### Binary Content with Base64 Encoding

Upload binary assets using base64-encoded content:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsS3ObjectSet
metadata:
  name: binary-assets
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.AwsS3ObjectSet.binary-assets
spec:
  region: eu-west-1
  bucket: my-assets-bucket
  objects:
    - key: images/favicon.ico
      contentBase64: AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAA...
      contentType: image/x-icon
      cacheControl: max-age=86400
    - key: data/seed.csv
      content: |
        id,name,email
        1,Alice,alice@example.com
        2,Bob,bob@example.com
      contentType: text/csv
```

### Using Foreign Key References

Reference an Planton-managed AwsS3Bucket instead of hardcoding the bucket name:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsS3ObjectSet
metadata:
  name: ref-objects
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsS3ObjectSet.ref-objects
spec:
  region: us-east-1
  bucket:
    valueFrom:
      kind: AwsS3Bucket
      name: my-bucket
      field: status.outputs.bucket_id
  objects:
    - key: deploy/manifest.json
      content: '{"version": "1.2.0", "timestamp": "2025-01-15T00:00:00Z"}'
      contentType: application/json
```

### Per-Object Tags and ACLs

Apply different tags and access controls per object within a single set:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsS3ObjectSet
metadata:
  name: mixed-access
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsS3ObjectSet.mixed-access
spec:
  region: us-east-1
  bucket: shared-bucket
  tags:
    team: platform
  objects:
    - key: public/index.html
      content: "<html><body>Public page</body></html>"
      contentType: text/html
      acl: public-read
      tags:
        visibility: public
    - key: internal/config.yaml
      content: |
        database:
          host: db.internal
          port: 5432
      contentType: application/x-yaml
      acl: private
      tags:
        visibility: internal
        sensitivity: high
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `object_etags` | `map<string, string>` | Map of object key to its ETag (content hash). The ETag changes when the object content changes, useful for cache invalidation. |
| `object_version_ids` | `map<string, string>` | Map of object key to its version ID. Only populated when the target bucket has versioning enabled. |

## Related Components

- [AwsS3Bucket](/docs/catalog/aws/awss3bucket) — provides the target bucket; can be referenced via `valueFrom` in the `bucket` field
- [AwsCloudFront](/docs/catalog/aws/awscloudfront) — serves objects from S3 via a CDN distribution
- [AwsLambda](/docs/catalog/aws/awslambda) — can be triggered by S3 object events in the target bucket
