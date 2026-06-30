---
title: "Presets"
description: "Ready-to-deploy configuration presets for Zero Trust Access Group"
type: "preset-list"
componentSlug: "zero-trust-access-group"
componentTitle: "Zero Trust Access Group"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-engineering-team"
    rank: "01"
    title: "Preset: Engineering team group"
    excerpt: "A reusable account-scoped group that matches your engineering staff by email domain, requires an allowed country, and excludes a known contractor account."
  - slug: "02-idp-group-with-mfa"
    rank: "02"
    title: "Preset: IdP group with MFA login method"
    excerpt: "An account-scoped group that matches an Okta group (federated through a configured identity provider) and additionally requires a hardware-key authentication method."
---

# Zero Trust Access Group Presets

Ready-to-deploy configuration presets for Zero Trust Access Group. Each preset is a complete manifest you can copy, customize, and deploy.
