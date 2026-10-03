# SigNoz

Deploys SigNoz — the open-source, OpenTelemetry-native observability platform: traces, metrics and logs in ONE application with one UI, stored in ClickHouse. The one-component alternative to composing a metrics stack, a log store, a trace store and a dashboard tool separately (kube-prometheus-stack + Grafana + Loki + Tempo). Both paths are first-class on this platform; pick per team taste.

The telemetry store is COMPOSED, never bundled: this component installs nothing ClickHouse-related. Point `spec.clickhouse` at a **ClickHouse** resource by reference (its Service, cluster name and auth Secret ride `valueFrom` references), or at any reachable ClickHouse by literals. The reason is operational, verified live: a chart-bundled database cannot uninstall cleanly — the database operator and its installation die in the same Helm release, and the installation's finalizer deadlocks waiting for an operator that is already gone. Composed, each side gets its own lifecycle, its own sizing, and clean teardown.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions (via the official `signoz` Helm chart — the chart version tracks the SigNoz application version in lockstep):

- **SigNoz server** — UI, API, rule evaluation and alerting in one consolidated binary, with a small PVC (default 1Gi) holding users, dashboards and alert rules in embedded SQLite. Telemetry lives in ClickHouse, not here. Deliberately single-instance: SQLite is single-writer, so there is no replica knob (the Postgres-backed HA store is an enterprise-edition feature).
- **SigNoz OpenTelemetry Collector** — the ingestion gateway every application sends telemetry to. Stateless; scales horizontally by replica count or an optional HorizontalPodAutoscaler.
- **Schema migrator** — sync and async migration jobs that create and evolve the `signoz_*` databases in the composed ClickHouse on install and upgrade.
- **Kubernetes Namespace** — created only when `createNamespace` is true; otherwise the namespace must already exist.

All Services stay ClusterIP — external exposure composes from Ingress / Gateway API kinds over the exported handles.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **A running ClickHouse** — typically a **ClickHouse** resource (with its **Altinity ClickHouse Operator**), deployed first and referenced from `spec.clickhouse`. The contract it must honor:
- **Version pairing** — run ClickHouse at the version the chart pairs with (25.12.5 at chart 0.133.0). Older servers fail the sync migration with `Unknown setting 'object_serialization_version'` and the install never becomes ready.
- **User posture** — the migrator runs `ON CLUSTER` DDL. On a ClickHouse resource, a user declared with NO grants carries unrestricted config-user access, which covers it; a constrained user needs `GRANT CLUSTER ON *.*` plus CREATE/DROP/INSERT/SELECT on the `signoz_*` databases.
- **Networks** — a ClickHouse user declared without `networks` is fenced to the ClickHouse pods and localhost; SigNoz's pods are rejected with what reads as a password failure. Declare the user's networks explicitly.
- **Same namespace** — the ClickHouse password rides a `secretKeyRef`, which reads Secrets only in SigNoz's own namespace. Co-locate SigNoz with its ClickHouse, or replicate the auth Secret.

## Deploy

### Console

Open the deployment store, find **SigNoz**, and click **Deploy**. The creation wizard walks you through namespace placement, the chart pin, the ClickHouse connection and credentials (the composition star), server state, URL & email, advanced env, the collector and its receivers, cluster & images, placement, and the helm-values escape hatch. Start from the **SigNoz for development** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSignoz
metadata:
  name: observability
  org: acme-corp
  env: dev
spec:
  namespace:
    value: telemetry
  clickhouse:
    host:
      valueFrom:
        kind: KubernetesClickHouse
        name: telemetry
        fieldPath: status.outputs.service_name
    clusterName:
      valueFrom:
        kind: KubernetesClickHouse
        name: telemetry
        fieldPath: status.outputs.cluster_name
    username: signoz
    passwordSecret:
      secretName:
        valueFrom:
          kind: KubernetesClickHouse
          name: telemetry
          fieldPath: status.outputs.auth_secret_name
      secretKey: signoz
```

```shell
planton apply -f signoz.yaml
```

This creates the smallest honest SigNoz: the whole platform (UI, API, alerting, the ingestion collector) against a composed ClickHouse named `telemetry` in the same namespace, wired through three references — no password anywhere in the manifest. An Infra Job tracks the provisioning in real time.

### InfraChart

Compose SigNoz behind its namespace and its telemetry store with references, and the InfraPipeline orders the deploys:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: telemetry-namespace
      fieldPath: spec.name
  clickhouse:
    host:
      valueFrom:
        kind: KubernetesClickHouse
        name: telemetry
        fieldPath: status.outputs.service_name
```

The InfraPipeline resolves the dependency graph: the namespace and the ClickHouse (with its operator) deploy before SigNoz — no extra sequencing needed.

## Key Configuration

These are the most important decisions when configuring a SigNoz deployment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The chart pin IS the SigNoz version** — `chartVersion` (default `0.133.0`) tracks the application version in lockstep. Editing the pin is the day-2 upgrade; the migrator runs its ClickHouse migrations on the way up.

**How data gets in** — applications point at the collector, not the server. OTLP gRPC (4317) and OTLP HTTP (4318) are always on; Jaeger (default on) and Zipkin (default off) receivers cover legacy tracers, and plain-HTTP log paths (JSON 8082, Heroku drain 8081, default on) cover log shippers. Disable the legacy doors once everything speaks OTLP.

**The server is single-instance by design** — users, dashboards and alert rules live in embedded SQLite on a small PVC (1Gi default; telemetry is in ClickHouse). SQLite is single-writer, so the community server runs exactly one replica — the COLLECTOR is what scales, by `otelCollector.replicas` or the autoscaling block (the HPA owns the count when enabled).

**Collector limits are configuration** — the collector's memory limiter derives its ceiling from the container memory LIMIT. Set limits in production or an ingest spike grows the pod until the node evicts it.

**Reach** — `server.externalUrl` builds the links inside alert notifications and invitation emails (without it they point at localhost); `server.smtp` enables email delivery with the password as a Secret reference. Exposure itself composes from Ingress / Gateway API kinds.

**Helm value overrides** — `helmValues` merges LAST over everything the typed fields render (Helm `-f` semantics, identical on both IaC engines): the hatch for collector pipeline overrides, migrator tuning, extra mounts. Never secrets — every credential on this spec rides a Secret reference.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesClickHouse** | `clickhouse.host` | `status.outputs.service_name` |
| **KubernetesClickHouse** | `clickhouse.clusterName` | `status.outputs.cluster_name` |
| **KubernetesClickHouse** | `clickhouse.passwordSecret.secretName` | `status.outputs.auth_secret_name` |
| **KubernetesStorageClass** | `server.storageClass` | `metadata.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the platform runs in | Application deployment manifests |
| `service` | The SigNoz UI/API Service (ClusterIP) | Ingress/Gateway composition |
| `kube_endpoint` | In-cluster FQDN for the UI/API | Internal service links |
| `port_forward_command` | Copy-paste `kubectl port-forward` for the UI | Local development access |
| `otel_collector_service` | The ingestion collector's Service | Ingress/Gateway composition for external ingest |
| `otlp_grpc_endpoint` | In-cluster OTLP gRPC endpoint (port 4317) | Application OTLP gRPC exporter configuration |
| `otlp_http_endpoint` | In-cluster OTLP HTTP endpoint (port 4318) | Application OTLP HTTP exporter configuration |
| `clickhouse_endpoint` | The composed ClickHouse endpoint SigNoz writes to | Downstream tooling that composes against the same telemetry store |

The outputs also mirror the declared ClickHouse connection back (`clickhouse_username` and the `clickhouse_password_secret` name + key) — passthroughs of what the spec declared, for tooling that follows the composition, never module-generated values.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Development** — the smallest honest SigNoz: the component's defaults against a composed ClickHouse named `telemetry` in the same namespace, wired through three references. Start from the **SigNoz for development** preset.

**Production** — verified TLS to ClickHouse, alert email over secret-safe SMTP, the external URL that makes alert links resolve, sized server resources, and an autoscaling ingestion collector. Start from the **SigNoz for production** preset.

## Works With

- [**ClickHouse**](/infra-catalog/kubernetes-click-house) — the composed telemetry store; deploy it first and wire it by reference.
- [**Altinity ClickHouse Operator**](/infra-catalog/kubernetes-altinity-operator) — the operator that reconciles that ClickHouse; the first link of the composition chain.
- [**OpenTelemetry Collector**](/infra-catalog/kubernetes-otel-collector) — daemonset-mode cluster telemetry (node logs, kubelet metrics, Kubernetes events) pointed at the exported OTLP endpoints.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — referenced placement; the InfraPipeline orders namespace-first.
- [**Kubernetes StorageClass**](/infra-catalog/kubernetes-storage-class) — explicit class for the server's state volume.
- [**Kubernetes Ingress**](/infra-catalog/kubernetes-ingress) and [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) — external exposure for the UI and the collector over the exported Service handles (everything stays ClusterIP by design).
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack), [**Grafana**](/infra-catalog/kubernetes-grafana), [**Grafana Loki**](/infra-catalog/kubernetes-loki), [**Grafana Tempo**](/infra-catalog/kubernetes-tempo) — the composed-stack alternative; run one path or the other per team taste.
