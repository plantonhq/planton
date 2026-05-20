# Cert-Manager with GKE Workload Identity

This preset installs cert-manager with GKE Workload Identity configured on the controller ServiceAccount. Required when using KubernetesClusterIssuer with the GCP Cloud DNS provider.

## When to Use

- You run GKE clusters with Workload Identity enabled
- You will create KubernetesClusterIssuer resources using GCP Cloud DNS

## Key Configuration Choices

- **GKE Workload Identity** (`workloadIdentity.gke`) -- binds the cert-manager ServiceAccount to a GCP service account for keyless authentication to Cloud DNS
- **GCP Service Account** -- must have `dns.admin` role on the project containing your DNS zones

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-gsa-email@your-project.iam.gserviceaccount.com>` | GCP service account with dns.admin role | GCP Console > IAM & Admin > Service Accounts |

## Related Presets

- **01-basic** -- Use when no workload identity is needed (Cloudflare-only)
- **03-eks-irsa** -- Use when running on EKS with AWS Route53
