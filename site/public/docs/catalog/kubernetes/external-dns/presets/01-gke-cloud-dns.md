---
title: "ExternalDNS on GKE with Google Cloud DNS"
description: "This preset deploys ExternalDNS on a GKE cluster to automatically manage DNS records in Google Cloud DNS. ExternalDNS watches Kubernetes Ingress and Service resources and creates/updates DNS records..."
type: "preset"
rank: "01"
presetSlug: "01-gke-cloud-dns"
componentSlug: "external-dns"
componentTitle: "External DNS"
provider: "kubernetes"
icon: "package"
order: 1
---

# ExternalDNS on GKE with Google Cloud DNS

This preset deploys ExternalDNS on a GKE cluster to automatically manage DNS records in Google Cloud DNS. ExternalDNS watches Kubernetes Ingress and Service resources and creates/updates DNS records to match.

## When to Use

- You run GKE and use Google Cloud DNS for your domain
- You want automatic DNS record creation when Ingress or Service resources are created
- GKE Workload Identity is enabled on your cluster

## Key Configuration Choices

- **GKE provider** -- uses Workload Identity for authentication; no service account keys needed
- **Default versions** -- uses ExternalDNS v0.19.0 and Helm chart 1.19.0 (proto defaults)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-gcp-project-id>` | GCP project hosting the DNS zone and GKE cluster | GCP Console > Project Settings |
| `<your-cloud-dns-zone-id>` | Cloud DNS zone ID for the managed domain | GCP Console > Cloud DNS |

## Related Presets

- **02-eks-route53** -- Use on EKS with AWS Route53
- **03-aks-azure-dns** -- Use on AKS with Azure DNS
- **04-cloudflare** -- Use with Cloudflare DNS on any cluster
