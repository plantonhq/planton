# API Worker with Custom Domain

Full-featured Worker with KV bindings, custom domain DNS routing, and environment variables. Use for production APIs that need storage, a custom hostname, and config. Script bundle must be pre-built and uploaded to R2.

## When to Use

- REST or GraphQL APIs running at the edge
- Workers needing KV for cache or session data
- Custom domain for API (e.g., api.example.com)

## Key Configuration Choices

- **kvBindings** (`kvBindings`) -- Each entry uses value wrapper; reference CloudflareKvNamespace namespace_id.
- **dns** (`dns`) -- enabled, zoneId, hostname; routePattern defaults to hostname/* if omitted.
- **env.variables** (`env.variables`) -- Non-sensitive config; use env.secrets for sensitive values.
- **scriptBundle** (`scriptBundle`) -- R2 bucket and path to the pre-built Worker bundle.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-account-id>` | Cloudflare account ID | Dashboard → Overview → Account ID |
| `<worker-name>` | Worker name (visible in dashboard) | Descriptive name (e.g., api-gateway) |
| `<r2-bucket-name>` | R2 bucket containing the script bundle | CloudflareR2Bucket bucket_name |
| `<script-bundle-path>` | Path to bundle in R2 (e.g., dist/worker.js) | Your build output path |
| `<kv-namespace-id>` | KV namespace ID to bind | CloudflareKvNamespace status.outputs.namespace_id |
| `<cloudflare-zone-id>` | Zone ID for DNS route | CloudflareDnsZone status.outputs.zone_id |
| `api.example.com` | Hostname for the Worker | Your desired API subdomain |

## Related Presets

- **02-minimal** -- Use when you only need a bare Worker with script bundle
