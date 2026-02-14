# Private Standard Bucket

This preset creates a private GCS bucket with uniform bucket-level access, versioning, public access prevention, and lifecycle rules to control version sprawl. It represents the standard configuration for application data, backups, and internal file storage.

## When to Use

- Application data storage (uploads, generated files, backups)
- Internal file storage that should never be publicly accessible
- Any bucket where accidental deletion protection via versioning is valuable

## Key Configuration Choices

- **Uniform bucket-level access** (`uniformBucketLevelAccessEnabled: true`) -- IAM-only access control, no legacy ACLs
- **Public access prevention enforced** -- blocks any public access regardless of IAM changes
- **Versioning enabled** -- protects against accidental deletion and overwrite
- **Lifecycle rules** -- automatically deletes non-current versions (max 3 kept, or 1 year old)
- **STANDARD storage class** -- lowest latency and highest availability; change to NEARLINE for infrequently accessed data

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<your-bucket-name>` | Globally unique bucket name (3-63 chars, lowercase) | Choose a unique name |
| `<gcp-region>` | Bucket location (e.g., `us-central1`, `US` for multi-region) | Your deployment region |

## Related Presets

- **02-static-website** -- Use for publicly accessible static website hosting
