---
title: "ExternalDNS on EKS with AWS Route53"
description: "This preset deploys ExternalDNS on an EKS cluster to automatically manage DNS records in AWS Route53. Authentication uses IRSA (IAM Roles for Service Accounts); an IAM role is auto-created if not..."
type: "preset"
rank: "02"
presetSlug: "02-eks-route53"
componentSlug: "external-dns"
componentTitle: "External DNS"
provider: "kubernetes"
icon: "package"
order: 2
---

# ExternalDNS on EKS with AWS Route53

This preset deploys ExternalDNS on an EKS cluster to automatically manage DNS records in AWS Route53. Authentication uses IRSA (IAM Roles for Service Accounts); an IAM role is auto-created if not explicitly provided.

## When to Use

- You run EKS and use AWS Route53 for your domain
- You want automatic DNS record creation when Ingress or Service resources are created

## Key Configuration Choices

- **EKS provider** -- uses IRSA for authentication; an IAM role is automatically created and bound
- **Default versions** -- uses ExternalDNS v0.19.0 and Helm chart 1.19.0 (proto defaults)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-route53-hosted-zone-id>` | Route53 hosted zone ID for the managed domain | AWS Console > Route53 > Hosted Zones |

## Related Presets

- **01-gke-cloud-dns** -- Use on GKE with Google Cloud DNS
- **03-aks-azure-dns** -- Use on AKS with Azure DNS
- **04-cloudflare** -- Use with Cloudflare DNS on any cluster
