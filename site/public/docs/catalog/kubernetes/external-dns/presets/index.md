---
title: "Presets"
description: "Ready-to-deploy configuration presets for External DNS"
type: "preset-list"
componentSlug: "external-dns"
componentTitle: "External DNS"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-gke-cloud-dns"
    rank: "01"
    title: "ExternalDNS on GKE with Google Cloud DNS"
    excerpt: "This preset deploys ExternalDNS on a GKE cluster to automatically manage DNS records in Google Cloud DNS. ExternalDNS watches Kubernetes Ingress and Service resources and creates/updates DNS records..."
  - slug: "02-eks-route53"
    rank: "02"
    title: "ExternalDNS on EKS with AWS Route53"
    excerpt: "This preset deploys ExternalDNS on an EKS cluster to automatically manage DNS records in AWS Route53. Authentication uses IRSA (IAM Roles for Service Accounts); an IAM role is auto-created if not..."
  - slug: "03-aks-azure-dns"
    rank: "03"
    title: "ExternalDNS on AKS with Azure DNS"
    excerpt: "This preset deploys ExternalDNS on an AKS cluster to automatically manage DNS records in Azure DNS. Authentication uses a user-assigned managed identity bound to the ExternalDNS service account."
  - slug: "04-cloudflare"
    rank: "04"
    title: "ExternalDNS with Cloudflare DNS"
    excerpt: "This preset deploys ExternalDNS with Cloudflare as the DNS provider. Works on any Kubernetes cluster (GKE, EKS, AKS, or self-managed). Authentication uses a Cloudflare API token."
---

# External DNS Presets

Ready-to-deploy configuration presets for External DNS. Each preset is a complete manifest you can copy, customize, and deploy.
