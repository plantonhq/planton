# KubernetesClusterIssuer

Creates a cert-manager ClusterIssuer for automated ACME TLS certificate issuance via DNS-01 challenges. Each instance manages one ClusterIssuer for one DNS domain.

## What Gets Created

- **ClusterIssuer** -- cert-manager ClusterIssuer CR named after the DNS domain
- **Cloudflare Secret** (Cloudflare only) -- Kubernetes Secret containing the API token in the cert-manager namespace

## Prerequisites

- cert-manager installed on the cluster (via KubernetesCertManager)
- For GCP/AWS/Azure: workload identity configured on the cert-manager ServiceAccount

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesClusterIssuer
metadata:
  name: my-cluster-issuer
spec:
  certManagerNamespace:
    value: cert-manager
  dnsDomain: example.com
  acme:
    email: admin@example.com
  cloudflare:
    apiToken: "<your-cloudflare-api-token>"
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `cluster_issuer_name` | Name of the ClusterIssuer (equals `dns_domain`) |
| `acme_account_key_secret_name` | ACME account private key Secret name |

## Related Components

- **KubernetesCertManager** -- installs the cert-manager controller
- **KubernetesIngressNginx** -- ingress controller that uses ClusterIssuers for TLS
