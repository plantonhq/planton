---
title: "Cert-Manager with EKS IRSA"
description: "This preset installs cert-manager with IAM Roles for Service Accounts (IRSA) configured on the controller ServiceAccount. Required when using KubernetesClusterIssuer with the AWS Route53 provider."
type: "preset"
rank: "03"
presetSlug: "03-eks-irsa"
componentSlug: "kubernetescertmanager"
componentTitle: "KubernetesCertManager"
provider: "kubernetes"
icon: "package"
order: 3
---

# Cert-Manager with EKS IRSA

This preset installs cert-manager with IAM Roles for Service Accounts (IRSA) configured on the controller ServiceAccount. Required when using KubernetesClusterIssuer with the AWS Route53 provider.

## When to Use

- You run EKS clusters with IRSA enabled
- You will create KubernetesClusterIssuer resources using AWS Route53

## Key Configuration Choices

- **EKS IRSA** (`workloadIdentity.eks`) -- binds the cert-manager ServiceAccount to an IAM role via OIDC for keyless authentication to Route53
- **IAM Role** -- must have permissions to modify Route53 DNS records

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-irsa-role-arn>` | IAM role ARN with Route53 permissions | AWS Console > IAM > Roles |

## Related Presets

- **01-basic** -- Use when no workload identity is needed (Cloudflare-only)
- **02-gke-workload-identity** -- Use when running on GKE with Cloud DNS
