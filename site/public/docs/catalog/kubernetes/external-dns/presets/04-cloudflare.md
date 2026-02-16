---
title: "ExternalDNS with Cloudflare DNS"
description: "This preset deploys ExternalDNS with Cloudflare as the DNS provider. Works on any Kubernetes cluster (GKE, EKS, AKS, or self-managed). Authentication uses a Cloudflare API token."
type: "preset"
rank: "04"
presetSlug: "04-cloudflare"
componentSlug: "external-dns"
componentTitle: "External DNS"
provider: "kubernetes"
icon: "package"
order: 4
---

# ExternalDNS with Cloudflare DNS

This preset deploys ExternalDNS with Cloudflare as the DNS provider. Works on any Kubernetes cluster (GKE, EKS, AKS, or self-managed). Authentication uses a Cloudflare API token.

## When to Use

- Your DNS zones are managed by Cloudflare regardless of your Kubernetes provider
- You want automatic DNS record creation when Ingress or Service resources are created
- You do not need Cloudflare proxy (orange cloud) for managed records

## Key Configuration Choices

- **Cloudflare provider** -- manages DNS records via the Cloudflare API
- **Proxy disabled** (`isProxied` not set) -- DNS records are DNS-only (grey cloud); enable `isProxied` for DDoS protection and CDN
- **Default versions** -- uses ExternalDNS v0.19.0 and Helm chart 1.19.0 (proto defaults)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-cloudflare-api-token>` | API token with Zone:Zone:Read and Zone:DNS:Edit permissions | Cloudflare dashboard > API Tokens |
| `<your-cloudflare-dns-zone-id>` | Cloudflare zone ID for the managed domain | Cloudflare dashboard > Websites > your domain > Overview |

## Related Presets

- **01-gke-cloud-dns** -- Use on GKE with Google Cloud DNS
- **02-eks-route53** -- Use on EKS with AWS Route53
- **03-aks-azure-dns** -- Use on AKS with Azure DNS
