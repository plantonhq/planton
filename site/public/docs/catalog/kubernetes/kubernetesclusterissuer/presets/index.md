---
title: "Presets"
description: "Ready-to-deploy configuration presets for KubernetesClusterIssuer"
type: "preset-list"
componentSlug: "kubernetesclusterissuer"
componentTitle: "KubernetesClusterIssuer"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-cloudflare"
    rank: "01"
    title: "ClusterIssuer with Cloudflare DNS-01 Challenge"
    excerpt: "This preset creates a ClusterIssuer that uses Cloudflare DNS for ACME DNS-01 certificate challenges via Let's Encrypt production. Cloudflare is the most common DNS provider for cert-manager due to..."
  - slug: "02-gcp-cloud-dns"
    rank: "02"
    title: "ClusterIssuer with GCP Cloud DNS"
    excerpt: "This preset creates a ClusterIssuer that uses Google Cloud DNS for ACME DNS-01 certificate challenges. Authentication uses GKE Workload Identity, so no service account keys are stored in the cluster...."
  - slug: "03-aws-route53"
    rank: "03"
    title: "ClusterIssuer with AWS Route53"
    excerpt: "This preset creates a ClusterIssuer that uses AWS Route53 for ACME DNS-01 certificate challenges. Authentication uses IAM Roles for Service Accounts (IRSA), so no AWS access keys are stored in the..."
---

# KubernetesClusterIssuer Presets

Ready-to-deploy configuration presets for KubernetesClusterIssuer. Each preset is a complete manifest you can copy, customize, and deploy.
