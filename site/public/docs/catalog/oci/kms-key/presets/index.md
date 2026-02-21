---
title: "Presets"
description: "Ready-to-deploy configuration presets for KMS Key"
type: "preset-list"
componentSlug: "kms-key"
componentTitle: "KMS Key"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-aes-256-hsm-auto-rotation"
    rank: "01"
    title: "AES-256 HSM Key with Auto-Rotation"
    excerpt: "This preset creates an AES-256 symmetric encryption key stored in an HSM (FIPS 140-2 Level 3) with automatic 90-day key rotation. This is the standard key for encrypting data at rest across OCI..."
  - slug: "02-rsa-4096-hsm-signing"
    rank: "02"
    title: "RSA-4096 HSM Signing Key"
    excerpt: "This preset creates an RSA-4096 asymmetric key stored in an HSM for digital signing and verification. Asymmetric keys are used when the signer and verifier are different entities -- the private key..."
---

# KMS Key Presets

Ready-to-deploy configuration presets for KMS Key. Each preset is a complete manifest you can copy, customize, and deploy.
