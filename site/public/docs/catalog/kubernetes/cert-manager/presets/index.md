---
title: "Presets"
description: "Ready-to-deploy configuration presets for Cert Manager"
type: "preset-list"
componentSlug: "cert-manager"
componentTitle: "Cert Manager"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-cloudflare"
    rank: "01"
    title: "Cert-Manager with Cloudflare DNS-01 Challenge"
    excerpt: "This preset deploys cert-manager with a Cloudflare DNS provider for DNS-01 ACME certificate challenges via Let's Encrypt production. Cloudflare is the most common DNS provider for cert-manager due to..."
  - slug: "02-gcp-cloud-dns"
    rank: "02"
    title: "Cert-Manager with GCP Cloud DNS"
    excerpt: "This preset deploys cert-manager with a Google Cloud DNS provider for DNS-01 ACME certificate challenges. Authentication uses GKE Workload Identity, so no service account keys are stored in the..."
  - slug: "03-aws-route53"
    rank: "03"
    title: "Cert-Manager with AWS Route53"
    excerpt: "This preset deploys cert-manager with an AWS Route53 DNS provider for DNS-01 ACME certificate challenges. Authentication uses IAM Roles for Service Accounts (IRSA), so no AWS access keys are stored..."
  - slug: "04-azure-dns"
    rank: "04"
    title: "Cert-Manager with Azure DNS"
    excerpt: "This preset deploys cert-manager with an Azure DNS provider for DNS-01 ACME certificate challenges. Authentication uses Azure Managed Identity, so no client secrets are stored in the cluster."
---

# Cert Manager Presets

Ready-to-deploy configuration presets for Cert Manager. Each preset is a complete manifest you can copy, customize, and deploy.
