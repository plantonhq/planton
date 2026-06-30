# Web Acceleration with HTTPS

This preset creates a CDN-accelerated domain for web content (images, small files, web pages) with HTTPS enabled via an Alibaba Cloud Certificate Management Service (CAS) certificate. A single domain-based origin pulls content over port 443. After creation, point a CNAME record at your DNS provider to the `cname` value from the stack outputs.

## When to Use

- Production websites and web applications that need global edge caching with HTTPS
- Accelerating images, CSS, JavaScript, and HTML pages served from an origin server
- Domains that already have a certificate provisioned in Alibaba Cloud CAS
- Mainland China web traffic where domestic-scope CDN provides the best latency

## Key Configuration Choices

- **Web CDN type** (`cdnType: web`) -- Optimized for small files, images, and web pages. Uses aggressive edge caching and connection reuse. Choose `download` for large file distribution or `video` for streaming.
- **Domestic scope** (`scope: domestic`) -- CDN edge nodes within mainland China only. This is the default and provides the best latency for China-based users. Change to `global` for worldwide coverage or `overseas` for non-China traffic.
- **Domain origin on port 443** (`type: domain`, `port: 443`) -- Origin pull uses HTTPS to protect data in transit between the CDN edge and your origin server. The origin domain must have a valid TLS certificate.
- **CAS certificate** (`certType: cas`) -- Uses a certificate managed in Alibaba Cloud Certificate Management Service, avoiding the need to upload and rotate certificate files manually. The `certRegion` defaults to `cn-hangzhou` for domestic certificates; use `ap-southeast-1` for international certificates.
- **Priority 20** (`priority: 20`) -- Standard primary origin priority. Add a standby origin with priority 30 for failover if needed.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region for provider API calls (CDN itself is global) | Any region, typically `cn-hangzhou` |
| `<your-cdn-domain>` | Domain name to accelerate (e.g., `cdn.example.com`) | Your domain registrar |
| `<your-origin-domain>` | Origin server domain (e.g., `origin.example.com`) | Your origin infrastructure |
| `<your-cas-certificate-id>` | Certificate ID from Alibaba Cloud CAS | CAS console |
| `<your-team>` | Team or business unit | Your organizational structure |

## Related Presets

- **02-oss-static-assets** -- Use when the origin is an OSS bucket for static website hosting or asset serving
