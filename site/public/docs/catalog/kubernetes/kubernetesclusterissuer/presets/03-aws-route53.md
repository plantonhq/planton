---
title: "ClusterIssuer with AWS Route53"
description: "This preset creates a ClusterIssuer that uses AWS Route53 for ACME DNS-01 certificate challenges. Authentication uses IAM Roles for Service Accounts (IRSA), so no AWS access keys are stored in the..."
type: "preset"
rank: "03"
presetSlug: "03-aws-route53"
componentSlug: "kubernetesclusterissuer"
componentTitle: "KubernetesClusterIssuer"
provider: "kubernetes"
icon: "package"
order: 3
---

# ClusterIssuer with AWS Route53

This preset creates a ClusterIssuer that uses AWS Route53 for ACME DNS-01 certificate challenges. Authentication uses IAM Roles for Service Accounts (IRSA), so no AWS access keys are stored in the cluster. Requires KubernetesCertManager deployed with `workload_identity.eks` configured.

## When to Use

- Your DNS domains are hosted on AWS Route53
- You run EKS clusters with IRSA enabled
- KubernetesCertManager is deployed with EKS IRSA configured

## Key Configuration Choices

- **IRSA authentication** -- cert-manager's ServiceAccount assumes an IAM role via OIDC; no long-lived credentials
- **ACME server** -- defaults to Let's Encrypt production; switch to staging URL for testing

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-domain.com>` | DNS domain managed by AWS Route53 | AWS Console > Route53 > Hosted Zones |
| `<your-acme-email@example.com>` | Email for Let's Encrypt registration | Your organization's ops email |
| `<your-aws-region>` | AWS region where Route53 is configured (e.g., `us-east-1`) | AWS Console > Route53 |

## Related Presets

- **01-cloudflare** -- Use when DNS domains are managed by Cloudflare
- **02-gcp-cloud-dns** -- Use when DNS domains are hosted on Google Cloud DNS
