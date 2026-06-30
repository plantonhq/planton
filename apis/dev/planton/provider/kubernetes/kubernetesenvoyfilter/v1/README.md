# Kubernetes Envoy Filter

Provision an Istio `EnvoyFilter` -- the namespaced, expert-only escape hatch that applies
low-level patches directly to the Envoy proxy configuration istiod generates for selected
workloads. Use it to add, merge, remove, insert, or replace Envoy listeners, filter chains,
network/HTTP filters, route configurations, virtual hosts, routes, and clusters when no
first-class Istio API expresses what you need.

> **Expert-only escape hatch.** EnvoyFilter patches Envoy's internal xDS API directly and the
> patch body is free-form JSON that istiod merges with **no** schema validation -- a malformed
> patch can break a workload's traffic. Reach for a typed Istio API (`VirtualService`,
> `DestinationRule`, `RequestAuthentication`, `Telemetry`, ...) first; use `EnvoyFilter` only
> for capabilities not yet modeled by a higher-level API.

## What Gets Created

- A namespaced `networking.istio.io/v1alpha3` `EnvoyFilter` custom resource. (EnvoyFilter is
  the only Istio API still served at `v1alpha3`; it has not graduated to `v1`.)
- An attachment scope -- either a `workload_selector` (pods/VMs by label) or `target_refs`
  (a Gateway/GatewayClass/Service/ServiceEntry), or neither (all workloads in the namespace) --
  plus an ordered list of `config_patches` and an optional `priority`.

## How a patch is targeted

Each entry in `config_patches` has three parts:

1. **`apply_to`** -- which kind of Envoy object the patch targets (`LISTENER`, `HTTP_FILTER`,
   `CLUSTER`, `ROUTE_CONFIGURATION`, ...).
2. **`match`** -- the conditions that select the specific object: a `context`
   (`SIDECAR_INBOUND` / `SIDECAR_OUTBOUND` / `GATEWAY` / `ANY`), an optional `proxy` match,
   and **at most one** of `listener`, `route_configuration`, or `cluster`.
3. **`patch`** -- the `operation` (`MERGE`, `ADD`, `INSERT_BEFORE`, `REPLACE`, ...) and the
   free-form `value` (the Envoy config fragment), plus an optional `filter_class`.

## Attachment: selector vs target_refs

`workload_selector` and `target_refs` are mutually exclusive (at most one). With neither set,
the filter applies to all workloads in its namespace -- or, when created in the Istio mesh root
namespace (e.g. `istio-system`), to all applicable workloads mesh-wide. Waypoint proxies
**require** `target_refs`; selector-based policies are ignored for waypoints.

## Prerequisites

- Istio CRDs installed on the cluster (`KubernetesIstioBaseCrds`).
- A running Istio control plane, istiod (`KubernetesIstio`), to translate the patches. The
  resource applies with only the CRDs present, but patches take effect only where istiod and a
  sidecar (or ambient) data plane are running.
- The target namespace (`KubernetesNamespace`).

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesEnvoyFilter
metadata:
  name: grpc-web-cors
spec:
  namespace:
    value: edge
  workload_selector:
    labels:
      app: grpc-web-gateway
  config_patches:
    - apply_to: HTTP_FILTER
      match:
        context: GATEWAY
        listener:
          filter_chain:
            filter:
              name: envoy.filters.network.http_connection_manager
              sub_filter:
                name: envoy.filters.http.router
      patch:
        operation: INSERT_BEFORE
        value:
          name: envoy.filters.http.cors
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.http.cors.v3.Cors
```

```bash
planton apply -f envoyfilter.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace the EnvoyFilter is created in. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `workload_selector.labels` | map | Pod/VM labels the patches apply to. Mutually exclusive with `target_refs`. |
| `target_refs` | list | Attach to specific resources (`group`/`kind`/`name`); max 16; no cross-namespace. Mutually exclusive with `workload_selector`. |
| `config_patches` | list | Ordered patches (see below). An empty list is a valid no-op. |
| `priority` | int | Patch-set ordering within a context (default 0; negatives first, positives last). |

### config_patches[] fields

| Field | Type | Description |
|-------|------|-------------|
| `apply_to` | string | `LISTENER`, `FILTER_CHAIN`, `NETWORK_FILTER`, `HTTP_FILTER`, `ROUTE_CONFIGURATION`, `VIRTUAL_HOST`, `HTTP_ROUTE`, `CLUSTER`, `EXTENSION_CONFIG`, `LISTENER_FILTER`, or `BOOTSTRAP` (deprecated). |
| `match.context` | string | `ANY` (default), `SIDECAR_INBOUND`, `SIDECAR_OUTBOUND`, `GATEWAY`. |
| `match.proxy` | object | `proxy_version` (RE2 regex) and/or exact `metadata` key-values. |
| `match.listener` | object | Listener match: `port_number`, `filter_chain` (`name`/`sni`/`transport_protocol`/`application_protocols`/`filter`→`sub_filter`/`destination_port`), `listener_filter`, `name`. |
| `match.route_configuration` | object | Route-config match: `port_number`, `port_name`, `gateway`, `vhost` (`name`/`domain_name`/`route`→`name`/`action`), `name`. |
| `match.cluster` | object | Cluster match: `port_number`, `service`, `subset`, `name`. |
| `patch.operation` | string | `MERGE`, `ADD`, `REMOVE`, `INSERT_BEFORE`, `INSERT_AFTER`, `INSERT_FIRST`, `REPLACE`. |
| `patch.value` | object | Free-form Envoy config (xDS JSON), merged via proto-merge. Not needed for `REMOVE`. |
| `patch.filter_class` | string | `AUTHN`, `AUTHZ`, or `STATS` (filter insertion point, used with `ADD`). |

At most one of `match.listener`, `match.route_configuration`, or `match.cluster` may be set.

## Examples

### Insert a CORS HTTP filter on a gateway

```yaml
spec:
  namespace:
    value: edge
  target_refs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: public-gateway
  config_patches:
    - apply_to: HTTP_FILTER
      match:
        context: GATEWAY
        listener:
          filter_chain:
            filter:
              name: envoy.filters.network.http_connection_manager
              sub_filter:
                name: envoy.filters.http.router
      patch:
        operation: INSERT_BEFORE
        value:
          name: envoy.filters.http.cors
```

### Merge a setting onto an outbound cluster

```yaml
spec:
  namespace:
    value: payments
  workload_selector:
    labels:
      app: checkout
  config_patches:
    - apply_to: CLUSTER
      match:
        context: SIDECAR_OUTBOUND
        cluster:
          service: reviews.default.svc.cluster.local
      patch:
        operation: MERGE
        value:
          connect_timeout: 5s
```

## Composing in Infra Charts

`namespace` is a `StringValueOrRef`, so it can reference a `KubernetesNamespace` output via
`valueFrom`. Neither `workload_selector.labels` nor any `target_refs` entry is a foreign key --
istiod resolves them at runtime, so they create no automatic DAG edge to a workload, gateway,
or service. To order this EnvoyFilter relative to what it patches, declare the dependency on
`metadata.relationships`:

```yaml
metadata:
  name: gateway-cors
  relationships:
    - kind: KubernetesGateway
      name: public-gateway
      type: depends_on
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: edge-ns
      fieldPath: spec.name
  target_refs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: public-gateway
  config_patches:
    - apply_to: HTTP_FILTER
      patch:
        operation: INSERT_BEFORE
        value:
          name: envoy.filters.http.cors
```

See `docs/README.md` for the full composability rationale and the escape-hatch posture.

## Stack Outputs

| Output | Description |
|--------|-------------|
| `envoy_filter_name` | Name of the created EnvoyFilter (equals metadata.name). |
| `namespace` | Namespace the EnvoyFilter was created in. |

## Related Components

- [Kubernetes Istio](kubernetesistio)
- [Kubernetes Istio Base CRDs](kubernetesistiobasecrds)
- [Kubernetes Namespace](kubernetesnamespace)
