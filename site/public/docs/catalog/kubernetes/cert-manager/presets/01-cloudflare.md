---
title: "Cert-Manager with Cloudflare DNS-01 Challenge"
description: "This preset deploys cert-manager with a Cloudflare DNS provider for DNS-01 ACME certificate challenges via Let's Encrypt production. Cloudflare is the most common DNS provider for cert-manager due to..."
type: "preset"
rank: "01"
presetSlug: "01-cloudflare"
componentSlug: "cert-manager"
componentTitle: "Cert Manager"
provider: "kubernetes"
icon: "package"
order: 1
---

# Cert-Manager with Cloudflare DNS-01 Challenge

This preset deploys cert-manager with a Cloudflare DNS provider for DNS-01 ACME certificate challenges via Let's Encrypt production. Cloudflare is the most common DNS provider for cert-manager due to its fast propagation and simple API token authentication.

## When to Use

- Your DNS zones are managed by Cloudflare
- You need automated TLS certificates from Let's Encrypt
- You want DNS-01 challenges (works for wildcard certificates and private clusters)

## Key Configuration Choices

- **ACME server** (Let's Encrypt production) -- issues trusted certificates; switch to staging URL for testing
- **DNS-01 challenge** -- proves domain ownership via DNS TXT records; unlike HTTP-01, works behind firewalls and supports wildcards
- **Cloudflare API token** -- scoped authentication; token needs `Zone:Zone:Read` and `Zone:DNS:Edit` permissions

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-acme-email@example.com>` | Email for Let's Encrypt registration and certificate expiry notifications | Your organization's ops email |
| `<your-domain.com>` | DNS zone that cert-manager will issue certificates for | Cloudflare dashboard > Websites |
| `<your-cloudflare-api-token>` | Cloudflare API token with Zone:Zone:Read and Zone:DNS:Edit | Cloudflare dashboard > API Tokens |

## Related Presets

- **02-gcp-cloud-dns** -- Use when DNS zones are hosted on Google Cloud DNS
- **03-aws-route53** -- Use when DNS zones are hosted on AWS Route53
- **04-azure-dns** -- Use when DNS zones are hosted on Azure DNS
