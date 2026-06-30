---
title: "External Secrets"
description: "External Secrets deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesexternalsecrets"
---

# Kubernetes External Secrets

Deploys the [External Secrets Operator](https://external-secrets.io/) (ESO) onto any Kubernetes cluster using its official Helm chart. ESO synchronizes secrets from cloud provider secret stores (Google Cloud Secret Manager, AWS Secrets Manager, Azure Key Vault) into native Kubernetes Secrets, keeping application pods decoupled from provider-specific APIs. The module creates a dedicated ServiceAccount wired to the correct cloud identity mechanism (GKE Workload Identity, EKS IRSA, or AKS Managed Identity), installs CRDs, and configures RBAC automatically.

## What Gets Created

When you deploy a KubernetesExternalSecrets resource, Planton provisions:

- **Kubernetes Namespace** — created if `createNamespace` is `true`
- **ServiceAccount** — a Kubernetes ServiceAccount annotated with the appropriate cloud identity binding (`iam.gke.io/gcp-service-account` for GKE, `eks.amazonaws.com/role-arn` for EKS, or `azure.workload.identity/client-id` for AKS)
- **External Secrets Helm Release** — the `external-secrets` chart (v0.9.20) from `https://charts.external-secrets.io`, which creates:
  - The ESO controller Deployment that watches `ExternalSecret` custom resources
  - Custom Resource Definitions (`ExternalSecret`, `SecretStore`, `ClusterSecretStore`, and others)
  - ClusterRole and ClusterRoleBinding for RBAC
  - Webhook components for validation and conversion

## Prerequisites

- **A Kubernetes cluster** with kubectl configured for access
- **A cloud secret store** with secrets you want to synchronize (Google Cloud Secret Manager, AWS Secrets Manager, or Azure Key Vault)
- **Cloud identity binding** already configured:
  - **GKE**: A Google Service Account with `roles/secretmanager.secretAccessor` and Workload Identity binding to the Kubernetes ServiceAccount
  - **EKS**: An IAM role with `secretsmanager:GetSecretValue` permission and an IRSA trust policy
  - **AKS**: A User-Assigned Managed Identity with `Key Vault Secrets User` role on the target Key Vault

## Quick Start

Create a file `external-secrets.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesExternalSecrets
metadata:
  name: eso
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesExternalSecrets.eso
spec:
  namespace:
    value: external-secrets
  createNamespace: true
  container:
    resources:
      requests:
        cpu: "50m"
        memory: "100Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"
  gke:
    projectId:
      value: my-gcp-project
    gsaEmail: eso-sa@my-gcp-project.iam.gserviceaccount.com
```

Deploy:

```shell
planton apply -f external-secrets.yaml
```

This installs the External Secrets Operator into the `external-secrets` namespace with GKE Workload Identity configured. The controller begins polling Google Cloud Secret Manager every 10 seconds by default.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the ESO deployment. Use `value` for a literal string or `valueFrom` to reference a KubernetesNamespace resource. | Required |
| `container` | `object` | Container resource configuration for the ESO controller. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `createNamespace` | `bool` | `false` | Create the namespace if it does not exist. |
| `pollIntervalSeconds` | `uint32` | `10` | How often the controller polls the backing secret store, in seconds. Must be greater than 0. Very small values can incur high cloud API costs. |
| `container.resources.requests.cpu` | `string` | `"50m"` | CPU request for the ESO controller container. |
| `container.resources.requests.memory` | `string` | `"100Mi"` | Memory request for the ESO controller container. |
| `container.resources.limits.cpu` | `string` | `"1000m"` | CPU limit for the ESO controller container. |
| `container.resources.limits.memory` | `string` | `"1Gi"` | Memory limit for the ESO controller container. |

**Provider Configuration** — exactly one of the following blocks must be set:

| Field | Type | Description |
|-------|------|-------------|
| `gke.projectId` | `StringValueOrRef` | GCP project hosting the secrets and the GKE cluster. Use `value` for a literal string or `valueFrom` to reference a GcpProject resource. Required when using `gke`. |
| `gke.gsaEmail` | `string` | Google Service Account email used via Workload Identity. Required when using `gke`. |
| `eks.region` | `string` | AWS region containing the secret store. Defaults to the cluster region if empty. |
| `eks.irsaRoleArnOverride` | `string` | Existing IAM role ARN for IRSA. If left blank, one is auto-created. |
| `aks.keyVaultResourceId` | `string` | Azure Key Vault resource ID that stores the secrets. |
| `aks.managedIdentityClientId` | `string` | Client ID of an existing User-Assigned Managed Identity to bind to ESO. |

> **Note on `StringValueOrRef` fields:** Fields typed as `StringValueOrRef` accept either a direct `value` string or a `valueFrom` block that references the output of another Planton resource.

## Examples

### GKE with Google Cloud Secret Manager

Deploy ESO on a GKE cluster with Workload Identity for Google Cloud Secret Manager access:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesExternalSecrets
metadata:
  name: eso-gke
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesExternalSecrets.eso-gke
spec:
  namespace:
    value: external-secrets
  createNamespace: true
  pollIntervalSeconds: 30
  container:
    resources:
      requests:
        cpu: "100m"
        memory: "128Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
  gke:
    projectId:
      value: my-gcp-project
    gsaEmail: eso-secrets@my-gcp-project.iam.gserviceaccount.com
```

### EKS with AWS Secrets Manager

Deploy ESO on an EKS cluster with IRSA for AWS Secrets Manager access. The `irsaRoleArnOverride` field lets you point to a pre-existing IAM role:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesExternalSecrets
metadata:
  name: eso-eks
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.KubernetesExternalSecrets.eso-eks
spec:
  namespace:
    value: external-secrets
  createNamespace: true
  pollIntervalSeconds: 15
  container:
    resources:
      requests:
        cpu: "100m"
        memory: "128Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"
  eks:
    region: us-east-1
    irsaRoleArnOverride: arn:aws:iam::123456789012:role/eso-secrets-role
```

### AKS with Azure Key Vault

Deploy ESO on an AKS cluster with Azure Workload Identity for Key Vault secret synchronization. Increase resource limits for a production workload with many ExternalSecret objects:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesExternalSecrets
metadata:
  name: eso-aks
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesExternalSecrets.eso-aks
spec:
  namespace:
    value: external-secrets
  createNamespace: true
  pollIntervalSeconds: 60
  container:
    resources:
      requests:
        cpu: "250m"
        memory: "256Mi"
      limits:
        cpu: "2000m"
        memory: "2Gi"
  aks:
    keyVaultResourceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.KeyVault/vaults/my-keyvault
    managedIdentityClientId: 11111111-1111-1111-1111-111111111111
```

### Referencing a KubernetesNamespace via valueFrom

Use `valueFrom` on the `namespace` field to reference a namespace managed by a separate KubernetesNamespace resource instead of hard-coding the name:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesExternalSecrets
metadata:
  name: eso-ref
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesExternalSecrets.eso-ref
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      metadata:
        name: secrets-ns
      fieldPath: spec.name
  container:
    resources:
      requests:
        cpu: "50m"
        memory: "100Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"
  gke:
    projectId:
      valueFrom:
        kind: GcpProject
        metadata:
          name: my-project
        fieldPath: status.outputs.project_id
    gsaEmail: eso-sa@my-gcp-project.iam.gserviceaccount.com
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the External Secrets Operator is deployed |
| `service` | `string` | Name of the Kubernetes Service for the ESO controller |
| `portForwardCommand` | `string` | Ready-to-run `kubectl port-forward` command for local access to the operator |
| `kubeEndpoint` | `string` | Cluster-internal endpoint for the ESO service (e.g., `eso.external-secrets.svc.cluster.local`) |
| `ingressEndpoint` | `string` | Public endpoint when ingress is configured |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — pre-create a namespace to reference via `valueFrom`
- [KubernetesSecret](/docs/catalog/kubernetes/secret) — manage Kubernetes Secrets directly when external secret stores are not needed
- [KubernetesHelmRelease](/docs/catalog/kubernetes/helm-release) — deploy arbitrary Helm charts if you need to customize the ESO installation beyond what this component offers
