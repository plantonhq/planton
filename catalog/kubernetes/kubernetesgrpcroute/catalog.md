# Kubernetes GRPCRoute

Creates a namespaced Kubernetes Gateway API `GRPCRoute` -- a route that matches **gRPC requests** by hostname (`:authority`), service/method, or header, optionally transforms them with filters, and forwards them to one or more backend Services through a Gateway. GRPCRoute is part of the Gateway API **standard channel** (served as `gateway.networking.k8s.io/v1`). This is the first-class way to expose a gRPC API behind a Gateway -- weighted canaries, header-based routing, and request mirroring included. This kind mirrors the upstream Gateway API `GRPCRoute` spec with full fidelity while adding proto validation, typed SDKs, and InfraChart composability.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A namespaced GRPCRoute** named after `metadata.name` in `spec.namespace`, attached to the Gateway listener(s) in `spec.parentRefs`, matching the `spec.hostnames` and the per-rule `matches`, and forwarding to the backends declared in its `spec.rules`.
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

The Gateway controller reconciles the route asynchronously: it reports per-parent `Accepted` / `ResolvedRefs` conditions in the route's status, which you observe with `kubectl` (these controller-managed values are intentionally not stored as outputs).

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Gateway API CRDs installed** -- deploy the `KubernetesGatewayApiCrds` component first. GRPCRoute is part of the standard channel, so the standard CRDs are sufficient (no experimental channel required).
- **A parent Gateway with an HTTP/2 listener** -- each `parentRefs` entry should resolve to a `KubernetesGateway` whose listener speaks HTTP/2 (h2c over `protocol: HTTP`, or HTTP/2 over `protocol: HTTPS`), with an `allowedRoutes` policy that admits this route. An HTTP/1-only listener cannot serve gRPC.
- **The target namespace exists** -- `spec.namespace` should resolve to a real `KubernetesNamespace`.
- **The backend gRPC Services exist** -- the `backendRefs` name in-cluster Services in the route's namespace (or in another namespace authorized by a `KubernetesReferenceGrant`).

## Deploy

### Console

Open the deployment store, find **Kubernetes GRPCRoute**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and three spec steps: **Namespace** (immutable), then **Routing** (the `hostnames` to match and the `parentRefs` Gateways to attach to), then **Rules** (per-rule matches, filters, and destination Services with weights). Start from the **gRPC Service Routing** or **gRPC Weighted Canary** preset in the [Presets](#presets) tab for a directly deployable configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGrpcRoute
metadata:
  name: my-grpc-route
  org: acme-corp
  env: prod
spec:
  namespace:
    value: prod-apps
  parentRefs:
    - name:
        value: prod-gateway
      sectionName: grpc
  hostnames:
    - api.example.com
  rules:
    - matches:
        - method:
            service: helloworld.Greeter
      backendRefs:
        - name:
            value: greeter
          port: 9000
```

```shell
planton apply -f grpc-route.yaml
```

This creates a GRPCRoute in `prod-apps` that attaches to the `grpc` listener of `prod-gateway`, matches calls to `helloworld.Greeter` on `api.example.com`, and forwards them to the `greeter` Service on port 9000. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the route's Gateway and backend Service by reference so the InfraPipeline orders the deploys:

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
      sectionName: grpc
  rules:
    - backendRefs:
        - name:
            valueFrom:
              kind: KubernetesService
              name: greeter-service
              fieldPath: status.outputs.service_name
          port: 9000
```

The InfraPipeline deploys the Gateway and the backend Service first, then creates the route against them.

## Key Configuration

These are the most important decisions when configuring a GRPCRoute. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Namespace** -- The `namespace` field is where the route lives. It is **immutable**, the default namespace for its backends, and the anchor for cross-namespace rules: parent Gateways and backend Services in this namespace attach without a `ReferenceGrant`; those elsewhere require one. Reference an existing `KubernetesNamespace` or type the name directly.

**Hostnames** -- The `hostnames` (0-16) match the request's `:authority` (Host) pseudo-header. A leading `*.` is a suffix match; a bare IP is never valid. Leave empty to accept every hostname the parent listeners permit.

**Parent Gateways** -- The `parentRefs` (0-32) attach this route to Gateway listeners. Use `sectionName` to target one named listener and/or `port` to pin a port. A parent in another namespace needs a `ReferenceGrant` there and a matching listener `allowedRoutes` policy. Each parent's `name` is a foreign key to `KubernetesGateway`: reference a Planton-managed Gateway (the route then deploys after it), or pass a literal name for a Gateway or ListenerSet created outside Planton.

**Matches are ORed, conditions ANDed** -- Each of a rule's `matches` (up to 64) selects requests by `method` (`service` and/or `method`, `Exact` or `RegularExpression`) and/or request `headers`. A request matches the rule if ANY one match is satisfied; within a match, every condition must hold. No matches means the rule applies to every request the parent admits -- a common surprise on shared listeners.

**Filters at rule level, for portability** -- `RequestHeaderModifier` / `ResponseHeaderModifier` (set/add/remove headers), `RequestMirror` (shadow traffic to another backend), or `ExtensionRef` (an implementation-specific filter). Rule-level filters apply to every backend; per-backend filters exist but are implementation-specific, so prefer rule-level when the route must survive a controller change.

**Backends split by weight** -- Each rule forwards to 1-16 `backendRefs`, each with a `name`, `port`, and optional `weight` for traffic splitting (a stable/canary split at 90/10; weight `0` drains a backend without removing it from the manifest -- the clean way to hold a backend ready for rollback).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesGateway** | `parentRefs[].name` | `status.outputs.gateway_name` |
| **KubernetesService** | `rules[].backendRefs[].name` | `status.outputs.service_name` |

Literal names cover Gateways and Services created outside Planton; cross-namespace references additionally require a `KubernetesReferenceGrant` in the target namespace.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `route_name` | Name of the created GRPCRoute (equals `metadata.name`) | Orders the route after its Gateway and backends in an InfraChart |
| `namespace` | The resolved namespace the route was created in | Same-namespace / ReferenceGrant rules for its parent and backend references |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**gRPC service routing** -- Match a public hostname and a gRPC service (and optionally a method), then forward to a backend gRPC Service. Start from the **gRPC Service Routing** preset.

**Weighted canary** -- Split gRPC traffic for a service across a stable and a canary backend by weight, the standard progressive-delivery pattern. Start from the **gRPC Weighted Canary** preset.

## Works With

- [**Kubernetes Gateway API CRDs**](/infra-catalog/kubernetes-gateway-api-crds) -- installs the Gateway API CRDs (standard channel is sufficient); deploy first (prerequisite).
- [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) -- the Gateway whose HTTP/2 listener this route attaches to (`parentRefs`); install first.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the namespace (`spec.namespace`) the route runs in.
- [**Kubernetes ReferenceGrant**](/infra-catalog/kubernetes-reference-grant) -- authorizes cross-namespace parent or backend references from this route.
- [**Kubernetes Service**](/infra-catalog/kubernetes-service) -- the backend gRPC workloads (`backendRefs`) that receive forwarded requests.
