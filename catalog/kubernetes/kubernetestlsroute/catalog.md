# Kubernetes TLSRoute

Creates a namespaced Kubernetes Gateway API `TLSRoute` -- a route that matches inbound TLS connections by their **SNI hostname** and forwards them, still encrypted, to one or more backend Services (TLS passthrough). The Gateway never decrypts the traffic: the backend terminates TLS itself. This is the standard way to expose services that must hold their own certificate (databases, mTLS services, or apps doing end-to-end TLS). This kind mirrors the upstream Gateway API `TLSRoute` (standard channel, `gateway.networking.k8s.io/v1` from Gateway API v1.6) spec with full fidelity while adding proto validation, typed SDKs, and InfraChart composability.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A namespaced TLSRoute** named after `metadata.name` in `spec.namespace`, attached to the Gateway listener(s) in `spec.parentRefs`, matching the `spec.hostnames` SNI names, and forwarding to the backends in its single rule.
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

The Gateway controller reconciles the route asynchronously: it reports per-parent `Accepted` / `ResolvedRefs` conditions in the route's status, which you observe with `kubectl` (these controller-managed values are intentionally not stored as outputs).

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Gateway API CRDs at v1.6.0+ installed** -- deploy the **Kubernetes Gateway API CRDs** component first. TLSRoute is served as `gateway.networking.k8s.io/v1` in the standard channel from v1.6.0 (it was experimental v1alpha2/v1alpha3 in earlier releases); older CRD releases will not register it.
- **A parent Gateway with a TLS passthrough listener** -- each `parentRefs` entry should resolve to a `KubernetesGateway` that has a listener of `protocol: TLS` with `tls.mode: Passthrough`, and an `allowedRoutes` policy that admits this route.
- **The target namespace exists** -- `spec.namespace` should resolve to a real `KubernetesNamespace`.
- **The backend Services exist** -- the `backendRefs` name in-cluster Services in the route's namespace (or in another namespace authorized by a `KubernetesReferenceGrant`).

## Deploy

### Console

Open the deployment store, find **Kubernetes TLSRoute**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and three spec steps: **Namespace** (immutable), then **Routing** (the SNI `hostnames` to match and the `parentRefs` Gateways to attach to), then **Backends** (the rule's destination Services and weights). Start from the **TLS Passthrough by SNI** or **TLS Weighted Backends** preset in the [Presets](#presets) tab for a directly deployable configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTlsRoute
metadata:
  name: my-tls-route
  org: acme-corp
  env: prod
spec:
  namespace:
    value: prod-apps
  parentRefs:
    - name:
        value: prod-gateway
      sectionName: tls-passthrough
  hostnames:
    - secure.example.com
  rules:
    - backendRefs:
        - name:
            value: secure-app
          port: 8443
```

```shell
planton apply -f tls-route.yaml
```

This creates a TLSRoute in `prod-apps` that attaches to the `tls-passthrough` listener of `prod-gateway`, matches connections whose SNI is `secure.example.com`, and forwards them, still encrypted, to the `secure-app` Service on port 8443. An Infra Job tracks the provisioning in real time.

### InfraChart

When the Gateway and backend Service are managed in the same InfraChart, wire the route's references with `valueFrom` so the resource graph carries the edges:

```yaml
spec:
  namespace:
    value: prod-apps
  parentRefs:
    - name:
        valueFrom:
          kind: KubernetesGateway
          name: prod-gateway
          fieldPath: status.outputs.gateway_name
      sectionName: tls-passthrough
  hostnames:
    - secure.example.com
  rules:
    - backendRefs:
        - name:
            valueFrom:
              kind: KubernetesService
              name: secure-app
              fieldPath: status.outputs.service_name
          port: 8443
```

The InfraPipeline deploys the referenced Gateway and Service first, then provisions the route against them.

## Key Configuration

These are the most important decisions when configuring a TLSRoute. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Namespace** -- The `namespace` field is where the route lives. It is **immutable**, the default namespace for its backends, and the anchor for cross-namespace rules: parent Gateways and backend Services in this namespace attach without a `ReferenceGrant`; those elsewhere require one. Reference an existing `KubernetesNamespace` or type the name directly.

**Hostnames** -- One or more SNI `hostnames` (at least one is required; upstream raised the cap from 16 to 1024 in the v1.6 release). Each is a DNS name, optionally a leading-wildcard (`*.example.com`). TLS routing matches **only** on SNI -- there is no path or header matching. Per RFC 6066, an SNI hostname can never be an IP address.

**Parent Gateways** -- The `parentRefs` (0-32) attach this route to Gateway listeners. Use `sectionName` to target one named listener and/or `port` to pin a port. A parent in another namespace needs a `ReferenceGrant` there and a matching listener `allowedRoutes` policy. Each parent's `name` is a foreign key to `KubernetesGateway`: reference a Planton-managed Gateway (the route then deploys after it), or pass a literal name for a Gateway or ListenerSet created outside Planton.

**Backends** -- The single rule's `backendRefs` (1-16) are the destination Services. Each has a `name`, a `port`, and an optional `weight` for traffic splitting (e.g. a canary at 90/10; weight `0` drains a backend). The backend terminates TLS -- the route passes the encrypted stream through unchanged.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|---|---|---|
| Kubernetes Namespace | `spec.namespace` | `spec.name` |
| Kubernetes Gateway | `spec.parentRefs[].name` | `status.outputs.gateway_name` |
| Kubernetes Service | `spec.rules[].backendRefs[].name` | `status.outputs.service_name` |

Literal names (`value:`) cover targets created outside Planton; cross-namespace references additionally require a Reference Grant in the target namespace.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `route_name` | Name of the created TLSRoute (equals `metadata.name`) | Orders the route after its Gateway and backends in an InfraChart |
| `namespace` | The resolved namespace the route was created in | Same-namespace / ReferenceGrant rules for its parent and backend references |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**TLS passthrough by SNI** -- Match a single SNI hostname and forward it, unmodified, to one backend that terminates TLS. Start from the **TLS Passthrough by SNI** preset.

**Weighted backends (canary)** -- Split passthrough traffic across two Services by weight for a canary or blue/green rollout. Start from the **TLS Weighted Backends** preset.

## Works With

- [**Kubernetes Gateway API CRDs**](/infra-catalog/kubernetes-gateway-api-crds) -- installs the Gateway API CRDs (v1.6.0+ standard channel carries TLSRoute); deploy first.
- [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) -- the Gateway whose TLS passthrough listener this route attaches to (`parentRefs`); install first.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the namespace (`spec.namespace`) the route runs in.
- [**Kubernetes ReferenceGrant**](/infra-catalog/kubernetes-reference-grant) -- authorizes cross-namespace parent or backend references from this route.
- [**Kubernetes Service**](/infra-catalog/kubernetes-service) -- the backend workloads (`backendRefs`) that terminate TLS and receive traffic.
