# kube-prometheus-stack

Deploys kube-prometheus-stack — the industry-standard cluster monitoring bundle — from the official `kube-prometheus-stack` Helm chart. One resource is the whole observability plane: the **Prometheus Operator** (the controller), an HA-capable **Prometheus** server on persistent volumes, an **Alertmanager**, the **kube-state-metrics** and **node-exporter** metric sources, the chart's curated Kubernetes alerting/recording rules, and a **bundled Grafana** pre-loaded with the matching dashboards.

The grain is deliberate: **one stack per cluster is the norm**. The chart installs the `monitoring.coreos.com` CRDs (ServiceMonitor, PodMonitor, PrometheusRule, Prometheus, Alertmanager, ...), which are cluster-scoped singletons — a second stack on the same cluster must set `skipCrds` and fence its discovery, an advanced posture, never the default. Every UI stays ClusterIP; external reachability composes from ingress and Gateway API kinds over the exported service handles.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Helm release** (official `kube-prometheus-stack` chart, default pin `91.8.2`, named `metadata.name`) — the operator Deployment, the Prometheus StatefulSet on a 50Gi PVC per replica (the chart's own emptyDir default is deliberately overridden), the Alertmanager StatefulSet on a 2Gi PVC, kube-state-metrics, the node-exporter DaemonSet, the curated PrometheusRule set, and the bundled Grafana with its pre-wired datasource and dashboards
- **The monitoring.coreos.com CRDs** — installed ONCE by Helm and never touched again: chart upgrades do NOT upgrade CRDs (pair version bumps across operator minors with `crdUpgradeJob`, a pre-upgrade hook that server-side-applies the new bundle), and uninstall KEEPS them, so ServiceMonitors and rules across the cluster survive removal of the stack
- **Grafana admin Secret** (`<name>-grafana`, keys `admin-user`/`admin-password`) — generated once at first install when `grafana.adminSecret` is empty, stable across upgrades, surfaced in the outputs
- **Kubernetes Namespace** — created only when `createNamespace` is true; otherwise the namespace must already exist

The modules pin the chart's fullname to `metadata.name` so child Services stay predictable (`<name>-prometheus`, `<name>-grafana`). **Keep the resource name at 26 characters or fewer** — the chart silently truncates its fullname past that, so both modules fail loudly on a longer name instead of letting the naming contract break.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **A StorageClass** for the Prometheus, Alertmanager, and (optional) Grafana volumes — most managed clusters provide a default; reference a **Kubernetes Storage Class** for explicit (SSD) placement. The TSDB rewards SSD-backed classes.
- **cert-manager** — only if you set `operator.admissionWebhooks.certManager`; the chart then renders an Issuer and Certificates for the webhook instead of its self-contained certgen Job.

## Deploy

### Console

Open the deployment store, find **kube-prometheus-stack**, and click **Deploy**. The creation wizard walks you through namespace placement (with the live naming-budget warning), the chart pin and the CRD lifecycle, the Prometheus server and its storage, the discovery fence, remote-write destinations, Alertmanager, the bundled Grafana, the operator and its admission webhooks, the exporters, the managed-cloud scraper set, the air-gap path, placement, and the Helm-values escape hatch. Start from the **Managed cloud preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKubePrometheusStack
metadata:
  name: cluster-metrics
  org: acme-corp
  env: prod
spec:
  namespace:
    value: monitoring
  createNamespace: true
```

```shell
planton apply -f kube-prometheus-stack.yaml
```

This empty-spec install is a complete monitoring plane: a single Prometheus with 10-day retention on a 50Gi volume, Alertmanager, both exporters, the curated rules, the bundled Grafana with generated credentials, and cluster-wide monitor discovery. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, compose the stack behind its namespace with a reference, and the InfraPipeline orders the deploys:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: monitoring-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline creates the namespace first, then installs the stack into it.

## Key Configuration

These are the most important decisions when configuring a kube-prometheus-stack deployment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Discovery is deliberately wider than the chart's own default** — unset, this component discovers EVERY ServiceMonitor, PodMonitor, PrometheusRule, Probe, and ScrapeConfig in the cluster, whoever created it. The chart's own default only discovers objects labeled by its release (upstream's most-tripped-over behavior); cluster-wide discovery is what makes every catalog kind's `service_monitor_enabled` toggle and any hand-authored monitor light up with zero wiring. Set `discovery: release_managed_only` to restore the chart's fence for deliberate multi-tenant ownership boundaries.

**On managed clouds, disable the scrapers that can never succeed** — on EKS/GKE/AKS the controller-manager, etcd, and scheduler run on provider machines the cluster network can never reach; leaving their scrapers on produces permanently-down targets and alerts that can never recover (kube-proxy's metrics port is often localhost-bound too). Disable them via `controlPlaneScrapers` and pair each with its rule group in `defaultRules.disabledGroups`, so the alert set stays truthful. The **Managed cloud preset** carries the exact set.

**Tune one curated alert without losing its group** — `defaultRules.disabledAlerts` leaves single alerts out and `defaultRules.alertOverrides` changes a kept alert's hold (`for`) or `severity`. Node pools that autoscale are the common case: the two overcommit alerts count today's nodes and fire on every burst the autoscaler absorbs, so the presets turn them off. For an alert that misreads work your cluster does on purpose (CI build pods, busy build machines), replace it: disable it here and declare the same alert name in a **KubernetesPrometheusRule** that leaves that work out. A name that matches no rule is skipped silently, so read the loaded rules back after a change.

**Alertmanager's default notifies NOBODY — declare `notifications`** — with nothing declared, everything routes to a `"null"` receiver: alerts are visible in the UIs and APIs but reach no one. `alertmanager.notifications` types the delivery: Discord, Pushover and webhook receivers, a routing tree with label matchers, and a **heartbeat** that posts the always-firing Watchdog alert to an outside dead-man's-switch monitor every minute, so a dead Alertmanager or cluster is noticed. Every credential is a managed-secret reference the module mounts from its own Secret; nothing lands in chart values. Messages lead with environment and component and never render a namespace, pod or description, so no alert names a customer. `configYaml` remains for integrations not typed yet (never inline a credential there). Mind the routing trap: the root receiver only takes alerts no child route claims.

**Remote write ships samples as they are scraped** — each destination carries one auth arm: basic auth (password in an existing Secret), a bearer-token Secret, **AWS SigV4 where NO keys means the pod's ambient IRSA identity** (the recommended keyless posture for Amazon Managed Prometheus; static keys are the both-or-neither fallback), or Azure AD managed identity. Name each destination when declaring more than one — the queue-metrics identity. `enableRemoteWriteReceiver` is the other direction: accept pushes FROM other Prometheus servers, turning this one into an aggregation point.

**Prometheus replicas duplicate, they never shard** — each replica scrapes and stores the FULL target set independently; 2 is the HA pair, and queries against the Service may see either replica. Deduplicated long-term storage is the Thanos surface, reached through `helmValues`. Retention is dual-bounded: `retention` (default 10d) by age and `retentionSize` by disk — size the latter BELOW the volume, because a full volume crash-loops instead of trimming.

**The bundled Grafana is the stack's dashboard console, not your fleet Grafana** — it arrives pre-wired with a datasource for this Prometheus and the curated dashboard set, with admin credentials generated once and exported. For a standalone, multi-datasource Grafana, disable it here and deploy **KubernetesGrafana** instead, pointed at this stack's exported Prometheus endpoint. Give the bundled one a `storage` block if people build dashboards by hand — the ephemeral default loses them with the pod.

**Nothing is exposed by default** — Prometheus, Alertmanager, and Grafana stay ClusterIP. Expose them by composing first-class kinds (KubernetesIngress, the Gateway API kinds) over the exported service handles; the stack never opens its own doors.

**`helmValues` merges last** — the escape hatch for chart surface beyond the typed fields (Thanos sidecar/ruler, windows monitoring, scrape classes, per-kind securityContexts). Anything here silently overrides the typed fields on every deploy; never put secrets in it, and leave `fullnameOverride` alone — the naming contract the outputs derive from depends on it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesStorageClass** | `prometheus.storageClass` | `metadata.name` |
| **KubernetesStorageClass** | `alertmanager.storageClass` | `metadata.name` |
| **KubernetesStorageClass** | `grafana.storage.storageClass` | `metadata.name` |

The stack's secret inputs — the bring-your-own Grafana `adminSecret`, remote-write basic-auth/bearer-token/SigV4 Secrets, and `imagePullSecrets` — are plain Secret name + key selectors resolved in the installation namespace, not typed references.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the stack runs in | Application deployment manifests |
| `release_name` | Helm release name (= metadata.name); every child name derives from it | Operational tooling |
| `prometheus_service` | The Prometheus Service (`<name>-prometheus`, port 9090) | ServiceMonitor-free scrape targets, diagnostics |
| `prometheus_endpoint` | In-cluster Prometheus URL | Grafana datasources, remote readers, API clients |
| `alertmanager_service` | The Alertmanager Service (`<name>-alertmanager`, port 9093). Empty when Alertmanager is disabled | External alert routing |
| `alertmanager_endpoint` | In-cluster Alertmanager URL. Empty when Alertmanager is disabled | Alert-source integrations |
| `grafana_service` | The bundled Grafana Service (`<name>-grafana`, port 80). Empty when the bundled Grafana is disabled | Ingress/Gateway exposure |
| `grafana_endpoint` | In-cluster Grafana URL. Empty when the bundled Grafana is disabled | Operator access, links |
| `grafana_admin_secret_name` | The Grafana admin-credentials Secret (`<name>-grafana`, keys `admin-user`/`admin-password`; echoes your own Secret's name when `adminSecret` is set). Empty when the bundled Grafana is disabled | Operator authentication |
| `prometheus_port_forward_command` | Copy-paste `kubectl port-forward` for the Prometheus UI | Local development access |
| `grafana_port_forward_command` | Copy-paste `kubectl port-forward` for Grafana. Empty when the bundled Grafana is disabled | Local development access |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev single cluster** — kind, a laptop lab, or a private single-node cluster: small volumes (10Gi/2Gi), 2-day retention, and the scrapers that cannot succeed on those platforms turned off with their rule groups paired. Start from the **Dev single cluster preset**.

**Managed cloud** — EKS / GKE / AKS production: a 50Gi Prometheus volume with 15-day retention, durable Alertmanager and Grafana state, and every provider-internal scraper off with its rule group disabled — the target list and alert set stay honest. Start from the **Managed cloud preset**.

**Production sized** — full HA: an HA Prometheus pair, a quorum-safe 3-replica Alertmanager gossip cluster, 30-day/80GiB dual-bounded retention on 100Gi volumes, and explicit resources on every component so the stack cannot be evicted by the workloads it watches. Start from the **Production sized preset**.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — referenced placement; the InfraPipeline orders namespace-first
- [**Kubernetes StorageClass**](/infra-catalog/kubernetes-storage-class) — SSD-backed classes for the TSDB and state volumes
- [**Kubernetes Secret**](/infra-catalog/kubernetes-secret) — bring-your-own Grafana credentials, remote-write credentials, image-pull Secrets
- [**Cert Manager**](/infra-catalog/kubernetes-cert-manager) — issues the operator's admission-webhook certificate when `certManager` is chosen
- [**Kubernetes Ingress**](/infra-catalog/kubernetes-ingress) — HTTP exposure over the exported Grafana and Prometheus handles (Gateway API kinds compose the same way)
- [**Grafana**](/infra-catalog/kubernetes-grafana) — the standalone, multi-datasource alternative to the bundled Grafana, consuming the exported `prometheus_endpoint`
- **Every catalog kind with a `service_monitor_enabled` toggle** — discovered automatically under the default cluster-wide discovery
