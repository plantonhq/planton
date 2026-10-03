# Kubernetes Gateway API CRDs

Installs the Kubernetes Gateway API Custom Resource Definitions on any Kubernetes cluster, enabling Gateway, HTTPRoute, GRPCRoute, and other next-generation ingress and service mesh resources. Supports both standard and experimental installation channels with version pinning for reproducible deployments.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Gateway API CRDs** -- cluster-scoped Custom Resource Definitions applied from the official Gateway API release manifest. As of v1.6 the standard channel installs GatewayClass, Gateway, ListenerSet, HTTPRoute, GRPCRoute, TLSRoute, TCPRoute, UDPRoute, ReferenceGrant, and BackendTLSPolicy. The experimental channel adds experimental resources (such as XBackendTrafficPolicy) plus experimental fields on the standard resources.
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Cluster admin access** required. CRDs are cluster-scoped resources that require elevated permissions to install.
- **No existing Gateway API CRDs** at the target version. If a different version is already installed, remove it first or update the `version` field to match.

## Deploy

### Console

Open the deployment store, find **Kubernetes Gateway API CRDs**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Gateway API CRDs** preset in the [Presets](#presets) tab to install the stable set of Gateway API resources.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGatewayApiCrds
metadata:
  name: gateway-api
  org: acme-corp
  env: prod
spec:
  version: v1.6.1
  installChannel:
    channel: standard
```

```shell
planton apply -f gateway-api-crds.yaml
```

This installs the standard channel Gateway API CRDs at version v1.6.1, enabling the full route family (Gateway, GatewayClass, ListenerSet, HTTP/GRPC/TLS/TCP/UDP routes, ReferenceGrant, BackendTLSPolicy) cluster-wide. No namespace is required since CRDs are cluster-scoped. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring Gateway API CRDs. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Version pinning** -- The `version` field controls which Gateway API release to install and defaults to `v1.6.1` when left empty -- the release the catalog's Gateway API kinds are designed against. Graduations matter: TCPRoute and UDPRoute are standard-channel only from v1.6.0, and ListenerSet from v1.5.0, so an older release narrows what those kinds can deploy. Pin explicitly for reproducible deployments.

**Installation channel** -- The `installChannel.channel` field selects between `standard` and `experimental`. As of v1.6 the standard channel carries the whole route family -- including TCPRoute, UDPRoute, TLSRoute, and GRPCRoute -- alongside ListenerSet and BackendTLSPolicy, and it is the channel the catalog's Gateway API kinds target. Choose experimental only for a specific experimental resource or field; those may break or be removed between releases.

**Ingress controller compatibility** -- Gateway API CRDs are controller-agnostic. Your cluster must have a Gateway API-compatible controller installed separately (e.g., Istio, Envoy Gateway, NGINX Gateway Fabric) to act on the Gateway and Route resources.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `installed_version` | Gateway API version that was installed | Version verification in downstream configurations |
| `installed_channel` | Installation channel used (standard or experimental) | Validation that expected channel is active |
| `installed_manifest_url` | URL of the release manifest that was applied | Auditing exactly which upstream manifest the cluster runs |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard channel installation** -- Installs the standard Gateway API resources at v1.6.1 (the full route family plus ListenerSet, ReferenceGrant, and BackendTLSPolicy). Suitable for production clusters. Start from the **Standard Gateway API CRDs** preset.

**Experimental channel** -- Only when a specific experimental resource or field is needed; experimental resources may break or be removed between releases. Start from the **Experimental Gateway API CRDs** preset.

## Works With

This kind is the family prerequisite: every Gateway API kind in the catalog deploys onto the CRDs it installs.

- [**Kubernetes GatewayClass**](/infra-catalog/kubernetes-gateway-class) -- binds a controller implementation to the API these CRDs define
- [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) -- the traffic-handling instance deployed onto these CRDs
- [**Kubernetes ListenerSet**](/infra-catalog/kubernetes-listener-set) -- merges additional listeners into an opted-in Gateway
- [**Kubernetes HTTPRoute**](/infra-catalog/kubernetes-http-route) -- the workhorse route kind
- [**Kubernetes GRPCRoute**](/infra-catalog/kubernetes-grpc-route) -- gRPC method-level routing
- [**Kubernetes TLSRoute**](/infra-catalog/kubernetes-tls-route) -- SNI-based routing for passthrough TLS
- [**Kubernetes TCPRoute**](/infra-catalog/kubernetes-tcp-route) -- raw TCP forwarding (standard-channel from v1.6.0)
- [**Kubernetes UDPRoute**](/infra-catalog/kubernetes-udp-route) -- raw UDP forwarding (standard-channel from v1.6.0)
- [**Kubernetes ReferenceGrant**](/infra-catalog/kubernetes-reference-grant) -- cross-namespace reference permissions for the family