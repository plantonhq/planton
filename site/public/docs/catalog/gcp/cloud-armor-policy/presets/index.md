---
title: "Presets"
description: "Ready-to-deploy configuration presets for Cloud Armor Policy"
type: "preset-list"
componentSlug: "cloud-armor-policy"
componentTitle: "Cloud Armor Policy"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-basic-ip-allowlist"
    rank: "01"
    title: "Preset: Basic IP Allowlist"
    excerpt: "Restrict access to your backend to known corporate or VPN IP ranges. All traffic from allowlisted CIDR blocks is permitted; everything else receives a 403. Ideal for internal dashboards, admin..."
  - slug: "02-rate-limiting-api"
    rank: "02"
    title: "Preset: Rate Limiting for API Endpoints"
    excerpt: "Protect APIs from abuse and DDoS with per-IP rate limiting and ban escalation. Traffic under the throttle limit is allowed; exceeding it returns 429. Persistent abusers crossing the ban threshold are..."
  - slug: "03-waf-owasp-protection"
    rank: "03"
    title: "Preset: WAF OWASP Protection"
    excerpt: "Protect web applications and APIs against OWASP Top 10–style attacks: SQL injection (SQLi), cross-site scripting (XSS), and Layer 7 DDoS. Uses Cloud Armor preconfigured WAF rules derived from OWASP..."
---

# Cloud Armor Policy Presets

Ready-to-deploy configuration presets for Cloud Armor Policy. Each preset is a complete manifest you can copy, customize, and deploy.
