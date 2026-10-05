# Kubernetes Kube Prometheus Stack

## When NOT to Use This

**One resource is ONE monitoring stack, and one stack per cluster is
the norm.** The chart installs the monitoring.coreos.com CRDs
(ServiceMonitor, PodMonitor, PrometheusRule, Prometheus, Alertmanager,
...), which are cluster-scoped singletons — a second stack on the same
cluster must set `skip_crds` and fence its discovery, an advanced
posture, not the default.

Also not the right kind when:

- **You only want dashboards** — deploy `KubernetesGrafana` and point
  it at existing datasources. The stack's bundled Grafana exists to
  serve THIS stack's dashboards; the standalone kind is the
  composition hub for many datasources, external state and HA.
- **You want logs or traces** — this is the metrics-and-alerting
  stack. Loki (logs) and Tempo (traces) are their own components;
  their Grafana datasources compose here or on a standalone Grafana.
- **You expect a managed metrics backend** — this runs the
  open-source stack ON the cluster. For Amazon Managed Prometheus /
  Azure Monitor / Grafana Cloud as the long-term store, deploy this
  stack for scraping and use `prometheus.remote_write` to ship
  samples there.
- **You expect a public endpoint out of the box** — every UI/API
  stays ClusterIP; exposure composes from first-class kinds
  (KubernetesIngress, Gateway API kinds) over the exported service
  handles.

## Overview

**KubernetesKubePrometheusStack** deploys the kube-prometheus-stack —
the industry-standard cluster monitoring bundle — from the official
`kube-prometheus-stack` Helm chart
(https://prometheus-community.github.io/helm-charts). One install
gives the Prometheus Operator, a Prometheus server declared through
it, an Alertmanager, kube-state-metrics and node-exporter, a curated
set of Kubernetes alerting/recording rules, and a bundled Grafana
pre-loaded with the matching dashboards.

**Key design points:**

- **Naming contract.** The modules pin the chart's fullname to
  `metadata.name`, so child names are deterministic:
  `<name>-operator`, `<name>-prometheus`, `<name>-alertmanager`,
  `<name>-grafana`. The chart silently truncates its fullname at 26
  characters — both modules fail loudly instead, so keep the resource
  name at 26 characters or fewer.
- **CRD lifecycle.** The CRDs ship in the chart's `crds` subchart:
  Helm installs them ONCE, never upgrades them, and KEEPS them on
  uninstall (ServiceMonitors and rules across the cluster survive
  removal of the stack). Pair chart upgrades that cross operator
  versions with `crd_upgrade_job`, or apply the new CRD bundle
  manually first.
- **Cluster-wide discovery by default.** This component discovers
  EVERY ServiceMonitor/PodMonitor/PrometheusRule/Probe/ScrapeConfig
  in the cluster — deliberately wider than the chart's release-fenced
  default (upstream's most-tripped-over behavior). That is what makes
  every other component's `service_monitor_enabled` toggle light up
  with zero wiring. `discovery: release_managed_only` restores the
  chart's fence for multi-tenant clusters.
- **Persistent by default.** The chart's own default for Prometheus
  and Alertmanager is an emptyDir — metrics and silences vanish on
  restart — so this component provisions PVCs by default
  (`disk_size`, per replica). `ephemeral: true` restores the chart
  posture for throwaway clusters.
- **Managed-cloud truth.** On EKS/GKE/AKS the controller-manager,
  scheduler and etcd are provider-internal and unreachable — leaving
  their scrapers on produces permanently-down targets and alerts that
  can never resolve. `control_plane_scrapers` toggles them off, and
  `default_rules.disabled_groups` silences the matching rule groups
  (the presets carry the exact set).
- **One alert at a time.** `default_rules.disabled_alerts` leaves single
  curated alerts out (the overcommit pair on autoscaled node pools), and
  `default_rules.alert_overrides` changes a kept alert's hold (`for`) or
  `severity`. Replace an alert that misreads your cluster with your own
  `KubernetesPrometheusRule` of the same name rather than only muting it.
- **Remote write covers the cloud backends.** Basic auth
  (Grafana Cloud/Mimir), bearer tokens, AWS SigV4 (Amazon Managed
  Prometheus — keyless via IRSA or static keys), and Azure AD managed
  identity. Passwords/tokens/keys ride existing Secrets; usernames
  declared for basic auth are materialized by the module into the
  `<name>-remote-write-auth` Secret because the Prometheus CRD wants
  BOTH halves from Secrets.
- **`helm_values` is the escape hatch** — additional chart values
  merged LAST over everything the typed fields render (Helm `-f`
  semantics, identical on both engines): Thanos sidecar/ruler,
  windows monitoring, scrape classes, per-component securityContexts.
  Never for secrets.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to install into — literal or a
  KubernetesNamespace reference (`create_namespace` to own it)

### Common

- **`spec.chart_version`**: chart pin (default `91.8.2`, pairing
  Prometheus Operator v0.94.1); CRD upgrades do NOT ride chart bumps
  — see `crd_upgrade_job` (a stack installed on an older pin needs it
  for the apply that moves it to this one)
- **`spec.prometheus`**: replicas (HA duplication, not sharding),
  `retention` / `retention_size`, `disk_size` + `storage_class`,
  `resources` (memory scales with active series), `external_labels`
  (set `cluster`), scrape/evaluation intervals, `discovery`,
  `remote_write`, `enable_remote_write_receiver`,
  `additional_scrape_configs` (the exotic-SD seam), scheduling
- **`spec.alertmanager`**: enabled, replicas (3 = quorum HA),
  retention, disk_size, `notifications` (typed delivery: Discord,
  Pushover and webhook receivers, a routing tree, a dead-man's-switch
  heartbeat; every credential a managed-secret reference the module
  mounts from its own Secret), or `config_yaml` for integrations not
  typed yet (never inline a credential there)
- **`spec.grafana`**: enabled (default true), `admin_secret`
  (existing) or chart-generated credentials in `<name>-grafana`,
  `default_dashboards_enabled`, storage, resources
- **`spec.operator`**: resources, `admission_webhooks` (default
  certgen hook Job; `cert_manager: true` for the cert-manager arm;
  `disabled` only where the certificate machinery cannot run)
- **`spec.exporters`**: kube-state-metrics and node-exporter toggles
  and resources
- **`spec.control_plane_scrapers` / `spec.default_rules`**: the
  managed-cloud pair — disable unreachable scrapers AND their rule
  groups together; `default_rules.disabled_alerts` and
  `alert_overrides` tune single curated alerts
- **`spec.image_registry` / `spec.image_pull_secrets` /
  `spec.helm_values`**: the air-gap path and the escape hatch

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the stack runs in |
| `release_name` | Helm release name (= `metadata.name`) |
| `prometheus_service` | The Prometheus Service (`<name>-prometheus`, port 9090) |
| `prometheus_endpoint` | In-cluster Prometheus URL — what Grafana datasources and API clients use |
| `alertmanager_service` / `alertmanager_endpoint` | Alertmanager handles (empty when disabled) |
| `grafana_service` / `grafana_endpoint` | Bundled-Grafana handles (empty when disabled) |
| `grafana_admin_secret_name` | The bundled Grafana's credentials Secret (`<name>-grafana`, keys `admin-user`/`admin-password`) |
| `prometheus_port_forward_command` / `grafana_port_forward_command` | Workstation access without composed exposure |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace); the storage classes reference
  KubernetesStorageClass.
- **A standalone `KubernetesGrafana` points at this stack** by
  referencing it in a datasource `url` — the FK resolves to
  `prometheus_endpoint`, the one-line wiring that gives dashboards
  cluster metrics.
- **Every other component's `service_monitor_enabled`** lights up
  against this stack's default cluster-wide discovery — deploy the
  stack once, and monitors across all namespaces are scraped without
  extra wiring.
- **Exposure composes, never embeds**: KubernetesIngress or a Gateway
  API route over `grafana_service` / `prometheus_service`.

## Examples

The smallest declarable stack is a namespace alone — every other
field has a working default (single persistent Prometheus, 10-day
retention, Alertmanager, bundled Grafana, both exporters,
cluster-wide discovery). A production shape on a managed cloud:

### Managed-cloud stack with remote write

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKubePrometheusStack
metadata:
  name: cluster-metrics
spec:
  namespace:
    value: monitoring
  create_namespace: true
  prometheus:
    replicas: 2
    disk_size: 100Gi
    retention: 15d
    retention_size: 80GiB
    external_labels:
      cluster: prod-us-east
    remote_write:
      - url: https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-abc/api/v1/remote_write
        sigv4:
          region: us-east-1
  alertmanager:
    replicas: 3
  control_plane_scrapers:
    kube_controller_manager: false
    kube_etcd: false
    kube_scheduler: false
    kube_proxy: false
  default_rules:
    disabled_groups:
      - etcd
      - kubeControllerManager
      - kubeSchedulerAlerting
      - kubeSchedulerRecording
      - kubeProxy
    disabled_alerts:
      - KubeCPUOvercommit
      - KubeMemoryOvercommit
    alert_overrides:
      - alert: KubePodNotReady
        for: 30m
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
