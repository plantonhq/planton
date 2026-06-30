---
title: "Presets"
description: "Ready-to-deploy configuration presets for Zero Trust Access Application"
type: "preset-list"
componentSlug: "zero-trust-access-application"
componentTitle: "Zero Trust Access Application"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-self-hosted-web-app"
    rank: "01"
    title: "Preset: Self-hosted web application"
    excerpt: "A self-hosted web app behind Cloudflare Access, protecting `dashboard.example.com` with a referenced Access policy and a 24-hour session."
  - slug: "02-saas-oidc-app"
    rank: "02"
    title: "Preset: SaaS application (OIDC)"
    excerpt: "Federate a SaaS application into Cloudflare Access over OIDC. Cloudflare acts as the identity provider; it issues the OAuth `client_id` / `client_secret` (exported as stack outputs) that you paste..."
---

# Zero Trust Access Application Presets

Ready-to-deploy configuration presets for Zero Trust Access Application. Each preset is a complete manifest you can copy, customize, and deploy.
