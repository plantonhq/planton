# KubernetesKubePrometheusStack

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesKubePrometheusStackSpec** deploys the kube-prometheus-stack —
the industry-standard cluster monitoring bundle — from the official
`kube-prometheus-stack` Helm chart
(https://prometheus-community.github.io/helm-charts). One install gives
the Prometheus Operator, a Prometheus server declared through it, an
Alertmanager, the kube-state-metrics and node-exporter metric sources,
a curated set of Kubernetes alerting/recording rules, and a bundled
Grafana pre-loaded with the matching dashboards.

GRAIN: one stack per cluster is the norm. The chart installs the
monitoring.coreos.com CRDs (ServiceMonitor, PodMonitor, PrometheusRule,
Prometheus, Alertmanager, ...) which are cluster-scoped singletons — a
second stack on the same cluster must set `skip_crds` and scope its
monitor discovery, which is an advanced posture, not the default.

NAMING: keep the resource name at 26 characters or fewer. The chart
SILENTLY truncates its fullname at 26 characters (its own headroom for
the longest child name it derives), and the modules pin the fullname to
the resource name so service names like `<name>-prometheus` and
`<name>-grafana` are predictable — a longer name would be truncated and
break that naming contract. Both modules fail loudly instead of letting
the chart truncate.

CRD LIFECYCLE: the CRDs ship in a subchart whose `crds/` directory Helm
installs ONCE and never touches again — chart upgrades do NOT upgrade
CRDs, and uninstall KEEPS them (ServiceMonitors and rules across the
cluster survive removal of the stack). When bumping `chart_version`
across operator minors, enable `crd_upgrade_job` for that upgrade (a
pre-upgrade hook that server-side-applies the new CRD bundle) or apply
the new CRDs manually before upgrading.

DISCOVERY: by default this component discovers EVERY ServiceMonitor,
PodMonitor, PrometheusRule, Probe and ScrapeConfig in the cluster —
deliberately wider than the chart's own default (which only discovers
objects labeled by its release, upstream's most-tripped-over behavior).
Cluster-wide discovery is what makes every catalog component's
`service_monitor_enabled` toggle and any user-authored monitor light up
without extra wiring. Set `discovery` to `release_managed_only` to get
the chart's fenced default back.

MANAGED CLOUDS: on EKS/GKE/AKS the control plane is invisible to the
cluster — the controller-manager, scheduler, etcd and (on some
distributions) kube-proxy scrape targets can never be reached, leaving
permanently-down targets and firing alerts. Disable those scrapers via
`control_plane_scrapers` (the per-cloud presets show the exact set).

EXPOSURE: every UI/API stays ClusterIP; expose Grafana or Prometheus
via first-class kinds (KubernetesIngress, Gateway API kinds) over the
exported service handles.

The typed fields below cover the chart's meaningful configuration
surface; `helm_values` remains as the escape hatch for chart values
beyond them (merged last, Helm `-f` semantics, identical on both
engines) — Thanos sidecar/ruler scale-out, windows monitoring, per-arm
webhook tuning, additional alertmanager templates — a safety valve,
never the primary interface.

## Example

```yaml
# Full-surface offline-proof manifest: exercises an HA Prometheus pair with
# retention/size trimming and a tuned PVC, external labels, all four
# remote-write auth arms (basic auth with the module-materialized username
# Secret, bearer token, keyless SigV4 with a role, Azure managed identity),
# the raw scrape-config seam, a 3-replica Alertmanager with a routed config,
# the bundled Grafana with an existing admin Secret and persistence,
# operator sizing with cert-manager-issued webhook certificates, exporter
# sizing, the managed-cloud scraper posture with its matching rule-group
# disables, the CRD upgrade hook, global registry + pull-secret plumbing,
# per-component scheduling, and an escape-hatch entry — so the offline tofu
# plan and pulumi preview proofs cover the full typed surface. Placeholder
# values; never applied to a real cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKubePrometheusStack
metadata:
  name: kps-hack
spec:
  namespace:
    value: kps-hack
  createNamespace: true
  chartVersion: 87.19.1
  crdUpgradeJob: true
  prometheus:
    replicas: 2
    retention: 30d
    retentionSize: 180GiB
    diskSize: 200Gi
    storageClass:
      value: fast-ssd
    resources:
      requests:
        cpu: "1"
        memory: 4Gi
      limits:
        cpu: "4"
        memory: 16Gi
    externalLabels:
      cluster: hack-cluster
      region: us-east-1
    scrapeInterval: 30s
    evaluationInterval: 1m
    discovery: all_monitors
    enableRemoteWriteReceiver: true
    remoteWrite:
      - url: https://prometheus-us-central1.grafana.net/api/prom/push
        name: grafana-cloud
        basicAuth:
          username: "123456"
          passwordSecret:
            name: grafana-cloud-credentials
            key: token
      - url: https://mimir.example.com/api/v1/push
        name: mimir
        bearerTokenSecret:
          name: mimir-credentials
          key: token
      - url: https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-hack/api/v1/remote_write
        name: amp
        sigv4:
          region: us-east-1
          roleArn: arn:aws:iam::123456789012:role/amp-remote-write
      - url: https://hack.eastus-1.metrics.ingest.monitor.azure.com/dataCollectionRules/dcr-hack/streams/Microsoft-PrometheusMetrics/api/v1/write
        name: azure-monitor
        azureAd:
          managedIdentityClientId: 11111111-2222-3333-4444-555555555555
    additionalScrapeConfigs: |
      - job_name: legacy-appliance
        static_configs:
          - targets:
              - 10.0.0.42:9100
    scheduling:
      nodeSelector:
        workload: monitoring
      tolerations:
        - key: dedicated
          operator: Equal
          value: monitoring
          effect: NoSchedule
      priorityClassName: system-cluster-critical
  alertmanager:
    replicas: 3
    retention: 240h
    diskSize: 5Gi
    storageClass:
      value: fast-ssd
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
    configYaml: |
      route:
        group_by:
          - namespace
        receiver: team-webhook
        routes:
          - receiver: "null"
            matchers:
              - alertname = "Watchdog"
      receivers:
        - name: "null"
        - name: team-webhook
          webhook_configs:
            - url_file: /etc/alertmanager/secrets/team-webhook/url
    scheduling:
      nodeSelector:
        workload: monitoring
  grafana:
    adminSecret:
      name: kps-hack-grafana-admin
      userKey: user
      passwordKey: password
    defaultDashboardsEnabled: true
    storage:
      size: 20Gi
      storageClass:
        value: fast-ssd
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: "1"
        memory: 1Gi
  operator:
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
    admissionWebhooks:
      certManager: true
    scheduling:
      nodeSelector:
        workload: monitoring
  exporters:
    kubeStateMetricsEnabled: true
    kubeStateMetricsResources:
      requests:
        cpu: 50m
        memory: 64Mi
      limits:
        cpu: 250m
        memory: 256Mi
    nodeExporterEnabled: true
    nodeExporterResources:
      requests:
        cpu: 50m
        memory: 32Mi
      limits:
        cpu: 250m
        memory: 128Mi
  controlPlaneScrapers:
    kubeControllerManager: false
    kubeEtcd: false
    kubeScheduler: false
    kubeProxy: false
  defaultRules:
    disabledGroups:
      - etcd
      - kubeControllerManager
      - kubeSchedulerAlerting
      - kubeSchedulerRecording
      - kubeProxy
  imageRegistry: mirror.example.com
  imagePullSecrets:
    - mirror-pull
  helmValues: |
    prometheus:
      prometheusSpec:
        walCompression: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `87.19.1` |  |
| `spec.skipCrds` | `bool` |  |  |  |
| `spec.crdUpgradeJob` | `bool` |  |  |  |
| `spec.prometheus` | `KubernetesKubePrometheusStackPrometheus` |  |  |  |
| `spec.prometheus.replicas` | `int32` |  | `1` |  |
| `spec.prometheus.retention` | `string` |  | `10d` |  |
| `spec.prometheus.retentionSize` | `string` |  |  |  |
| `spec.prometheus.diskSize` | `string` |  | `50Gi` |  |
| `spec.prometheus.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.prometheus.ephemeral` | `bool` |  |  |  |
| `spec.prometheus.resources` | `ContainerResources` |  |  |  |
| `spec.prometheus.resources.limits` | `CpuMemory` |  |  |  |
| `spec.prometheus.resources.limits.cpu` | `string` |  |  |  |
| `spec.prometheus.resources.limits.memory` | `string` |  |  |  |
| `spec.prometheus.resources.requests` | `CpuMemory` |  |  |  |
| `spec.prometheus.resources.requests.cpu` | `string` |  |  |  |
| `spec.prometheus.resources.requests.memory` | `string` |  |  |  |
| `spec.prometheus.externalLabels` | `map<string, string>` |  |  |  |
| `spec.prometheus.scrapeInterval` | `string` |  |  |  |
| `spec.prometheus.evaluationInterval` | `string` |  |  |  |
| `spec.prometheus.discovery` | `enum` |  | `all_monitors` |  |
| `spec.prometheus.remoteWrite` | `[]KubernetesKubePrometheusStackRemoteWrite` |  |  |  |
| `spec.prometheus.remoteWrite[].url` | `string` | yes |  |  |
| `spec.prometheus.remoteWrite[].name` | `string` |  |  |  |
| `spec.prometheus.remoteWrite[].basicAuth` | `KubernetesKubePrometheusStackRemoteWriteBasicAuth` |  |  |  |
| `spec.prometheus.remoteWrite[].basicAuth.username` | `string` | yes |  |  |
| `spec.prometheus.remoteWrite[].basicAuth.passwordSecret` | `KubernetesKubePrometheusStackSecretKeyRef` | yes |  |  |
| `spec.prometheus.remoteWrite[].basicAuth.passwordSecret.name` | `string` | yes |  |  |
| `spec.prometheus.remoteWrite[].basicAuth.passwordSecret.key` | `string` | yes |  |  |
| `spec.prometheus.remoteWrite[].bearerTokenSecret` | `KubernetesKubePrometheusStackSecretKeyRef` |  |  |  |
| `spec.prometheus.remoteWrite[].bearerTokenSecret.name` | `string` | yes |  |  |
| `spec.prometheus.remoteWrite[].bearerTokenSecret.key` | `string` | yes |  |  |
| `spec.prometheus.remoteWrite[].sigv4` | `KubernetesKubePrometheusStackRemoteWriteSigv4` |  |  |  |
| `spec.prometheus.remoteWrite[].sigv4.region` | `string` | yes |  |  |
| `spec.prometheus.remoteWrite[].sigv4.roleArn` | `string` |  |  |  |
| `spec.prometheus.remoteWrite[].sigv4.accessKeySecret` | `KubernetesKubePrometheusStackSecretKeyRef` |  |  |  |
| `spec.prometheus.remoteWrite[].sigv4.accessKeySecret.name` | `string` | yes |  |  |
| `spec.prometheus.remoteWrite[].sigv4.accessKeySecret.key` | `string` | yes |  |  |
| `spec.prometheus.remoteWrite[].sigv4.secretKeySecret` | `KubernetesKubePrometheusStackSecretKeyRef` |  |  |  |
| `spec.prometheus.remoteWrite[].sigv4.secretKeySecret.name` | `string` | yes |  |  |
| `spec.prometheus.remoteWrite[].sigv4.secretKeySecret.key` | `string` | yes |  |  |
| `spec.prometheus.remoteWrite[].azureAd` | `KubernetesKubePrometheusStackRemoteWriteAzureAd` |  |  |  |
| `spec.prometheus.remoteWrite[].azureAd.managedIdentityClientId` | `string` | yes |  |  |
| `spec.prometheus.remoteWrite[].azureAd.cloud` | `string` |  |  |  |
| `spec.prometheus.enableRemoteWriteReceiver` | `bool` |  |  |  |
| `spec.prometheus.additionalScrapeConfigs` | `string` |  |  |  |
| `spec.prometheus.scheduling` | `KubernetesKubePrometheusStackScheduling` |  |  |  |
| `spec.prometheus.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.prometheus.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.prometheus.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.prometheus.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.prometheus.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.prometheus.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.prometheus.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.prometheus.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.alertmanager` | `KubernetesKubePrometheusStackAlertmanager` |  |  |  |
| `spec.alertmanager.enabled` | `bool` |  | `true` |  |
| `spec.alertmanager.replicas` | `int32` |  | `1` |  |
| `spec.alertmanager.retention` | `string` |  | `120h` |  |
| `spec.alertmanager.diskSize` | `string` |  | `2Gi` |  |
| `spec.alertmanager.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.alertmanager.ephemeral` | `bool` |  |  |  |
| `spec.alertmanager.resources` | `ContainerResources` |  |  |  |
| `spec.alertmanager.resources.limits` | `CpuMemory` |  |  |  |
| `spec.alertmanager.resources.limits.cpu` | `string` |  |  |  |
| `spec.alertmanager.resources.limits.memory` | `string` |  |  |  |
| `spec.alertmanager.resources.requests` | `CpuMemory` |  |  |  |
| `spec.alertmanager.resources.requests.cpu` | `string` |  |  |  |
| `spec.alertmanager.resources.requests.memory` | `string` |  |  |  |
| `spec.alertmanager.configYaml` | `string` |  |  |  |
| `spec.alertmanager.scheduling` | `KubernetesKubePrometheusStackScheduling` |  |  |  |
| `spec.alertmanager.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.alertmanager.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.alertmanager.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.alertmanager.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.alertmanager.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.alertmanager.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.alertmanager.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.alertmanager.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.grafana` | `KubernetesKubePrometheusStackGrafana` |  |  |  |
| `spec.grafana.enabled` | `bool` |  | `true` |  |
| `spec.grafana.adminSecret` | `KubernetesKubePrometheusStackGrafanaAdminSecret` |  |  |  |
| `spec.grafana.adminSecret.name` | `string` | yes |  |  |
| `spec.grafana.adminSecret.userKey` | `string` |  | `admin-user` |  |
| `spec.grafana.adminSecret.passwordKey` | `string` |  | `admin-password` |  |
| `spec.grafana.defaultDashboardsEnabled` | `bool` |  | `true` |  |
| `spec.grafana.storage` | `KubernetesKubePrometheusStackGrafanaStorage` |  |  |  |
| `spec.grafana.storage.size` | `string` |  | `10Gi` |  |
| `spec.grafana.storage.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.grafana.resources` | `ContainerResources` |  |  |  |
| `spec.grafana.resources.limits` | `CpuMemory` |  |  |  |
| `spec.grafana.resources.limits.cpu` | `string` |  |  |  |
| `spec.grafana.resources.limits.memory` | `string` |  |  |  |
| `spec.grafana.resources.requests` | `CpuMemory` |  |  |  |
| `spec.grafana.resources.requests.cpu` | `string` |  |  |  |
| `spec.grafana.resources.requests.memory` | `string` |  |  |  |
| `spec.operator` | `KubernetesKubePrometheusStackOperator` |  |  |  |
| `spec.operator.resources` | `ContainerResources` |  |  |  |
| `spec.operator.resources.limits` | `CpuMemory` |  |  |  |
| `spec.operator.resources.limits.cpu` | `string` |  |  |  |
| `spec.operator.resources.limits.memory` | `string` |  |  |  |
| `spec.operator.resources.requests` | `CpuMemory` |  |  |  |
| `spec.operator.resources.requests.cpu` | `string` |  |  |  |
| `spec.operator.resources.requests.memory` | `string` |  |  |  |
| `spec.operator.admissionWebhooks` | `KubernetesKubePrometheusStackAdmissionWebhooks` |  |  |  |
| `spec.operator.admissionWebhooks.disabled` | `bool` |  |  |  |
| `spec.operator.admissionWebhooks.certManager` | `bool` |  |  |  |
| `spec.operator.scheduling` | `KubernetesKubePrometheusStackScheduling` |  |  |  |
| `spec.operator.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.operator.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.operator.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.operator.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.operator.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.operator.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.operator.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.operator.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.exporters` | `KubernetesKubePrometheusStackExporters` |  |  |  |
| `spec.exporters.kubeStateMetricsEnabled` | `bool` |  | `true` |  |
| `spec.exporters.kubeStateMetricsResources` | `ContainerResources` |  |  |  |
| `spec.exporters.kubeStateMetricsResources.limits` | `CpuMemory` |  |  |  |
| `spec.exporters.kubeStateMetricsResources.limits.cpu` | `string` |  |  |  |
| `spec.exporters.kubeStateMetricsResources.limits.memory` | `string` |  |  |  |
| `spec.exporters.kubeStateMetricsResources.requests` | `CpuMemory` |  |  |  |
| `spec.exporters.kubeStateMetricsResources.requests.cpu` | `string` |  |  |  |
| `spec.exporters.kubeStateMetricsResources.requests.memory` | `string` |  |  |  |
| `spec.exporters.nodeExporterEnabled` | `bool` |  | `true` |  |
| `spec.exporters.nodeExporterResources` | `ContainerResources` |  |  |  |
| `spec.exporters.nodeExporterResources.limits` | `CpuMemory` |  |  |  |
| `spec.exporters.nodeExporterResources.limits.cpu` | `string` |  |  |  |
| `spec.exporters.nodeExporterResources.limits.memory` | `string` |  |  |  |
| `spec.exporters.nodeExporterResources.requests` | `CpuMemory` |  |  |  |
| `spec.exporters.nodeExporterResources.requests.cpu` | `string` |  |  |  |
| `spec.exporters.nodeExporterResources.requests.memory` | `string` |  |  |  |
| `spec.controlPlaneScrapers` | `KubernetesKubePrometheusStackControlPlaneScrapers` |  |  |  |
| `spec.controlPlaneScrapers.kubeApiServer` | `bool` |  | `true` |  |
| `spec.controlPlaneScrapers.kubelet` | `bool` |  | `true` |  |
| `spec.controlPlaneScrapers.kubeControllerManager` | `bool` |  | `true` |  |
| `spec.controlPlaneScrapers.coreDns` | `bool` |  | `true` |  |
| `spec.controlPlaneScrapers.kubeEtcd` | `bool` |  | `true` |  |
| `spec.controlPlaneScrapers.kubeScheduler` | `bool` |  | `true` |  |
| `spec.controlPlaneScrapers.kubeProxy` | `bool` |  | `true` |  |
| `spec.defaultRules` | `KubernetesKubePrometheusStackDefaultRules` |  |  |  |
| `spec.defaultRules.enabled` | `bool` |  | `true` |  |
| `spec.defaultRules.disabledGroups` | `[]string` |  |  |  |
| `spec.imageRegistry` | `string` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install into. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the resource.
When false, the namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "87.19.1" — chart 87.19.1 pairs
with Prometheus Operator v0.92.1). Versions must exist as SERVED
charts in the repository index
(https://prometheus-community.github.io/helm-charts). KNOW THIS:
bumping the chart version does NOT upgrade the CRDs (see the CRD
LIFECYCLE note above) — pair upgrades across operator minors with
`crd_upgrade_job`.

- default: `87.19.1`

### spec.skipCrds

`bool`

Skip installing the monitoring.coreos.com CRDs. Set ONLY when the
CRDs are owned elsewhere (a second stack on the cluster, or a
GitOps-managed CRD bundle). With the CRDs absent the install fails —
this is a bring-your-own-CRDs arm, not a lighter install.

### spec.crdUpgradeJob

`bool`

Run the chart's CRD upgrade hook (a pre-install/pre-upgrade Job that
server-side-applies the chart's CRD bundle). Chart default: off.
Enable for chart upgrades that cross operator versions — Helm never
upgrades the `crds/`-directory CRDs on its own. The hook runs kubectl
from `registry.k8s.io/kubectl` (tag = the cluster's Kubernetes
version); override its images via `helm_values` for air-gapped
registries.

### spec.prometheus

`KubernetesKubePrometheusStackPrometheus`

The Prometheus server (rendered as a Prometheus CR the operator
reconciles into a StatefulSet). Empty = single replica, 10-day
retention, a 50Gi persistent volume and cluster-wide discovery.

- rule: ephemeral: true runs on emptyDir — a non-default disk_size or a storage_class must not be set with it

### spec.prometheus.replicas

`int32` · optional (explicit presence)

Number of Prometheus replicas. Each replica scrapes and stores the
FULL target set independently (this is HA duplication, not sharding)
— queries against the Service may see either replica. 2 is the HA
pair; deduplicated long-term storage across replicas is the Thanos
surface (`helm_values`).

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.prometheus.retention

`string` · optional (explicit presence)

How long samples are kept, as a Prometheus duration (e.g. "10d",
"36h", "2w"). Retention ends when EITHER this or `retention_size` is
reached.

- default: `10d`
- rule: {"string":{"pattern":"^[0-9]+(ms|s|m|h|d|w|y)$"}}

### spec.prometheus.retentionSize

`string`

Maximum on-disk size of the TSDB, as a Prometheus byte size (e.g.
"45GiB", "500MB"). Size it BELOW the volume's `disk_size` — when the
volume itself fills up, Prometheus crash-loops instead of trimming.
Empty = no size-based trimming.

- rule: retention_size must be a Prometheus byte size like '45GiB' or '500MB' (units B, KB, MB, GB, TB, PB, EB, KiB, MiB, GiB, TiB, PiB, EiB)

### spec.prometheus.diskSize

`string` · optional (explicit presence)

Size of the persistent TSDB volume PER replica (e.g. "50Gi").
KNOW THIS: the chart's own default is an emptyDir — every metric
vanishes on pod restart — so this component provisions a
PersistentVolumeClaim by default instead. Kubernetes cannot shrink
PVCs; plan for growth. Ignored when `ephemeral` is true.

- default: `50Gi`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?\\s?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.prometheus.storageClass

`string | valueFrom`

Storage class for the TSDB volumes. Accepts a literal name or a
reference to a KubernetesStorageClass resource. Empty = the cluster's
default storage class. Ignored when `ephemeral` is true.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.prometheus.ephemeral

`bool`

Run WITHOUT persistent storage (the chart's own emptyDir default):
all metrics vanish with each pod restart. Throwaway dev/test
clusters only.

### spec.prometheus.resources

`ContainerResources`

CPU and memory for the Prometheus container. Memory scales with
active series count — undersized limits are the most common cause of
a crash-looping Prometheus on busy clusters. Empty = no
requests/limits (the chart default).

### spec.prometheus.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.prometheus.resources.limits.cpu

`string`

### spec.prometheus.resources.limits.memory

`string`

### spec.prometheus.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.prometheus.resources.requests.cpu

`string`

### spec.prometheus.resources.requests.memory

`string`

### spec.prometheus.externalLabels

`map<string, string>`

Labels attached to every sample and alert leaving this Prometheus
(federation, remote write, Alertmanager). Convention: a `cluster`
label naming the cluster — multi-cluster backends key on it.

### spec.prometheus.scrapeInterval

`string`

Default interval between scrapes, as a Prometheus duration (e.g.
"30s"). Empty = the Prometheus default (30s). Individual monitors
may override per target.

- rule: scrape_interval must be a Prometheus duration like '30s' or '1m'

### spec.prometheus.evaluationInterval

`string`

Default interval between rule evaluations, as a Prometheus duration.
Empty = the operator default (30s).

- rule: evaluation_interval must be a Prometheus duration like '30s' or '1m'

### spec.prometheus.discovery

`enum` · optional (explicit presence)

Which ServiceMonitor/PodMonitor/PrometheusRule/Probe/ScrapeConfig
objects this Prometheus discovers. Default `all_monitors` — see the
DISCOVERY note on the spec.

- default: `all_monitors`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `kubernetes_kube_prometheus_stack_monitor_discovery_unspecified` -- Unspecified. Defaults to all_monitors.
- `all_monitors` -- Discover every ServiceMonitor/PodMonitor/PrometheusRule/Probe/ ScrapeConfig in the cluster, whoever created it — what makes other components' service_monitor toggles and user-authored monitors work with zero extra wiring. The component default.
- `release_managed_only` -- The chart's own fenced default: discover only objects carrying this release's label. For multi-tenant clusters running several Prometheus servers with deliberate ownership boundaries.

### spec.prometheus.remoteWrite

`[]KubernetesKubePrometheusStackRemoteWrite`

Ship samples to remote/long-term backends as they are scraped —
Grafana Cloud/Mimir, Amazon Managed Prometheus (SigV4), Azure
Monitor, Thanos Receive, VictoriaMetrics.

### spec.prometheus.remoteWrite[].url

`string` · required

The remote write URL (e.g.
"https://aps-workspaces.us-east-1.amazonaws.com/workspaces/ws-.../api/v1/remote_write").

- rule: {"required":true}

### spec.prometheus.remoteWrite[].name

`string`

Name for this remote-write queue — shows up in Prometheus'
queue metrics; required when declaring more than one destination.

### spec.prometheus.remoteWrite[].basicAuth

`KubernetesKubePrometheusStackRemoteWriteBasicAuth`

HTTP basic auth. The password is read from an existing Secret.

### spec.prometheus.remoteWrite[].basicAuth.username

`string` · required

Username (e.g. the Grafana Cloud metrics instance ID).

- rule: {"required":true}

### spec.prometheus.remoteWrite[].basicAuth.passwordSecret

`KubernetesKubePrometheusStackSecretKeyRef` · required

Existing Secret holding the password.

- rule: {"required":true}

### spec.prometheus.remoteWrite[].basicAuth.passwordSecret.name

`string` · required

Secret name. The Secret must live in the stack's namespace —
Prometheus mounts it from there.

- rule: {"required":true}

### spec.prometheus.remoteWrite[].basicAuth.passwordSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.prometheus.remoteWrite[].bearerTokenSecret

`KubernetesKubePrometheusStackSecretKeyRef`

Bearer-token auth (e.g. Grafana Cloud API tokens). The token is
read from an existing Secret.

### spec.prometheus.remoteWrite[].bearerTokenSecret.name

`string` · required

Secret name. The Secret must live in the stack's namespace —
Prometheus mounts it from there.

- rule: {"required":true}

### spec.prometheus.remoteWrite[].bearerTokenSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.prometheus.remoteWrite[].sigv4

`KubernetesKubePrometheusStackRemoteWriteSigv4`

AWS SigV4 signing — Amazon Managed Prometheus. With no keys
declared the pod's ambient identity signs (IRSA on EKS — the
keyless posture).

- rule: declare both access_key_secret and secret_key_secret, or neither (keyless)

### spec.prometheus.remoteWrite[].sigv4.region

`string` · required

AWS region of the destination workspace (e.g. "us-east-1").

- rule: {"required":true}

### spec.prometheus.remoteWrite[].sigv4.roleArn

`string`

IAM role to assume before signing. Empty = sign as the pod's own
identity (IRSA — the recommended keyless posture on EKS).

### spec.prometheus.remoteWrite[].sigv4.accessKeySecret

`KubernetesKubePrometheusStackSecretKeyRef`

Existing Secret holding a static access key ID. Declare together
with `secret_key_secret` only when ambient identity is unavailable —
prefer keyless.

### spec.prometheus.remoteWrite[].sigv4.accessKeySecret.name

`string` · required

Secret name. The Secret must live in the stack's namespace —
Prometheus mounts it from there.

- rule: {"required":true}

### spec.prometheus.remoteWrite[].sigv4.accessKeySecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.prometheus.remoteWrite[].sigv4.secretKeySecret

`KubernetesKubePrometheusStackSecretKeyRef`

Existing Secret holding the static secret access key.

### spec.prometheus.remoteWrite[].sigv4.secretKeySecret.name

`string` · required

Secret name. The Secret must live in the stack's namespace —
Prometheus mounts it from there.

- rule: {"required":true}

### spec.prometheus.remoteWrite[].sigv4.secretKeySecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.prometheus.remoteWrite[].azureAd

`KubernetesKubePrometheusStackRemoteWriteAzureAd`

Azure AD managed-identity auth — Azure Monitor managed Prometheus.

### spec.prometheus.remoteWrite[].azureAd.managedIdentityClientId

`string` · required

Client ID of the managed identity that writes to the Azure Monitor
workspace (workload identity on AKS).

- rule: {"required":true}

### spec.prometheus.remoteWrite[].azureAd.cloud

`string`

Azure cloud. Empty = "AzurePublic" (others: "AzureChina",
"AzureGovernment").

### spec.prometheus.enableRemoteWriteReceiver

`bool`

Accept remote-write pushes from OTHER Prometheus servers or agents on
this server's `/api/v1/write` (turns this Prometheus into an
aggregation point).

### spec.prometheus.additionalScrapeConfigs

`string`

Raw Prometheus scrape_config entries (YAML list) appended to the
generated configuration — the seam for exotic SD mechanisms no
monitor object models. Entries here are outside the operator's
validation: a syntax error stalls config reload. Prefer ScrapeConfig
objects (discovered per `discovery`) when they can express the job.

### spec.prometheus.scheduling

`KubernetesKubePrometheusStackScheduling`

Scheduling for the Prometheus pods.

### spec.prometheus.scheduling.nodeSelector

`map<string, string>`

Node selector for the component's pods.

### spec.prometheus.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the component's pods.

### spec.prometheus.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.prometheus.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.prometheus.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.prometheus.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.prometheus.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.prometheus.scheduling.priorityClassName

`string`

Priority class name for the component's pods.

### spec.alertmanager

`KubernetesKubePrometheusStackAlertmanager`

Alertmanager (rendered as an Alertmanager CR). Deployed by default —
disable only when alerts are routed to an external Alertmanager.

- rule: ephemeral: true runs on emptyDir — a non-default disk_size or a storage_class must not be set with it

### spec.alertmanager.enabled

`bool` · optional (explicit presence)

Deploy Alertmanager. Default true. Disable only when this
Prometheus sends alerts to an Alertmanager it does not own
(declare the external endpoints via `helm_values`).

- default: `true`

### spec.alertmanager.replicas

`int32` · optional (explicit presence)

Number of Alertmanager replicas. Multiple replicas gossip into one
notification cluster (deduplicated alerts); 3 is the quorum-safe HA
count.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.alertmanager.retention

`string` · optional (explicit presence)

How long Alertmanager keeps its notification state, as a Prometheus
duration (silences and dedup state survive restarts through the data
volume).

- default: `120h`
- rule: {"string":{"pattern":"^[0-9]+(ms|s|m|h|d|w|y)$"}}

### spec.alertmanager.diskSize

`string` · optional (explicit presence)

Size of the persistent state volume PER replica (silences,
notification log). The chart's own default is an emptyDir — silences
vanish on restart — so this component provisions a small PVC by
default. Ignored when `ephemeral` is true.

- default: `2Gi`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?\\s?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.alertmanager.storageClass

`string | valueFrom`

Storage class for the state volumes. Empty = the cluster's default
class. Ignored when `ephemeral` is true.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.alertmanager.ephemeral

`bool`

Run WITHOUT persistent state (the chart's own emptyDir default):
silences and the notification log vanish with each pod restart.

### spec.alertmanager.resources

`ContainerResources`

CPU and memory for the Alertmanager container. Empty = no
requests/limits (the chart default). Alertmanager is light — small
requests suffice.

### spec.alertmanager.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.alertmanager.resources.limits.cpu

`string`

### spec.alertmanager.resources.limits.memory

`string`

### spec.alertmanager.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.alertmanager.resources.requests.cpu

`string`

### spec.alertmanager.resources.requests.memory

`string`

### spec.alertmanager.configYaml

`string`

The Alertmanager configuration document (route/receivers/
inhibit_rules) as YAML — where notification destinations (Slack,
PagerDuty, email, webhooks) are declared. Empty = the chart's
default: everything routes to a "null" receiver (alerts are visible
in the UIs and APIs but notify nobody), with the always-firing
Watchdog alert routed separately as the dead-man's-switch hook.
KNOW THIS: webhook URLs and API keys in this document are
credentials — reference them with Alertmanager's `_file` fields and
mount the Secret via `helm_values`, or manage the whole document as
an AlertmanagerConfig object instead of inlining tokens here.

### spec.alertmanager.scheduling

`KubernetesKubePrometheusStackScheduling`

Scheduling for the Alertmanager pods.

### spec.alertmanager.scheduling.nodeSelector

`map<string, string>`

Node selector for the component's pods.

### spec.alertmanager.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the component's pods.

### spec.alertmanager.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.alertmanager.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.alertmanager.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.alertmanager.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.alertmanager.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.alertmanager.scheduling.priorityClassName

`string`

Priority class name for the component's pods.

### spec.grafana

`KubernetesKubePrometheusStackGrafana`

The bundled Grafana, pre-provisioned with the stack's dashboards and
a datasource pointing at this Prometheus. Deployed by default. For a
standalone Grafana composed with multiple datasources, disable this
and deploy KubernetesGrafana instead.

### spec.grafana.enabled

`bool` · optional (explicit presence)

Deploy the bundled Grafana. Default true — it arrives pre-wired
with a datasource for this Prometheus and the stack's dashboard set.

- default: `true`

### spec.grafana.adminSecret

`KubernetesKubePrometheusStackGrafanaAdminSecret`

Read the admin credentials from an existing Secret. Empty = the
chart generates a random admin password ONCE at first install
(stable across upgrades) and keeps it in its own
`<name>-grafana` Secret — keys `admin-user` / `admin-password`;
the Secret name lands in the stack outputs.

### spec.grafana.adminSecret.name

`string` · required

Secret name (a reference to an existing Kubernetes Secret, never
secret material). Must exist BEFORE the install.

- rule: {"required":true}

### spec.grafana.adminSecret.userKey

`string` · optional (explicit presence)

Key holding the admin username. Empty = "admin-user".

- default: `admin-user`

### spec.grafana.adminSecret.passwordKey

`string` · optional (explicit presence)

Key holding the admin password. Empty = "admin-password".

- default: `admin-password`

### spec.grafana.defaultDashboardsEnabled

`bool` · optional (explicit presence)

Provision the stack's curated dashboard set (Kubernetes cluster,
node, workload, Prometheus and Alertmanager dashboards). Default
true.

- default: `true`

### spec.grafana.storage

`KubernetesKubePrometheusStackGrafanaStorage`

Persistent storage for Grafana's own state (UI-authored dashboards,
users, preferences). Empty = ephemeral (the chart default) — safe
when everything is provisioned as code, wrong if people build
dashboards in this UI by hand.

### spec.grafana.storage.size

`string` · optional (explicit presence)

Volume size as a Kubernetes quantity (e.g. "10Gi").

- default: `10Gi`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?\\s?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.grafana.storage.storageClass

`string | valueFrom`

Storage class for the PVC. Empty = the cluster's default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.grafana.resources

`ContainerResources`

CPU and memory for the Grafana container. Empty = no requests/limits
(the chart default).

### spec.grafana.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.grafana.resources.limits.cpu

`string`

### spec.grafana.resources.limits.memory

`string`

### spec.grafana.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.grafana.resources.requests.cpu

`string`

### spec.grafana.resources.requests.memory

`string`

### spec.operator

`KubernetesKubePrometheusStackOperator`

The Prometheus Operator itself (the controller that reconciles the
Prometheus/Alertmanager CRs and admission-validates rules).

### spec.operator.resources

`ContainerResources`

CPU and memory for the operator container. Empty = no
requests/limits (the chart default).

### spec.operator.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.operator.resources.limits.cpu

`string`

### spec.operator.resources.limits.memory

`string`

### spec.operator.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.operator.resources.requests.cpu

`string`

### spec.operator.resources.requests.memory

`string`

### spec.operator.admissionWebhooks

`KubernetesKubePrometheusStackAdmissionWebhooks`

The operator's admission webhooks — they validate PrometheusRule and
AlertmanagerConfig objects at admission so a malformed rule can never
stall Prometheus' config reload.

- rule: cert_manager configures the webhooks' certificate — it cannot be combined with disabled: true

### spec.operator.admissionWebhooks.disabled

`bool`

Turn the admission webhooks OFF entirely. Rules then fail at
Prometheus config-reload time instead of at kubectl-apply time —
accept only where the webhook's certificate machinery cannot run.

### spec.operator.admissionWebhooks.certManager

`bool`

Let cert-manager issue and inject the webhook's serving certificate
(an Issuer + Certificates rendered by the chart) instead of the
default self-contained certgen hook Job. Requires
KubernetesCertManager on the cluster.

### spec.operator.scheduling

`KubernetesKubePrometheusStackScheduling`

Scheduling for the operator pod.

### spec.operator.scheduling.nodeSelector

`map<string, string>`

Node selector for the component's pods.

### spec.operator.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the component's pods.

### spec.operator.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.operator.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.operator.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.operator.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.operator.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.operator.scheduling.priorityClassName

`string`

Priority class name for the component's pods.

### spec.exporters

`KubernetesKubePrometheusStackExporters`

The bundled metric sources: kube-state-metrics (object-state metrics)
and node-exporter (per-node host metrics, a DaemonSet). Both on by
default — the stack's own dashboards and rules assume them.

### spec.exporters.kubeStateMetricsEnabled

`bool` · optional (explicit presence)

Deploy kube-state-metrics (Deployment) — turns Kubernetes object
state (deployments, pods, PVCs, jobs, ...) into metrics. The
kubernetes-apps rules and dashboards are built on it. Default true.

- default: `true`

### spec.exporters.kubeStateMetricsResources

`ContainerResources`

CPU and memory for kube-state-metrics. Empty = chart defaults.

### spec.exporters.kubeStateMetricsResources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.exporters.kubeStateMetricsResources.limits.cpu

`string`

### spec.exporters.kubeStateMetricsResources.limits.memory

`string`

### spec.exporters.kubeStateMetricsResources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.exporters.kubeStateMetricsResources.requests.cpu

`string`

### spec.exporters.kubeStateMetricsResources.requests.memory

`string`

### spec.exporters.nodeExporterEnabled

`bool` · optional (explicit presence)

Deploy node-exporter (DaemonSet on every node) — host-level CPU,
memory, disk, filesystem and network metrics. The node rules and
dashboards are built on it. Default true.

- default: `true`

### spec.exporters.nodeExporterResources

`ContainerResources`

CPU and memory for each node-exporter pod. Empty = chart defaults.

### spec.exporters.nodeExporterResources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.exporters.nodeExporterResources.limits.cpu

`string`

### spec.exporters.nodeExporterResources.limits.memory

`string`

### spec.exporters.nodeExporterResources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.exporters.nodeExporterResources.requests.cpu

`string`

### spec.exporters.nodeExporterResources.requests.memory

`string`

### spec.controlPlaneScrapers

`KubernetesKubePrometheusStackControlPlaneScrapers`

Scrape toggles for the Kubernetes control-plane components. All on by
default — the right shape for self-managed clusters (kubeadm, kind,
datacenter). On managed clouds disable the unreachable ones (see the
MANAGED CLOUDS note above; the per-cloud presets carry the exact set).

### spec.controlPlaneScrapers.kubeApiServer

`bool` · optional (explicit presence)

Scrape the API server (reachable everywhere — managed clouds expose
it through the kubernetes Service). Default true.

- default: `true`

### spec.controlPlaneScrapers.kubelet

`bool` · optional (explicit presence)

Scrape the kubelets (reachable everywhere — kubelets run on YOUR
nodes). Default true.

- default: `true`

### spec.controlPlaneScrapers.kubeControllerManager

`bool` · optional (explicit presence)

Scrape kube-controller-manager. Managed clouds (EKS/GKE/AKS):
disable — it is provider-internal and unreachable. Default true.

- default: `true`

### spec.controlPlaneScrapers.coreDns

`bool` · optional (explicit presence)

Scrape CoreDNS (runs as pods in kube-system — reachable everywhere).
Default true.

- default: `true`

### spec.controlPlaneScrapers.kubeEtcd

`bool` · optional (explicit presence)

Scrape etcd. Managed clouds: disable — etcd is provider-internal.
Default true.

- default: `true`

### spec.controlPlaneScrapers.kubeScheduler

`bool` · optional (explicit presence)

Scrape kube-scheduler. Managed clouds: disable — provider-internal.
Default true.

- default: `true`

### spec.controlPlaneScrapers.kubeProxy

`bool` · optional (explicit presence)

Scrape kube-proxy. Reachable on most distributions (it runs on your
nodes) BUT its metrics port binds to localhost by default on several
managed platforms and on clusters whose CNI replaces kube-proxy
entirely (Cilium kube-proxy-replacement) there is nothing to scrape
— disable in those postures. Default true.

- default: `true`

### spec.defaultRules

`KubernetesKubePrometheusStackDefaultRules`

The chart's curated Kubernetes alerting/recording rules (rendered as
PrometheusRule objects).

### spec.defaultRules.enabled

`bool` · optional (explicit presence)

Render the curated PrometheusRule set (Kubernetes alerts +
recording rules the dashboards depend on). Default true.

- default: `true`

### spec.defaultRules.disabledGroups

`[]string`

Names of rule GROUPS to leave out, from the chart's defaultRules.rules
map (e.g. "etcd", "kubeProxy", "kubeControllerManager",
"kubeSchedulerAlerting", "windows"). Pair with the
control_plane_scrapers you disable — a scraper without its rule
group silences the alerts that could never fire truthfully.

### spec.imageRegistry

`string`

Registry that replaces the registry part of EVERY image the stack
pulls (operator, prometheus, alertmanager, exporters, grafana,
webhook-certgen sidecars) — the air-gap/private-mirror path. Empty =
each image's upstream registry (quay.io, registry.k8s.io, docker.io,
ghcr.io).

### spec.imagePullSecrets

`[]string`

Names of existing image-pull Secrets applied to every workload the
stack creates (chart `global.imagePullSecrets`). The Secrets must
already exist in the namespace.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged LAST
over everything the typed fields render (Helm `-f` semantics,
identical on both engines). For the chart surface beyond the typed
fields (Thanos sidecar/ruler, windows monitoring, ingress-per-replica,
scrape classes, additional alertmanager template files, per-component
securityContexts, ...) — never the substitute for them. Do not put
secrets here; secret material belongs in the typed secret references.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesKubePrometheusStack, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the stack runs in. |
| `status.outputs.release_name` | `string` | Helm release name (= metadata.name). The modules pin the chart's fullname to it, so every child name below derives from it. |
| `status.outputs.prometheus_service` | `string` | name of the Prometheus Service (`<name>-prometheus`, port 9090). |
| `status.outputs.prometheus_endpoint` | `string` | in-cluster Prometheus endpoint — the URL Grafana datasources, remote readers and API clients use, e.g. http://monitoring-prometheus.observability.svc.cluster.local:9090 |
| `status.outputs.alertmanager_service` | `string` | name of the Alertmanager Service (`<name>-alertmanager`, port 9093). Empty when alertmanager is disabled. |
| `status.outputs.alertmanager_endpoint` | `string` | in-cluster Alertmanager endpoint, e.g. http://monitoring-alertmanager.observability.svc.cluster.local:9093. Empty when alertmanager is disabled. |
| `status.outputs.grafana_service` | `string` | name of the bundled Grafana Service (`<name>-grafana`, port 80). Empty when the bundled grafana is disabled. |
| `status.outputs.grafana_endpoint` | `string` | in-cluster Grafana endpoint, e.g. http://monitoring-grafana.observability.svc.cluster.local. Empty when the bundled grafana is disabled. |
| `status.outputs.grafana_admin_secret_name` | `string` | name of the Secret holding the bundled Grafana's admin credentials — `<name>-grafana`, keys `admin-user` / `admin-password` (the chart generates it once and keeps it stable across upgrades; when spec.grafana.admin_secret points at an existing Secret, that name is echoed here instead). Empty when the bundled grafana is disabled. |
| `status.outputs.prometheus_port_forward_command` | `string` | command to port-forward the Prometheus UI to a developer laptop, e.g. kubectl port-forward svc/monitoring-prometheus -n observability 9090:9090 |
| `status.outputs.grafana_port_forward_command` | `string` | command to port-forward the bundled Grafana to a developer laptop, e.g. kubectl port-forward svc/monitoring-grafana -n observability 3000:80. Empty when the bundled grafana is disabled. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.prometheus.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.alertmanager.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.grafana.storage.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesGrafana | `spec.datasources[].url` | `status.outputs.prometheus_endpoint` |
| KubernetesLoki | `spec.ruler.alertmanagerUrl` | `status.outputs.alertmanager_endpoint` |
| KubernetesTempo | `spec.metricsGenerator.remoteWriteUrl` | `status.outputs.prometheus_endpoint` |

## See Also

- [Overview](../README.md)
