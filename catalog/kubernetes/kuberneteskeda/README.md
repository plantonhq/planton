# Kubernetes KEDA

## When NOT to Use This

**One installation per cluster.** KEDA registers the cluster-wide
`v1beta1.external.metrics.k8s.io` APIService — a singleton, and Kubernetes
allows only one external metrics provider. Check whether the cluster
already runs KEDA (or another external-metrics adapter such as
Prometheus-adapter serving external metrics) before adding this component;
a second installation would fight the first over the registration.

Also not the right kind when:

- **You only scale on CPU/memory** — that is plain HorizontalPodAutoscaler
  territory backed by metrics-server; KEDA's value is real-world signals
  (queue depth, stream lag, cron schedules, cloud metrics) and
  scale-to-zero.
- **You want to declare WHAT scales** — the scaling declarations
  (ScaledObject, ScaledJob, TriggerAuthentication) are KEDA custom
  resources deployed per workload, alongside the workloads they scale
  (KubernetesManifest carries them today). This component installs and
  configures the ENGINE.
- **You want HTTP request-based scaling** — that is the separate KEDA HTTP
  add-on (its own chart and interceptor architecture), not part of this
  installation.

## Overview

**KubernetesKeda** installs KEDA — Kubernetes Event-Driven Autoscaling —
from the official Helm chart (`keda` at
`https://kedacore.github.io/charts`). KEDA scales workloads on real-world
signals instead of only CPU/memory: its operator watches
ScaledObject/ScaledJob resources, drives the workload's HPA (including
scale-to-ZERO, which plain HPA cannot do), and serves the
`external.metrics.k8s.io` API the HPA controller reads. 70+ scalers cover
queues, streams, databases, cron schedules, and cloud metric sources.

The typed spec covers the chart's meaningful configuration surface, with a
`helm_values` escape hatch (merged last, Helm `-f` semantics, identical on
both engines) for anything beyond it.

**Key design points:**

- **The release name is fixed to `keda`** — one installation per cluster is
  an upstream constraint (the external-metrics APIService is a cluster
  singleton), so the release name never derives from `metadata.name`. The
  chart names its components fixed too: `keda-operator`,
  `keda-operator-metrics-apiserver`, `keda-admission-webhooks`.
- **CRDs are kept on uninstall by default** — the chart has no native keep
  knob, so the module stamps `helm.sh/resource-policy: keep` onto the CRDs
  through `crds.additionalAnnotations`. Without it, a plain uninstall
  cascade-deletes every ScaledObject/ScaledJob/TriggerAuthentication in the
  cluster.
- **Extra replicas are warm standbys, not capacity** — the operator and
  metrics server lead-elect/serve one at a time; per upstream HA guidance a
  second replica buys failover speed, not throughput.
- **The install waits for real readiness** — both engines install
  atomically (300s) and wait for the components to become Available: a
  ServiceMonitor rendered without the Prometheus operator CRDs, or broken
  internal TLS wiring, fails THIS deploy instead of surfacing later as
  ScaledObjects that never scale.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: installation namespace (`keda` is the upstream
  convention) — literal or a KubernetesNamespace reference

### Common

- **`spec.create_namespace`**: create (and own) the namespace with the
  release — usually true for a dedicated `keda` namespace
- **`spec.chart_version`**: pinned chart version (default `2.20.1`, which
  ships KEDA 2.20.1 — chart and app versions move together)
- **`spec.crds`**: `install` (chart default true) and `keep_on_uninstall`
  (default true — the guard rail against cascade-deleting every scaling
  declaration in the cluster)
- **`spec.watch_namespace`**: empty (chart default) watches ALL namespaces
  — the normal cluster-wide posture; set one namespace to fence KEDA into
  a single team's space
- **`spec.operator` / `spec.metrics_server`**: replicas (chart default 1;
  extras are leader-elected standbys) and container resources (chart
  defaults: requests 100m/100Mi, limits 1/1000Mi)
- **`spec.webhooks`**: admission validation of ScaledObjects at apply time
  — `enabled` (chart default true), `failure_policy` (`Ignore` chart
  default, or `Fail` to reject applies while the webhook is unreachable),
  replicas, resources
- **`spec.pod_identity`**: ambient cloud identity for scalers —
  `aws_irsa` (role ARN), `azure_workload_identity` (client + tenant ID), or
  `gcp_workload_identity` (service-account email); per-trigger
  authentication beyond it lives in TriggerAuthentication resources next
  to the workloads
- **`spec.certificates`**: internal TLS (operator ↔ metrics server ↔
  webhooks) — `operator` (chart default; the operator self-generates
  certificates and patches the APIService caBundle) or `cert_manager` with
  an optional existing Issuer/ClusterIssuer reference
- **`spec.http_timeout_ms`**: default timeout for scalers that reach
  external services over raw HTTP (chart default 3000)
- **`spec.priority_class_name` / `node_selector` / `tolerations`**:
  scheduling for all KEDA components — the autoscaling engine should
  outlive workload evictions
- **`spec.prometheus`**: KEDA's OWN telemetry (scaler loop latencies,
  trigger errors) and optional ServiceMonitors (require the Prometheus
  operator CRDs — the release fails without them)
- **`spec.helm_values`**: escape hatch for chart values beyond the typed
  fields — never the primary interface

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Installation namespace (where the engine runs — ScaledObjects live next to their workloads) |
| `release_name` | Helm release name (always `keda`) |
| `operator_service_account_name` | Always `keda-operator` — the subject cloud-side keyless bindings (IRSA trust policies, GCP WI bindings, Entra federated credentials) are written against |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind KubernetesNamespace,
  field path `spec.name`).
- **`spec.certificates.cert_manager_issuer`** names its grain by arm:
  `issuer.name` is a foreign key to a KubernetesIssuer
  (`status.outputs.issuer_name`), `cluster_issuer.name` to a
  KubernetesClusterIssuer (`status.outputs.cluster_issuer_name`) — the whole
  chain (cert-manager → issuer → KEDA) then deploys in dependency order.
- **Cloud-side keyless identity** closes over the
  `operator_service_account_name` output: IRSA trust policies, Entra
  federated credentials, and GCP Workload Identity bindings all name the
  `keda-operator` service account in the installation namespace.
- **ScaledObjects need no reference to this component** — they are
  cluster-visible custom resources the operator watches; deploy them with
  the workloads they scale once the engine is up.

## Examples

### Minimal (cluster-wide engine, upstream defaults)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKeda
metadata:
  name: keda
spec:
  namespace:
    value: keda
  createNamespace: true
```

### EKS with IRSA for AWS scalers

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKeda
metadata:
  name: keda
spec:
  namespace:
    value: keda
  createNamespace: true
  podIdentity:
    awsIrsa:
      enabled: true
      roleArn: arn:aws:iam::111111111111:role/keda-scalers
  webhooks:
    failurePolicy: Fail
```

### HA production (standby replicas, cert-manager TLS, telemetry)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKeda
metadata:
  name: keda
spec:
  namespace:
    value: keda
  createNamespace: true
  operator:
    replicas: 2
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: "1"
        memory: 512Mi
  metricsServer:
    replicas: 2
  webhooks:
    replicas: 2
  priorityClassName: system-cluster-critical
  httpTimeoutMs: 5000
  certificates:
    type: cert_manager
    certManagerIssuer:
      clusterIssuer:
        name:
          valueFrom:
            name: platform-ca
  prometheus:
    enabled: true
    serviceMonitor: true # requires the Prometheus operator CRDs
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
