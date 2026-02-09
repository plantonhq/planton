# Overview

The AWS S3 Object Set API resource provides a declarative interface for uploading and managing one or more objects in an Amazon S3 bucket. By abstracting the complexity of individual S3 object uploads, this resource allows you to manage configuration files, static assets, seed data, and other content alongside your infrastructure.

## Why We Created This API Resource

Managing S3 objects alongside infrastructure requires coordinating bucket creation with object uploads. This resource solves that by:

- **Declarative Object Management**: Define objects as code alongside your infrastructure, ensuring they are created, updated, or removed consistently.
- **Foreign Key Integration**: Reference an `AwsS3Bucket` component directly, so the bucket name is resolved automatically when both resources are managed together.
- **Batch Uploads**: Upload multiple objects to the same bucket in a single deployment, reducing configuration duplication.
- **Content Flexibility**: Support both inline text content (UTF-8) and base64-encoded binary content for any file type.

## Key Features

### Bucket Reference via Foreign Key

- **Literal Bucket Name**: Provide the bucket name directly as a string value.
- **Component Reference**: Reference an `AwsS3Bucket` component by name; the bucket ID is resolved from `status.outputs.bucket_id` automatically.

### Multi-Object Support

- Upload one or more objects per deployment.
- Each object has its own key, content, content type, caching settings, and tags.
- Set-level tags are merged with object-level tags (object tags take precedence).

### Content Sources

- **Inline Text** (`content`): For configuration files, JSON, YAML, HTML, and other text formats.
- **Base64 Binary** (`content_base64`): For images, compiled assets, or any binary data.

### Object Metadata

- **Content Type**: Set the MIME type for correct browser/client handling.
- **Cache Control**: Configure caching headers for CDN and browser caching.
- **Content Encoding**: Declare pre-compressed content (gzip, brotli).
- **ACL**: Set per-object access control (private, public-read, etc.).

## Stack Outputs

- **object_etags**: Map of object key to ETag (content hash) for cache invalidation.
- **object_version_ids**: Map of object key to version ID (when bucket versioning is enabled).

## Benefits

- **Infrastructure as Code**: Manage S3 objects declaratively alongside buckets and other resources.
- **Consistency**: Ensure objects are always in sync with infrastructure deployments.
- **Simplicity**: Single resource manages multiple objects with shared defaults.
- **Flexibility**: Supports text and binary content, per-object metadata, and tag inheritance.
