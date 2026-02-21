---
title: "Presets"
description: "Ready-to-deploy configuration presets for KMS Key"
type: "preset-list"
componentSlug: "kms-key"
componentTitle: "KMS Key"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Encryption Key"
    excerpt: "This preset creates a KMS key with all defaults: AES-256 symmetric encryption, software-based protection, no automatic rotation, and no deletion protection. This is the simplest configuration,..."
  - slug: "02-production-with-rotation"
    rank: "02"
    title: "Production Encryption Key with Rotation"
    excerpt: "This preset creates a production-grade KMS key with annual automatic rotation and deletion protection enabled. It is the recommended starting point for any KMS key that protects production data --..."
  - slug: "03-asymmetric-signing"
    rank: "03"
    title: "Asymmetric Signing Key"
    excerpt: "This preset creates an RSA-2048 asymmetric key for digital signature generation and verification. The private key never leaves KMS; only the public key can be exported for external verification."
---

# KMS Key Presets

Ready-to-deploy configuration presets for KMS Key. Each preset is a complete manifest you can copy, customize, and deploy.
