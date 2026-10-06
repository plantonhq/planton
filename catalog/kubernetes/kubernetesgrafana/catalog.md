# Grafana

Deploys Grafana — the observability dashboard — from the official `grafana` Helm chart. Standalone Grafana is the **composition hub**: point it at any mix of datasources — a KubernetesKubePrometheusStack, external Prometheus or Mimir, Loki, Tempo, ClickHouse, Postgres — and run it independently of any one of them. If all you need is dashboards for ONE kube-prometheus-stack, that stack's bundled Grafana (on by default there) is the simpler path; this kind is for the Grafana that outlives and spans its datasources.

Everything that matters is provisioned as code: `datasources` and `communityDashboards` render Grafana's provisioning files, present from first boot and re-rendered on every restart. The one genuinely stateful thing — dashboards built by hand in the UI, users, preferences — lives in an embedded SQLite database on local disk, and where THAT lives is the central decision this spec asks you to make.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Helm release** (official `grafana` chart, pinned `12.8.0` — ships Grafana 13.1.1 — named `metadata.name`) — the Grafana Deployment, a ClusterIP Service (port 80 → container 3000), provisioning ConfigMaps for the declared datasources and community dashboards, the dashboard-discovery sidecar (on by default), and any declared plugins installed at startup
- **Admin credentials Secret** (`<name>`, keys `admin-user` / `admin-password`) — generated ONCE at first install and stable across upgrades; skipped when `adminSecret` points at an existing Secret you own (that name is echoed in the outputs instead)
- **Sign-in Secret** (`<name>-sso`) — only when `auth.google` or `auth.genericOauth` is declared; holds the client secrets Grafana reads as environment variables
- **Agent reader** — only when `agentReader` is declared: a ServiceAccount, Role and RoleBinding (`<name>-agent-reader`), a scripts ConfigMap (`<name>-agent-reader-script`), and a Job (`<name>-agent-reader-<8 hex>`) that keeps a Viewer service account in Grafana and its one token in the `<name>-agent-reader` Secret (keys `token`, `generation`), which the Job writes and the ServiceAccount owns
- **PersistentVolumeClaim** — only when `storage` is declared; a ReadWriteOnce volume under the embedded database
- **ServiceMonitor** — only when `serviceMonitorEnabled` is true (requires the Prometheus Operator CRDs)
- **Kubernetes Namespace** — created only when `createNamespace` is true; otherwise the namespace must already exist

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **An external Postgres or MySQL** — only when `database` is declared. The database itself must exist before the install (Grafana creates its tables, never the database), and its password must sit in an existing Secret in Grafana's namespace. Running more than one replica REQUIRES this block.
- **The Prometheus Operator CRDs** — only if you set `serviceMonitorEnabled`; deploy a **kube-prometheus-stack** first.
- **A StorageClass** — only when `storage` is declared; empty means the cluster's default class.

## Deploy

### Console

Open the deployment store, find **Grafana**, and click **Deploy**. The creation wizard walks you through namespace placement, the chart pin, the state-and-replicas decision, admin credentials, datasources, dashboards, plugins, the server URL and auth behavior, outbound email, observability, resources, placement, and the Helm-values escape hatch. Start from the **Dev dashboards preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGrafana
metadata:
  name: dev-grafana
  org: acme-corp
  env: dev
spec:
  namespace:
    value: dev-grafana
  createNamespace: true
  datasources:
    - name: Prometheus
      url:
        value: http://monitoring-prometheus.observability.svc.cluster.local:9090
      isDefault: true
```

```shell
planton apply -f grafana.yaml
```

This creates the smallest useful Grafana: one ephemeral replica, the chart-generated admin credentials in the `dev-grafana` Secret (name and a port-forward command land in the outputs), and a Prometheus datasource present from first boot because it is provisioned as code, not clicked together. An Infra Job tracks the provisioning in real time.

### InfraChart

Compose Grafana behind its namespace with a reference, and the InfraPipeline orders the deploys:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: observability-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then provisions Grafana into it.

## Key Configuration

These are the most important decisions when configuring a Grafana instance. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The state decision comes first** — Grafana keeps hand-built dashboards, users and preferences in an embedded SQLite database on the pod's local disk, and the chart default is EPHEMERAL: a pod restart — a node drain, an upgrade, an eviction — erases everything hand-made. Provisioned datasources and dashboards survive (they are configuration, re-rendered on every boot); nothing built by hand does. That default is honest and correct exactly as long as the UI is a window, not a workshop. Three postures: stay ephemeral, declare `storage` (a 10Gi ReadWriteOnce volume — ONE stateful instance), or declare `database` (external Postgres/MySQL carries ALL state — the HA path).

**Replicas above 1 REQUIRE the database** — each pod's embedded SQLite is private, so a scaled deployment without an external database splits dashboards and sessions across pods; the spec refuses the combination. The `storage` volume is ReadWriteOnce and backs exactly one replica — the spec refuses that combination too. When a single stateful instance needs HA later, the move is a Grafana database export/import into the `database` posture, not a bigger replica number.

**Datasources are code, wired by reference** — each `datasources` entry renders into Grafana's provisioning file, present from first boot. The `url` is a foreign key: a `valueFrom` reference to a **kube-prometheus-stack** resolves its exported Prometheus endpoint — the one-line wiring that points this Grafana at the cluster's metrics and orders it after the stack. Declare `isDefault` on exactly one entry.

**Other teams ship dashboards without touching this resource** — the dashboard sidecar (on by default) discovers any ConfigMap labeled `grafana_dashboard: "1"` anywhere in the cluster. That label is the composition contract: components and teams publish dashboards by creating labeled ConfigMaps, never by editing this spec.

**Community dashboards import by grafana.com ID** — `communityDashboards` pulls dashboards at install (1860 is Node Exporter Full), each bound to a declared datasource's name. Pin `revision` for reproducible installs; an unpinned entry floats on whatever revision is latest that day, so two environments deployed a week apart can render different dashboards from the same spec.

**Admin credentials are chart-generated once** — a random admin password lands in the `<name>` Secret (keys `admin-user` / `admin-password`) at first install and stays stable across upgrades; the Secret name is exported in the outputs. Point `adminSecret` at an existing Secret to bring your own, so rotation follows your team's machinery. Credentials never appear in rendered Helm values.

**Plugins install at startup, by ID** — list plugin IDs in `plugins`, optionally with a version (`"grafana-oncall-app 1.3.0"`). Know this: Grafana 13 moved the once-core `elasticsearch` and `cloudwatch` datasource plugins out of the core image — declaring a datasource of those types means listing the plugin here too. The modules enable the chart's bundled-plugin shadowing so installs succeed on the read-only image directory.

**Exposure composes from first-class kinds** — the Service stays ClusterIP by design; no exposure fields exist on this spec. Compose an HTTP route (Gateway API kinds) over the exported `service` handle, and set `server.rootUrl` to the public URL when you do — sign-in redirect URLs, alert links and rendered images embed it.

**Sign-in through Google or any OAuth provider** — `auth.google` puts Grafana behind a Google Workspace, `auth.generic_oauth` behind Okta, Microsoft Entra ID, Keycloak or any OpenID Connect provider. The client secret is picked from Planton's secrets, never pasted, and never reaches Grafana's configuration file. A Google sign-in that would let any Google account create a user is refused until `allowedDomains` names your domain. Once sign-in is declared, Grafana's own authentication screen can no longer change it.

**`helmValues` merges last** — the escape hatch for chart surface beyond the typed fields (LDAP in grafana.ini, the image renderer, alerting provisioning, extra sidecars), with Helm `-f` semantics. Never put secrets in it: the chart refuses to render secrets into its config ConfigMap, and every typed credential rides Secrets and environment expansion instead.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesStorageClass** | `storage.storageClass` | `status.outputs.storage_class_name` |
| **KubernetesPostgres** | `database.host` | `status.outputs.kube_endpoint` |
| **KubernetesKubePrometheusStack** | `datasources[].url` | `status.outputs.prometheus_endpoint` |

Secret references (`adminSecret`, `database.passwordSecret`, datasource basic-auth, `smtp.credentialsSecretName`) name existing Kubernetes Secrets directly — they are plain names, not ValueFromRef foreign keys.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace Grafana runs in | Application deployment manifests |
| `release_name` | Helm release name (= metadata.name) | Operational tooling |
| `service` | The Grafana Service (port 80 → container 3000) | Composed HTTP exposure |
| `endpoint` | In-cluster endpoint, e.g. `http://dashboards.observability.svc.cluster.local` | In-cluster API clients, embedded links |
| `admin_secret_name` | The Secret holding the admin credentials (keys `admin-user` / `admin-password`) — chart-generated, or your own name echoed back | Sign-in, credential automation |
| `port_forward_command` | Copy-paste `kubectl port-forward` for the UI | Local development access |
| `agent_reader_token_secret` | The Secret key holding agent teammates' read-only token (`<name>-agent-reader`, key `token`) — unset unless `agentReader` is declared and not disabled | Grafana's MCP server (`GRAFANA_SERVICE_ACCOUNT_TOKEN`), read at each start |
| `agent_reader_job_name` | The Job keeping the agent teammates' account and token — `kubectl logs job/<this>` explains each run | Diagnosing the agent reader |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev dashboards** — ephemeral (the chart default): one replica, chart-generated admin credentials, a single provisioned Prometheus datasource — a real dashboard endpoint for a dev loop without production ceremony. Start from the **Dev dashboards preset**.

**Persistent team instance** — `storage` backs a 10Gi volume: the single stateful instance most teams want, where hand-built dashboards survive restarts, the datasource is wired by reference to the metrics stack, and the public root URL is set for composed exposure. Start from the **Persistent team preset**.

**HA with an external database** — two replicas behind one Service with ALL state in Postgres, a team-owned admin Secret, and the ServiceMonitor on — the production posture, and the only honest way to scale Grafana. Start from the **HA external database preset**.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — referenced placement; the InfraPipeline orders namespace-first.
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack) — its exported Prometheus endpoint is the classic first datasource; its operator CRDs are the ServiceMonitor prerequisite.
- [**PostgreSQL**](/infra-catalog/kubernetes-postgres) and [**MySQL**](/infra-catalog/kubernetes-mysql) — the external state database behind the HA posture, wired by its read-write endpoint.
- [**Kubernetes StorageClass**](/infra-catalog/kubernetes-storage-class) — explicit placement for the state volume.
- [**Kubernetes Secret**](/infra-catalog/kubernetes-secret) — bring-your-own admin credentials, the database password, datasource basic-auth, SMTP credentials — always by reference, never material.
- [**Kubernetes HTTPRoute**](/infra-catalog/kubernetes-http-route) — HTTP exposure composed over the exported `service` handle, with `server.rootUrl` set to match.
