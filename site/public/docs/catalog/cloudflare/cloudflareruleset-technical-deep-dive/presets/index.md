---
title: "Presets"
description: "Ready-to-deploy configuration presets for CloudflareRuleset — Technical Deep Dive"
type: "preset-list"
componentSlug: "cloudflareruleset-technical-deep-dive"
componentTitle: "CloudflareRuleset — Technical Deep Dive"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-origin-rule"
    rank: "01"
    title: "Origin Rule — Split Traffic Between Origins"
    excerpt: "Route requests to different origin servers based on URL path. The default origin (configured in DNS) handles marketing/static paths, while an Origin Rule overrides the origin for application paths."
  - slug: "02-waf-managed"
    rank: "02"
    title: "Managed WAF — Cloudflare + OWASP Rulesets"
    excerpt: "Enable Cloudflare's managed WAF rulesets to protect against common web attacks including XSS, SQL injection, and OWASP Top 10 vulnerabilities."
  - slug: "03-cache-settings"
    rank: "03"
    title: "Cache Settings — Static Assets + API Bypass"
    excerpt: "Configure Cloudflare's edge caching with aggressive TTLs for static assets and explicit cache bypass for dynamic API endpoints."
  - slug: "04-rate-limit"
    rank: "04"
    title: "Preset: Rate limiting rules"
    excerpt: "Throttle abusive traffic with the `http_ratelimit` phase. Each rule counts requests per a set of characteristics (buckets) over a period and applies its action when the threshold is exceeded."
  - slug: "05-config-settings"
    rank: "05"
    title: "Preset: Configuration rules (set_config)"
    excerpt: "Override zone settings per request with the `http_config_settings` phase. A `set_config` rule applies settings (SSL mode, security level, Rocket Loader, Polish, auto-minify, email obfuscation, …)..."
  - slug: "06-advanced-cache-key"
    rank: "06"
    title: "Preset: Advanced cache key and Cache Reserve"
    excerpt: "Take full control of the cache key with the `http_request_cache_settings` phase: include only the query parameters, cookies, headers, and user attributes that should vary the cached object, and..."
  - slug: "07-bulk-redirect"
    rank: "07"
    title: "Bulk Redirect — Redirect From a List"
    excerpt: "Apply a large set of URL redirects from a reusable Bulk Redirect list. The list holds the source → target entries (managed independently as `CloudflareListItem` resources), and an account-level..."
---

# CloudflareRuleset — Technical Deep Dive Presets

Ready-to-deploy configuration presets for CloudflareRuleset — Technical Deep Dive. Each preset is a complete manifest you can copy, customize, and deploy.
