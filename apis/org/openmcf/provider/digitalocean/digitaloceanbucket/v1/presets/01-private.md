# Private Bucket with Versioning

This preset creates a private DigitalOcean Spaces bucket with versioning enabled. Objects are not publicly accessible; use signed URLs or IAM for access. Versioning protects against accidental overwrites and supports compliance requirements.

## When to Use

- Storing backups, logs, or sensitive data
- Application assets that require IAM or signed URL access
- Workloads needing object version history for recovery

## Key Configuration Choices

- **Private access** (`accessControl: PRIVATE`) -- no public read; all access via credentials or signed URLs.
- **Versioning enabled** (`versioningEnabled: true`) -- preserves previous object versions; useful for backups and audit trails.
- **Tags** (`storage`, `backups`) -- for organization and lifecycle policies (if configured separately).
- **Bucket name** (`bucketName`) -- must be DNS-compatible (lowercase, hyphens, 3-63 chars); must be globally unique.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `nyc3` | Target DigitalOcean region slug | [DigitalOcean Regions API](https://docs.digitalocean.com/reference/api/api-reference/#tag/Regions) |
| `my-private-bucket` | Unique bucket name (3-63 chars, DNS-compatible) | Choose a unique name; must not conflict with other Spaces buckets globally |

## Related Presets

- **02-public-static-website** -- Use instead when hosting public static assets (JS, CSS, images)
