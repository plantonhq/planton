---
title: "Presets"
description: "Ready-to-deploy configuration presets for External Secrets"
type: "preset-list"
componentSlug: "external-secrets"
componentTitle: "External Secrets"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-gke-secrets-manager"
    rank: "01"
    title: "External Secrets on GKE with Google Cloud Secret Manager"
    excerpt: "This preset deploys the External Secrets Operator (ESO) on a GKE cluster with Google Cloud Secret Manager as the backing store. Authentication uses GKE Workload Identity via a Google service account."
  - slug: "02-eks-secrets-manager"
    rank: "02"
    title: "External Secrets on EKS with AWS Secrets Manager"
    excerpt: "This preset deploys the External Secrets Operator (ESO) on an EKS cluster with AWS Secrets Manager as the backing store. Authentication uses IRSA (IAM Roles for Service Accounts); an IAM role is..."
  - slug: "03-aks-key-vault"
    rank: "03"
    title: "External Secrets on AKS with Azure Key Vault"
    excerpt: "This preset deploys the External Secrets Operator (ESO) on an AKS cluster with Azure Key Vault as the backing store. Authentication uses a user-assigned managed identity bound to the ESO service..."
---

# External Secrets Presets

Ready-to-deploy configuration presets for External Secrets. Each preset is a complete manifest you can copy, customize, and deploy.
