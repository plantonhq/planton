---
title: "Wildcard Domain DNS-Validated Certificate"
description: "This preset provisions a wildcard ACM certificate that covers all subdomains of a domain, plus the apex domain itself as a Subject Alternative Name (SAN). A single wildcard certificate can secure..."
type: "preset"
rank: "02"
presetSlug: "02-wildcard-domain"
componentSlug: "certificate"
componentTitle: "Certificate"
provider: "aws"
icon: "package"
order: 2
---

# Wildcard Domain DNS-Validated Certificate

This preset provisions a wildcard ACM certificate that covers all subdomains of a domain, plus the apex domain itself as a Subject Alternative Name (SAN). A single wildcard certificate can secure `app.example.com`, `api.example.com`, `www.example.com`, and `example.com` without needing separate certificates.

## When to Use

- Multiple subdomains under the same apex domain that all need HTTPS
- Microservice architectures where each service has its own subdomain
- Simplifying certificate management by using one certificate for an entire domain

## Key Configuration Choices

- **Wildcard primary** (`primaryDomainName: *.example.com`) -- Covers all first-level subdomains (e.g., `app.example.com`, `api.example.com`)
- **Apex as SAN** (`alternateDomainNames: [example.com]`) -- Wildcard certificates do not cover the bare apex domain; adding it as a SAN ensures full coverage
- **DNS validation** (`validationMethod: DNS`) -- Fully automated when using Route53; validates both the wildcard and apex domain

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-domain.com>` | Your apex domain (e.g., `example.com`); used in both wildcard and SAN | Your domain registrar or DNS provider |
| `<route53-hosted-zone-id>` | ID of the Route53 hosted zone matching your domain | AWS Route53 console or `AwsRoute53Zone` status outputs |

## Related Presets

- **01-single-domain-dns** -- Use instead when you only need a certificate for one specific domain or subdomain
