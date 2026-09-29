# ClusterIssuer with GCP Cloud DNS

This preset creates a ClusterIssuer that uses Google Cloud DNS for ACME DNS-01 certificate challenges. Authentication uses GKE Workload Identity, so no service account keys are stored in the cluster. Requires KubernetesCertManager deployed with `workload_identity.gke` configured.

## When to Use

- Your DNS domains are hosted on Google Cloud DNS
- You run GKE clusters with Workload Identity enabled
- KubernetesCertManager is deployed with GKE Workload Identity configured

## Key Configuration Choices

- **GCP Workload Identity** -- cert-manager's ServiceAccount is bound to a GCP service account; no JSON key files needed
- **ACME server** -- defaults to Let's Encrypt production; switch to staging URL for testing

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-email@example.com>` | Email for Let's Encrypt registration and expiry notifications | Your organization's ops email |
| `<gcp-project-id>` | GCP project containing the DNS zone | GCP Console > Project Settings |

## Related Presets

- **01-cloudflare** -- Use when DNS domains are managed by Cloudflare
- **03-aws-route53** -- Use when DNS domains are hosted on AWS Route53
