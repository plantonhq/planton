---
title: "Presets"
description: "Ready-to-deploy configuration presets for Auth0Action"
type: "preset-list"
componentSlug: "auth0action"
componentTitle: "Auth0Action"
provider: "auth0"
icon: "package"
order: 200
presets:
  - slug: "01-post-login-custom-claims"
    rank: "01"
    title: "Preset: Post-Login Custom Claims"
    excerpt: "Enrich ID and access tokens with custom claims after successful authentication. This is the most common Auth0 Action pattern — nearly every production Auth0 tenant needs custom claims for role-based..."
  - slug: "02-pre-registration-domain-allowlist"
    rank: "02"
    title: "Preset: Pre-Registration Domain Allowlist"
    excerpt: "Restrict user registration to specific email domains. Users with disallowed domains receive a clear denial message and are blocked from creating an account."
  - slug: "03-credentials-exchange-audit-log"
    rank: "03"
    title: "Preset: Credentials Exchange Audit Log"
    excerpt: "Log every machine-to-machine (M2M) token exchange to an external audit endpoint for compliance and security monitoring."
---

# Auth0Action Presets

Ready-to-deploy configuration presets for Auth0Action. Each preset is a complete manifest you can copy, customize, and deploy.
