---
title: "Presets"
description: "Ready-to-deploy configuration presets for KMS Vault"
type: "preset-list"
componentSlug: "kms-vault"
componentTitle: "KMS Vault"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-shared-vault"
    rank: "01"
    title: "Shared Vault"
    excerpt: "This preset creates an OCI KMS Vault with the default vault type, which uses a shared HSM partition. Shared vaults provide FIPS 140-2 Level 3 certified key storage at lower cost than dedicated..."
  - slug: "02-dedicated-vault"
    rank: "02"
    title: "Dedicated Vault"
    excerpt: "This preset creates an OCI KMS Vault with the virtual private vault type, which allocates a dedicated HSM partition exclusively for your tenancy. Dedicated vaults provide higher cryptographic..."
---

# KMS Vault Presets

Ready-to-deploy configuration presets for KMS Vault. Each preset is a complete manifest you can copy, customize, and deploy.
