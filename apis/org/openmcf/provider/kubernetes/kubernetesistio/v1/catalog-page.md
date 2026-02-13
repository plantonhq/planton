# Kubernetes Istio

Deploys the Istio service mesh on Kubernetes using three official Istio Helm charts (base, istiod, and gateway, pinned to version 1.22.3), with configurable resource limits for the Istiod control plane, optional namespace creation for both `istio-system` and `istio-ingress`, and an ingress gateway exposed as a LoadBalancer service.

## What Gets Created

When you deploy a KubernetesIstio resource, OpenMCF provisions:

- **Namespaces** — `istio-system` and `istio-ingress` are created only when `createNamespace` is `true`; if your cluster already has these namespaces, leave the flag as `false`
- **Helm Release (istio/base)** — installs Istio CRDs and cluster-scoped resources from the `base` chart at `https://istio-release.storage.googleapis.com/charts`, pinned to version 1.22.3, with atomic rollback and a 3-minute timeout
- **Helm Release (istiod)** — deploys the Istio control plane from the `istiod` chart in the system namespace, with configurable CPU and memory resource limits/requests for the pilot container
- **Helm Release (istio-gateway)** — deploys the Istio ingress gateway from the `gateway` chart in the `istio-ingress` namespace, exposed as a `LoadBalancer` service type

All Helm release names are prefixed with `metadata.name` (e.g., `my-mesh-base`, `my-mesh-istiod`, `my-mesh-gateway`) to avoid conflicts when multiple Istio instances share a cluster.

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists (`istio-system` and `istio-ingress`), or set `createNamespace` to `true`
- **Sufficient cluster RBAC** for creating CRDs, MutatingWebhookConfigurations, and LoadBalancer services

## Quick Start

Create a file `istio.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesIstio
metadata:
  name: my-mesh
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesIstio.my-mesh
spec:
  namespace: istio-system
  createNamespace: true
  container: {}
```

Deploy:

```shell
openmcf apply -f istio.yaml
```

This creates an Istio service mesh with three Helm releases (base, istiod, ingress gateway) using default resource limits (1000m CPU / 1Gi memory limits, 50m CPU / 100Mi memory requests for istiod). The ingress gateway is exposed as a LoadBalancer in the `istio-ingress` namespace.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the Istio system components (istiod and base). Defaults to `istio-system` when left empty. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `container` | `object` | Container specification for the Istiod control plane. Pass `{}` to accept all defaults. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates both the `istio-system` and `istio-ingress` namespaces before deploying Helm releases. |
| `container.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for the Istiod control plane pod. |
| `container.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for the Istiod control plane pod. |
| `container.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for the Istiod control plane pod. |
| `container.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for the Istiod control plane pod. |

## Examples

### Development Instance with Minimal Resources

A lightweight Istio mesh for development or testing with reduced control plane resource allocations:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesIstio
metadata:
  name: dev-mesh
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesIstio.dev-mesh
spec:
  namespace: istio-system
  createNamespace: true
  container:
    resources:
      limits:
        cpu: "500m"
        memory: "512Mi"
      requests:
        cpu: "50m"
        memory: "64Mi"
```

### Production Instance with Higher Resource Limits

A production Istio deployment with increased CPU and memory for the Istiod control plane to handle a larger number of sidecars and configuration updates:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesIstio
metadata:
  name: prod-mesh
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesIstio.prod-mesh
spec:
  namespace: istio-system
  createNamespace: true
  container:
    resources:
      limits:
        cpu: "2000m"
        memory: "2Gi"
      requests:
        cpu: "250m"
        memory: "512Mi"
```

### Using Foreign Key References

Reference an OpenMCF-managed namespace instead of hardcoding the name. The `namespace` field supports `valueFrom` to resolve the namespace name from another resource at deploy time:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesIstio
metadata:
  name: platform-mesh
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesIstio.platform-mesh
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: istio-namespace
      field: spec.name
  container:
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where istiod and the base chart are deployed |
| `service` | `string` | Kubernetes Service name for the Istiod control plane (format: `{name}-istiod`) |
| `port_forward_command` | `string` | kubectl port-forward command for local access to the Istiod debug port on `localhost:15014` |
| `kube_endpoint` | `string` | Cluster-internal FQDN for the Istiod xDS port (e.g., `my-mesh-istiod.istio-system.svc.cluster.local:15012`) |
| `ingress_endpoint` | `string` | Cluster-internal FQDN for the ingress gateway (e.g., `my-mesh-gateway.istio-ingress.svc.cluster.local:80`) |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesHelmRelease](/docs/catalog/kubernetes/kuberneteshelmrelease) — alternative for deploying individual Helm charts when full Istio mesh setup is not needed
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — application deployments that can be injected with Istio sidecar proxies
