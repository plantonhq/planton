# Kubernetes Gateway API CRDs

Installs the Kubernetes Gateway API Custom Resource Definitions (CRDs) on a target Kubernetes cluster. The Gateway API is the next-generation, role-oriented API for managing ingress and service mesh traffic, replacing the legacy Ingress resource with richer routing primitives such as Gateway, HTTPRoute, GRPCRoute, and ReferenceGrant. This component fetches the official CRD manifests from the upstream `kubernetes-sigs/gateway-api` releases and applies them directly to the cluster.

## What Gets Created

When you deploy a KubernetesGatewayApiCrds resource, OpenMCF provisions:

- **Standard Channel CRDs** — `GatewayClass`, `Gateway`, `HTTPRoute`, and `ReferenceGrant` custom resource definitions, enabling the core Gateway API surface
- **Experimental Channel CRDs** (when `installChannel` is set to `experimental`) — all standard CRDs plus `TCPRoute`, `UDPRoute`, `TLSRoute`, and `GRPCRoute` experimental custom resource definitions

No namespaced workloads are created. The CRDs are cluster-scoped and make the Gateway API resource types available for any namespace in the cluster.

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **Cluster-admin privileges** on the target cluster, because CRD installation requires cluster-wide write access
- **Network access** from the deployment runner to `https://github.com/kubernetes-sigs/gateway-api/releases/download` to fetch CRD manifests

## Quick Start

Create a file `gateway-api-crds.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGatewayApiCrds
metadata:
  name: gateway-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesGatewayApiCrds.gateway-api
spec: {}
```

Deploy:

```shell
openmcf apply -f gateway-api-crds.yaml
```

This installs the standard-channel Gateway API CRDs at the default version (v1.2.1) on the cluster configured in your environment.

## Configuration Reference

### Required Fields

This component has no strictly required spec fields. An empty `spec: {}` installs the standard-channel CRDs at the default version.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `version` | `string` | `v1.2.1` | Gateway API release version to install. Must match the pattern `v<major>.<minor>.<patch>` with an optional pre-release suffix (e.g., `v1.3.0`, `v1.2.1-rc1`). |
| `installChannel.channel` | `enum` | `standard` | CRD installation channel. `standard` installs Gateway, GatewayClass, HTTPRoute, and ReferenceGrant. `experimental` adds TCPRoute, UDPRoute, TLSRoute, and GRPCRoute. |

## Examples

### Standard CRDs at Default Version

Installs the stable Gateway API CRDs using all defaults:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGatewayApiCrds
metadata:
  name: gateway-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesGatewayApiCrds.gateway-api
spec: {}
```

### Experimental Channel with Specific Version

Installs all Gateway API CRDs, including the experimental route types, at a pinned version:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGatewayApiCrds
metadata:
  name: gateway-api-experimental
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesGatewayApiCrds.gateway-api-experimental
spec:
  version: "v1.3.0"
  installChannel:
    channel: experimental
```

### Target a Specific GKE Cluster

Installs the standard CRDs on a named GKE cluster in a production environment:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGatewayApiCrds
metadata:
  name: gateway-api-prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesGatewayApiCrds.gateway-api-prod
spec:
  targetCluster:
    clusterKind: GcpGkeCluster
    clusterName: prod-cluster
  version: "v1.2.1"
  installChannel:
    channel: standard
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `installedVersion` | `string` | Gateway API version that was installed (e.g., `v1.2.1`, `v1.3.0`) |
| `installedChannel` | `string` | Installation channel that was used (`standard` or `experimental`) |
| `installedManifestUrl` | `string` | Full URL of the Gateway API CRD bundle that was applied (encodes version + channel, e.g., `.../releases/download/v1.5.1/experimental-install.yaml`) |

## Related Components

- [KubernetesHelmRelease](/docs/catalog/kubernetes/kuberneteshelmrelease) — deploy a Gateway API controller (such as Envoy Gateway or Istio) after the CRDs are in place
- [KubernetesManifest](/docs/catalog/kubernetes/kubernetesmanifest) — apply Gateway and HTTPRoute manifests that reference the installed CRDs
- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — create namespaces for Gateway API controller workloads
