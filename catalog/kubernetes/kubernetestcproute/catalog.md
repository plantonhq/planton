# Kubernetes TCPRoute

Creates a namespaced Kubernetes Gateway API `TCPRoute` -- a route that forwards **raw TCP connections** arriving on a Gateway listener to one or more backend Services. TCPRoute is a **layer-4** route: it has no hostnames, no matches, and no filters. A connection on the parent listener's port is simply forwarded to the rule's backends. This is the standard way to expose a non-HTTP TCP service (a database, a message broker, a custom protocol) through a Gateway. This component mirrors the upstream Gateway API `TCPRoute` (GA, `gateway.networking.k8s.io/v1`) spec with full fidelity while adding proto validation, typed SDKs, and InfraChart composability.

> **Standard channel from v1.6.** `TCPRoute` graduated to GA and is served as `gateway.networking.k8s.io/v1` in the standard channel from Gateway API v1.6.0 (it was an experimental v1alpha2 resource in earlier releases). Deploy `KubernetesGatewayApiCrds` at v1.6.0+ first.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A namespaced TCPRoute** named after `metadata.name` in `spec.namespace`, attached to the Gateway listener(s) in `spec.parentRefs`, forwarding connections to the backends declared in its `spec.rules`.
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

The Gateway controller reconciles the route asynchronously: it reports per-parent `Accepted` / `ResolvedRefs` conditions in the route's status, which you observe with `kubectl` (these controller-managed values are intentionally not stored as outputs).

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Gateway API CRDs at v1.6.0+ installed** -- deploy the `KubernetesGatewayApiCrds` component first. TCPRoute is standard-channel only from v1.6.0; older CRD releases will not register it.
- **A parent Gateway with a TCP listener** -- each `parentRefs` entry should resolve to a `KubernetesGateway` that has a listener of `protocol: TCP`, and an `allowedRoutes` policy that admits this route.
- **The target namespace exists** -- `spec.namespace` should resolve to a real `KubernetesNamespace`.
- **The backend Services exist** -- the `backendRefs` name in-cluster Services in the route's namespace (or in another namespace authorized by a `KubernetesReferenceGrant`).

## Deploy

### Console

Open the deployment store, find **Kubernetes TCPRoute**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and three spec steps: **Namespace** (immutable), then **Gateways** (the `parentRefs` Gateways to attach to), then **Rules** (the route rules and their destination Services and weights). Start from the **TCP Port Forwarding** or **TCP Weighted Backends** preset in the [Presets](#presets) tab for a directly deployable configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTcpRoute
metadata:
  name: my-tcp-route
  org: acme-corp
  env: prod
spec:
  namespace:
    value: prod-apps
  parentRefs:
    - name:
        value: prod-gateway
      sectionName: tcp-postgres
  rules:
    - backendRefs:
        - name:
            value: postgres
          port: 5432
```

```shell
planton apply -f tcp-route.yaml
```

This creates a TCPRoute in `prod-apps` that attaches to the `tcp-postgres` listener of `prod-gateway` and forwards every connection arriving on that listener to the `postgres` Service on port 5432. An Infra Job tracks the provisioning in real time.

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
      sectionName: tcp-postgres
  rules:
    - backendRefs:
        - name:
            valueFrom:
              kind: KubernetesService
              name: postgres
              fieldPath: status.outputs.service_name
          port: 5432
```

The InfraPipeline deploys the referenced Gateway and Service first, then provisions the route against them.

## Key Configuration

These are the most important decisions when configuring a TCPRoute. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Namespace** -- The `namespace` field is where the route lives. It is **immutable**, the default namespace for its backends, and the anchor for cross-namespace rules: parent Gateways and backend Services in this namespace attach without a `ReferenceGrant`; those elsewhere require one. Reference an existing `KubernetesNamespace` or type the name directly.

**Parent Gateways** -- The `parentRefs` (0-32) attach this route to Gateway listeners. Use `sectionName` to target one named TCP listener and/or `port` to pin a port -- the listener's port decides which connections arrive, since TCP does no matching. A parent in another namespace needs a `ReferenceGrant` there and a matching listener `allowedRoutes` policy. Each parent's `name` is a foreign key to `KubernetesGateway`: reference a Planton-managed Gateway (the route then deploys after it), or pass a literal name for a Gateway or ListenerSet created outside Planton.

**Rules** -- A TCPRoute has 1-16 `rules`, each with an optional `name` and `backendRefs` (1-16). Each backend's `name` is a foreign key to `KubernetesService` (reference a Planton-managed Service or pass a literal in-cluster name), with a `port` and an optional `weight` for traffic splitting (e.g. a primary/standby split at 90/10; weight `0` drains a backend). There is no matching -- every connection on the listener is forwarded to the rule's backends.

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
| `route_name` | Name of the created TCPRoute (equals `metadata.name`) | Orders the route after its Gateway and backends in an InfraChart |
| `namespace` | The resolved namespace the route was created in | Same-namespace / ReferenceGrant rules for its parent and backend references |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**TCP port forwarding** -- Forward all connections arriving on a Gateway's TCP listener to a single backend Service (a database, broker, or custom protocol). Start from the **TCP Port Forwarding** preset.

**Weighted backends** -- Split TCP connections across two Services by weight, for example a primary and a standby. Start from the **TCP Weighted Backends** preset.

## Works With

- [**Kubernetes Gateway API CRDs**](/infra-catalog/kubernetes-gateway-api-crds) -- installs the Gateway API CRDs (v1.6.0+ standard channel carries TCPRoute); deploy first.
- [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) -- the Gateway whose TCP listener this route attaches to (`parentRefs`); install first.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the namespace (`spec.namespace`) the route runs in.
- [**Kubernetes ReferenceGrant**](/infra-catalog/kubernetes-reference-grant) -- authorizes cross-namespace parent or backend references from this route.
- [**Kubernetes Service**](/infra-catalog/kubernetes-service) -- the backend workloads (`backendRefs`) that receive forwarded connections.
