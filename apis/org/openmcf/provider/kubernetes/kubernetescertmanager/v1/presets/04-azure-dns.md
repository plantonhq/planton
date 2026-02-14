# Cert-Manager with Azure DNS

This preset deploys cert-manager with an Azure DNS provider for DNS-01 ACME certificate challenges. Authentication uses Azure Managed Identity, so no client secrets are stored in the cluster.

## When to Use

- Your DNS zones are hosted on Azure DNS
- You run AKS clusters with Managed Identity enabled
- You need automated TLS certificates from Let's Encrypt

## Key Configuration Choices

- **Managed Identity** -- cert-manager authenticates via a user-assigned managed identity with DNS Zone Contributor role; no client secrets needed
- **ACME server** (Let's Encrypt production) -- issues trusted certificates; switch to staging URL for testing
- **DNS-01 challenge** -- proves domain ownership via Azure DNS TXT records

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-acme-email@example.com>` | Email for Let's Encrypt registration and certificate expiry notifications | Your organization's ops email |
| `<your-domain.com>` | DNS zone managed by Azure DNS | Azure Portal > DNS Zones |
| `<your-azure-subscription-id>` | Azure subscription containing the DNS zone | Azure Portal > Subscriptions |
| `<your-dns-resource-group>` | Resource group containing the DNS zone | Azure Portal > DNS Zones > Resource Group |
| `<your-managed-identity-client-id>` | Client ID of the managed identity with DNS Zone Contributor role | Azure Portal > Managed Identities |

## Related Presets

- **01-cloudflare** -- Use when DNS zones are managed by Cloudflare
- **02-gcp-cloud-dns** -- Use when DNS zones are hosted on Google Cloud DNS
- **03-aws-route53** -- Use when DNS zones are hosted on AWS Route53
