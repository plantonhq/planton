# Kubernetes GatewayClass

Creates a cluster-scoped Kubernetes Gateway API `GatewayClass` that identifies the controller (Istio, Envoy Gateway, NGINX Gateway Fabric, and others) responsible for managing Gateways of that class. GatewayClass is the infrastructure-provider layer of the Gateway API role model -- the root resource a `KubernetesGateway` references by name. This kind mirrors the upstream Gateway API v1 `GatewayClass` spec with full fidelity while adding proto validation, typed SDKs, and InfraChart composability.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A cluster-scoped GatewayClass** named after `metadata.name`, with the specified `controllerName` and optional `parametersRef` and `description`. The matching Gateway API controller observes the GatewayClass and sets its `Accepted` status condition.
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

No namespaced workloads are created -- GatewayClass is cluster-scoped.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Gateway API CRDs installed** -- deploy the `KubernetesGatewayApiCrds` component first. GatewayClass is a Gateway API type and will not register without the CRDs.
- **A running Gateway API controller** -- Istio, Envoy Gateway, NGINX Gateway Fabric, Cilium, etc. -- whose identity matches the `controllerName` you choose. Without a matching controller, the GatewayClass is created but never `Accepted`.

## Deploy

### Console

Open the deployment store, find **Kubernetes GatewayClass**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the two spec steps: **Controller** (the immutable `controllerName`) and **Parameters & Description** (both optional). Start from the **Istio GatewayClass** or **Envoy Gateway GatewayClass** preset in the [Presets](#presets) tab for a directly deployable configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGatewayClass
metadata:
  name: istio
  org: acme-corp
  env: prod
spec:
  controllerName: istio.io/gateway-controller
  description: "Production ingress via Istio"
```

```shell
planton apply -f gateway-class.yaml
```

This creates a GatewayClass named `istio` bound to the Istio Gateway controller. The name becomes the value Gateways reference via `spec.gatewayClassName`. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a GatewayClass. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Controller name** -- The `controllerName` field is a domain-prefixed path identifying the controller that manages Gateways of this class (e.g. `istio.io/gateway-controller`, `gateway.envoyproxy.io/gatewayclass-controller`, `gateway.nginx.org/nginx-gateway-controller`). It is **immutable**: the Gateway API admission webhook rejects changes, so switching controllers means recreating the GatewayClass. Copy the value verbatim from your controller's documentation.

**Parameters reference** -- The optional `parametersRef` points to a controller-specific resource (a ConfigMap or a custom resource such as Envoy Gateway's `EnvoyProxy`) that holds class-wide configuration. Set `kind` and `name` (and `namespace` for namespace-scoped referents); leave it unset for a controller-default class.

**Description** -- An optional human-friendly note (max 64 characters) that documents the class's purpose in lists and reviews.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. It does require the Gateway API CRDs and a matching controller to be present on the cluster (operational prerequisites, not Infra Component references).

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `gateway_class_name` | Name of the created GatewayClass (equals `metadata.name`) | `KubernetesGateway.spec.gatewayClassName` -- the foreign-key target that orders GatewayClass before Gateway |
| `controller_name` | The controller managing this class (echoes `spec.controllerName`) | Observability and confirming which implementation owns the class |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Istio GatewayClass** -- Binds the class to `istio.io/gateway-controller` for clusters running Istio with the Gateway API controller enabled. The standard choice for production ingress and service-mesh traffic. Start from the **Istio GatewayClass** preset.

**Envoy Gateway GatewayClass** -- Binds the class to `gateway.envoyproxy.io/gatewayclass-controller` for a lightweight, Envoy-based data plane without a full mesh. Optionally attach an `EnvoyProxy` resource via `parametersRef` for advanced tuning. Start from the **Envoy Gateway GatewayClass** preset.

## Works With

- [**Kubernetes Gateway API CRDs**](/infra-catalog/kubernetes-gateway-api-crds) -- installs the Gateway API CRDs (prerequisite, install first)
- [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) -- references this class via `gatewayClassName` to define listeners and entry points
- [**Kubernetes HTTPRoute**](/infra-catalog/kubernetes-http-route) -- routes HTTP traffic through a Gateway of this class
- [**Kubernetes GRPCRoute**](/infra-catalog/kubernetes-grpc-route) -- routes gRPC traffic through a Gateway of this class
- [**Kubernetes TLSRoute**](/infra-catalog/kubernetes-tls-route) -- routes passthrough TLS through a Gateway of this class
- [**Kubernetes TCPRoute**](/infra-catalog/kubernetes-tcp-route) -- routes raw TCP through a Gateway of this class
