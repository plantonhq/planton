# Cloudflare Ruleset

Provision and manage Cloudflare Rulesets using OpenMCF's unified API — Origin Rules, WAF Custom Rules, Cache Rules, Redirect Rules, Transform Rules, and more.

## Overview

Cloudflare Rulesets are ordered collections of rules that execute during specific phases of HTTP request processing on the Cloudflare edge network. A single `CloudflareRuleset` component can model any phase — the `phase` and rule `action` fields determine the behavior.

Common use cases include:

- **Origin Rules** (`http_request_origin`) — Override the origin server for matching requests (e.g., split traffic between origins based on URL path)
- **WAF Custom Rules** (`http_request_firewall_custom`) — Block, challenge, or log requests matching security expressions
- **Cache Rules** (`http_request_cache_settings`) — Configure per-path caching behavior, edge TTL, and browser TTL
- **Redirect Rules** (`http_request_dynamic_redirect`) — Redirect requests with custom status codes
- **Transform Rules** (`http_request_transform`, `http_response_headers_transform`) — Rewrite URLs or modify headers

This component follows the **80/20 principle**: it exposes the action parameters needed for the most common ruleset types while keeping the API manageable. Niche settings (autominify, polish, rocket loader, etc.) are intentionally excluded.

## Key Features

- **All 24 Phases**: Full coverage of Cloudflare's request processing pipeline
- **13 Action Types**: route, block, challenge, execute, skip, redirect, rewrite, set_cache_settings, and more
- **Origin Override**: Route traffic to different origins with custom Host headers and SNI
- **Managed WAF Integration**: Execute Cloudflare's managed rulesets with per-rule and per-category overrides
- **Cache Control**: Edge TTL, browser TTL, status-code-specific TTLs, and serve-stale configuration
- **Validation**: Built-in protobuf validation for required fields, enum ranges, and mutual-exclusion constraints
- **Zone or Account Scope**: Create rulesets at zone level (single domain) or account level (all domains)

## Prerequisites

1. **Cloudflare Zone** (for zone-level rulesets): Use the `CloudflareDnsZone` component or provide a zone ID
2. **API Token**: Cloudflare API token with appropriate permissions (Zone > Zone Settings > Edit for most phases)
3. **Cloudflare Proxy Enabled**: Origin Rules and most request-phase rulesets require Cloudflare proxy (orange cloud) to be enabled on the DNS record
4. **OpenMCF CLI**: Install from [openmcf.org](https://openmcf.org)

## Quick Start

### Origin Rule

Route non-marketing traffic to a Kubernetes origin:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareRuleset
metadata:
  name: planton-origin-routing
spec:
  zoneId: "<your-zone-id>"
  rulesetKind: zone
  phase: http_request_origin
  name: "Route app traffic to K8s"
  rules:
    - expression: >-
        not (
          http.request.uri.path eq "/" or
          http.request.uri.path starts_with "/docs" or
          http.request.uri.path starts_with "/features" or
          http.request.uri.path starts_with "/_site"
        )
      action: route
      description: "Route non-marketing paths to K8s origin"
      actionParameters:
        hostHeader: "planton.ai"
        origin:
          host: "<k8s-lb-hostname>"
          port: 443
```

Deploy:

```bash
planton apply -f ruleset.yaml
```

### WAF Custom Rule

Block requests from specific IPs:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareRuleset
metadata:
  name: block-bad-actors
spec:
  zoneId: "<your-zone-id>"
  rulesetKind: zone
  phase: http_request_firewall_custom
  name: "Block malicious traffic"
  rules:
    - expression: 'ip.src eq 192.0.2.1'
      action: block
      description: "Block known bad actor"
      actionParameters:
        response:
          statusCode: 403
          content: "Access denied"
          contentType: "text/plain"
```

### Cache Rule

Override caching for static assets:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareRuleset
metadata:
  name: cache-static-assets
spec:
  zoneId: "<your-zone-id>"
  rulesetKind: zone
  phase: http_request_cache_settings
  name: "Cache static assets aggressively"
  rules:
    - expression: 'http.request.uri.path starts_with "/assets"'
      action: set_cache_settings
      description: "Cache assets for 24 hours at edge"
      actionParameters:
        cache: true
        edgeTtl:
          mode: "override_origin"
          defaultTtl: 86400
        browserTtl:
          mode: "override_origin"
          defaultTtl: 3600
```

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `zoneId` | StringValueOrRef | One of zone/account | Cloudflare Zone ID |
| `accountId` | string | One of zone/account | Cloudflare Account ID |
| `rulesetKind` | enum | No (default: zone) | `zone`, `custom`, `managed`, `root` |
| `phase` | enum | Yes | Processing phase (24 values) |
| `name` | string | Yes | Human-readable ruleset name |
| `description` | string | No | Informative description |
| `rules` | repeated Rule | Yes (min 1) | Ordered list of rules |

### Rule Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ref` | string | No | Stable reference ID (prevents rule recreation) |
| `expression` | string | Yes | Cloudflare wirefilter expression |
| `action` | enum | Yes | Action to perform (13 values) |
| `description` | string | No | Rule description |
| `enabled` | bool | No (default: true) | Whether rule is active |
| `actionParameters` | message | Depends on action | Action-specific configuration |

### Action Parameters (by action type)

| Action | Key Parameters |
|--------|---------------|
| `route` | `hostHeader`, `origin { host, port }`, `sni { value }` |
| `block` | `response { statusCode, content, contentType }` |
| `rewrite` | `uri { path, query }`, `headers` (map) |
| `redirect` | `fromValue { targetUrl, statusCode, preserveQueryString }` |
| `skip` | `phases`, `products`, `ruleset`, `rulesets` |
| `execute` | `id`, `overrides { action, enabled, categories, rules }` |
| `set_cache_settings` | `cache`, `edgeTtl`, `browserTtl`, `serveStale` |

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `ruleset_id` | string | Cloudflare-assigned ruleset ID |
| `version` | string | Current ruleset version |
| `zone_id` | string | Zone ID (pass-through) |
| `phase` | string | Phase (pass-through) |

## Phases Reference

| Phase | Purpose |
|-------|---------|
| `http_request_origin` | Origin Rules (override origin server) |
| `http_request_firewall_custom` | WAF Custom Rules |
| `http_request_firewall_managed` | WAF Managed Rulesets |
| `http_request_cache_settings` | Cache Rules |
| `http_request_dynamic_redirect` | Dynamic Redirect Rules |
| `http_request_redirect` | Bulk Redirect Rules |
| `http_request_transform` | URL Rewrite Rules |
| `http_response_headers_transform` | Response Header Modification |
| `http_ratelimit` | Rate Limiting Rules |
| `http_request_late_transform` | Late Transform Rules |
| `http_config_settings` | Configuration Rules |
| `http_custom_errors` | Custom Error Responses |

See `spec.proto` for the complete list of 24 phases.

## Expression Language

Rules use Cloudflare's wirefilter expression language. Common fields:

- `http.request.uri.path` — Request path (e.g., `"/api"`)
- `http.host` — Request hostname
- `ip.src` — Client IP address
- `ip.geoip.country` — Client country code
- `cf.threat_score` — Cloudflare threat score (0-100)
- `http.request.method` — HTTP method

Operators: `eq`, `ne`, `starts_with`, `ends_with`, `contains`, `in`, `not`, `and`, `or`

Example: `not (http.request.uri.path starts_with "/static" or http.request.uri.path eq "/")`

## Best Practices

1. **One ruleset per phase per zone**: Cloudflare allows only one custom ruleset per phase per zone. Multiple rules go in the same ruleset.
2. **Use `ref` for stable rule identity**: Prevents Terraform from recreating rules when their position changes in the list.
3. **Order matters**: Rules are evaluated top-to-bottom. Place more specific rules before general ones.
4. **Test expressions**: Use the Cloudflare dashboard's expression builder to validate expressions before deploying.
5. **Proxy required**: Most request-phase rulesets require the orange cloud (Cloudflare proxy) to be enabled on the DNS record.

## Related Components

- **CloudflareDnsZone** — Manage the DNS zone that this ruleset applies to
- **CloudflareDnsRecord** — Manage DNS records (must be proxied for rulesets to take effect)
