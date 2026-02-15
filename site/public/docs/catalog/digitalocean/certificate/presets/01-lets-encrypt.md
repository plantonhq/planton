---
title: "Let's Encrypt Certificate"
description: "This preset creates a free, auto-renewing SSL certificate from Let's Encrypt via DigitalOcean. Supports multiple domains and wildcards. DigitalOcean handles renewal automatically; use the certificate..."
type: "preset"
rank: "01"
presetSlug: "01-lets-encrypt"
componentSlug: "certificate"
componentTitle: "Certificate"
provider: "digitalocean"
icon: "package"
order: 1
---

# Let's Encrypt Certificate

This preset creates a free, auto-renewing SSL certificate from Let's Encrypt via DigitalOcean. Supports multiple domains and wildcards. DigitalOcean handles renewal automatically; use the certificate name in load balancers for HTTPS termination. Ideal for production websites.

## When to Use

- Production HTTPS for public websites
- Cost-free SSL with automatic renewal
- Multiple domains or subdomains on one certificate
- Load balancer SSL termination

## Key Configuration Choices

- **Let's Encrypt** (`type: lets_encrypt`, `letsEncrypt`) -- free, auto-renewed; DigitalOcean performs ACME validation.
- **Domains** (`domains`) -- list of FQDNs to include (e.g., `example.com`, `www.example.com`). Wildcards supported (e.g., `*.example.com`) but require DNS validation.
- **Auto-renew enabled** (`disableAutoRenew: false`) -- DigitalOcean renews before expiry; keep enabled for production.
- **Certificate name** (`certificateName`) -- used when referencing this cert in load balancers; use a stable name for IaC.
- **Use certificate name in LB** -- reference by name (not ID) so IaC state survives renewals.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `example.com`, `www.example.com` | Domains to include in the certificate | Your registered domain and desired subdomains |
| `my-lets-encrypt-cert` | Human-readable certificate identifier | Choose a descriptive name; used in load balancer config |

## Related Presets

- **02-custom** -- Use when you have an existing certificate (e.g., from enterprise CA, purchased cert)
