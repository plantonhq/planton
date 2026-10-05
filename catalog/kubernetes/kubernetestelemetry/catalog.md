# Istio Telemetry

Defines an Istio Telemetry resource: a namespaced configuration of *how the mesh observes* the workloads it selects. It controls distributed tracing (sampling rate, reporting provider, custom span tags), metrics (provider selection and per-metric tag dimensions), and access logging (provider, on/off, and a CEL filter) -- without touching application code. This is how you dial trace sampling up or down, prune high-cardinality metric tags, and log only the requests that matter.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A Telemetry resource** -- a namespaced Istio policy that shapes traces, metrics, and access logs for the workloads it selects.
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

istiod merges the resource with parent-scope configuration at runtime and programs the data plane -- there is no controller-reconciled status to wait on.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Istio installed** -- the Istio CRDs (deploy **Istio Base CRDs** first) must be present and the Istio control plane (istiod) running. Telemetry is only generated where istiod is active.
- **Telemetry providers configured** -- the tracing/metrics/logging providers you reference by name must be declared in the mesh's `MeshConfig` extension providers. Leave providers empty to use the mesh defaults.
- **Target namespace exists** -- the resource is created in a specific namespace; reference an existing one or create it first.

## Deploy

### Console

Open the deployment store, find **Istio Telemetry**, and click **Deploy**. The creation wizard walks you through the namespace, the scope, and the three observability signals (tracing, metrics, access logging), with guidance at each step. Start from the **Mesh-Wide Trace Sampling** or **Prometheus Metric Dimensions** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTelemetry
metadata:
  name: prod-apps-observability
  org: acme-corp
  env: prod
spec:
  namespace:
    value: prod-apps
  tracing:
    - randomSamplingPercentage: 10
      customTags:
        cluster_name:
          environment:
            name: ISTIO_META_CLUSTER_ID
  metrics:
    - overrides:
        - match:
            metric: REQUEST_COUNT
          tagOverrides:
            request_path:
              operation: REMOVE
  accessLogging:
    - filter:
        expression: response.code >= 500
```

```shell
planton apply -f telemetry.yaml
```

This samples 10% of traces and tags every span with the cluster ID, drops the high-cardinality `request_path` dimension from request counts, and logs only server 5xx responses -- a sane, cost-aware observability baseline for a whole namespace. An Infra Job tracks the provisioning in real time.

### InfraChart

When attaching the configuration to a Gateway managed in the same InfraChart, wire the target reference with `valueFrom` so the policy deploys after its gateway:

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
          name: prod-gateway
          fieldPath: status.outputs.gateway_name
  accessLogging:
    - filter:
        expression: response.code >= 500
```

The InfraPipeline deploys the namespace and Gateway first, then creates the Telemetry resource against them.

## Key Configuration

These are the most important decisions when configuring a Telemetry resource. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope is one of three shapes, and empty means everything** -- no selector and no target references applies the configuration to every workload in its namespace (or the whole mesh, from the Istio root namespace); a `selector` matches pod labels; `targetRefs` attaches to a Gateway, Service, or ServiceEntry. At most one of `selector` and `targetRefs` may be set -- the spec enforces it. Waypoint proxies honor ONLY target references; label selectors are silently ignored by waypoints.

**The selector is a runtime label match, not a dependency edge** -- istiod matches pod labels live, so no automatic ordering exists between this configuration and the workloads it observes. Order it after them with `metadata.relationships`. A target reference's `name`, by contrast, IS a foreign key defaulting to a Gateway: wiring it with `valueFrom` orders this configuration after its gateway, while a literal `value:` covers Services, ServiceEntries, and resources created outside Planton.

**Merging rewards least-specific-first** -- istiod merges multiple rules with later (more specific) entries overriding earlier ones, so list broad matches first. Custom span tags are the exception: they fully REPLACE any tags inherited from a parent configuration -- they are never merged, so a child that sets one tag drops every inherited tag.

**Sampling percentage is the cost lever** -- `randomSamplingPercentage` (0-100, in 0.01% increments) decides what fraction of requests produce spans when no prior sampling decision exists; upstream defaults to 0% when unset. Sample low (1-10%) in production and rely on trace-context propagation for completeness -- 100% sampling on a busy mesh overwhelms the trace backend and its bill.

**Metric tag overrides need the right operation** -- `UPSERT` requires a CEL value expression while `REMOVE` takes none. Removing a high-cardinality dimension (a request path, a user ID) from a standard metric is the single most effective way to keep the time-series database lean.

**An empty resource changes nothing** -- every signal is optional; workloads inherit the mesh defaults until a rule says otherwise. That makes an incremental rollout safe: start with one signal, verify, add the next.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|---|---|---|
| Kubernetes Namespace | `spec.namespace` | `spec.name` |
| Kubernetes Gateway | `spec.targetRefs[].name` | `status.outputs.gateway_name` |

The workload `selector` is a plain runtime label match and carries no dependency edge -- order the configuration after the workloads it observes with `metadata.relationships`.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `telemetry_name` | Name of the created Telemetry resource (equals `metadata.name`) | Ordering resources that depend on the config being in place |
| `namespace` | The namespace the resource was created in | Confirming where the config applies |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Cost-aware tracing** -- sample a small percentage of requests so traces stay useful without overwhelming the trace backend or its bill. Start from the **Mesh-Wide Trace Sampling** preset.

**Metric cardinality control** -- REMOVE a high-cardinality tag (like a request path or user ID) from a standard metric to keep your time-series database lean, or UPSERT a CEL-derived dimension you want to slice by. Start from the **Prometheus Metric Dimensions** preset.

**Error-only access logs** -- attach a CEL filter such as `response.code >= 500` so only failing requests are logged, keeping log volume and cost down while preserving the signal you need to debug.

## Works With

- [**Istio Base CRDs**](/infra-catalog/kubernetes-istio-base-crds) -- installs the Telemetry CRD; a prerequisite together with a running istiod.
- [**Istio**](/infra-catalog/kubernetes-istio) -- the control plane that merges Telemetry resources and programs the data plane; its `MeshConfig` declares the providers referenced here.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the namespace (`spec.namespace`) the configuration is created in and scopes to.
- [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) -- a natural target reference: observe a gateway's traffic with its own sampling and logging rules.
- [**Istio Service Entry**](/infra-catalog/kubernetes-service-entry) -- another target-reference kind, for observing traffic to mesh-external services.
