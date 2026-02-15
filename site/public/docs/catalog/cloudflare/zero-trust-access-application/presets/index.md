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
  - slug: "01-company-wide-email"
    rank: "01"
    title: "Company-Wide Email Domain Access"
    excerpt: "Allows access to a protected hostname for anyone with an email from your company domain (e.g., @company.com). Simple Zero Trust policy for internal tools when the entire organization should have..."
  - slug: "02-team-google-groups"
    rank: "02"
    title: "Team Access with Google Groups + MFA"
    excerpt: "Restricts access to a hostname to specific Google Workspace groups and requires multi-factor authentication. Use for sensitive internal tools where only certain teams should have access and MFA is..."
---

# Zero Trust Access Application Presets

Ready-to-deploy configuration presets for Zero Trust Access Application. Each preset is a complete manifest you can copy, customize, and deploy.
