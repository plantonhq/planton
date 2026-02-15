---
title: "Cert-Manager with GCP Cloud DNS"
description: "This preset deploys cert-manager with a Google Cloud DNS provider for DNS-01 ACME certificate challenges. Authentication uses GKE Workload Identity, so no service account keys are stored in the..."
type: "preset"
rank: "02"
presetSlug: "02-gcp-cloud-dns"
componentSlug: "cert-manager"
componentTitle: "Cert Manager"
provider: "kubernetes"
icon: "package"
order: 2
---

# Cert-Manager with GCP Cloud DNS

This preset deploys cert-manager with a Google Cloud DNS provider for DNS-01 ACME certificate challenges. Authentication uses GKE Workload Identity, so no service account keys are stored in the cluster.

## When to Use

- Your DNS zones are hosted on Google Cloud DNS
- You run GKE clusters with Workload Identity enabled
- You need automated TLS certificates from Let's Encrypt

## Key Configuration Choices

- **GCP Workload Identity** -- cert-manager's Kubernetes ServiceAccount is bound to a GCP service account; no JSON key files needed
- **ACME server** (Let's Encrypt production) -- issues trusted certificates; switch to staging URL for testing
- **DNS-01 challenge** -- proves domain ownership via Cloud DNS TXT records

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-acme-email@example.com>` | Email for Let's Encrypt registration and certificate expiry notifications | Your organization's ops email |
| `<your-domain.com>` | DNS zone managed by Google Cloud DNS | GCP Console > Cloud DNS |
| `<your-gcp-project-id>` | GCP project containing the DNS zone | GCP Console > Project Settings |
| `<your-gsa-email@your-project.iam.gserviceaccount.com>` | GCP service account with `dns.admin` role on the project | GCP Console > IAM & Admin > Service Accounts |

## Related Presets

- **01-cloudflare** -- Use when DNS zones are managed by Cloudflare
- **03-aws-route53** -- Use when DNS zones are hosted on AWS Route53
- **04-azure-dns** -- Use when DNS zones are hosted on Azure DNS
