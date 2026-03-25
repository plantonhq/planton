# Cache Settings — Static Assets + API Bypass

Configure Cloudflare's edge caching with aggressive TTLs for static assets and explicit cache bypass for dynamic API endpoints.

## When to Use

- Sites with a mix of static assets and dynamic API responses
- When origin Cache-Control headers are missing or need overriding
- Reducing origin load by caching immutable assets at Cloudflare's edge

## Key Configuration Choices

- **`phase: http_request_cache_settings`** — Cache Rules phase
- **Edge TTL 86400s (24 hours)** — Static assets are cached at Cloudflare's edge for 24 hours, reducing origin requests
- **Browser TTL 3600s (1 hour)** — Clients cache assets locally for 1 hour, balancing freshness with performance
- **`cache: false` for `/api`** — Explicitly bypasses cache for API endpoints to ensure fresh responses

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-zone-id>` | Zone ID for your domain | Cloudflare dashboard > Overview > Zone ID |

## Related Presets

- **01-origin-rule** — Route traffic to the correct origin before cache evaluation
- **02-waf-managed** — WAF rules execute before cache rules in the pipeline
