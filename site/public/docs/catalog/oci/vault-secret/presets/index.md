---
title: "Presets"
description: "Ready-to-deploy configuration presets for Vault Secret"
type: "preset-list"
componentSlug: "vault-secret"
componentTitle: "Vault Secret"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-explicit-credential"
    rank: "01"
    title: "Explicit Credential"
    excerpt: "This preset creates an OCI Vault Secret with user-provided base64-encoded content and a 90-day version expiry rule. This is the standard pattern for storing application credentials, API keys,..."
  - slug: "02-auto-generated-passphrase"
    rank: "02"
    title: "Auto-Generated Passphrase with Rotation"
    excerpt: "This preset creates an OCI Vault Secret where OCI automatically generates a 32-character passphrase and rotates it every 30 days against an Autonomous Database. The rotation process generates a new..."
---

# Vault Secret Presets

Ready-to-deploy configuration presets for Vault Secret. Each preset is a complete manifest you can copy, customize, and deploy.
