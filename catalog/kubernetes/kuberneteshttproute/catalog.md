# Kubernetes HTTPRoute

Creates a namespaced Kubernetes Gateway API `HTTPRoute` -- a route that matches **HTTP requests** by hostname, path, header, query parameter, or method, optionally transforms them with filters (header changes, redirects, URL rewrites, request mirroring, CORS), and forwards them to one or more backend Services through a Gateway. HTTPRoute is the richest and most widely used route in the Gateway API **standard channel** (served as `gateway.networking.k8s.io/v1`). It is the modern, first-class replacement for Ingress -- host/path routing, weighted canaries, header-based routing, redirects, and rewrites all in one resource. This component mirrors the upstream Gateway API `HTTPRoute` spec with full fidelity while adding proto validation, typed SDKs, and InfraChart composability.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A namespaced HTTPRoute** named after `metadata.name` in `spec.namespace`, attached to the Gateway listener(s) in `spec.parentRefs`, matching the `spec.hostnames` and the per-rule `matches`, and forwarding to the backends declared in its `spec.rules` (or terminating with a redirect).
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

The Gateway controller reconciles the route asynchronously: it reports per-parent `Accepted` / `ResolvedRefs` conditions in the route's status, which you observe with `kubectl` (these controller-managed values are intentionally not stored as outputs).

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Gateway API CRDs installed** -- deploy the `KubernetesGatewayApiCrds` component first. HTTPRoute is part of the standard channel, so the standard CRDs are sufficient (no experimental channel required).
- **A parent Gateway with an HTTP/HTTPS listener** -- each `parentRefs` entry should resolve to a `KubernetesGateway` whose listener serves HTTP or HTTPS, with an `allowedRoutes` policy that admits this route.
- **The target namespace exists** -- `spec.namespace` should resolve to a real `KubernetesNamespace`.
- **The backend Services exist** -- the `backendRefs` name in-cluster Services in the route's namespace (or in another namespace authorized by a `KubernetesReferenceGrant`).

## Deploy

### Console

Open the deployment store, find **Kubernetes HTTPRoute**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and three spec steps: **Namespace** (immutable), then **Routing** (the `hostnames` to match and the `parentRefs` Gateways to attach to), then **Rules** (per-rule matches, filters, and destination Services with weights). Start from the **Host + Path Routing** or **Weighted Canary Split** preset in the [Presets](#presets) tab for a directly deployable configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesHttpRoute
metadata:
  name: my-web-route
  org: acme-corp
  env: prod
spec:
  namespace:
    value: prod-apps
  parentRefs:
    - name:
        value: prod-gateway
      sectionName: https
  hostnames:
    - app.example.com
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /api
      backendRefs:
        - name:
            value: api
          port: 8080
```

```shell
planton apply -f http-route.yaml
```

This creates an HTTPRoute in `prod-apps` that attaches to the `https` listener of `prod-gateway`, matches requests to `app.example.com` under `/api`, and forwards them to the `api` Service on port 8080. An Infra Job tracks the provisioning in real time.

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
      sectionName: https
  rules:
    - backendRefs:
        - name:
            valueFrom:
              kind: KubernetesService
              name: api-service
              fieldPath: status.outputs.service_name
          port: 8080
```

The InfraPipeline deploys the Gateway and the backend Service first, then creates the route against them.

## Key Configuration

These are the most important decisions when configuring an HTTPRoute. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Namespace** -- The `namespace` field is where the route lives. It is **immutable**, the default namespace for its backends, and the anchor for cross-namespace rules: parent Gateways and backend Services in this namespace attach without a `ReferenceGrant`; those elsewhere require one. Reference an existing `KubernetesNamespace` or type the name directly.

**Hostnames** -- The `hostnames` (0-16) match the request's `Host` header. A leading `*.` is a suffix match; a bare IP is never valid. Leave empty to accept every hostname the parent listeners permit.

**Parent Gateways** -- The `parentRefs` (0-32) attach this route to Gateway listeners. Use `sectionName` to target one named listener and/or `port` to pin a port. A parent in another namespace needs a `ReferenceGrant` there and a matching listener `allowedRoutes` policy. Each parent's `name` is a foreign key to `KubernetesGateway`: reference a Planton-managed Gateway (the route then deploys after it), or pass a literal name for a Gateway or ListenerSet created outside Planton.

**Matches are ORed, conditions ANDed** -- Each of a rule's `matches` (up to 64) selects requests by `path` (`Exact` / `PathPrefix` / `RegularExpression`), request `headers`, `queryParams`, and/or `method`. A request matches the rule if ANY one match is satisfied; within a match, every condition must hold. No matches means a prefix match on `/` -- the rule catches everything the parent admits.

**Filters at rule level, for portability** -- `RequestHeaderModifier` / `ResponseHeaderModifier` (set/add/remove headers), `RequestRedirect` (respond with an HTTP redirect -- e.g. force HTTPS), `URLRewrite` (rewrite the host or path while forwarding), `RequestMirror` (shadow traffic to another backend), `CORS` (cross-origin headers), or `ExtensionRef` (an implementation-specific filter). Rule-level filters apply to every backend; per-backend filters exist but are implementation-specific, so prefer rule-level when the route must survive a controller change. A rule that only redirects needs no backend.

**Backends split by weight** -- Each rule forwards to 1-16 `backendRefs`, each with a `name`, `port`, and optional `weight` for traffic splitting (a stable/canary split at 90/10; weight `0` drains a backend without removing it from the manifest -- the clean way to hold a backend ready for rollback).

**Timeouts are per rule** -- optional `request` and `backendRequest` durations (GEP-2257 strings such as `10s`). `backendRequest` cannot exceed `request` (validated, unless `request` is `0s` which disables the limit); leave both unset to inherit the implementation's defaults.

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
| `route_name` | Name of the created HTTPRoute (equals `metadata.name`) | Orders the route after its Gateway and backends in an InfraChart |
| `namespace` | The resolved namespace the route was created in | Same-namespace / ReferenceGrant rules for its parent and backend references |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Host + path routing** -- Match a public hostname and a path prefix, then forward to a backend Service. The standard pattern for exposing a web application behind a Gateway. Start from the **Host + Path Routing** preset.

**Weighted canary** -- Split traffic across a stable and a canary backend by weight, the foundation of progressive delivery. Start from the **Weighted Canary Split** preset.

**Redirect to HTTPS** -- A rule with only a `RequestRedirect` filter (`scheme: https`, `statusCode: 308`) upgrades plain HTTP to HTTPS with no backend.

**Prefix rewrite** -- Pair a `PathPrefix` match with a `URLRewrite` `ReplacePrefixMatch` to strip a routing prefix (e.g. expose `/api` externally but forward `/` to the backend).

## Works With

- [**Kubernetes Gateway API CRDs**](/infra-catalog/kubernetes-gateway-api-crds) -- installs the Gateway API CRDs (standard channel is sufficient); deploy first (prerequisite).
- [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) -- the Gateway whose HTTP/HTTPS listener this route attaches to (`parentRefs`); install first.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the namespace (`spec.namespace`) the route runs in.
- [**Kubernetes ReferenceGrant**](/infra-catalog/kubernetes-reference-grant) -- authorizes cross-namespace parent or backend references from this route.
- [**Kubernetes Service**](/infra-catalog/kubernetes-service) -- the backend workloads (`backendRefs`) that receive forwarded requests.
