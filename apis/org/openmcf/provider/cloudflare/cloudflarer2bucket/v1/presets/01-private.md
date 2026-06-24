# Private R2 Bucket

Creates a private R2 bucket with no public access. Data is accessible only via Workers, API tokens, or the Cloudflare Dashboard. Use for backups, logs, or any object storage that must stay private.

## When to Use

- Backups, archives, or internal artifacts
- Logs and data that should not be publicly accessible
- Object storage accessed only via Workers or API

## Key Configuration Choices

- **publicAccess: false** (`publicAccess: false`) -- No public URL; access via Workers or credentials only.
- **location** (`location: auto`) -- Auto lets Cloudflare choose; or wnam, enam, weur, eeur, apac, oc.
- **bucketName** (`bucketName`) -- DNS-compatible, 3–63 chars, lowercase alphanumeric and hyphens.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<bucket-name>` | Unique bucket name | Choose DNS-safe name (e.g., app-backups-prod) |
| `<cloudflare-account-id>` | Cloudflare account ID | Dashboard → Overview → Account ID |
