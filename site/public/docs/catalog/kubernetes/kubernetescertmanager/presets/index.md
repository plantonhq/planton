---
title: "Presets"
description: "Ready-to-deploy configuration presets for KubernetesCertManager"
type: "preset-list"
componentSlug: "kubernetescertmanager"
componentTitle: "KubernetesCertManager"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-basic"
    rank: "01"
    title: "Basic Cert-Manager Installation"
    excerpt: "This preset installs cert-manager with default settings. No workload identity is configured -- suitable for clusters where ClusterIssuers will use Cloudflare (API token secrets) or where workload..."
  - slug: "02-gke-workload-identity"
    rank: "02"
    title: "Cert-Manager with GKE Workload Identity"
    excerpt: "This preset installs cert-manager with GKE Workload Identity configured on the controller ServiceAccount. Required when using KubernetesClusterIssuer with the GCP Cloud DNS provider."
  - slug: "03-eks-irsa"
    rank: "03"
    title: "Cert-Manager with EKS IRSA"
    excerpt: "This preset installs cert-manager with IAM Roles for Service Accounts (IRSA) configured on the controller ServiceAccount. Required when using KubernetesClusterIssuer with the AWS Route53 provider."
---

# KubernetesCertManager Presets

Ready-to-deploy configuration presets for KubernetesCertManager. Each preset is a complete manifest you can copy, customize, and deploy.
