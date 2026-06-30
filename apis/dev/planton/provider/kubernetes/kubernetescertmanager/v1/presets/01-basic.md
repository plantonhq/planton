# Basic Cert-Manager Installation

This preset installs cert-manager with default settings. No workload identity is configured -- suitable for clusters where ClusterIssuers will use Cloudflare (API token secrets) or where workload identity will be configured later.

## When to Use

- You only need Cloudflare-based ClusterIssuers (no workload identity required)
- You want to install cert-manager first and configure workload identity later
- You are running on a cluster without cloud-provider workload identity support

## Key Configuration Choices

- **No workload identity** -- cert-manager SA has no cloud identity annotations
- **Default versions** -- uses the default cert-manager and Helm chart versions

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.

## Related Presets

- **02-gke-workload-identity** -- Use when you need GCP Cloud DNS authentication
- **03-eks-irsa** -- Use when you need AWS Route53 authentication
