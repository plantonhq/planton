# KubernetesCertManager

Installs the cert-manager controller on a Kubernetes cluster for automated TLS certificate management. Handles Helm deployment, CRDs, and optional workload identity configuration.

## What Gets Created

- **Namespace** (optional) -- target namespace for cert-manager
- **ServiceAccount** -- with optional workload identity annotations
- **Helm Release** -- cert-manager chart with CRDs and DNS resolver configuration

## Prerequisites

- A Kubernetes cluster (GKE, EKS, AKS, or any conformant cluster)

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesCertManager
metadata:
  name: my-cert-manager
spec:
  namespace:
    value: cert-manager
  createNamespace: true
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Namespace where cert-manager was deployed |
| `release_name` | Helm release name |
| `service_account_name` | Controller ServiceAccount name |

## Related Components

- **KubernetesClusterIssuer** -- creates ClusterIssuers (deploy after cert-manager)
- **KubernetesIngressNginx** -- ingress controller that uses ClusterIssuers for TLS
