---
title: "Cert Manager"
description: "Cert Manager deployment documentation"
icon: "package"
order: 100
componentName: "kubernetescertmanager"
---

# Kubernetes Cert Manager

Deploys cert-manager on Kubernetes using the official Jetstack Helm chart (cert-manager v1.19.1) with support for Google Cloud DNS (GKE Workload Identity), AWS Route53 (IRSA), Azure DNS (Managed Identity), and Cloudflare as DNS-01 ACME challenge solvers, automatic ServiceAccount creation with workload-identity annotations, one ClusterIssuer per DNS zone for clear per-domain certificate management, optional namespace creation, and configurable cert-manager and Helm chart versions.

## What Gets Created

When you deploy a KubernetesCertManager resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **ServiceAccount** — a dedicated Kubernetes ServiceAccount annotated with workload-identity bindings for the selected DNS provider (GKE Workload Identity, EKS IRSA, AKS Workload Identity, or no annotation for Cloudflare)
- **Helm Release (cert-manager)** — deploys cert-manager from the `cert-manager` chart at `https://charts.jetstack.io`, pinned to version v1.19.1 by default, with CRDs installed, atomic rollback enabled, cleanup-on-fail, and a 3-minute timeout; DNS-01 recursive nameservers are set to `1.1.1.1:53` and `8.8.8.8:53` for reliable propagation checks
- **Cloudflare API Token Secret** — when using a Cloudflare provider, a Kubernetes Secret named `{name}-{providerName}-credentials` is created in the target namespace containing the API token
- **ClusterIssuer (per DNS zone)** — one ClusterIssuer is created for each DNS zone across all configured providers, named after the domain itself (e.g., `example.com`), with a Let's Encrypt ACME solver and a private key secret named `letsencrypt-{domain}-account-key`

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **An ACME account email** for Let's Encrypt registration and certificate expiry notifications
- **DNS provider access** — one or more of the following depending on your provider:
  - **GCP Cloud DNS**: a GCP project with Cloud DNS enabled and a Google Service Account with `dns.admin` role; Workload Identity must be configured on the GKE cluster
  - **AWS Route53**: an AWS Route53 hosted zone and an IAM Role ARN with permissions to modify Route53 records
  - **Azure DNS**: an Azure subscription with DNS zones and a Managed Identity with DNS Zone Contributor role
  - **Cloudflare**: an API token with `Zone:Zone:Read` and `Zone:DNS:Edit` permissions

## Quick Start

Create a file `cert-manager.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCertManager
metadata:
  name: my-cert-manager
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesCertManager.my-cert-manager
spec:
  namespace: cert-manager
  createNamespace: true
  acme:
    email: certs@example.com
  dnsProviders:
    - name: cloudflare-prod
      dnsZones:
        - example.com
      cloudflare:
        apiToken: cf-api-token-value
```

Deploy:

```shell
openmcf apply -f cert-manager.yaml
```

This creates a cert-manager instance in the `cert-manager` namespace with a ClusterIssuer for `example.com` using Cloudflare DNS-01 challenges and Let's Encrypt production as the ACME server.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the cert-manager deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `acme` | `object` | Global ACME configuration for certificate issuance. | Required |
| `acme.email` | `string` | ACME account email for registration and expiry notifications from the Certificate Authority. | Required |
| `dnsProviders` | `list` | List of DNS provider configurations. Each provider manages one or more DNS zones. The module creates a ClusterIssuer per zone. | At least 1 item |
| `dnsProviders[].name` | `string` | Unique identifier for the provider config. Used for generating Kubernetes Secret names. | Required |
| `dnsProviders[].dnsZones` | `list` | DNS zones this provider manages. cert-manager uses this provider's solver for matching certificate requests. | At least 1 item |
| `dnsProviders[].provider` | `oneof` | Exactly one of `gcpCloudDns`, `awsRoute53`, `azureDns`, or `cloudflare` must be set. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `kubernetesCertManagerVersion` | `string` | `v1.19.1` | cert-manager container image tag. Minimum version v1.16.4 is enforced for Cloudflare API compatibility. |
| `helmChartVersion` | `string` | `v1.19.1` | Helm chart version for the cert-manager chart from Jetstack. |
| `skipInstallSelfSignedIssuer` | `bool` | `false` | When `true`, skips installation of the default self-signed ClusterIssuer. |
| `acme.server` | `string` | `https://acme-v02.api.letsencrypt.org/directory` | ACME server URL. Use the staging URL (`https://acme-staging-v02.api.letsencrypt.org/directory`) for testing. |

**GCP Cloud DNS Provider** (`gcpCloudDns`):

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `gcpCloudDns.projectId` | `string` | GCP project ID that contains the Cloud DNS zones. | Required |
| `gcpCloudDns.serviceAccountEmail` | `string` | GCP Service Account email for Workload Identity. Must have the `dns.admin` role on the specified project. | Required |

**AWS Route53 Provider** (`awsRoute53`):

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `awsRoute53.region` | `string` | AWS region where Route53 is configured. | Required |
| `awsRoute53.roleArn` | `string` | IAM Role ARN for IRSA. Must have permissions to modify Route53 records in the specified zones. | Required |

**Azure DNS Provider** (`azureDns`):

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `azureDns.subscriptionId` | `string` | Azure subscription ID that contains the DNS zones. | Required |
| `azureDns.resourceGroup` | `string` | Azure resource group containing the DNS zones. | Required |
| `azureDns.clientId` | `string` | Managed Identity Client ID. Must have DNS Zone Contributor role on the specified resource group. | Required |

**Cloudflare Provider** (`cloudflare`):

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `cloudflare.apiToken` | `string` | Cloudflare API token for DNS-01 challenge authentication. Required permissions: `Zone:Zone:Read` and `Zone:DNS:Edit`. Stored as a Kubernetes Secret. | Required |

> **Note on `valueFrom`**: The `namespace` field is a `StringValueOrRef` type. You can provide a literal string value directly, or use `valueFrom` to reference the output of another OpenMCF resource. See the foreign key reference example below.

## Examples

### Cloudflare with Multiple Zones

Deploy cert-manager with a single Cloudflare provider managing two DNS zones:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCertManager
metadata:
  name: cf-cert-manager
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesCertManager.cf-cert-manager
spec:
  namespace: cert-manager
  createNamespace: true
  acme:
    email: certs@example.com
  dnsProviders:
    - name: cloudflare-prod
      dnsZones:
        - example.com
        - example.org
      cloudflare:
        apiToken: my-cloudflare-api-token
```

This creates two ClusterIssuers (`example.com` and `example.org`), each backed by a Cloudflare DNS-01 solver. A Kubernetes Secret named `cf-cert-manager-cloudflare-prod-credentials` is created to hold the API token.

### GCP Cloud DNS with Workload Identity

Deploy cert-manager on a GKE cluster to issue certificates for a Google Cloud DNS zone:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCertManager
metadata:
  name: gke-cert-manager
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesCertManager.gke-cert-manager
spec:
  namespace: cert-manager
  createNamespace: true
  kubernetesCertManagerVersion: "v1.19.1"
  acme:
    email: platform-team@mycompany.com
    server: https://acme-v02.api.letsencrypt.org/directory
  dnsProviders:
    - name: gcp-prod
      dnsZones:
        - mycompany.com
      gcpCloudDns:
        projectId: my-gcp-project
        serviceAccountEmail: cert-manager@my-gcp-project.iam.gserviceaccount.com
```

The module creates a ServiceAccount annotated with `iam.gke.io/gcp-service-account: cert-manager@my-gcp-project.iam.gserviceaccount.com` and a ClusterIssuer for `mycompany.com` using the Google Cloud DNS solver.

### AWS Route53 with IRSA

Deploy cert-manager on an EKS cluster with IAM Roles for Service Accounts:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCertManager
metadata:
  name: eks-cert-manager
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesCertManager.eks-cert-manager
spec:
  namespace: cert-manager
  createNamespace: true
  helmChartVersion: "v1.19.1"
  acme:
    email: devops@mycompany.com
  dnsProviders:
    - name: aws-prod
      dnsZones:
        - mycompany.io
      awsRoute53:
        region: us-east-1
        roleArn: arn:aws:iam::123456789012:role/cert-manager-route53
```

The ServiceAccount is annotated with `eks.amazonaws.com/role-arn` for IRSA authentication. A ClusterIssuer for `mycompany.io` is created using the Route53 solver in `us-east-1`.

### Multi-Provider with Staging ACME

Deploy cert-manager with multiple DNS providers and the Let's Encrypt staging server for testing:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCertManager
metadata:
  name: multi-cert-manager
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesCertManager.multi-cert-manager
spec:
  namespace: cert-manager
  createNamespace: true
  skipInstallSelfSignedIssuer: true
  acme:
    email: platform@mycompany.com
    server: https://acme-staging-v02.api.letsencrypt.org/directory
  dnsProviders:
    - name: cloudflare-public
      dnsZones:
        - public.example.com
      cloudflare:
        apiToken: cf-public-token
    - name: gcp-internal
      dnsZones:
        - internal.example.com
      gcpCloudDns:
        projectId: my-gcp-project
        serviceAccountEmail: cert-manager@my-gcp-project.iam.gserviceaccount.com
```

This creates two ClusterIssuers: `public.example.com` backed by Cloudflare and `internal.example.com` backed by Google Cloud DNS. Both use the Let's Encrypt staging server, which is useful for testing without hitting production rate limits.

### Using Foreign Key References

Reference OpenMCF-managed resources instead of hardcoding values:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCertManager
metadata:
  name: platform-cert-manager
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesCertManager.platform-cert-manager
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: cert-manager-ns
      field: spec.name
  createNamespace: false
  acme:
    email: certs@example.com
  dnsProviders:
    - name: cloudflare-prod
      dnsZones:
        - example.com
      cloudflare:
        apiToken: my-cloudflare-api-token
```

This example references an OpenMCF-managed namespace rather than embedding a literal value.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where cert-manager is deployed |
| `releaseName` | `string` | Helm release name for the cert-manager deployment |
| `clusterIssuerNames` | `list(string)` | Names of the ClusterIssuers created (one per DNS zone, named after the domain) |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — provides the target namespace via `valueFrom` reference
- [KubernetesExternalDns](/docs/catalog/kubernetes/external-dns) — manages DNS records for services and ingresses; pairs well with cert-manager for full DNS automation
- [KubernetesIngressNginx](/docs/catalog/kubernetes/ingress-nginx) — ingress controller that uses cert-manager ClusterIssuers to provision TLS certificates
- [KubernetesIstio](/docs/catalog/kubernetes/istio) — service mesh that can consume cert-manager certificates for mTLS and gateway TLS
- [KubernetesHelmRelease](/docs/catalog/kubernetes/helm-release) — alternative for deploying Helm charts with custom configurations
