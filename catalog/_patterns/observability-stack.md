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
| Assembled | The cluster already runs kube-prometheus-stack (most do); teams want Grafana; pieces must scale or be swapped independently; monitoring CRDs (ServiceMonitor et al.) are expected by other kinds |
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
  is kube-dns, not CoreDNS, so `core_dns` goes off there too, and a
  KubernetesPodMonitor on `k8s-app: kube-dns` in kube-system reads its
  sidecar's `metrics` port (10054), whose probes answer "is cluster DNS
  answering, and how fast?". After install, every active target reading
  `up` -- and every monitor finding at least one target -- is the check
  that the posture is right.
- **Turn the cloud's managed collection off once the agent runs.** GKE
  turns on Managed Service for Prometheus and its kube-state, cAdvisor,
  kubelet and DCGM packages by default, each billed per sample, while the
  agent already collects the same signals. On the cluster's
  GcpGkeCluster set `monitoring.managed_prometheus_enabled: false` and
  `monitoring.components: [SYSTEM_COMPONENTS]` together in one update --
  an empty `components` list changes nothing, so the packages stay until
  the free system components are named alone.
- **A curated alert that misreads the cluster is replaced, not muted.**
  Every alert that fires on normal work teaches people to stop reading
  the channel before the real page arrives. On autoscaled node pools
  turn `KubeCPUOvercommit` and `KubeMemoryOvercommit` off
  (`default_rules.disabled_alerts`): they count today's nodes. For an
  alert that misreads work done on purpose (CI build pods, a build
  machine at full CPU), disable it and declare the same alert name in a
  KubernetesPrometheusRule whose expression leaves that work out
  (`kube_pod_owner{owner_kind!~"Job|TaskRun"}`, every pod on a build
  machine, or `unless` the machine's build taint from
  `kube_node_spec_taint`), keeping the
  upstream `for`, severity and the `namespace` label Alertmanager's
  inhibitions read. Then read the loaded rules back
  (`/api/v1/rules?type=alert`): a name that matched nothing changed
  nothing.
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

### Alert rules of your own, as declared objects

The stack's default rules watch Kubernetes; the rules that say a product
is hurt are yours. Declare each set as a
[KubernetesPrometheusRule](../kubernetes/kubernetesprometheusrule/GUIDE.md),
never through the stack's spec or `helm_values`: it is a node of its own
on the diagram, composes with its namespace, and deploys on either
engine with every upstream setting.

- **Rules live where the data is complete.** Put them beside each
  cluster's agent stack, which loads every rule object under its default
  `all_monitors` discovery. A receiver that only stores remote-written
  series (a hub on `release_managed_only`) loads only objects labelled
  `release: <its release_name>` -- and evaluating there pages twice and on
  partial data, so give a rule that label only on purpose.
- **The labels on a rule are the route; the annotations are the page.**
  Every alerting rule carries `severity` (the pager route matches
  `page`), and `component`; `environment` and `cluster` arrive through
  `prometheus.external_labels`. Its annotations carry a `summary` and a
  `runbook_url` whose first line is the first action. Keep their text
  static: a `{{ $labels.namespace }}` in an annotation can carry a
  customer's name into a message, and a templating chart engine that
  renders the manifest mangles it. The subject belongs in labels
  (`node`, `pod`), which dashboards read and messages never render.
- **One object per owner.** Prometheus refuses a rule file with one bad
  rule and the operator drops the whole object, so a mistake in one
  team's rules must not silence another's. The kind refuses the two
  shapes that cause it (record and alert together or neither; a recording
  rule with alert-only fields) before the apply.
- **Record, then alert.** Burn-rate alerts read recorded error ratios:
  the recording rules sit earlier in the same group, the alert reads
  cheap series, and the dashboards read the same series. The kind's
  `01-error-budget-burn-alerts` and `02-recording-rules` presets carry
  that pair.

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPrometheusRule
metadata:
  name: api-slo-alerts
  relationships:
    - kind: KubernetesKubePrometheusStack
      name: cluster-metrics
      type: depends_on
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: observability-ns
      fieldPath: spec.name
  groups:
    - name: api-slo-alerts
      interval: 30s
      labels:
        component: api
      rules:
        - alert: ApiErrorBudgetFastBurn
          expr: job:slo_errors_per_request:ratio_rate1h{job="api"} > (14.4 * 0.001) and job:slo_errors_per_request:ratio_rate5m{job="api"} > (14.4 * 0.001)
          for: 2m
          keep_firing_for: 5m
          labels:
            severity: page
          annotations:
            summary: The API is burning its 30-day error budget 14x too fast
            runbook_url: https://runbooks.example.com/api-error-budget-burn
```

### Scraping as declared objects

What a cluster's agent scrapes beyond Kubernetes itself is declared the
same way: a [KubernetesServiceMonitor](../kubernetes/kubernetesservicemonitor/GUIDE.md)
for a workload whose Service names its metrics port, a
[KubernetesPodMonitor](../kubernetes/kubernetespodmonitor/GUIDE.md) for
pods no Service exposes (a database operator's instances, a DaemonSet's
exporters). Never a raw scrape config in the stack's `helm_values`, and
never a kind's own monitor toggle where the monitor needs settings the
toggle doesn't carry.

- **Put the monitor where it can install: beside the workload, or as a
  class on the agent.** A monitor needs the Prometheus operator's CRDs,
  and those arrive with the agent stack. Under the agent's default
  `all_monitors` discovery every monitor in the cluster loads with no
  label; a hub that only receives remote-written series never scrapes, so
  a monitor never carries its `release` label.
  - A workload whose composition installs **after** the agent carries its
    own monitor beside it: the component's own switch
    (`service_monitor_enabled` on KubernetesOpenBao, KubernetesValkey,
    KubernetesOpenFga, KubernetesTemporal) or a KubernetesServiceMonitor in
    its namespace.
  - A component installed **before** the agent -- the cluster's own gateway,
    istiod, cert-manager, external-dns, the database operator -- cannot: a
    monitor, or a switch that renders one, in the cluster's composition
    fails a fresh cluster's install the day it is rebuilt, because the CRD
    does not exist yet. Watch it from the agent's composition with one
    **class monitor** per kind of component: `namespace_selector: {any:
    true}` and a selector every instance carries (every Istio gateway's
    `gateway.networking.k8s.io/gateway-class-name: istio`, every
    CloudNativePG instance's `cnpg.io/podRole: instance`). One monitor then
    reads every instance on the cluster, including the one an environment
    adds tomorrow, with no list to keep.
- **A namespace's network policy must admit the agent.** A namespace that
  admits only its own pods silently starves every monitor of it: the
  target reads down with a dial timeout, and the stack's `TargetDown`
  posts ten minutes later. Admit the agent's Prometheus pods (the
  namespace selector `kubernetes.io/metadata.name: <agent namespace>` and
  the pod selector `app.kubernetes.io/name: prometheus` in **one** peer)
  on the metrics ports only, and apply the policy before the monitor, so
  no target is ever down. On a shared cluster, tell environments apart by
  namespace: relabeling a series' `environment` to the namespace's
  environment gives the label two meanings next to the cluster-wide
  signals that keep the agent's external `environment`.
- **Read what a component's series carry before you keep them.** The
  first scrape of a new component is where its surprises show:
  - **A series that already has a label wins over the agent's external
    label of the same name.** OpenBao labels every series `cluster` with
    its own cluster id, so the hub never learns which Kubernetes cluster
    the vault runs on: drop it with `metric_relabelings` `{regex: cluster,
    action: labeldrop}`.
  - **A component's own switch keeps every series it serves.** Temporal
    serves its latencies per operation and per task queue, about 150,000
    series in one idle environment: a keep list and a histogram thinned
    to a few bounds (`metric_relabelings` dropping the other `le`
    values) bring it to a few thousand. Count before and after with
    `scrape_samples_post_metric_relabeling`.
  - **A metric that names its subject's namespace needs `honor_labels`.**
    cert-manager's certificate series carry the certificate's own
    `namespace`; without `honor_labels: true` it becomes
    `exported_namespace` beside the scrape target's, and an alert on a
    certificate is placed in cert-manager's namespace instead of the
    environment it serves.
  - **A gateway's gRPC requests include long-lived streams.** A p95
    over every request through an Istio gateway reads the streams'
    lifetimes (tens of seconds); read front-door latency on
    `request_protocol="http"` or per backend.
- **Selectors match labels, not references.** A ServiceMonitor's
  `selector` matches the Service's labels and a PodMonitor's matches the
  pods'. The diagram draws no edge to either, so a reviewer checks the
  match. `job_label` names the label whose value becomes `job`; pick one
  that reads the same in every environment, or every dashboard's `job`
  filter breaks between them.
- **Credentials are references.** A bearer token, basic-auth halves,
  OAuth2 client credentials and TLS material are Secret and ConfigMap
  references in the monitor's namespace, so the graph creates them first.
  The operator skips a monitor whose Secret it can't read, in silence.
- **The operator skips a broken monitor whole.** Two authentication methods
  on one endpoint, a client certificate without its key, or a relabeling
  step that breaks its action's rules make the operator drop the whole
  object. The kinds refuse those shapes before the apply; the Prometheus's
  `/targets` page (job `serviceMonitor/<ns>/<name>/<n>`) is where a missing
  Secret shows.
- **Bound every target.** `sample_limit` turns a cardinality explosion
  into one failed scrape, and `metric_relabelings` with `action: drop`
  keep unread series out of storage and off the remote-write bill.

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPodMonitor
metadata:
  name: orders-db
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: orders-ns
      fieldPath: spec.name
  job_label: cnpg.io/cluster
  pod_target_labels:
    - cnpg.io/instanceRole
  sample_limit: 100000
  selector:
    match_labels:
      cnpg.io/cluster: orders-db
  pod_metrics_endpoints:
    - port: metrics
      interval: 30s
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
  default kind is the stack) until other clusters send metrics; then
  the hub gets a receiving Prometheus of its own ("Several clusters,
  one hub" below), and only then.
- **The hub brings its own front door.** Instead of editing the
  cluster's Gateway for every hostname, the Gateway admits listener sets
  (`allowed_listeners`) and the hub attaches a `KubernetesListenerSet`
  with its own certificate, one per hostname, so a browser never reuses
  another hostname's connection and meets a 404. external-dns writes a
  record for a route on a listener set only with
  `gateway_listener_sets: true` (beside a `gateway-*` source); without it
  the route reads Accepted and no record appears.

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

## Several clusters, one hub

When other clusters report to the hub, three pieces join it, and every
cluster keeps its own agent and Alertmanager (rules run next to complete
data, so a broken link never changes what alerts):

- **A receiving Prometheus in the hub, never the agent's.** If the hub
  cluster's agent received the others' samples, its standard rules would
  run over them too and every remote alert would post twice. The
  receiver is a second `KubernetesKubePrometheusStack` that brings
  nothing the agent already runs: `skip_crds`, `discovery:
  release_managed_only`, Alertmanager, Grafana, both exporters, every
  `control_plane_scrapers` entry and `default_rules` off, and
  `enable_remote_write_receiver`. Two settings still ride `helm_values`
  (no typed field yet): `prometheusOperator.enabled: false`, because the
  agent's operator (it watches every namespace) runs this Prometheus and
  a second operator would fight it, and
  `prometheus.prometheusSpec.tsdb.outOfOrderTimeWindow` sized to what a
  sender can resend. Without that window Prometheus accepts a sample
  only up to about an hour older than its newest, while a sender keeps
  about two hours to resend after an outage of the door or the link. Any
  scraper left on in the receiver's release is scraped twice, because
  the agent discovers monitors cluster-wide. Its own release's
  self-monitor is the exception worth keeping: the agent reads it too and
  alerts on the hub.
- **Every agent writes to it, the hub's own cluster included,** so one
  query covers the estate. Each agent's `external_labels` (`cluster`,
  `environment`) ride the remote write onto every series, added only
  where a series lacks the label, so a per-pod environment wins. The
  address is a literal in the agent, never a reference: the hub depends
  on the agent (its priority, its operator), so a reference back loops.
  Log lines need the same stamp from the collector: a `resource`
  processor inserting `k8s.cluster.name` and
  `deployment.environment.name`, which Loki indexes as labels by default
  (`k8s_cluster_name`, `deployment_environment_name`).
- **A telemetry door: its own Gateway, so its own load balancer.** A JWT
  check on a Gateway refuses every bearer token that is not one of its
  JWTs, which would break whatever else that Gateway serves, and a flood
  of telemetry must never slow the platform's door. The door lives in
  the hub's namespace beside its routes and backends (an Istio policy's
  `target_refs` cannot cross namespaces), with exactly three exact-path
  rules: `/api/v1/write` to the receiver's 9090, `/otlp/v1/logs` to
  Loki's gateway on 80, `/v1/traces` to Tempo's 4318. Istio creates the
  door's pods itself; point the Gateway's `infrastructure.parametersRef`
  at a `KubernetesConfigMap` whose `deployment` key overlays the pod
  spec, so the door runs at the monitoring priority and yields like the
  rest of monitoring.
- **One issuer, one key per cluster.** A `KubernetesRequestAuthentication`
  with one rule: the door's hostname as issuer and audience, and an
  inline key set holding one key per cluster (key id = the cluster's
  name; an inline key set requires the issuer). Each cluster's token
  names itself as subject, so its principal reads `<issuer>/<cluster>`,
  and an ALLOW `KubernetesAuthorizationPolicy` lists those principals for
  POST on the three paths. The check alone lets a request with no token
  through; the ALLOW policy is what refuses it (403), while a token the
  key set cannot verify gets 401. The check strips the token before the
  stores see it (`forward_original_token` stays off). Revoking a cluster
  is deleting its key and principal; nothing needs to keep the signing
  key once its one token is signed.
- **The token reaches a sender through types the catalog has:** a
  `KubernetesSecret` from the managed secret, read by remote write's
  `bearer_token_secret` and by the collector's `env_from_secrets` as an
  `Authorization: Bearer ${env:...}` header. Both name the Secret as a
  plain string, so each needs an explicit `depends_on`.
- **The hub's Loki is sized for every cluster catching up at once,** not
  for the steady rate: `limits.ingestion_rate_mb` above the summed
  catch-up (12 serves a few clusters) and `ingestion_burst_size_mb` above
  every collector's largest batch (24 against a 4 MiB `max_size`), because
  Loki refuses a push larger than its burst every time. Each collector
  batches only from its on-disk queue, capped in bytes, and blocks when
  that queue is full (the `KubernetesOtelCollector` guide), and each has
  `service_monitor_enabled`, so the hub sees a sender's queue fill before
  any line is late.

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRequestAuthentication
metadata:
  name: telemetry-jwt
spec:
  namespace:
    value: observability-hub
  target_refs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name:
        valueFrom:
          name: telemetry-gateway
  jwt_rules:
    - issuer: telemetry.example.com
      audiences:
        - telemetry.example.com
      jwks: '{"keys":[{"kty":"RSA","kid":"cluster-a","use":"sig","alg":"RS256","n":"<modulus>","e":"AQAB"}]}'
---
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAuthorizationPolicy
metadata:
  name: telemetry-allow
spec:
  namespace:
    value: observability-hub
  target_refs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name:
        valueFrom:
          name: telemetry-gateway
  action: ALLOW
  rules:
    - from:
        - source:
            request_principals:
              - telemetry.example.com/cluster-a
      to:
        - operation:
            methods:
              - POST
            paths:
              - /api/v1/write
              - /otlp/v1/logs
              - /v1/traces
```

Proving the door takes four requests: no token (403), a token signed by
a key the door never saw, with identical claims (401), the real token
with an empty body (the store's own 400 or 422, so the door let it
through), and the real token on any other path or method (403).

**A cluster joins a running hub in one order:**
1. Mint its token.
2. Put its key and its principal in the door, and re-apply the hub.
3. Write the token where its agent reads it.
4. Install the agent.
5. Only after the agent's first heartbeat, list the cluster with
   whatever watches heartbeats.

Each step needs the one before it. A watcher told first reports the
cluster lost before it ever reported.

A cluster leaves, or is rebuilt, in the reverse order: out of the
heartbeat list, then its agent, then the cluster. Removing the cluster
under a declared agent leaves resources for a cluster that no longer
exists. A rebuilt cluster with the same name keeps its token, so only
the agent and the heartbeat entry come back.

**Every node, tainted pools included:** each cluster's log collector
tolerates every `NoSchedule` taint (the `KubernetesOtelCollector` guide).
Without that toleration, a build or GPU pool's logs never leave its
nodes, and the daemonset still reads complete.

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
  zero; show fullness now for cattle. A cluster's memory a week ahead
  does answer "will it run out", once the range holds days of samples:
  withhold it until then and say so in the cell (the `KubernetesGrafana`
  guide has the query shape).
- **Once several clusters write into one Prometheus, every join and
  grouping carries `cluster`.** Node addresses, namespaces and pod names
  repeat across clusters, so `on(instance)` or `on(namespace, pod)` alone
  matches two clusters' series and the query fails, and an "All" view
  sums clusters that should be compared. Lead with one row per cluster
  (reserved, used now, the busiest node's peak in the range, memory used
  a week ahead) rather than a blended number.
- **Count sparse events as totals, not per-interval charts.**
  `increase(x[$__interval])` over a window of one or two scrapes sees only
  increments between its own samples; a sender on a steady rhythm (an
  Alertmanager heartbeat every two minutes) lands every increment between
  windows and reads zero all hour while Prometheus holds dozens. A table
  of totals over `$__range` per channel, sent and failed, answers "did it
  go out"; filter to the integrations in use so a silent pager shows as
  zero, not as absent.
- **Name a component by its container across metrics and logs.** The
  container name (`postgres`, `openfga`, `temporal-history`) is the same
  in kube-state-metrics, cAdvisor and Loki's `k8s_container_name`, so one
  mapping joins a component's restarts, out-of-memory kills and error
  lines in a row; its workloads come from the controllers that survive
  scaling to zero.
- **An exporter that labels what it probes keeps its labels.** An
  outside watcher reports each probed environment in `environment`;
  scraping it with a static `environment` of its own moves that to
  `exported_environment`. `honor_labels: true` on that scrape job keeps
  the probe's environment, and the static label still fills series that
  carry none.
- **Roll pods up to the workload that owns them** from
  `kube_pod_owner`: a ReplicaSet's name less its last segment is its
  Deployment, other owners are named as they are, and a pod owned by
  nothing or by its node is its own workload. Operators act on the
  workload, not on a pod hash.
- **Provisioned dashboards are read-only**, even for an Admin, and the
  rest of the rules (delimiters in an Infra Chart, `schemaVersion`,
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
