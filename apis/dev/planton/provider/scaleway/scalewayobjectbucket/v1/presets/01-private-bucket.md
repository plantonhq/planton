# Private Object Bucket

This preset creates a minimal Scaleway Object Storage bucket with default settings. The bucket is private (no public access), has no versioning, and cannot be destroyed while it contains objects. This is the simplest and most common bucket configuration.

## When to Use

- General-purpose file storage (uploads, documents, backups)
- Application asset storage (images, videos, user-generated content)
- Quick setup when versioning and lifecycle rules are not needed

## Key Configuration Choices

- **Paris region** (`region: fr-par`) -- primary Scaleway region; change to `nl-ams` or `pl-waw` for Amsterdam or Warsaw
- **Versioning disabled** (`versioningEnabled: false`) -- no object version history; reduces storage costs
- **Force destroy disabled** (`forceDestroy: false`) -- the bucket cannot be deleted while it contains objects; prevents accidental data loss
- **No lifecycle rules** -- objects are retained indefinitely; add lifecycle rules for automatic archival or expiration
- **No CORS** -- no cross-origin access; add CORS rules for web application frontends that need to upload directly to the bucket

## Placeholders to Replace

No placeholders -- this preset is ready to deploy as-is. The bucket name is derived from `metadata.name`.

## Related Presets

- **02-versioned-lifecycle** -- Use instead for production data that needs version history, lifecycle transitions to Glacier, and multi-part upload cleanup
