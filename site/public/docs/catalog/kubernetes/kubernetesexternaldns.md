---
title: "External DNS"
description: "External DNS deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesexternaldns"
---

# Kubernetes External DNS

Deploys ExternalDNS on Kubernetes using the official Helm chart (external-dns v1.19.0) from kubernetes-sigs, with support for Google Cloud DNS (GKE), AWS Route53 (EKS), Azure DNS (AKS), and Cloudflare as DNS providers, automatic ServiceAccount creation with workload-identity annotations, optional namespace creation, and configurable ExternalDNS and Helm chart versions.

## What Gets Created

When you deploy a KubernetesExternalDns resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **ServiceAccount** — a dedicated Kubernetes ServiceAccount annotated with workload-identity bindings for the selected DNS provider (GKE Workload Identity, EKS IRSA, AKS Workload Identity, or no annotation for Cloudflare)
- **Helm Release (ExternalDNS)** — deploys ExternalDNS from the `external-dns` chart at `https://kubernetes-sigs.github.io/external-dns/`, pinned to version 1.19.0 by default, with atomic rollback enabled, cleanup-on-fail, and a 3-minute timeout
- **Cloudflare API Token Secret** — when using the Cloudflare provider, a Kubernetes Secret named `{name}-cloudflare-api-token` is created in the target namespace and mounted as the `CF_API_TOKEN` environment variable

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **DNS provider access** — one of the following depending on your provider:
  - **GKE**: a GCP project with Cloud DNS enabled and a Google Service Account with `dns.admin` role; Workload Identity must be configured on the GKE cluster
  - **EKS**: an AWS Route53 hosted zone and either an existing IRSA role ARN or IAM permissions for auto-creation
  - **AKS**: an Azure DNS zone and optionally a Managed Identity client ID for workload identity
  - **Cloudflare**: an API token with `Zone:Zone:Read` and `Zone:DNS:Edit` permissions

## Quick Start

Create a file `external-dns.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesExternalDns
metadata:
  name: my-external-dns
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesExternalDns.my-external-dns
spec:
  namespace: external-dns
  createNamespace: true
  cloudflare:
    apiToken: cf-api-token-value
    dnsZoneId: zone-id-value
```

Deploy:

```shell
openmcf apply -f external-dns.yaml
```

This creates an ExternalDNS instance in the `external-dns` namespace configured to manage DNS records in Cloudflare, using ExternalDNS v0.19.0 and Helm chart version 1.19.0.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the ExternalDNS deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| Provider config (one of) | `object` | Exactly one of `gke`, `eks`, `aks`, or `cloudflare` must be set. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `externalDnsVersion` | `string` | `v0.19.0` | ExternalDNS container image tag. |
| `helmChartVersion` | `string` | `1.19.0` | Helm chart version for the external-dns chart. |

**GKE Provider** (`gke`):

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `gke.projectId` | `string` | GCP project hosting the DNS zone and GKE cluster. Can reference a GcpProject resource via `valueFrom`. | Required |
| `gke.dnsZoneId` | `string` | GCP Cloud DNS zone ID for ExternalDNS to manage. Can reference a GcpDnsZone resource via `valueFrom`. | Required |

**EKS Provider** (`eks`):

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `eks.route53ZoneId` | `string` | AWS Route53 hosted zone ID for ExternalDNS to manage. Can reference an AwsRoute53Zone resource via `valueFrom`. | Required |
| `eks.irsaRoleArnOverride` | `string` | Existing IAM role ARN for IRSA. If blank, the role is expected to be auto-created externally. | Optional |

**AKS Provider** (`aks`):

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `aks.dnsZoneId` | `string` | Azure DNS zone ID for ExternalDNS to manage. Can reference an AzureDnsZone resource via `valueFrom`. | Required |
| `aks.managedIdentityClientId` | `string` | Azure Managed Identity client ID for workload identity authentication. | Optional |

**Cloudflare Provider** (`cloudflare`):

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `cloudflare.apiToken` | `string` | Cloudflare API token with `Zone:Zone:Read` and `Zone:DNS:Edit` permissions. Stored as a Kubernetes Secret. | Required |
| `cloudflare.dnsZoneId` | `string` | Cloudflare DNS zone ID to manage. Can reference a CloudflareDnsZone resource via `valueFrom`. | Required |
| `cloudflare.isProxied` | `bool` | Enable Cloudflare proxy (orange cloud) for managed DNS records, routing traffic through Cloudflare's edge network for DDoS protection, WAF, and CDN. Default: `false`. | Optional |

> **Note on `valueFrom`**: Fields marked "Can reference ... via `valueFrom`" are `StringValueOrRef` types. You can provide a literal string value directly, or use `valueFrom` to reference the output of another OpenMCF resource. See the foreign key reference example below.

## Examples

### GKE with Google Cloud DNS

Deploy ExternalDNS on a GKE cluster to manage records in a Google Cloud DNS zone:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesExternalDns
metadata:
  name: gke-external-dns
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesExternalDns.gke-external-dns
spec:
  namespace: external-dns
  createNamespace: true
  gke:
    projectId: my-gcp-project
    dnsZoneId: my-dns-zone
```

The module creates a ServiceAccount annotated with `iam.gke.io/gcp-service-account: gke-external-dns@my-gcp-project.iam.gserviceaccount.com` and configures ExternalDNS to use the `google` provider scoped to the specified zone.

### EKS with AWS Route53

Deploy ExternalDNS on an EKS cluster with an existing IRSA role:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesExternalDns
metadata:
  name: eks-external-dns
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesExternalDns.eks-external-dns
spec:
  namespace: external-dns
  createNamespace: true
  externalDnsVersion: "v0.15.1"
  helmChartVersion: "1.15.0"
  eks:
    route53ZoneId: Z0123456789ABCDEF
    irsaRoleArnOverride: arn:aws:iam::123456789012:role/external-dns-irsa
```

This pins ExternalDNS to v0.15.1 and the Helm chart to 1.15.0. The ServiceAccount is annotated with `eks.amazonaws.com/role-arn` for IRSA authentication.

### Cloudflare with Proxy Enabled

Deploy ExternalDNS to manage Cloudflare DNS records with the proxy (orange cloud) enabled:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesExternalDns
metadata:
  name: cf-external-dns
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesExternalDns.cf-external-dns
spec:
  namespace: external-dns
  createNamespace: true
  cloudflare:
    apiToken: my-cloudflare-api-token
    dnsZoneId: abc123def456
    isProxied: true
```

The module creates a Kubernetes Secret for the API token, configures ExternalDNS to watch Services, Ingress, Gateway API HTTPRoutes, and Istio Gateways, and enables the `--cloudflare-proxied` flag so all managed records are proxied through Cloudflare's edge network.

### Using Foreign Key References

Reference OpenMCF-managed resources instead of hardcoding values:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesExternalDns
metadata:
  name: platform-external-dns
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesExternalDns.platform-external-dns
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: dns-namespace
      field: spec.name
  createNamespace: false
  gke:
    projectId:
      valueFrom:
        kind: GcpProject
        name: platform-project
        field: status.outputs.project_id
    dnsZoneId:
      valueFrom:
        kind: GcpDnsZone
        name: platform-zone
        field: status.outputs.zone_id
```

This example references an OpenMCF-managed namespace, GCP project, and DNS zone rather than embedding literal values.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where ExternalDNS is deployed |
| `releaseName` | `string` | Helm release name for the ExternalDNS deployment |
| `solverSa` | `string` | Kubernetes ServiceAccount name used by ExternalDNS |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesHelmRelease](/docs/catalog/kubernetes/kuberneteshelmrelease) — alternative for deploying Helm charts with custom configurations
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — application deployments that ExternalDNS can create DNS records for
- [KubernetesService](/docs/catalog/kubernetes/kubernetesservice) — services annotated with ExternalDNS hostname annotations to trigger DNS record creation
