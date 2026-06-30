# KubernetesClusterIssuer

> Declarative ACME ClusterIssuer management for cert-manager on any Kubernetes cluster

## Overview

KubernetesClusterIssuer creates a cert-manager [ClusterIssuer](https://cert-manager.io/docs/concepts/issuer/) for a single DNS domain using ACME DNS-01 challenges. Each instance manages one ClusterIssuer, keeping certificate authority configuration independent from the cert-manager controller installation.

This component is designed to work alongside **KubernetesCertManager**, which installs the cert-manager controller and optionally configures workload identity for cloud DNS authentication. While KubernetesCertManager handles the controller lifecycle, KubernetesClusterIssuer handles the issuer lifecycle -- allowing you to add, remove, or reconfigure issuers without touching the controller.

## Prerequisites

- **cert-manager must be installed** on the target cluster (via KubernetesCertManager or manually)
- For GCP Cloud DNS, AWS Route53, or Azure DNS providers: the cert-manager ServiceAccount must be configured with the appropriate workload identity (done via KubernetesCertManager's `workload_identity` config)
- For Cloudflare: no workload identity needed -- this component creates the API token Secret directly

## Quick Start

### Cloudflare DNS

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesClusterIssuer
metadata:
  name: example-com-issuer
spec:
  certManagerNamespace:
    value: cert-manager
  dnsDomain: example.com
  acme:
    email: admin@example.com
  cloudflare:
    apiToken: "<your-cloudflare-api-token>"
```

### GCP Cloud DNS

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesClusterIssuer
metadata:
  name: example-com-issuer
spec:
  certManagerNamespace:
    value: cert-manager
  dnsDomain: example.com
  acme:
    email: admin@example.com
  gcpCloudDns:
    projectId: my-gcp-project
```

### Deploy

```bash
planton pulumi up --manifest cluster-issuer.yaml --stack org/project/env
```

## How It Works

1. **Cloudflare provider**: Creates a Kubernetes Secret in the cert-manager namespace containing the API token, then creates a ClusterIssuer referencing that secret
2. **GCP/AWS/Azure providers**: Creates only the ClusterIssuer CR -- authentication is handled by workload identity on the cert-manager ServiceAccount (configured via KubernetesCertManager)
3. The ClusterIssuer is named after the `dns_domain` value (e.g., `example.com`), matching the convention all Planton ingress components use to derive issuer names from hostnames

## Naming Convention

The ClusterIssuer Kubernetes resource is named after the `dns_domain` field. This is critical because all Planton ingress-enabled components (KubernetesDeployment, KubernetesArgocd, KubernetesKafka, etc.) derive the issuer name from the ingress hostname by stripping the first label:

```
Ingress hostname: argocd.example.com
Derived issuer:   example.com  (matches dns_domain)
```

## Multiple Domains

Create one KubernetesClusterIssuer per domain:

```yaml
# example.com
apiVersion: kubernetes.planton.dev/v1
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
---
# internal.example.net
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesClusterIssuer
metadata:
  name: internal-example-net
spec:
  certManagerNamespace:
    value: cert-manager
  dnsDomain: internal.example.net
  acme:
    email: admin@example.com
  gcpCloudDns:
    projectId: my-project
```

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `certManagerNamespace` | StringValueOrRef | Yes | Namespace where cert-manager is installed |
| `dnsDomain` | string | Yes | DNS domain for the ClusterIssuer (becomes the k8s resource name) |
| `acme.email` | string | Yes | ACME registration email |
| `acme.server` | string | No | ACME server URL (default: Let's Encrypt production) |
| `cloudflare.apiToken` | string | If Cloudflare | Cloudflare API token |
| `gcpCloudDns.projectId` | string | If GCP | GCP project containing Cloud DNS zone |
| `awsRoute53.region` | string | If AWS | AWS region for Route53 |
| `azureDns.subscriptionId` | string | If Azure | Azure subscription ID |
| `azureDns.resourceGroup` | string | If Azure | Resource group containing DNS zone |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `cluster_issuer_name` | Name of the created ClusterIssuer (equals `dns_domain`) |
| `acme_account_key_secret_name` | Name of the ACME account key Secret |

## Related Components

- **KubernetesCertManager** -- installs the cert-manager controller (prerequisite)
- **KubernetesIngressNginx** -- ingress controller that uses ClusterIssuers for TLS
- **KubernetesExternalDns** -- DNS record management for ingress hostnames
