# OpenTelemetry Collector

Declares one OpenTelemetry Collector -- a telemetry pipeline (receivers, processors, exporters wired into service pipelines) that the cluster's OpenTelemetry Operator reconciles into a running workload. This kind is a DECLARATION, not an install: the engine is the separately deployed KubernetesOtelOperator, and one collector per pipeline shape is the grain -- a per-node log daemonset, a scalable traces gateway, an OTLP fan-in front door, or an injected sidecar. The pipeline document is the product: the operator validates it at admission, derives the collector Service and its ports from the declared receivers, and exports the in-cluster OTLP endpoints applications point at.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **OpenTelemetryCollector CR** -- the declaration the operator reconciles into the mode's workload:
  - `deployment` (the default) -- a scalable gateway/fan-in Deployment, sized by `replicas` or the operator-managed `autoscaler`
  - `daemonset` -- one collector per node, for log files and host/kubelet metrics that only exist node-locally
  - `statefulset` -- stable pod identities for the target allocator and persistent sending queues
  - `sidecar` -- NO standalone workload; the operator injects the collector into pods annotated `sidecar.opentelemetry.io/inject`
- **Derived Services** (all modes except sidecar) -- the collector Service carrying the ports the operator derives from the declared receivers, a headless Service for per-pod gRPC addressing, and a monitoring Service (port 8888) exposing the collector's own metrics
- **Rendered ConfigMap** -- the pipeline document, shipped verbatim; credentials never ride it (they load from Secrets as environment variables and are referenced as `${env:VAR_NAME}`)
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **KubernetesOtelOperator (HARD prerequisite)** -- a running operator on the cluster. This resource is a declaration the operator reconciles; without it nothing admits, converts, or rolls out.
- **RBAC for cluster-state pipelines** -- pipelines using `k8sattributes`, `kubeletstats` or `k8s_cluster` read cluster state the operator's default ServiceAccount cannot. Compose a KubernetesServiceAccount + KubernetesRbac and name the account in `serviceAccount` -- permission problems surface as receiver errors in the collector logs, never as pod failures.
- **Secrets before pods** -- every Secret named in `envFromSecrets` must exist in the namespace when the collector starts, or the pods hold in `CreateContainerConfigError`.

## Deploy

### Console

Open the deployment store, find **OpenTelemetry Collector**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Cluster-logs-to-Loki preset** for the per-node log pattern, the **Traces-gateway-to-Tempo preset** for the fixed-replica gateway, or the **OTLP fan-in preset** for the autoscaled front door in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOtelCollector
metadata:
  name: traces-gateway
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "observability"
  createNamespace: true
  configYaml: |
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: 0.0.0.0:4317
          http:
            endpoint: 0.0.0.0:4318
    processors:
      memory_limiter:
        check_interval: 1s
        limit_mib: 400
        spike_limit_mib: 100
      batch: {}
    exporters:
      otlp:
        endpoint: tempo.observability.svc.cluster.local:4317
        tls:
          insecure: true
    service:
      pipelines:
        traces:
          receivers: [otlp]
          processors: [memory_limiter, batch]
          exporters: [otlp]
  replicas: 2
  resources:
    requests:
      cpu: 200m
      memory: 512Mi
    limits:
      memory: 512Mi
```

```shell
planton apply -f traces-gateway.yaml
```

This declares a two-replica traces gateway in the `observability` namespace: applications push OTLP spans to the exported endpoints, and the collector ships them to Tempo through the `memory_limiter` and `batch` processors. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the collector to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: observability-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then declares the collector into it.

## Key Configuration

These are the most important decisions when configuring the OpenTelemetry Collector. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The pipeline document IS the product** -- `configYaml` is required and ships VERBATIM: receivers, processors, exporters and the service pipelines wiring them, on the collector's own open contract (the component registry is unbounded by design -- typing a subset would fence off the ecosystem). The operator's admission webhook validates its shape at apply time and derives the Service ports from the declared receivers. Keep the standard `otlp` receiver on 4317/4318 -- the exported OTLP endpoints assume it.

**Credentials ride references, never the document** -- The document lands in a rendered ConfigMap, so tokens never belong in it. Load existing Secrets whole with `envFromSecrets` and reference their keys as `${env:VAR_NAME}`; rotating a credential means updating the Secret, never the spec. This spec carries Secret NAMES only, which is why they render plainly.

**The mode gates what applies** -- `daemonset` and `sidecar` reject `replicas` and `autoscaler` (one per node / inside the targets); `sidecar` also rejects tolerations and a priority class (the target pods own their scheduling). These are the spec's own admission rules -- the wizard clears the dials on a mode switch and the API rejects violating manifests.

**One owner of the replica count** -- Declare `replicas` (empty = 1) OR the operator-managed `autoscaler` (min/max bounds, CPU/memory utilization targets), never both. The autoscaler's targets are percentages of the CPU/memory REQUESTS -- without real requests the HPA has nothing to measure against.

**The image dial is per-collector** -- `image` empty means the operator injects its default collector image (the fleet-wide dial lives on the operator's `defaultCollectorImage`). Set it here only to diverge THIS collector -- a contrib build for `filelog`/`k8sattributes`, a vendor distribution. The operator cannot know what an image ships: a missing component surfaces at RUNTIME in the collector logs.

**Size memory WITH the memory_limiter** -- The container memory limit and the pipeline's `memory_limiter` are ONE decision: `limit_mib` a step below the limit (the classic pairing: 512Mi limit, `limit_mib: 400`) makes overload degrade visibly -- refused requests, back-pressure, counted drops -- instead of an OOM-kill that takes the telemetry with it.

**The daemonset log pattern costs root** -- Container runtimes write pod log files readable only by root, and the default collector image runs non-root: without `podSecurityContext.runAsUser: 0` the filelog receiver reports permission errors while the pods stay Running/Ready -- logs simply never arrive. Pair root with read-only log mounts (`/var/log/pods` hostPath) plus a writable checkpoint hostPath so offsets survive restarts.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef (every Service-derived output is empty in sidecar mode -- no standalone workload exists):

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the collector runs in | Locating the pipeline for diagnostics |
| `collector_name` | The OpenTelemetryCollector CR name -- every child name derives from it | kubectl inspection of the reconciled workload |
| `service` | The collector Service (`<name>-collector`) | Network-policy allowances for telemetry traffic |
| `otlp_grpc_endpoint` | In-cluster OTLP gRPC ingest endpoint (`<service>:4317`) | Applications' OTLP exporter configuration |
| `otlp_http_endpoint` | In-cluster OTLP HTTP ingest endpoint (`http://<service>:4318`) | Applications' OTLP/HTTP exporter configuration |
| `headless_service` | The headless Service -- per-pod addressing | Load-balancing-aware gRPC clients |
| `monitoring_service` | The monitoring Service (port 8888) -- the collector's own metrics | Prometheus scrape targets; watching the memory_limiter's refused/dropped series |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Per-node log collection** -- Daemonset mode, the filelog receiver over `/var/log/pods`, `k8sattributes` enrichment, a control-plane toleration, `runAsUser: 0`, and an `otlphttp` exporter into a Loki gateway. Start from the **Cluster-logs-to-Loki preset**.

**Traces gateway** -- A fixed two-replica deployment applications push OTLP spans to, with the memory limit and `memory_limiter` in deliberate agreement, exporting to Tempo. Start from the **Traces-gateway-to-Tempo preset**.

**OTLP fan-in front door** -- One autoscaled endpoint for everything: apps configure a single OTLP exporter, and this resource decides where each signal lands -- changed here without touching a single application. Start from the **OTLP fan-in preset**.

## Works With

- [**OpenTelemetry Operator**](/infra-catalog/kubernetes-otel-operator) -- the HARD prerequisite: the engine that reconciles this declaration
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the collector
- [**Kubernetes ServiceAccount**](/infra-catalog/kubernetes-service-account) + [**Kubernetes RBAC**](/infra-catalog/kubernetes-rbac) -- the identity composition for cluster-state pipelines (`k8sattributes`, `kubeletstats`)
- [**Grafana Loki**](/infra-catalog/kubernetes-loki) -- a common logs backend: its gateway ingests OTLP at the `/otlp` route
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack) -- scrapes the collector's own metrics through the monitoring Service
