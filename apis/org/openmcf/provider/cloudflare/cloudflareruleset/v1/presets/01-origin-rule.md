# Origin Rule — Split Traffic Between Origins

Route requests to different origin servers based on URL path. The default origin (configured in DNS) handles marketing/static paths, while an Origin Rule overrides the origin for application paths.

## When to Use

- Serving a marketing site and an application from the same domain using different backends
- Migrating from a subdomain (`app.example.com`) to path-based routing (`example.com/app`)
- A/B testing between different backends based on URL patterns

## Key Configuration Choices

- **`phase: http_request_origin`** — Origin Rules phase, the only phase where the `route` action is valid
- **`ruleset_kind: zone`** — Zone-level ruleset (applies to a single domain)
- **`host_header`** (`spec.rules[].actionParameters.hostHeader`) — Overrides the Host header sent to the origin so the backend receives the correct hostname
- **`origin.port: 443`** — HTTPS to the backend; Cloudflare re-initiates TLS to the origin

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-zone-id>` | Zone ID for your domain | Cloudflare dashboard > Overview > Zone ID |
| `<your-domain>` | The domain name (e.g., `planton.ai`) | Your DNS configuration |
| `<backend-hostname>` | Origin server hostname or IP | Your infrastructure (K8s LB, ALB, etc.) |

## Related Presets

- **02-waf-managed** — Add WAF protection to the same zone
- **03-cache-settings** — Configure caching for the static paths
