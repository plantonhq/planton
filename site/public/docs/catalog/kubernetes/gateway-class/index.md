---
title: "Gateway Class"
description: "Gateway Class deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesgatewayclass"
---

# Kubernetes Gateway Class

Creates a cluster-scoped Kubernetes Gateway API `GatewayClass` that identifies the controller (Istio, Envoy Gateway, NGINX Gateway Fabric, and others) responsible for managing Gateways of that class. GatewayClass is the infrastructure-provider layer of the Gateway API role model -- the root resource that a `KubernetesGateway` references by name. This component mirrors the upstream Gateway API v1 `GatewayClass` spec with full fidelity while adding proto validation, typed SDKs, and InfraChart composability.

## What Gets Created

When you deploy a KubernetesGatewayClass resource, Planton provisions:

- **A cluster-scoped GatewayClass custom resource** named after `metadata.name`, with the specified `controllerName` and optional `parametersRef` and `description`.

No namespaced workloads are created. The matching Gateway API controller observes the GatewayClass and sets its `Accepted` status condition.

## Prerequisites

- **Gateway API CRDs** installed on the target cluster -- deploy the `KubernetesGatewayApiCrds` component first (it is registered as a prerequisite of this kind).
- **A running Gateway API controller** (Istio, Envoy Gateway, NGINX Gateway Fabric, etc.) whose identity matches `controllerName`.
- **Kubernetes credentials** configured via the Planton provider config.

## Quick Start

Create a file `gateway-class.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesGatewayClass
metadata:
  name: istio
spec:
  controllerName: istio.io/gateway-controller
  description: "Istio gateway controller for production ingress"
```

Deploy:

```shell
planton apply -f gateway-class.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `controllerName` | `string` | Domain-prefixed path identifying the controller (e.g. `istio.io/gateway-controller`). Immutable once created. 1-253 characters. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind (e.g. `GcpGkeCluster`, `AwsEksCluster`). |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `parametersRef.group` | `string` | `""` | API group of the controller-specific parameters resource. |
| `parametersRef.kind` | `string` | `""` | Kind of the parameters resource (e.g. `ConfigMap`, `EnvoyProxy`). |
| `parametersRef.name` | `string` | — | Name of the parameters resource (required when `parametersRef` is set). |
| `parametersRef.namespace` | `string` | — | Namespace of the parameters resource; set only for namespace-scoped resources. |
| `description` | `string` | — | Human-friendly description (max 64 characters). |

## Examples

### Istio GatewayClass

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesGatewayClass
metadata:
  name: istio
spec:
  controllerName: istio.io/gateway-controller
  description: "Istio gateway controller"
```

### Envoy Gateway with a parameters reference

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesGatewayClass
metadata:
  name: envoy-gateway
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: custom-proxy-config
    namespace: envoy-gateway-system
  description: "Envoy Gateway with custom proxy config"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `gatewayClassName` | `string` | Name of the created GatewayClass (equals `metadata.name`). Reference this from `KubernetesGateway.spec.gatewayClassName`. |
| `controllerName` | `string` | The controller managing this GatewayClass. |

## Related Components

- [KubernetesGatewayApiCrds](/docs/catalog/kubernetes/gateway-api-crds) — installs the Gateway API CRDs (prerequisite)
- [KubernetesGateway](/docs/catalog/kubernetes/gateway) — references this class via `gatewayClassName` to define listeners and entry points
- [KubernetesHttpRoute](/docs/catalog/kubernetes/http-route) — routes HTTP traffic through a Gateway of this class
