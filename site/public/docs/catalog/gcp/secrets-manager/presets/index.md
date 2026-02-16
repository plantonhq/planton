---
title: "Presets"
description: "Ready-to-deploy configuration presets for Secrets Manager"
type: "preset-list"
componentSlug: "secrets-manager"
componentTitle: "Secrets Manager"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-application-secrets"
    rank: "01"
    title: "Application Secrets"
    excerpt: "This preset creates a set of common application secrets in Google Cloud Secret Manager. The secrets are created as empty shells -- secret values must be populated separately via the GCP console,..."
---

# Secrets Manager Presets

Ready-to-deploy configuration presets for Secrets Manager. Each preset is a complete manifest you can copy, customize, and deploy.
