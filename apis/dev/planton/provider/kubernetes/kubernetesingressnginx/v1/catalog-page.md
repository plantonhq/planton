# Kubernetes Ingress Nginx

Deploys the ingress-nginx controller on Kubernetes using the upstream Helm chart (default version 4.11.1), with provider-specific load balancer configuration for GKE, EKS, and AKS, optional internal load balancer mode, configurable chart version, and optional namespace creation.

## What Gets Created

When you deploy a KubernetesIngressNginx resource, Planton provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Helm Release (ingress-nginx)** — deploys the ingress-nginx controller from `https://kubernetes.github.io/kubernetes-ingress-nginx`, pinned to the specified `chartVersion` (default 4.11.1), with atomic rollback enabled, cleanup on failure, wait-for-jobs, and a 180-second timeout; the controller service is set to type `LoadBalancer` with the default ingress class enabled and `watchIngressWithoutClass` turned on
- **Load Balancer Annotations** — provider-specific annotations applied to the controller service based on the selected provider config (`gke`, `eks`, or `aks`) and the `internal` flag

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **Cloud provider load balancer support** — the target cluster must support `LoadBalancer`-type services (GKE, EKS, AKS, or equivalent)
- **Static IP / subnet resources** pre-created if referencing them in provider-specific configuration (e.g., GKE static IP, EKS subnets)

## Quick Start

Create a file `ingress-nginx.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesIngressNginx
metadata:
  name: my-ingress
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesIngressNginx.my-ingress
spec:
  namespace: ingress-nginx
  createNamespace: true
```

Deploy:

```shell
planton apply -f ingress-nginx.yaml
```

This creates an ingress-nginx controller in the `ingress-nginx` namespace with the default chart version (4.11.1), an external `LoadBalancer` service, the default ingress class enabled, and no provider-specific annotations.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the ingress-nginx deployment. Accepts a literal string or a `valueFrom` reference to a KubernetesNamespace resource (see `spec.name` on the referenced resource). | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying the Helm release. |
| `chartVersion` | `string` | `4.11.1` | Upstream ingress-nginx Helm chart version tag. |
| `internal` | `bool` | `false` | When `true`, configures the controller service with an internal load balancer. The default (`false`) produces an external load balancer. |
| `gke.staticIpName` | `string` | — | Name of a pre-existing reserved static IP address to assign to the GKE load balancer. |
| `gke.subnetworkSelfLink` | `string` | — | Subnetwork self-link for internal load balancers on GKE. |
| `eks.additionalSecurityGroupIds` | `string[]` | — | Security group IDs to attach to the AWS load balancer in addition to the controller-managed group. Each entry accepts a literal string or a `valueFrom` reference to an AwsSecurityGroup resource. |
| `eks.subnetIds` | `string[]` | — | Subnet IDs where the ELB/NLB should be placed. Leave empty to let AWS select subnets automatically. Each entry accepts a literal string or a `valueFrom` reference to an AwsVpc resource. |
| `eks.irsaRoleArnOverride` | `string` | — | Existing IAM role ARN for IRSA. If empty, the stack can auto-create and wire up a role. |
| `aks.managedIdentityClientId` | `string` | — | Client ID of a user-assigned managed identity for Azure Workload Identity binding on the controller ServiceAccount. |
| `aks.publicIpName` | `string` | — | Name of a pre-existing Azure public IP resource to reuse for the load balancer. |

> **Note on `valueFrom`:** Fields of type `StringValueOrRef` (such as `namespace`, `eks.additionalSecurityGroupIds`, and `eks.subnetIds`) accept either a literal string value or a `valueFrom` block that references another Planton resource's output field. See the [Foreign Key References](#using-foreign-key-references) example below.

## Examples

### External Load Balancer on GKE with Static IP

Deploy ingress-nginx on a GKE cluster using a reserved static IP for the external load balancer:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesIngressNginx
metadata:
  name: gke-external
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesIngressNginx.gke-external
spec:
  namespace: ingress-nginx
  createNamespace: true
  chartVersion: "4.11.1"
  gke:
    staticIpName: prod-ingress-ip
```

### Internal Load Balancer on EKS

Deploy an internal-only ingress-nginx controller on EKS, pinned to specific subnets and with additional security groups:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesIngressNginx
metadata:
  name: eks-internal
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesIngressNginx.eks-internal
spec:
  namespace: ingress-system
  createNamespace: true
  internal: true
  eks:
    additionalSecurityGroupIds:
      - sg-0123456789abcdef0
      - sg-abcdef0123456789a
    subnetIds:
      - subnet-aaa111
      - subnet-bbb222
```

### AKS with Managed Identity

Deploy ingress-nginx on AKS with Azure Workload Identity and a pre-existing public IP:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesIngressNginx
metadata:
  name: aks-ingress
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesIngressNginx.aks-ingress
spec:
  namespace: ingress-nginx
  createNamespace: true
  aks:
    managedIdentityClientId: 12345678-abcd-efgh-ijkl-123456789abc
    publicIpName: prod-ingress-pip
```

### Using Foreign Key References

Reference an Planton-managed namespace and EKS security groups from other resources instead of hardcoding values:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesIngressNginx
metadata:
  name: platform-ingress
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesIngressNginx.platform-ingress
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: platform-namespace
      field: spec.name
  createNamespace: false
  internal: true
  eks:
    additionalSecurityGroupIds:
      - valueFrom:
          kind: AwsSecurityGroup
          name: ingress-sg
          fieldPath: status.outputs.security_group_id
    subnetIds:
      - valueFrom:
          kind: AwsSubnet
          name: platform-public-subnet-a
          fieldPath: status.outputs.subnet_id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where ingress-nginx is deployed |
| `release_name` | `string` | Helm release name (matches `metadata.name`) |
| `service_name` | `string` | Kubernetes Service name for the ingress-nginx controller (format: `{name}-controller`) |
| `service_type` | `string` | Service type, typically `LoadBalancer` |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesHelmRelease](/docs/catalog/kubernetes/kuberneteshelmrelease) — alternative for deploying arbitrary Helm charts when the ingress-nginx component does not cover your use case
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — application workloads that use Ingress resources routed through the ingress-nginx controller
- [KubernetesService](/docs/catalog/kubernetes/kubernetesservice) — backend services exposed via Ingress rules
