---
title: "Presets"
description: "Ready-to-deploy configuration presets for KMS Key"
type: "preset-list"
componentSlug: "kms-key"
componentTitle: "KMS Key"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-symmetric-encryption"
    rank: "01"
    title: "Preset: Symmetric Encryption Key (CMEK)"
    excerpt: "Use this preset when you need a standard customer-managed encryption key (CMEK) for encrypting data in GCP services like BigQuery, Spanner, CloudSQL, GCS, GKE, PubSub, or AlloyDB."
  - slug: "02-hsm-symmetric-encryption"
    rank: "02"
    title: "Preset: HSM-Protected Symmetric Encryption Key"
    excerpt: "Use this preset when you need a customer-managed encryption key that is protected by hardware security modules (HSM). This is required for compliance scenarios that mandate FIPS 140-2 Level 3..."
  - slug: "03-asymmetric-signing"
    rank: "03"
    title: "Preset: Asymmetric Signing Key"
    excerpt: "Use this preset when you need a key for digital signatures -- signing build artifacts, container images, JWTs, or any data that requires integrity verification."
---

# KMS Key Presets

Ready-to-deploy configuration presets for KMS Key. Each preset is a complete manifest you can copy, customize, and deploy.
