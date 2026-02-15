---
title: "ExternalDNS on AKS with Azure DNS"
description: "This preset deploys ExternalDNS on an AKS cluster to automatically manage DNS records in Azure DNS. Authentication uses a user-assigned managed identity bound to the ExternalDNS service account."
type: "preset"
rank: "03"
presetSlug: "03-aks-azure-dns"
componentSlug: "external-dns"
componentTitle: "External DNS"
provider: "kubernetes"
icon: "package"
order: 3
---

# ExternalDNS on AKS with Azure DNS

This preset deploys ExternalDNS on an AKS cluster to automatically manage DNS records in Azure DNS. Authentication uses a user-assigned managed identity bound to the ExternalDNS service account.

## When to Use

- You run AKS and use Azure DNS for your domain
- You want automatic DNS record creation when Ingress or Service resources are created

## Key Configuration Choices

- **AKS provider** -- uses Azure Managed Identity for authentication; no client secrets needed
- **Default versions** -- uses ExternalDNS v0.19.0 and Helm chart 1.19.0 (proto defaults)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-azure-dns-zone-id>` | Azure DNS zone resource ID | Azure Portal > DNS Zones |
| `<your-managed-identity-client-id>` | Client ID of the managed identity with DNS Zone Contributor role | Azure Portal > Managed Identities |

## Related Presets

- **01-gke-cloud-dns** -- Use on GKE with Google Cloud DNS
- **02-eks-route53** -- Use on EKS with AWS Route53
- **04-cloudflare** -- Use with Cloudflare DNS on any cluster
