# Cert-Manager with AWS Route53

This preset deploys cert-manager with an AWS Route53 DNS provider for DNS-01 ACME certificate challenges. Authentication uses IAM Roles for Service Accounts (IRSA), so no AWS access keys are stored in the cluster.

## When to Use

- Your DNS zones are hosted on AWS Route53
- You run EKS clusters with IRSA enabled
- You need automated TLS certificates from Let's Encrypt

## Key Configuration Choices

- **IRSA authentication** -- cert-manager's Kubernetes ServiceAccount assumes an IAM role via OIDC; no long-lived AWS credentials needed
- **ACME server** (Let's Encrypt production) -- issues trusted certificates; switch to staging URL for testing
- **DNS-01 challenge** -- proves domain ownership via Route53 TXT records

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-acme-email@example.com>` | Email for Let's Encrypt registration and certificate expiry notifications | Your organization's ops email |
| `<your-domain.com>` | DNS zone managed by AWS Route53 | AWS Console > Route53 > Hosted Zones |
| `<your-aws-region>` | AWS region where Route53 is configured (e.g., `us-east-1`) | AWS Console > Route53 |
| `<your-irsa-role-arn>` | IAM role ARN with Route53 permissions for IRSA | AWS Console > IAM > Roles |

## Related Presets

- **01-cloudflare** -- Use when DNS zones are managed by Cloudflare
- **02-gcp-cloud-dns** -- Use when DNS zones are hosted on Google Cloud DNS
- **04-azure-dns** -- Use when DNS zones are hosted on Azure DNS
