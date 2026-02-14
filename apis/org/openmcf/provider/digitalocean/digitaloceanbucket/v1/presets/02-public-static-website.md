# Public Static Website Bucket

This preset creates a public-read DigitalOcean Spaces bucket suitable for hosting static websites, CDN origins, or publicly served assets (JS, CSS, images). Objects are readable by anyone with the URL. Use with a CDN or custom domain for production sites.

## When to Use

- Static site hosting (HTML, JS, CSS, images)
- CDN origin for cached assets
- Public documentation or media files
- JAMstack deployments (build output to bucket)

## Key Configuration Choices

- **Public read** (`accessControl: PUBLIC_READ`) -- objects are publicly accessible; suitable for static assets.
- **Versioning disabled** (`versioningEnabled: false`) -- typical for static sites where overwrites are intentional.
- **Tags** (`static-site`, `cdn`) -- for organization and billing visibility.
- **Bucket name** (`bucketName`) -- use a subdomain-friendly name if serving via custom domain (e.g., `assets-myapp`).

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `nyc3` | Target DigitalOcean region slug | [DigitalOcean Regions API](https://docs.digitalocean.com/reference/api/api-reference/#tag/Regions) |
| `my-static-website` | Unique bucket name (3-63 chars, DNS-compatible) | Choose a unique name; consider region endpoint format for custom domains |

## Related Presets

- **01-private** -- Use for sensitive data or when access must be restricted to authenticated users
