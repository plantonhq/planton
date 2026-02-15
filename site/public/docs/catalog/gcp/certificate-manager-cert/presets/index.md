---
title: "Presets"
description: "Ready-to-deploy configuration presets for Certificate Manager Cert"
type: "preset-list"
componentSlug: "certificate-manager-cert"
componentTitle: "Certificate Manager Cert"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-single-domain-dns"
    rank: "01"
    title: "Single Domain Managed Certificate"
    excerpt: "This preset creates a Google-managed TLS certificate for a single domain using DNS validation via Cloud DNS. The certificate is automatically provisioned and renewed by Google Certificate Manager."
  - slug: "02-wildcard-domain"
    rank: "02"
    title: "Wildcard Domain Certificate"
    excerpt: "This preset creates a Google-managed wildcard TLS certificate (`*.example.com`) with the apex domain (`example.com`) as a Subject Alternative Name. This covers all subdomains under a single..."
---

# Certificate Manager Cert Presets

Ready-to-deploy configuration presets for Certificate Manager Cert. Each preset is a complete manifest you can copy, customize, and deploy.
