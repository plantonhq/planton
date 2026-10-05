# Cloudflare Web Analytics Site

## Overview

`CloudflareWebAnalyticsSite` creates a Web Analytics (RUM) site: Cloudflare's privacy-first real-user monitoring for one website, measured by a small JavaScript beacon. No cookies, no cross-site tracking, free on every plan -- and it works whether or not the site proxies through Cloudflare. Include/exclude rules narrow what gets measured. A plain CRUD object -- real create, update, delete.

## Key Features

- **Two ways to identify a site** -- by hostname (any site; you embed the snippet) or by Cloudflare zone (Cloudflare can inject the snippet at the edge)
- **Automatic installation** -- with a zone, `auto_install` puts the beacon on every proxied page with no code change
- **Measurement rules** -- include or exclude traffic by host and path, folded into this component as `rules[]`
- **The lite beacon** -- a smaller script with a reduced metric set for pages that are latency-sensitive
- **Ready-to-embed snippet** -- the exact script tag is an output, alongside the site token it carries

## Use Cases

**Ideal for:**

- Privacy-conscious page analytics without a third-party tracker or a cookie banner
- Measuring a site that is not behind Cloudflare, with a manually embedded snippet
- Excluding admin, checkout, or internal paths from measurement

**Not ideal for:**

- Server-side traffic analytics -- Cloudflare's zone analytics and `CloudflareLogpushJob` cover requests the edge saw, including non-browser traffic
- Product event analytics -- RUM measures page performance and visits, not custom business events

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |
| `host` or `zone_tag` | string / reference | Exactly one | Hostname to measure, or the Cloudflare zone to measure. |

### Key Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `auto_install` | bool | Cloudflare injects the beacon at the edge (proxied zones only). |
| `enabled` | bool | Whether measurement is active (Cloudflare's default is enabled). |
| `lite` | bool | Serve the lightweight beacon variant. |
| `rules` | list | Include/exclude rows: `host`, `paths[]`, `inclusive`, `is_paused`. |

### Outputs

| Field | Description |
|-------|-------------|
| `site_tag` | The Cloudflare-assigned site tag (its identity in every RUM API path) |
| `site_token` | The beacon's measurement token |
| `snippet` | The ready-to-embed script tag (carries the token) |
| `ruleset_id` | The parent object the measurement rules live under |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareWebAnalyticsSite
metadata:
  name: www-example-rum
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  zone_tag:
    valueFrom:
      kind: CloudflareDnsZone
      name: example-com
      fieldPath: status.outputs.zone_id
  auto_install: true
  rules:
    - paths:
        - "/admin/*"
      inclusive: false
    - paths:
        - "/*"
      inclusive: true
```

## Prerequisites

- **A Cloudflare API token** with BOTH Account Settings Write AND Account Settings Read permissions -- measured: Write alone creates sites but every read-back (including the provider's own refresh) answers 403, so deploys break mid-lifecycle without the Read half
- For `auto_install`: a zone proxied through Cloudflare (orange-clouded)

## Destroy Semantics

Real delete for both the site and its rules. Measurement stops and historical analytics for the site are no longer reachable. If `auto_install` was on, the injected beacon disappears with the site.

## Related Kinds

- [Cloudflare DNS Zone](/docs/catalog/cloudflare/cloudflarednszone) -- the zone a zone-measured site points at
- [Cloudflare Logpush Job](/docs/catalog/cloudflare/cloudflarelogpushjob) -- edge-side request logs, the server's view of the same traffic
- [Cloudflare Notification Policy](/docs/catalog/cloudflare/cloudflarenotificationpolicy) -- the web-analytics metrics alert family

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
