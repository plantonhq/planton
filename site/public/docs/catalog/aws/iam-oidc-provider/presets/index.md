---
title: "Presets"
description: "Ready-to-deploy configuration presets for IAM OIDC Provider"
type: "preset-list"
componentSlug: "iam-oidc-provider"
componentTitle: "IAM OIDC Provider"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-eks-irsa"
    rank: "01"
    title: "EKS IRSA OIDC Provider"
    excerpt: "This preset registers an EKS cluster's OIDC issuer as a trusted IAM identity provider, enabling IAM Roles for Service Accounts (IRSA). It references the cluster directly, so the issuer URL is..."
  - slug: "02-github-actions"
    rank: "02"
    title: "GitHub Actions OIDC Provider"
    excerpt: "This preset registers GitHub Actions as a trusted OIDC identity provider, enabling keyless deployments from CI. Workflows assume an IAM deploy role with a short-lived token on every run, removing..."
  - slug: "03-generic-issuer"
    rank: "03"
    title: "Generic OIDC Provider with Explicit Thumbprint"
    excerpt: "This preset registers any standards-compliant OIDC issuer as a trusted identity provider, with an explicit root-CA thumbprint. Use it for self-hosted or partner issuers whose certificate authority is..."
---

# IAM OIDC Provider Presets

Ready-to-deploy configuration presets for IAM OIDC Provider. Each preset is a complete manifest you can copy, customize, and deploy.
