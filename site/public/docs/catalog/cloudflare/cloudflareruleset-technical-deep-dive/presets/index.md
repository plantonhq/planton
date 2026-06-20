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
---

# CloudflareRuleset — Technical Deep Dive Presets

Ready-to-deploy configuration presets for CloudflareRuleset — Technical Deep Dive. Each preset is a complete manifest you can copy, customize, and deploy.
