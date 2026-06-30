---
title: "Presets"
description: "Ready-to-deploy configuration presets for Origin CA Certificate"
type: "preset-list"
componentSlug: "origin-ca-certificate"
componentTitle: "Origin CA Certificate"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-generated-key"
    rank: "01"
    title: "Preset: Generated Key (recommended)"
    excerpt: "The recommended default: the module generates an RSA key + CSR for your hostnames and returns the signed certificate together with the (sensitive) private key. A downstream origin can mount both..."
  - slug: "02-byo-csr"
    rank: "02"
    title: "Preset: Bring Your Own CSR"
    excerpt: "For teams that already manage their own key material. You supply a PEM-encoded CSR; the module requests the certificate for that exact CSR and generates no key, so your private key never leaves your..."
---

# Origin CA Certificate Presets

Ready-to-deploy configuration presets for Origin CA Certificate. Each preset is a complete manifest you can copy, customize, and deploy.
