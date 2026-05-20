# KubernetesCertManager

> Install and manage cert-manager on any Kubernetes cluster

## Overview

KubernetesCertManager installs the [cert-manager](https://cert-manager.io/) controller on a Kubernetes cluster via Helm. It handles the controller lifecycle -- CRDs, ServiceAccount, DNS resolver configuration, and optional workload identity for cloud DNS authentication.

**ClusterIssuer management is handled separately** by the **KubernetesClusterIssuer** component. This decoupling allows you to add, remove, or reconfigure certificate issuers independently from the cert-manager controller.

## Quick Start

### Basic Installation

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCertManager
metadata:
  name: cert-manager
spec:
  namespace:
    value: cert-manager
  createNamespace: true
```

### With GKE Workload Identity

Required when using KubernetesClusterIssuer with GCP Cloud DNS:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCertManager
metadata:
  name: cert-manager
spec:
  namespace:
    value: cert-manager
  createNamespace: true
  workloadIdentity:
    gke:
      serviceAccountEmail: "cert-manager@my-project.iam.gserviceaccount.com"
```

### Deploy

```bash
openmcf pulumi up --manifest cert-manager.yaml --stack org/project/env
```

## What Gets Deployed

- **Namespace** (optional) -- creates the target namespace if `createNamespace` is true
- **ServiceAccount** -- with optional workload identity annotations for cloud DNS authentication
- **Helm Release** -- cert-manager chart from Jetstack with CRDs, configured DNS resolvers

## Workload Identity

The `workload_identity` field configures the cert-manager controller's ServiceAccount to authenticate with cloud DNS APIs. This is required when using KubernetesClusterIssuer with cloud DNS providers (GCP Cloud DNS, AWS Route53, Azure DNS). Not needed for Cloudflare, which uses API token secrets.

| Cloud | Field | Annotation |
|-------|-------|------------|
| GKE | `workloadIdentity.gke.serviceAccountEmail` | `iam.gke.io/gcp-service-account` |
| EKS | `workloadIdentity.eks.roleArn` | `eks.amazonaws.com/role-arn` |
| AKS | `workloadIdentity.aks.clientId` | `azure.workload.identity/client-id` |

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `namespace` | StringValueOrRef | Yes | Kubernetes namespace for cert-manager |
| `createNamespace` | bool | No | Create the namespace if it doesn't exist |
| `kubernetesCertManagerVersion` | string | No | cert-manager image tag (default: v1.19.1) |
| `helmChartVersion` | string | No | Helm chart version (default: v1.19.1) |
| `skipInstallSelfSignedIssuer` | bool | No | Skip the default self-signed issuer |
| `workloadIdentity` | WorkloadIdentityConfig | No | Cloud identity for DNS authentication |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Namespace where cert-manager was deployed |
| `release_name` | Helm release name |
| `service_account_name` | Controller ServiceAccount name |

## Next Steps

After installing cert-manager, create ClusterIssuers using the **KubernetesClusterIssuer** component:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesClusterIssuer
metadata:
  name: example-com
spec:
  certManagerNamespace:
    value: cert-manager
  dnsDomain: example.com
  acme:
    email: admin@example.com
  cloudflare:
    apiToken: "<token>"
```

## Related Components

- **KubernetesClusterIssuer** -- creates ClusterIssuers for specific DNS domains
- **KubernetesIngressNginx** -- ingress controller that uses ClusterIssuers for TLS
- **KubernetesExternalDns** -- DNS record management for ingress hostnames
