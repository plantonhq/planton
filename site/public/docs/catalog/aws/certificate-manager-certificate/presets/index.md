---
title: "Presets"
description: "Ready-to-deploy configuration presets for Certificate Manager Certificate"
type: "preset-list"
componentSlug: "certificate-manager-certificate"
componentTitle: "Certificate Manager Certificate"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-single-domain-dns"
    rank: "01"
    title: "Single Domain DNS-Validated Certificate"
    excerpt: "This preset provisions an ACM certificate for a single domain using automated DNS validation via Route53. DNS validation is the recommended method because it requires no manual intervention -- ACM..."
  - slug: "02-wildcard-domain"
    rank: "02"
    title: "Wildcard Domain DNS-Validated Certificate"
    excerpt: "This preset provisions a wildcard ACM certificate that covers all subdomains of a domain, plus the apex domain itself as a Subject Alternative Name (SAN). A single wildcard certificate can secure..."
---

# Certificate Manager Certificate Presets

Ready-to-deploy configuration presets for Certificate Manager Certificate. Each preset is a complete manifest you can copy, customize, and deploy.
