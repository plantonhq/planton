---
title: "S3objectset"
description: "S3objectset deployment documentation"
icon: "package"
order: 100
componentName: "awss3objectset"
---

# AWS S3 Object Set - Research Documentation

## Introduction

Amazon S3 (Simple Storage Service) is the foundational object storage service in AWS, used for everything from static website hosting to data lake storage. While S3 *buckets* are the containers, S3 *objects* are the actual data units stored within them. Managing objects programmatically is a critical need for infrastructure automation -- configuration files, static assets, seed data, TLS certificates, and application resources all need to be uploaded to S3 as part of deployment workflows.

The `AwsS3ObjectSet` component addresses the gap between "bucket exists" and "bucket has the right content." It provides a declarative way to manage a collection of S3 objects within a single bucket, tightly integrated with OpenMCF's resource graph via foreign key references to `AwsS3Bucket`.

The core problem this component solves: infrastructure tools create buckets easily, but populating them with initial content requires separate scripts, manual uploads, or ad-hoc automation. By bringing object management into the same declarative model as bucket creation, `AwsS3ObjectSet` ensures that content and infrastructure are deployed atomically.

## Evolution and Historical Context

### The Manual Era

In the early days of S3, objects were uploaded via the AWS Console or simple HTTP PUT requests. Teams manually uploaded files after bucket creation, leading to "empty bucket" problems where infrastructure was deployed but not functional until someone remembered to upload the config files.

### The Scripting Era

As DevOps practices matured, teams wrote shell scripts using the AWS CLI (`aws s3 cp`, `aws s3 sync`) to upload objects after infrastructure provisioning. This worked but created a two-phase deployment: first create infrastructure, then run scripts to populate it. These scripts were fragile, poorly integrated with infrastructure state, and difficult to track for drift detection.

### The IaC Integration Era

Terraform introduced `aws_s3_object` (formerly `aws_s3_bucket_object`) and Pulumi provided `s3.BucketObject`, allowing objects to be managed as infrastructure resources. This was a significant improvement -- objects could be declared alongside buckets, tracked in state, and managed through the same lifecycle. However, managing many objects still requires significant boilerplate, and coordinating bucket references between modules remains tedious.

### The OpenMCF Approach

OpenMCF's `AwsS3ObjectSet` takes the IaC approach further by providing:
- A single resource that manages multiple objects (reducing boilerplate)
- Foreign key integration for bucket references (reducing coordination complexity)
- Unified tagging with inheritance (reducing repetition)
- Content flexibility with inline text and base64 binary support

## Deployment Methods Landscape

### Level 0: Manual (AWS Console)

**Workflow:**
1. Navigate to S3 in the AWS Console
2. Select the target bucket
3. Click "Upload"
4. Drag and drop files or browse to select them
5. Configure metadata (content type, cache control, etc.)
6. Set permissions
7. Click "Upload"

**Pros:**
- Visual interface, easy for one-off uploads
- Supports drag-and-drop for convenience
- No tooling required

**Cons:**
- Not repeatable or auditable
- Tedious for multiple files with different settings
- No version control for what was uploaded
- No drift detection
- Cannot be automated in CI/CD pipelines

**Verdict:** Suitable only for ad-hoc debugging or one-time uploads. Not appropriate for production infrastructure.

### Level 1: AWS CLI

**Example commands:**
```bash
# Upload a single file
aws s3 cp config/app.json s3://my-bucket/config/app.json \
  --content-type application/json \
  --cache-control max-age=3600

# Upload with inline content
echo '{"key": "value"}' | aws s3 cp - s3://my-bucket/config/app.json \
  --content-type application/json

# Upload multiple files
aws s3 sync ./assets/ s3://my-bucket/assets/ \
  --exclude "*.tmp"

# Upload with tags
aws s3api put-object \
  --bucket my-bucket \
  --key config/app.json \
  --body config/app.json \
  --content-type application/json \
  --tagging "environment=production&team=platform"
```

**Pros:**
- Scriptable and automatable
- Supports all S3 object features
- Can be integrated into CI/CD pipelines
- `s3 sync` handles multiple files efficiently

**Cons:**
- No state tracking (cannot detect drift)
- Scripts become complex for different content types and settings
- No dependency management with bucket creation
- Error handling requires custom logic
- Credentials must be managed separately

**Verdict:** Good for simple automation and CI/CD integration. Falls short for infrastructure-as-code workflows where state tracking and dependency management matter.

### Level 2: Terraform

**Example configuration:**
```hcl
resource "aws_s3_object" "app_config" {
  bucket       = aws_s3_bucket.main.id
  key          = "config/app.json"
  content      = jsonencode({
    database = "postgres"
    port     = 5432
  })
  content_type = "application/json"
  
  tags = {
    environment = "production"
  }
}

resource "aws_s3_object" "index_html" {
  bucket       = aws_s3_bucket.main.id
  key          = "index.html"
  content      = file("${path.module}/files/index.html")
  content_type = "text/html"
  cache_control = "max-age=300"
  
  tags = {
    environment = "production"
  }
}

# For binary files
resource "aws_s3_object" "favicon" {
  bucket         = aws_s3_bucket.main.id
  key            = "favicon.ico"
  content_base64 = filebase64("${path.module}/files/favicon.ico")
  content_type   = "image/x-icon"
}
```

**Pros:**
- Full state tracking and drift detection
- Dependency management (waits for bucket creation)
- Plan/apply workflow for review
- Supports all S3 object features
- Version controlled

**Cons:**
- Each object is a separate resource block (verbose for many objects)
- Referencing bucket across modules requires output passing
- Large binary files bloat state
- Content changes trigger full resource replacement
- No native "upload multiple objects" primitive

**Verdict:** Production-grade approach for managing S3 objects as infrastructure. The verbosity for multiple objects is the main drawback.

### Level 3: Pulumi

**Example (Go):**
```go
bucket, _ := s3.NewBucket(ctx, "main", &s3.BucketArgs{})

s3.NewBucketObject(ctx, "app-config", &s3.BucketObjectArgs{
    Bucket:      bucket.ID(),
    Key:         pulumi.String("config/app.json"),
    Content:     pulumi.String(`{"database": "postgres", "port": 5432}`),
    ContentType: pulumi.String("application/json"),
    Tags: pulumi.StringMap{
        "environment": pulumi.String("production"),
    },
})

s3.NewBucketObject(ctx, "index-html", &s3.BucketObjectArgs{
    Bucket:       bucket.ID(),
    Key:          pulumi.String("index.html"),
    Content:      pulumi.String("<html>...</html>"),
    ContentType:  pulumi.String("text/html"),
    CacheControl: pulumi.String("max-age=300"),
})
```

**Pros:**
- Full state tracking
- Type safety with native language constructs
- Loops and conditionals for managing multiple objects
- First-class dependency resolution
- Testable with unit tests

**Cons:**
- Requires programming language knowledge
- Each object still needs explicit creation
- Cross-stack references need explicit exports/imports

**Verdict:** Excellent for teams comfortable with general-purpose programming languages. Loop constructs naturally handle multiple objects.

### Other Methods

**Ansible:**
- `amazon.aws.s3_object` module supports upload/download
- Good for configuration management but less suitable for infrastructure-as-code
- No state tracking; relies on idempotency

**Crossplane:**
- `Object` resource in the AWS provider
- Good for Kubernetes-native workflows
- Each object is a separate CR; verbose for many objects

## Comparative Analysis

- **Manual Console**: No automation, no state, no versioning. Development/debugging only.
- **AWS CLI**: Scriptable but no state tracking. Good for CI/CD file sync.
- **Terraform**: Full state, drift detection, verbose per-object. Production standard.
- **Pulumi**: Full state, type-safe, loops for batching. Production standard.
- **OpenMCF**: Full state via Terraform/Pulumi, foreign key references, batch objects, tag inheritance. Minimal configuration for common patterns.

## The OpenMCF Approach

### Design Philosophy

`AwsS3ObjectSet` applies the 80/20 principle: expose the 20% of S3 object configuration that covers 80% of use cases, while keeping the API surface clean and approachable.

### Why "ObjectSet" Instead of "Object"

A single `AwsS3Object` component per object would create excessive resource proliferation. Most real-world use cases involve uploading a *group* of related objects to the same bucket (e.g., a set of config files, a collection of static assets). The "set" model:
- Reduces the number of OpenMCF resources to manage
- Shares common configuration (bucket, region, tags) across all objects
- Maps naturally to how teams think about "deploying content to a bucket"

### Foreign Key Integration

The `bucket` field uses `StringValueOrRef` with `default_kind = AwsS3Bucket`:

```yaml
# Literal bucket name (for pre-existing buckets)
bucket:
  value: my-existing-bucket

# Reference to an AwsS3Bucket component (resolved automatically)
bucket:
  valueFrom:
    name: my-s3-bucket
```

This pattern mirrors how `KubernetesDeployment` references `KubernetesNamespace`, providing seamless cross-component wiring.

### Fields Included (and Why)

| Field | Rationale |
|-------|-----------|
| `bucket` | Required target. Foreign key enables component wiring. |
| `aws_region` | Required for provider configuration. |
| `objects[].key` | Required S3 object path. |
| `objects[].content` | Inline text for config files, HTML, JSON, YAML. |
| `objects[].content_base64` | Binary support for images, compiled assets. |
| `objects[].content_type` | MIME type affects browser handling and CDN behavior. |
| `objects[].cache_control` | Critical for static websites and CDN integration. |
| `objects[].content_encoding` | Pre-compressed content support (gzip/brotli). |
| `objects[].acl` | Per-object access control for mixed-access patterns. |
| `objects[].tags` | Per-object governance with set-level inheritance. |
| `tags` (set-level) | Common tags applied to all objects. |

### Fields Excluded (and Why)

| Excluded Feature | Rationale |
|-----------------|-----------|
| `source` (file path) | OpenMCF runs remotely; local file paths don't translate. Inline content and base64 cover all cases. |
| `server_side_encryption` | Bucket-level encryption (configured in AwsS3Bucket) applies to all objects automatically. |
| `storage_class` | Per-object storage class overrides are rare; bucket-level lifecycle rules handle transitions. |
| `website_redirect` | Niche feature; can be added later if needed. |
| `object_lock` | Governance feature typically set at bucket level, not per-object. |
| `metadata` (custom) | Rarely used; tags serve the same governance purpose. |

## Implementation Landscape

### Pulumi Module Architecture

The Pulumi module iterates over the `objects` list and creates one `s3.BucketObject` per entry:

```
main.go          - Entry point, loads StackInput
module/
  locals.go      - Extracts spec fields, merges tags
  main.go        - Creates BucketObject resources in a loop
  outputs.go     - Collects ETags and version IDs into maps
```

Key implementation details:
- Bucket name is resolved from the foreign key before reaching the module
- Tags are merged: set-level tags serve as defaults, object-level tags override
- ETag and version ID are collected per-object for the outputs map

### Terraform Module Architecture

The Terraform module uses `for_each` over the objects list:

```
variables.tf     - Input variables (metadata, spec)
locals.tf        - Object map creation, tag merging
main.tf          - aws_s3_object resource with for_each
outputs.tf       - ETag and version ID maps
provider.tf      - AWS provider configuration
```

Key implementation details:
- Objects are keyed by their `key` field for stable `for_each` iteration
- `content` and `content_base64` are mutually exclusive via conditional expressions
- Tag merging uses `merge()` function with object tags taking precedence

## Production Best Practices

### Content Management

- **Use `content` for text files**: Configuration files, HTML, CSS, JSON, YAML. This keeps content readable and diff-able.
- **Use `content_base64` for binary files**: Images, compiled assets, certificates. Keep binary content small (< 1MB) to avoid bloating IaC state.
- **Large files**: For files > 1MB, consider using CI/CD pipelines with `aws s3 cp` instead of managing them as IaC resources.

### Tagging Strategy

- Set common tags at the set level (`tags` on `AwsS3ObjectSetSpec`)
- Override per-object only when needed (e.g., different `visibility` tags)
- Include standard organizational tags: `environment`, `team`, `project`, `cost-center`

### Cache Control

- **Static assets** (CSS, JS, images): `max-age=31536000` (1 year) with content-hash in filename
- **HTML pages**: `max-age=300` (5 minutes) or `no-cache` for always-fresh content
- **Config files**: `no-cache` or omit (private, not cached)

### Security

- Default to `private` ACL (inherits bucket default)
- Use `public-read` only for intentionally public content
- Rely on bucket-level encryption (configured in `AwsS3Bucket`) for encryption at rest
- Never store secrets as S3 objects; use AWS Secrets Manager instead

### Cost Optimization

- S3 PUT requests are charged per-request; batch related objects in a single `AwsS3ObjectSet`
- Use appropriate content types to enable compression at CDN level
- Consider lifecycle rules on the bucket for objects that become stale

## Conclusion

`AwsS3ObjectSet` bridges the gap between bucket provisioning and content management in OpenMCF. It provides a clean, declarative interface for managing multiple S3 objects with shared configuration, foreign key integration, and tag inheritance. The component handles the most common use cases -- config files, static assets, and seed data -- while keeping the API surface minimal and approachable.

### When to Use This Component

- Deploying configuration files alongside infrastructure
- Setting up initial content for static websites
- Managing seed data for applications
- Uploading TLS certificates or other small artifacts
- Any scenario where S3 objects should be managed declaratively

### When NOT to Use This Component

- Large file uploads (> 1MB) -- use CI/CD pipelines instead
- Frequently changing content -- use application-level upload logic
- Thousands of objects -- use `aws s3 sync` in CI/CD pipelines
- Object content generated at runtime -- use application code

### References

- [AWS S3 Documentation](https://docs.aws.amazon.com/s3/)
- [Terraform aws_s3_object Resource](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_object)
- [Pulumi AWS S3 BucketObject](https://www.pulumi.com/registry/packages/aws/api-docs/s3/bucketobject/)
- [S3 Best Practices](https://docs.aws.amazon.com/AmazonS3/latest/userguide/optimizing-performance.html)
