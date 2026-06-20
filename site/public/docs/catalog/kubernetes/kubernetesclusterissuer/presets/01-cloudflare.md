---
title: "ClusterIssuer with Cloudflare DNS-01 Challenge"
description: "This preset creates a ClusterIssuer that uses Cloudflare DNS for ACME DNS-01 certificate challenges via Let's Encrypt production. Cloudflare is the most common DNS provider for cert-manager due to..."
type: "preset"
rank: "01"
presetSlug: "01-cloudflare"
componentSlug: "kubernetesclusterissuer"
componentTitle: "KubernetesClusterIssuer"
provider: "kubernetes"
icon: "package"
order: 1
---

# ClusterIssuer with Cloudflare DNS-01 Challenge

This preset creates a ClusterIssuer that uses Cloudflare DNS for ACME DNS-01 certificate challenges via Let's Encrypt production. Cloudflare is the most common DNS provider for cert-manager due to its fast DNS propagation and simple API token authentication.

## When to Use

- Your DNS domains are managed by Cloudflare
- You need automated TLS certificates from Let's Encrypt
- You want DNS-01 challenges (works for wildcard certificates and private clusters)

## Key Configuration Choices

- **ACME server** (`server: https://acme-v02.api.letsencrypt.org/directory`) -- Let's Encrypt production; switch to staging URL for testing
- **DNS-01 challenge** -- proves domain ownership via DNS TXT records; works behind firewalls and supports wildcards
- **Cloudflare API token** -- scoped authentication; token needs `Zone:Zone:Read` and `Zone:DNS:Edit` permissions

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-domain.com>` | DNS domain to issue certificates for | Cloudflare dashboard > Websites |
| `<your-acme-email@example.com>` | Email for Let's Encrypt registration and expiry notifications | Your organization's ops email |
| `<your-cloudflare-api-token>` | Cloudflare API token with Zone:Zone:Read and Zone:DNS:Edit | Cloudflare dashboard > API Tokens |

## Related Presets

- **02-gcp-cloud-dns** -- Use when DNS domains are hosted on Google Cloud DNS
- **03-aws-route53** -- Use when DNS domains are hosted on AWS Route53
