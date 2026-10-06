# Kubernetes Grafana

## When NOT to Use This

**One resource is ONE standalone Grafana** — the composition hub you
point at any mix of datasources and run independently of all of them.

Not the right kind when:

- **All you need is dashboards for one kube-prometheus-stack** — its
  bundled Grafana (on by default there, pre-wired with the stack's
  datasource and dashboards) is the simpler path. Deploy this kind
  when Grafana outlives any single stack: multiple datasources,
  external state, HA.
- **You expect metrics to exist because Grafana does** — Grafana
  renders data, it stores none. A KubernetesKubePrometheusStack (or
  Loki, Tempo, a database) is the data layer; its endpoint composes
  here as a datasource.
- **You expect a public endpoint out of the box** — the Service is
  ClusterIP; exposure composes from first-class kinds
  (KubernetesIngress, Gateway API kinds) over the service handle. Set
  `server.root_url` when you do, or OAuth redirects and rendered
  links point at localhost.
- **You expect hand-built dashboards to survive by default** — the
  chart default is EPHEMERAL: Grafana's embedded SQLite state
  vanishes on pod restart. Declare `storage` (single instance) or
  `database` (external Postgres/MySQL — also the HA requirement)
  before anyone builds anything by hand.

## Overview

**KubernetesGrafana** deploys Grafana — the observability dashboard —
from the official `grafana` Helm chart
(https://grafana-community.github.io/helm-charts, the chart's current
home; the old grafana.github.io repository stopped serving new
versions at 10.5.x). Chart pinned 12.8.0, shipping Grafana 13.1.1.

**The credentials contract**: the chart generates a random admin
password ONCE at first install (stable across upgrades) into its own
`<name>` Secret — keys `admin-user` / `admin-password` — unless
`admin_secret` points at an existing Secret. Credentials never appear
in rendered Helm values; the Secret name lands in the outputs.

**Key design points:**

- **State is the design decision.** Grafana keeps UI-authored
  dashboards, users and preferences in an embedded SQLite database on
  local disk. Three postures: ephemeral (the default — right when
  everything is provisioned as code), `storage` (a ReadWriteOnce PVC
  — one stateful instance), or `database` (external Postgres/MySQL —
  durable AND the requirement for `replicas > 1`; SQLite cannot be
  shared, so the spec enforces it). The database password rides an
  existing Secret through environment expansion — never the rendered
  config.
- **Provisioning as code.** `datasources` render Grafana's
  provisioning files — present from first boot. Each datasource URL
  is a foreign key defaulting to KubernetesKubePrometheusStack (its
  Prometheus endpoint) — the one-line wiring to cluster metrics.
  Datasource basic-auth passwords ride Secrets through environment
  expansion.
- **The dashboard sidecar is the composition contract.** On by
  default, it discovers any ConfigMap labeled `grafana_dashboard: "1"`
  cluster-wide — other components and teams ship dashboards to this
  Grafana without ever touching its spec. `community_dashboards`
  imports grafana.com dashboards by ID (pin `revision` for
  reproducible installs).
- **Grafana 13 plugin truth.** Several once-core datasource plugins
  (elasticsearch, cloudwatch) moved out of the core image — list them
  in `plugins` to keep using them; the modules enable the chart's
  bundled-plugin shadowing so installs succeed on the read-only image
  directory.
- **Sign-in is typed** — `auth.google` or `auth.generic_oauth`, with
  the client secret as a `$secret/` reference the modules write into
  their own `<name>-sso` Secret. Declaring either locks Grafana's
  Administration > Authentication screen so the manifest stays the only
  source of truth, and a rotated secret rolls the pods on the next apply.
- **Agent teammates read it through MCP** — `agent_reader` keeps a
  Viewer service account and its one token in the `<name>-agent-reader`
  Secret for Grafana's MCP server (`mcp-grafana --disable-write`): a
  short Job mints it after the release is Ready, `token_generation`
  replaces it, and `disabled` refuses it at once.
- **`helm_values` is the escape hatch** — merged LAST (Helm `-f`
  semantics, identical engines) for LDAP, the image renderer, alerting
  provisioning, extra sidecars. Never for secrets:
  the chart refuses to render secrets into its config ConfigMap, and
  every typed credential rides Secrets + environment expansion
  instead.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to install into — literal or a
  KubernetesNamespace reference (`create_namespace` to own it)

### Common

- **`spec.chart_version`**: chart pin (default `12.8.0` = Grafana
  13.1.1)
- **`spec.replicas`**: instance count — above 1 REQUIRES `database`
- **`spec.admin_secret`**: existing credentials Secret
  (`name`, `user_key`, `password_key`); empty = chart-generated
- **`spec.storage`**: PVC size + class for single-instance
  persistence (mutually exclusive with `replicas > 1`)
- **`spec.database`**: engine (postgres/mysql), host (literal or a
  KubernetesPostgres reference), name, user, `password_secret`,
  `ssl_mode`
- **`spec.datasources`**: name, type (default prometheus), url
  (literal or a KubernetesKubePrometheusStack reference),
  `is_default`, `uid`, `basic_auth` (password via Secret),
  `json_data`
- **`spec.dashboard_sidecar_enabled` / `spec.community_dashboards` /
  `spec.plugins`**: the dashboard/plugin surface
- **`spec.server.root_url` / `spec.auth` / `spec.smtp`**: public URL,
  anonymous access and login-form behavior, outbound email
  (credentials via Secret)
- **`spec.service_monitor_enabled`**: scrape Grafana's own /metrics
  (requires the Prometheus Operator CRDs)
- **`spec.agent_reader`**: agent teammates' read-only way in —
  `service_account_name` (default `agent-reader`), `token_generation`
  (raise to replace the token), `disabled` (refuse every token), and
  the Job's `image`
- **`spec.image` / `spec.helm_values`**: the air-gap path and the
  escape hatch

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace Grafana runs in |
| `release_name` | Helm release name (= `metadata.name`) |
| `service` | The Grafana Service (port 80 → container 3000; = the release name) |
| `endpoint` | In-cluster URL (`http://<name>.<ns>.svc.cluster.local`) |
| `admin_secret_name` | The credentials Secret — chart-owned `<name>`, or the referenced Secret echoed |
| `port_forward_command` | Workstation access without composed exposure |
| `agent_reader_token_secret` | Agent teammates' read-only token — `<name>-agent-reader`, key `token`; unset unless `agent_reader` is declared and not disabled |
| `agent_reader_job_name` | The Job keeping that account and token (`<name>-agent-reader-<8 hex>`) |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace); `storage.storage_class` references a
  KubernetesStorageClass; `database.host` accepts a KubernetesPostgres
  reference (its read-write endpoint); each datasource `url` accepts
  a KubernetesKubePrometheusStack reference (its Prometheus
  endpoint).
- **Other components ship dashboards** by creating ConfigMaps labeled
  `grafana_dashboard: "1"` — discovered cluster-wide by the sidecar,
  no coupling to this resource.
- **Exposure composes, never embeds**: a KubernetesIngress or Gateway
  API route over `service`, with `server.root_url` set to the public
  URL.

## Examples

The smallest declarable Grafana is a namespace alone — chart-generated
credentials, ephemeral state, no datasources. The composed shape:

### Dashboards over a metrics stack, persistent

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGrafana
metadata:
  name: dashboards
spec:
  namespace:
    value: observability
  create_namespace: true
  storage:
    size: 10Gi
  datasources:
    - name: Prometheus
      # The url is a foreign key: the reference resolves to the named
      # stack's prometheus_endpoint output (kind defaults to
      # KubernetesKubePrometheusStack). A literal `value:` works too.
      url:
        value_from:
          name: cluster-metrics
      is_default: true
  server:
    root_url: https://grafana.example.com
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
