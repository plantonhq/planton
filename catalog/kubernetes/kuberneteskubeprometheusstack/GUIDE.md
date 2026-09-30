# KubernetesKubePrometheusStack Guide

The judgment this guide carries: this stack is closer to cluster
infrastructure than to an app — one per cluster, and HALF THE CATALOG's
monitoring toggles silently depend on the CRDs it installs. Whether to
run it at all is the assembled-vs-all-in-one decision, which lives in the
[observability-stack pattern](../../_patterns/observability-stack.md).

## The serviceMonitor seam — the trap that spans the catalog

Components across the catalog carry a `serviceMonitor` (or metrics)
toggle — ingress-nginx, databases, operators. Every one of those toggles
creates a ServiceMonitor, and ServiceMonitor is a CRD THIS stack installs:
enable any component's monitor before the stack exists and THAT
component's release fails to install. When an architecture turns on any
monitoring toggle anywhere, this stack (or its CRDs) must already be on
the cluster — deploy it early, in the shared-cluster chart.

## One per cluster, by CRD physics

The monitoring CRDs are cluster-scoped singletons; a second stack must
skip CRDs and scope its discovery — an advanced posture the reference
page describes, never the default. Treat the stack like the operators:
one install, shared-cluster layer, its own namespace (`createNamespace: true` is
the [namespace-ownership pattern](../../_patterns/namespace-ownership.md)'s
sole-tenant case). Note the reference page's 26-character name budget —
the chart silently truncates longer fullnames.

## Alerts that reach a person

Out of the box Alertmanager notifies nobody. Monitoring exists only when
a deliberately fired alert has reached someone, so declare
`alertmanager.notifications` with the stack, not later, and prove it the
same day. Five things decide whether it works:

- **Credentials are managed-secret references.** Discord webhook URLs,
  Pushover keys and bearer tokens are `$secret/<slug>` values; the
  module writes them into its own Secret, mounts it, and renders `_file`
  fields, so no credential lands in chart values. A rotated secret is
  picked up on the next notification with no restart (proven live).
- **Wire the heartbeat on day one.** The always-firing Watchdog alert,
  posted to a monitor OUTSIDE the cluster every `heartbeat.interval`, is
  the only thing that notices when Alertmanager or the whole cluster
  dies. Keep the interval well under the monitor's staleness window (1m
  for a 5-minute window).
- **Pages and channels split by label, and `continue_matching` has a
  trap.** Route `severity=page` (only on rules where a person must act
  now) to a pager receiver. The root receiver takes only alerts no child
  route claims, so a continuing page route with nothing after it never
  reaches the channel: give the pager receiver both integrations
  (Pushover plus the channel's Discord), or follow it with a catch-all
  channel route.
- **Messages carry no customer names.** Every title is `[<environment>]
  <component>: <alertname>`, and the body is the alert's
  `customer_impact` annotation (else `summary`) plus `runbook_url`.
  `namespace`, `pod` and `description` are never rendered, because on a
  shared cluster a namespace can name a customer. Put `environment` on
  every alert through `prometheus.external_labels`, and `component` on
  your own rules.
- **Pushover's emergency priority repeats until acknowledged**, even
  after the alert resolves: Alertmanager cannot cancel it. Whoever holds
  the phone acknowledges in the app.

Typed Discord delivery needs chart 88 or later (Prometheus Operator
v0.93): older operators re-parse the configuration and refuse Discord's
`webhook_url_file`, so Alertmanager never starts even though `amtool
check-config` passes. When Alertmanager does not appear, read the
Alertmanager resource's `Reconciled` condition first; the operator's
reason is there.

Prove it: `kubectl exec` into the Alertmanager pod and run `amtool alert
add alertname=Drill severity=warning environment=<env>
--alertmanager.url=http://localhost:9093`; it must arrive in the
channel. Then `amtool config routes test
--config.file=/etc/alertmanager/config_out/alertmanager.env.yaml
severity=page environment=<env>` must name the pager.

## The bundled Grafana boundary

The stack ships a Grafana pre-loaded with its dashboards — enough when
this stack is the only datasource. The moment dashboards must read Loki,
Tempo, or anything else, the standalone
[KubernetesGrafana](../kubernetesgrafana/GUIDE.md) is the
composition hub; run it INSTEAD of the bundled one, not beside it.

## On the diagram

The stack renders as a shared-layer node; a standalone Grafana draws a
datasource edge into its `prometheus_endpoint`, and Loki's alert routing
can draw an edge into its Alertmanager. The serviceMonitor dependency,
like every prerequisite-shaped coupling
([operator-prerequisite pattern](../../_patterns/operator-prerequisite.md)),
draws NOTHING — reviewers check
for the stack node whenever any component's monitoring toggle is on.

## Pairs well with

- KubernetesGrafana — the hub, when datasources go beyond this stack.
- KubernetesLoki — log alerts routed through this Alertmanager
  (`ruler.alertmanagerUrl`).
- Every kind with a serviceMonitor toggle — they all assume this stack.
