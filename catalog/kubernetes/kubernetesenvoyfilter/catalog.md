# Istio Envoy Filter

Defines an Istio EnvoyFilter: a namespaced, expert-only customization of the raw Envoy proxy configuration that Istio generates for the workloads it selects. It applies an ordered list of patches -- merge, add, remove, insert, or replace -- to low-level Envoy objects (listeners, filter chains, network and HTTP filters, route configurations, virtual hosts, routes, and clusters). This is the escape hatch you reach for only when a capability is not yet exposed by a higher-level Istio API.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EnvoyFilter** -- a namespaced `networking.istio.io/v1alpha3` policy that patches the generated Envoy configuration for the workloads it selects. istiod merges the patches into the xDS configuration it pushes to the matched proxies.

The resource is pure configuration: no pods, no Services. Its effect exists entirely inside the proxies istiod programs.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Istio installed** -- the Istio CRDs (deploy **Istio Base CRDs**) must be present and the Istio control plane (istiod) running. Patches are only honored where istiod is active.
- **Target namespace exists** -- the resource is created in a specific namespace; reference an existing one or create it first.

## Deploy

### Console

Open the deployment store, find **Istio Envoy Filter**, and click **Deploy**. The creation wizard walks you through the namespace, the attachment (namespace-wide, workload selector, or target references), and the config patches. Start from the **Tune an outbound cluster (MERGE)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesEnvoyFilter
metadata:
  name: reviews-idle-timeout
  org: acme-corp
  env: prod
spec:
  namespace:
    value: prod-apps
  workloadSelector:
    labels:
      app: reviews
  configPatches:
    - applyTo: CLUSTER
      match:
        context: SIDECAR_OUTBOUND
        cluster:
          service: reviews.prod.svc.cluster.local
      patch:
        operation: MERGE
        value:
          common_http_protocol_options:
            idle_timeout: 30s
```

```shell
planton apply -f envoy-filter.yaml
```

This sets a 30-second idle timeout on the outbound cluster for the `reviews` service, applied only to the `reviews` workload's sidecars in `prod-apps`. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the target reference to the gateway managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: prod-apps-namespace
      fieldPath: spec.name
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name:
        valueFrom:
          kind: KubernetesGateway
          name: public-gateway
          fieldPath: status.outputs.gateway_name
```

The InfraPipeline deploys the namespace and gateway first, then provisions the EnvoyFilter against the resolved names.

## Key Configuration

These are the most important decisions when configuring an EnvoyFilter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Patches are unvalidated -- match narrowly, test narrowly** -- the patch value is free-form configuration (a `google.protobuf.Struct`) that istiod merges into the generated config with no schema validation. A malformed patch can break the patched workload's traffic outright. Match objects as specifically as possible, and prove a patch on a narrow `workloadSelector` before widening it to a namespace or the mesh.

**Attachment: selector or target references, never both** -- `workloadSelector` matches pods by label; `targetRefs` attach to a Gateway, Service, or ServiceEntry (at most 16, no cross-namespace). The spec enforces the exclusivity. Omit both and the patches apply namespace-wide -- or mesh-wide when the resource lives in the Istio root namespace, which turns a typo into a mesh-wide outage. Waypoint proxies honor only `targetRefs`; label selectors are ignored by waypoints.

**Target reference names order the graph** -- a `targetRefs[].name` is a foreign key defaulting to a Gateway reference: wiring it with `valueFrom` orders this resource after the gateway it patches. A literal `value:` covers Services, ServiceEntries, and resources created outside Planton. The `workloadSelector`, by contrast, is a plain runtime label match with no dependency edge -- order this resource after the workloads it patches with `metadata.relationships`.

**Prefer ADD with a filter class over INSERT operations** -- `ADD` with a `filterClass` (AUTHN / AUTHZ / STATS) lets istiod place the filter at a stable position in the chain. The `INSERT_BEFORE` / `INSERT_AFTER` operations pin against specific filter names that can change between Istio versions -- the classic way an EnvoyFilter silently breaks on upgrade.

**REMOVE takes no value; everything else does** -- `REMOVE` deletes the matched object outright. Every other operation carries the configuration fragment to merge, insert, or replace.

**Priority decides cross-resource ordering** -- root-namespace patch sets always apply before workload-namespace ones; within the same namespace, resources sort by `priority` (default 0; negatives first), then patches apply in `configPatches` list order. Leave priority unset unless two EnvoyFilters genuinely contend for the same object.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesGateway** | `targetRefs[].name` | `status.outputs.gateway_name` |

### What This Kind Provides

An EnvoyFilter is a policy resource consumed by istiod -- there is no controller status worth exporting. `status.outputs` carries only the resource identity (`envoy_filter_name`, `namespace`), both echoes of the manifest; downstream resources have nothing to consume from it.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Insert a stock Envoy HTTP filter at a gateway** -- add a native Envoy filter (CORS for gRPC-Web, ext_authz, rate limiting) into a gateway's HTTP connection manager when no first-class Istio API covers it on your version. Start from the **Add a CORS HTTP filter to a gateway (gRPC-Web)** preset.

**Tune an upstream cluster** -- MERGE connection-pool or timeout settings onto a generated cluster that the typed DestinationRule API does not expose. Start from the **Tune an outbound cluster (MERGE)** preset.

**Remove a generated object** -- REMOVE a filter or listener Istio adds by default when you need a leaner proxy configuration; the narrowest-possible match keeps the removal from hitting objects you did not intend.

## Works With

- [**Istio Base CRDs**](/infra-catalog/kubernetes-istio-base-crds) -- the prerequisite CRDs the EnvoyFilter kind is defined by
- [**Istio**](/infra-catalog/kubernetes-istio) -- the control plane (istiod) that merges the patches into proxy configuration
- [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) -- the most common `targetRefs` attachment point for gateway-level filters
- [**Istio Service Entry**](/infra-catalog/kubernetes-service-entry) -- an attachment target for patching egress to external hosts
- [**Istio Destination Rule**](/infra-catalog/kubernetes-destination-rule) -- the typed API to prefer for cluster-level tuning; reach for EnvoyFilter only for what it does not expose
- [**Istio Telemetry**](/infra-catalog/kubernetes-telemetry) -- the typed API to prefer for observability configuration
- [**Istio Authorization Policy**](/infra-catalog/kubernetes-authorization-policy) -- the typed API to prefer for access control
