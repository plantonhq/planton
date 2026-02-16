---
title: "Presets"
description: "Ready-to-deploy configuration presets for Secrets Manager"
type: "preset-list"
componentSlug: "secrets-manager"
componentTitle: "Secrets Manager"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-application-secrets"
    rank: "01"
    title: "Application Secrets"
    excerpt: "This preset creates a set of secrets in AWS Secrets Manager for a typical application. It provisions empty secret placeholders for database credentials, API keys, and TLS certificates. Secret values..."
---

# Secrets Manager Presets

Ready-to-deploy configuration presets for Secrets Manager. Each preset is a complete manifest you can copy, customize, and deploy.
