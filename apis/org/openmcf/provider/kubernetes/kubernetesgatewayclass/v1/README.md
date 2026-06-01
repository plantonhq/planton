# KubernetesGatewayClass

> Declarative management of Kubernetes Gateway API GatewayClass resources

## Overview

KubernetesGatewayClass creates a cluster-scoped Gateway API [GatewayClass](https://gateway-api.sigs.k8s.io/api-types/gatewayclass/) that identifies the controller (Istio, Envoy Gateway, NGINX Gateway Fabric, etc.) responsible for managing Gateways of that class. It is the infrastructure-provider layer of the Gateway API role model: a GatewayClass is to a Gateway what a StorageClass is to a PersistentVolume.

This component mirrors the upstream Gateway API v1 `GatewayClass` spec with 100% fidelity, so any value you can express in raw Gateway API YAML, you can express here -- with proto validation, typed SDKs, and InfraChart composability on top.

## Prerequisites

- The **Gateway API CRDs** must already be installed on the target cluster. Use the `KubernetesGatewayApiCrds` component (registered as a prerequisite of this kind).
- The chosen controller (for example Istio or Envoy Gateway) must be running in the cluster for the GatewayClass to be accepted.

## Quick Start

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGatewayClass
metadata:
  name: istio
spec:
  controllerName: istio.io/gateway-controller
  description: "Istio gateway controller for production ingress"
```

Deploy:

```bash
openmcf pulumi up --manifest gateway-class.yaml --stack org/project/env
```

## How It Works

1. The module creates a cluster-scoped `GatewayClass` custom resource named after `metadata.name`.
2. The Gateway API controller matching `controllerName` observes the GatewayClass and sets its `Accepted` status condition.
3. A `KubernetesGateway` then references this class by name via `spec.gatewayClassName`, and the controller provisions the Gateway's data plane.

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `targetCluster` | KubernetesClusterSelector | No | Target cluster (resolved from provider config when omitted) |
| `controllerName` | string | Yes | Domain-prefixed path identifying the controller (e.g. `istio.io/gateway-controller`). Immutable once created. |
| `parametersRef` | ParametersReference | No | Reference to a controller-specific parameters resource (ConfigMap or implementation CRD) |
| `parametersRef.group` | string | No | API group of the referent (empty for the core group) |
| `parametersRef.kind` | string | No | Kind of the referent (e.g. `ConfigMap`) |
| `parametersRef.name` | string | Yes (within parametersRef) | Name of the referent |
| `parametersRef.namespace` | string | No | Namespace of the referent; set only for namespace-scoped resources |
| `description` | string | No | Human-friendly description (max 64 characters) |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `gateway_class_name` | Name of the created GatewayClass (equals `metadata.name`). Reference this from `KubernetesGateway.spec.gateway_class_name`. |
| `controller_name` | The controller managing this GatewayClass |

## Related Components

- **KubernetesGatewayApiCrds** -- installs the Gateway API CRDs (prerequisite)
- **KubernetesGateway** -- references this class via `gatewayClassName` to define listeners and entry points
- **KubernetesHttpRoute / KubernetesGrpcRoute** -- attach routes to a Gateway of this class
