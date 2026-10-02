---
kinds:
  - KubernetesNamespace
  - KubernetesKubePrometheusStack
  - KubernetesPriorityClass
  - KubernetesGrafana
  - KubernetesLoki
  - KubernetesTempo
  - KubernetesOtelCollector
  - KubernetesSignoz
  - KubernetesClickHouse
---

# Observability: the Assembled Stack vs the All-in-One

"Give me observability" has two honest answers in this catalog, and the
choice spans five or more kinds. This pattern is the comparison's single
home; each kind's `GUIDE.md` carries only its own judgment.

## The two shapes

- **The assembled stack** — best-of-breed pieces wired together:
  KubernetesKubePrometheusStack (metrics, alerting, the monitoring CRDs),
  KubernetesLoki (logs) fed by a KubernetesOtelCollector in daemonset
  mode, KubernetesTempo (traces) fed OTLP by apps or a collector, and
  KubernetesGrafana as the query hub reading all three. Each piece scales,
  upgrades, and fails independently; each is swappable; the operational
  surface is four components.
- **The all-in-one** — KubernetesSignoz: traces, metrics and logs in one
  UI, one query engine, one alert system — backed by a composed
  KubernetesClickHouse (plus its KubernetesAltinityOperator). One product
  to learn and operate, but the telemetry store is a real database whose
  lifecycle you own, and swapping any single concern means leaving the
  product.

| Choose | When |
|---|---|
| Assembled | The cluster already runs kube-prometheus-stack (most do); teams want Grafana; pieces must scale or be swapped independently; monitoring CRDs (ServiceMonitor et al.) are expected by other components |
| Signoz | One team wants one tool for all three signals; ClickHouse expertise exists (or the ClickHouse pair is composed anyway); minimizing the number of moving products outweighs per-piece flexibility |

Neither is a workaround — both are first-class. What is NOT first-class:
proposing pieces from both shapes for the same signals (two metrics
pipelines, two alert systems) without saying why.

## The assembled wiring, typed end to end

Every join in the assembled stack is a real reference the platform
validates and draws — this composition is fully `valueFrom`-wired:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesNamespace
metadata:
  name: observability
spec:
  name: observability
---
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKubePrometheusStack
metadata:
  name: monitoring
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: observability
      fieldPath: spec.name
---
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesLoki
metadata:
  name: logs
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: observability
      fieldPath: spec.name
---
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTempo
metadata:
  name: traces
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: observability
      fieldPath: spec.name
---
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGrafana
metadata:
  name: dashboards
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: observability
      fieldPath: spec.name
  datasources:
    - name: prometheus
      url:
        valueFrom:
          kind: KubernetesKubePrometheusStack
          name: monitoring
          fieldPath: status.outputs.prometheus_endpoint
    - name: loki
      type: loki
      url:
        valueFrom:
          kind: KubernetesLoki
          name: logs
          fieldPath: status.outputs.gateway_endpoint
    - name: tempo
      type: tempo
      url:
        valueFrom:
          kind: KubernetesTempo
          name: traces
          fieldPath: status.outputs.http_endpoint
```

Ingestion completes the picture: a KubernetesOtelCollector in daemonset
mode ships cluster logs to Loki's `gateway_endpoint`, applications (or a
gateway collector) send OTLP to Tempo's `otlp_grpc_endpoint`, and metric
scraping is declared through the stack's ServiceMonitor machinery. Loki's
`ruler.alertmanagerUrl` can reference the stack's Alertmanager, so even
log-driven alerts route through the one alerting system.

## Alerts that reach a person, on every cluster

Dashboards tell you why; alerts tell you THAT, and only alerts wake
anyone. Build the alerting half first and prove it before the first
dashboard exists.

- **One stack per cluster, each with its own Alertmanager.** Rules
  evaluate next to complete data and each cluster pages on its own, so
  a network blip to a central hub never causes a false page and a dead
  hub never silences production. A central Grafana (and its long-term
  store) is where signals are READ, never what pages.
- **Delivery is typed and its credentials are managed secrets.**
  `alertmanager.notifications` declares Discord, Pushover or webhook
  receivers whose URLs, tokens and keys are `$secret/<slug>`
  references; nothing sensitive sits in chart values.
- **Split by severity, not by volume.** Only alerts where a person must
  act now carry `severity=page` and reach a phone; everything else
  posts to a team channel and is read in the morning. A page route
  that continues does not reach the root receiver: give the pager
  receiver the channel's integration too.
- **Watch the watcher from outside.** The stack's always-firing
  Watchdog alert, sent as `notifications.heartbeat` to a monitor
  running outside every cluster, is the only signal that survives the
  cluster (or Alertmanager) dying. The monitor pages on silence.
  Install the stack first and confirm its heartbeat arrives, THEN tell
  the monitor to expect that cluster: expected before the first
  heartbeat, it opens with a false "cluster silent" alert.
- **Monitoring yields to the workload.** Give Prometheus, Alertmanager
  and the operator a KubernetesPriorityClass BELOW the platform's (for
  example -1, with preemption `Never`) through their `scheduling`
  blocks, so under pressure monitoring is evicted first and never
  evicts anything. Stay at -10 or above: the cluster autoscaler treats
  lower-priority pods as expendable and never adds a node for them.
- **Scrape only what the cluster can show you.** On a managed control
  plane (GKE, EKS, AKS) the controller manager, scheduler and etcd are
  the provider's, and kube-proxy's metrics are usually unreachable:
  turn those `control_plane_scrapers` off together with their
  `default_rules.disabled_groups`, or the cluster carries targets that
  are down forever and alerts that can never clear. GKE's cluster DNS
  is kube-dns, not CoreDNS, so `core_dns` goes off there too. After
  install, every active target reading `up` is the check that the
  posture is right.
- **No alert names a customer.** Messages render environment,
  component, summary and runbook only; a namespace on a shared cluster
  can be a customer's name.
- **Proven, not assumed.** Fire a synthetic alert with `amtool alert
  add` and watch it arrive; stop Alertmanager and watch the outside
  monitor page. Until both have happened, the cluster is not
  monitored.

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPriorityClass
metadata:
  name: observability
spec:
  name: observability
  value: -1
  preemption_policy: never
  description: Monitoring yields to the workloads it watches.
---
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKubePrometheusStack
metadata:
  name: prod-metrics
spec:
  namespace:
    valueFrom:
      name: observability
  prometheus:
    external_labels:
      environment: prod
      cluster: prod-cluster
    scheduling:
      priority_class_name: observability
  alertmanager:
    scheduling:
      priority_class_name: observability
    notifications:
      receivers:
        - name: channel
          discord:
            - webhook_url:
                value: $secret/discord-alerts-webhook-url
        - name: pager
          pushover:
            - token:
                value: $secret/pushover-app-token
              user_key:
                value: $secret/pushover-on-call-user-key
          discord:
            - webhook_url:
                value: $secret/discord-alerts-webhook-url
      route:
        receiver: channel
        routes:
          - matchers:
              - label: severity
                value: page
            receiver: pager
      heartbeat:
        url: https://watcher.example.com/heartbeat/prod-cluster
        bearer_token:
          value: $secret/heartbeat-token
  operator:
    scheduling:
      priority_class_name: observability
  grafana:
    enabled: false
  # A GKE cluster: the managed control plane's scrapers and their rule
  # groups are off together, and kube-dns replaces CoreDNS.
  control_plane_scrapers:
    kube_controller_manager: false
    kube_etcd: false
    kube_scheduler: false
    kube_proxy: false
    core_dns: false
  default_rules:
    disabled_groups:
      - etcd
      - kubeControllerManager
      - kubeSchedulerAlerting
      - kubeSchedulerRecording
      - kubeProxy
```

## Logs and traces outside the cluster

A hub's logs and traces are the evidence an incident review reads weeks
later, often after the cluster was rebuilt or moved. Keep them in an object
store that lives one environment above the hub, so destroying the hub never
destroys them. On Cloudflare that is the `r2` arm of `KubernetesLoki` and
`KubernetesTempo`: the bucket is a `CloudflareR2Bucket` declared where it
outlives the hub, referenced by name, and the key pair of a token scoped to
that one bucket arrives from secrets. The modules compose the S3 host,
region and addressing, and keep the pair in their own Secret.

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareR2Bucket
metadata:
  name: hub-logs
spec:
  bucketName: example-hub-logs
  accountId: 0123456789abcdef0123456789abcdef
  location: enam
  publicAccess: false
---
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesLoki
metadata:
  name: logs
spec:
  namespace:
    value: observability
  storage:
    r2:
      account_id:
        valueFrom:
          kind: CloudflareR2Bucket
          name: hub-logs
          fieldPath: status.outputs.account_id
      bucket:
        valueFrom:
          kind: CloudflareR2Bucket
          name: hub-logs
          fieldPath: status.outputs.bucket_name
      credentials:
        access_key_id:
          value: $secret/hub-logs-writer-access-key-id
        secret_access_key:
          value: $secret/hub-logs-writer-secret-access-key
  retention_period: 720h
```

Three things decide whether it holds: the token is scoped to the one bucket
(Object Read & Write), the bucket's lifecycle expiry runs later than the
workload's retention, and the location hint is chosen where the cluster
runs, because R2 honours it only at creation.

## Who can open the hub

Grafana shows every system at once, so who can sign in is part of the
design. Declare it on the kind: `auth.google` for a Google Workspace,
`auth.generic_oauth` for any OpenID Connect provider, the client secret
from secrets, and `server.root_url` set to the address people open (the
redirect URI is `<root_url>/login/google`). Once sign-in is declared,
Grafana's own authentication screen can no longer change it, so the
manifest stays the one record of who gets in.

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGrafana
metadata:
  name: hub
spec:
  namespace:
    value: observability
  server:
    root_url: https://grafana.example.com
  auth:
    google:
      client_id: 123456789-example.apps.googleusercontent.com
      client_secret:
        value: $secret/hub-google-signin-client-secret
      allowed_domains:
        - example.com
      hosted_domain: example.com
      role_attribute_path: "email == 'lead@example.com' && 'Admin' || 'Viewer'"
```

Who a Google client admits is decided at Google by its consent screen: an
Internal screen admits only the Workspace that owns the project, an
External one any Google account (in Testing, only listed test users).
Grafana's own gate is `allowed_domains`, matched against the email; the
kind refuses a Google sign-in that allows sign-up with none.

## A hub beside a cluster's agent

The hub usually lives on a cluster that already runs its own
monitoring agent (the stack and its Alertmanager, from "Alerts that
reach a person"). Split the work by lifecycle, not by component:

- **Collection belongs to the agent; storage and reading to the hub.**
  The daemonset log collector joins the agent, the same composition
  every cluster runs, with only its destination differing (the hub's
  Loki `otlp_push_endpoint` in-cluster, a public telemetry door
  elsewhere). The hub holds Loki, Tempo and Grafana in its own
  namespace, so rebuilding one never removes the other.
- **Grafana reads the cluster's own Prometheus** (the datasource's
  default kind is the stack) until other clusters send metrics; a
  second stack on one cluster needs `skip_crds`, so add a receiving
  store only when there is something to receive.
- **The hub brings its own front door.** Instead of editing the
  cluster's Gateway for every hostname, the Gateway admits listener sets
  (`allowed_listeners`) and the hub attaches a `KubernetesListenerSet`
  with its own certificate, one per hostname, so a browser never reuses
  another hostname's connection and meets a 404. external-dns writes a
  record for a route on a listener set only when its chart value
  `enableGatewayListenerSets` is on (no typed field yet, so it rides
  `helm_values`); without it the route reads Accepted and no record
  appears.

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesListenerSet
metadata:
  name: grafana-door
spec:
  namespace:
    value: observability-hub
  parentRef:
    kind: Gateway
    namespace: istio-ingress
    name:
      valueFrom:
        kind: KubernetesGateway
        name: cluster-gateway
        fieldPath: status.outputs.gateway_name
  listeners:
    - name: https
      hostname: grafana.example.com
      port: 443
      protocol: HTTPS
      tls:
        mode: Terminate
        certificateRefs:
          - name:
              valueFrom:
                kind: KubernetesCertificate
                name: grafana-cert
                fieldPath: status.outputs.secret_name
      allowedRoutes:
        namespaces:
          from: Same
---
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesHttpRoute
metadata:
  name: grafana-route
spec:
  namespace:
    value: observability-hub
  hostnames:
    - grafana.example.com
  parentRefs:
    - kind: ListenerSet
      name:
        valueFrom:
          kind: KubernetesListenerSet
          name: grafana-door
          fieldPath: status.outputs.listener_set_name
  rules:
    - backendRefs:
        - name:
            valueFrom:
              kind: KubernetesGrafana
              name: hub
              fieldPath: status.outputs.service
          port: 80
      matches:
        - path:
            type: PathPrefix
            value: /
```

The certificate's Secret lives in the listener set's own namespace.
Install order on a fresh cluster: the Gateway's and external-dns's
switches, then the agent (its collector holds lines on the node's disk
until Loki answers), then the hub.

## Dashboards as code

A Grafana nobody provisioned fills with hand-made screens no one can
review. Ship each dashboard as a `KubernetesConfigMap` labeled
`grafana_dashboard` (the Grafana sidecar loads it from any namespace)
and keep the screens in their own composition beside the hub, so a
panel change never re-plans Grafana, Loki or Tempo:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesConfigMap
metadata:
  name: capacity-dashboard
spec:
  name: dashboard-capacity
  namespace:
    valueFrom:
      name: observability-hub-ns
  labels:
    grafana_dashboard: "1"
  data:
    capacity.json: |
      {
        "title": "Will a cluster run out of room this week?",
        "uid": "capacity",
        "schemaVersion": 42,
        "panels": []
      }
```

- **One question per dashboard, one per panel.** The title is the
  question an operator asks mid-incident, and each panel's description
  is the question it answers. That is also what an agent reading the
  dashboard through Grafana's API gets.
- **Every panel returns data on a healthy cluster** (`or vector(0)` on
  counts, sorted lists rather than filtered ones), so a blank panel
  means broken and a checker can say so. Grafana answers "no data" as a
  200 with an empty frame.
- **Datasources by pinned uid** in every panel, and a `$cluster`
  variable whose "All" is `.*`: the same files serve one cluster's
  Prometheus and a hub Prometheus that several clusters write into.
- **"How full is the node" comes from the node exporter**
  (`node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes`), not
  the containers' working set. The kubelet that reports container memory
  is the first thing a starving node stops serving: on a night nodes sat
  at 99%, the container sum read 58% and then went silent. Compare that
  against what pods reserve, because the autoscaler sees only
  reservations.
- **Forecast only what has history.** A week's linear forecast from a
  node or build volume minutes old predicts hundreds of gigabytes below
  zero; show fullness now for cattle, and trends over the dashboard's
  range.
- **Provisioned dashboards are read-only**, even for an Admin, and the
  rest of the rules (delimiters in an infra chart, `schemaVersion`,
  catching a hand-made copy) are in the `KubernetesGrafana` guide,
  "Dashboards as code".

## On the diagram

The assembled shape renders as a hub: Grafana with three datasource edges
into the stack, Loki and Tempo, the collector's edge into Loki, and every
component's namespace edge into the shared observability namespace — the
telemetry topology is reviewable at a glance. The Signoz shape renders
smaller — Signoz plus its ClickHouse (and the operator in the shared
layer) — with application OTLP converging on one ingestion gateway.

## When the answer is BOTH

Signoz for one product team's application telemetry can coexist with a
cluster-level kube-prometheus-stack (platform components' ServiceMonitors
need the stack's CRDs regardless). That is a scoped decision, not
double-tooling — say which signals go where.

## See also

- `KubernetesLoki` and `KubernetesTempo` guides, "outside the cluster: R2"
- `KubernetesGrafana` guide, "Who can open Grafana"

Each kind's guide beside its reference page: the stack (CRD singleton +
the serviceMonitor seam), Grafana (hub vs bundled; ephemeral state), Loki
(nothing ships logs by itself), Tempo (the replica/storage floor), Signoz
(the composed-ClickHouse chain), and the collector (modes and RBAC).
