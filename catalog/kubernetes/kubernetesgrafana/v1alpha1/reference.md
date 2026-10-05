# KubernetesGrafana

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesGrafanaSpec** deploys Grafana — the observability
dashboard — from the official `grafana` Helm chart
(https://grafana-community.github.io/helm-charts).

WHEN TO USE THIS KIND: standalone Grafana is the composition hub —
point it at any mix of datasources (a KubernetesKubePrometheusStack,
external Prometheus/Mimir, Loki, Tempo, ClickHouse, Postgres) and run
it independently of any one of them. If all you need is dashboards for
ONE kube-prometheus-stack, its bundled Grafana (on by default there)
is the simpler path.

CREDENTIALS: the chart generates a random admin password ONCE at first
install (stable across upgrades) into its own `<name>` Secret — keys
`admin-user` / `admin-password` — unless `admin_secret` points at an
existing Secret. Credentials never appear in rendered Helm values.

STATE: Grafana keeps UI-authored dashboards, users and preferences in
an embedded SQLite database on local disk. The chart's default is
EPHEMERAL — everything hand-made vanishes on pod restart. Declare
`storage` for a single stateful instance, or `database` (external
Postgres/MySQL) which is also the REQUIREMENT for running more than
one replica — SQLite cannot be shared, so `replicas > 1` without a
database splits sessions and dashboards across pods.

PROVISIONING AS CODE: `datasources` below renders Grafana's datasource
provisioning file, and `community_dashboards` imports dashboards by
grafana.com ID — the declarative path that survives pod restarts
without persistence. The dashboard SIDECAR (on by default) discovers
any ConfigMap labeled `grafana_dashboard: "1"` cluster-wide — the
contract by which other components and teams ship dashboards to this
Grafana without touching its spec.

SIGN-IN: `auth.google` and `auth.generic_oauth` put Grafana behind a
company's identity provider. The client secret is a managed-secret
reference the modules write into their own `<name>-sso` Secret; once
either is declared the manifest owns sign-in and Grafana's
Administration > Authentication screen can no longer change it.

EXPOSURE: the service stays ClusterIP; expose via first-class kinds
(KubernetesIngress, Gateway API kinds) over the exported service
handle. Set `server.root_url` to the public URL when composing
exposure — sign-in redirects and rendered links depend on it.

The typed fields below cover the chart's meaningful configuration
surface; `helm_values` remains as the escape hatch for chart values
beyond them (merged last, Helm `-f` semantics, identical on both
engines) — LDAP, the image renderer, alerting provisioning, extra
sidecars — a safety valve, never the primary interface. Never put
secret material in `helm_values`: the chart refuses to render secrets
into its config ConfigMap, and the typed fields wire every credential
through Secrets and environment variables instead.

## Example

```yaml
# Full-surface offline-proof manifest: exercises an HA pair on an external
# Postgres database (password via Secret + env expansion), an existing
# admin Secret, three provisioned datasources (a literal Prometheus URL, a
# basic-auth Mimir with the $__env password fold, and a Loki with typed
# jsonData), the dashboard sidecar, pinned community dashboards, plugins
# (including a Grafana-13 moved-out-of-core datasource plugin), server
# root_url + anonymous viewing, Google and generic OAuth sign-in (the
# module-owned `<name>-sso` Secret, the admin-screen lock and the
# credentials checksum), SMTP with a credentials Secret, the
# ServiceMonitor toggle, a private-mirror image with a pull secret,
# scheduling, and an escape-hatch entry — so the offline tofu plan and
# pulumi preview proofs cover the full typed surface. Placeholder values;
# never applied to a real cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGrafana
metadata:
  name: grafana-hack
spec:
  namespace:
    value: grafana-hack
  createNamespace: true
  chartVersion: 12.8.0
  replicas: 2
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: "1"
      memory: 1Gi
  adminSecret:
    name: grafana-hack-admin
    userKey: user
    passwordKey: password
  database:
    engine: postgres
    host:
      value: grafana-db-rw.data.svc.cluster.local:5432
    name: grafana
    user: grafana
    passwordSecret:
      name: grafana-hack-db
      key: password
    sslMode: require
  datasources:
    - name: Prometheus
      url:
        value: http://monitoring-prometheus.observability.svc.cluster.local:9090
      isDefault: true
      uid: prom-main
      jsonData: |
        httpMethod: POST
    - name: External Mimir
      type: prometheus
      url:
        value: https://mimir.example.com/prometheus
      basicAuth:
        username: tenant-1
        passwordSecret:
          name: mimir-credentials
          key: password
    - name: Loki
      type: loki
      url:
        value: http://loki-gateway.observability.svc.cluster.local
      jsonData: |
        maxLines: 1000
  dashboardSidecarEnabled: true
  communityDashboards:
    - gnetId: 1860
      revision: 37
      datasource: Prometheus
    - gnetId: 15757
      revision: 43
      datasource: Prometheus
  plugins:
    - grafana-clock-panel
    - elasticsearch
  server:
    rootUrl: https://grafana.example.com
  auth:
    anonymousEnabled: true
    anonymousOrgRole: Viewer
    google:
      clientId: 123456789-hack.apps.googleusercontent.com
      clientSecret:
        value: placeholder-google-client-secret
      allowedDomains:
        - example.com
      hostedDomain: example.com
      roleAttributePath: "email == 'lead@example.com' && 'Admin' || 'Viewer'"
    genericOauth:
      name: Okta
      clientId: 0oa1hack
      clientSecret:
        value: placeholder-okta-client-secret
      authUrl: https://id.example.com/oauth2/v1/authorize
      tokenUrl: https://id.example.com/oauth2/v1/token
      apiUrl: https://id.example.com/oauth2/v1/userinfo
      groupsAttributePath: groups
      allowedGroups:
        - platform
  smtp:
    host: smtp.example.com:587
    fromAddress: grafana@example.com
    fromName: Grafana Alerts
    credentialsSecretName: grafana-hack-smtp
  serviceMonitorEnabled: true
  image:
    repository: mirror.example.com/grafana/grafana
    pullSecretName: mirror-pull
  scheduling:
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: dedicated
        operator: Equal
        value: monitoring
        effect: NoSchedule
    priorityClassName: system-cluster-critical
  helmValues: |
    revisionHistoryLimit: 5
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `12.8.0` |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.adminSecret` | `KubernetesGrafanaAdminSecret` |  |  |  |
| `spec.adminSecret.name` | `string` | yes |  |  |
| `spec.adminSecret.userKey` | `string` |  | `admin-user` |  |
| `spec.adminSecret.passwordKey` | `string` |  | `admin-password` |  |
| `spec.storage` | `KubernetesGrafanaStorage` |  |  |  |
| `spec.storage.size` | `string` |  | `10Gi` |  |
| `spec.storage.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.database` | `KubernetesGrafanaDatabase` |  |  |  |
| `spec.database.engine` | `enum` | yes |  |  |
| `spec.database.host` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.kube_endpoint`) |
| `spec.database.name` | `string` | yes |  |  |
| `spec.database.user` | `string` | yes |  |  |
| `spec.database.passwordSecret` | `KubernetesGrafanaSecretKeyRef` | yes |  |  |
| `spec.database.passwordSecret.name` | `string` | yes |  |  |
| `spec.database.passwordSecret.key` | `string` | yes |  |  |
| `spec.database.sslMode` | `string` |  |  |  |
| `spec.datasources` | `[]KubernetesGrafanaDatasource` |  |  |  |
| `spec.datasources[].name` | `string` | yes |  |  |
| `spec.datasources[].type` | `string` |  | `prometheus` |  |
| `spec.datasources[].url` | `string \| valueFrom` | yes |  | KubernetesKubePrometheusStack (`status.outputs.prometheus_endpoint`) |
| `spec.datasources[].isDefault` | `bool` |  |  |  |
| `spec.datasources[].uid` | `string` |  |  |  |
| `spec.datasources[].basicAuth` | `KubernetesGrafanaDatasourceBasicAuth` |  |  |  |
| `spec.datasources[].basicAuth.username` | `string` | yes |  |  |
| `spec.datasources[].basicAuth.passwordSecret` | `KubernetesGrafanaSecretKeyRef` | yes |  |  |
| `spec.datasources[].basicAuth.passwordSecret.name` | `string` | yes |  |  |
| `spec.datasources[].basicAuth.passwordSecret.key` | `string` | yes |  |  |
| `spec.datasources[].jsonData` | `string` |  |  |  |
| `spec.dashboardSidecarEnabled` | `bool` |  | `true` |  |
| `spec.communityDashboards` | `[]KubernetesGrafanaCommunityDashboard` |  |  |  |
| `spec.communityDashboards[].gnetId` | `int32` |  |  |  |
| `spec.communityDashboards[].revision` | `int32` |  |  |  |
| `spec.communityDashboards[].datasource` | `string` | yes |  |  |
| `spec.plugins` | `[]string` |  |  |  |
| `spec.server` | `KubernetesGrafanaServer` |  |  |  |
| `spec.server.rootUrl` | `string` |  |  |  |
| `spec.auth` | `KubernetesGrafanaAuth` |  |  |  |
| `spec.auth.anonymousEnabled` | `bool` |  |  |  |
| `spec.auth.anonymousOrgRole` | `string` |  | `Viewer` |  |
| `spec.auth.disableLoginForm` | `bool` |  |  |  |
| `spec.auth.google` | `KubernetesGrafanaGoogleSignIn` |  |  |  |
| `spec.auth.google.clientId` | `string` | yes |  |  |
| `spec.auth.google.clientSecret` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.auth.google.allowedDomains` | `[]string` |  |  |  |
| `spec.auth.google.hostedDomain` | `string` |  |  |  |
| `spec.auth.google.allowSignUp` | `bool` |  | `true` |  |
| `spec.auth.google.autoLogin` | `bool` |  |  |  |
| `spec.auth.google.roleAttributePath` | `string` |  |  |  |
| `spec.auth.google.roleAttributeStrict` | `bool` |  |  |  |
| `spec.auth.genericOauth` | `KubernetesGrafanaGenericOAuthSignIn` |  |  |  |
| `spec.auth.genericOauth.name` | `string` |  | `OAuth` |  |
| `spec.auth.genericOauth.clientId` | `string` | yes |  |  |
| `spec.auth.genericOauth.clientSecret` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.auth.genericOauth.authUrl` | `string` | yes |  |  |
| `spec.auth.genericOauth.tokenUrl` | `string` | yes |  |  |
| `spec.auth.genericOauth.apiUrl` | `string` | yes |  |  |
| `spec.auth.genericOauth.scopes` | `[]string` |  |  |  |
| `spec.auth.genericOauth.emailAttributePath` | `string` |  |  |  |
| `spec.auth.genericOauth.loginAttributePath` | `string` |  |  |  |
| `spec.auth.genericOauth.nameAttributePath` | `string` |  |  |  |
| `spec.auth.genericOauth.roleAttributePath` | `string` |  |  |  |
| `spec.auth.genericOauth.roleAttributeStrict` | `bool` |  |  |  |
| `spec.auth.genericOauth.groupsAttributePath` | `string` |  |  |  |
| `spec.auth.genericOauth.allowedGroups` | `[]string` |  |  |  |
| `spec.auth.genericOauth.allowedDomains` | `[]string` |  |  |  |
| `spec.auth.genericOauth.allowSignUp` | `bool` |  | `true` |  |
| `spec.auth.genericOauth.autoLogin` | `bool` |  |  |  |
| `spec.auth.genericOauth.usePkce` | `bool` |  | `true` |  |
| `spec.smtp` | `KubernetesGrafanaSmtp` |  |  |  |
| `spec.smtp.host` | `string` | yes |  |  |
| `spec.smtp.fromAddress` | `string` |  |  |  |
| `spec.smtp.fromName` | `string` |  |  |  |
| `spec.smtp.credentialsSecretName` | `string` |  |  |  |
| `spec.smtp.skipVerify` | `bool` |  |  |  |
| `spec.serviceMonitorEnabled` | `bool` |  |  |  |
| `spec.image` | `KubernetesGrafanaImage` |  |  |  |
| `spec.image.repository` | `string` |  |  |  |
| `spec.image.tag` | `string` |  |  |  |
| `spec.image.pullSecretName` | `string` |  |  |  |
| `spec.scheduling` | `KubernetesGrafanaScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.scheduling.priorityClassName` | `string` |  |  |  |
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

Helm chart version to install (e.g. "12.8.0" — chart 12.8.0 ships
Grafana 13.1.1). Versions must exist as SERVED charts in the
repository index (https://grafana-community.github.io/helm-charts —
the chart's current home; the old grafana.github.io repository
stopped serving new versions at 10.5.x).

- default: `12.8.0`

### spec.replicas

`int32` · optional (explicit presence)

Number of Grafana replicas. More than 1 REQUIRES `database` — the
embedded SQLite state cannot be shared between pods (enforced
below). With a database, replicas serve behind one Service for HA.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.resources

`ContainerResources`

CPU and memory for the Grafana container. Empty = no requests/limits
(the chart default).

### spec.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.resources.limits.cpu

`string`

### spec.resources.limits.memory

`string`

### spec.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.resources.requests.cpu

`string`

### spec.resources.requests.memory

`string`

### spec.adminSecret

`KubernetesGrafanaAdminSecret`

Read the admin credentials from an existing Secret. Empty = the
chart generates them (see the CREDENTIALS note on the spec).

### spec.adminSecret.name

`string` · required

Secret name (a reference to an existing Kubernetes Secret, never
secret material). Must exist BEFORE the install — the chart wires
it at template time.

- rule: {"required":true}

### spec.adminSecret.userKey

`string` · optional (explicit presence)

Key holding the admin username. Empty = "admin-user".

- default: `admin-user`

### spec.adminSecret.passwordKey

`string` · optional (explicit presence)

Key holding the admin password. Empty = "admin-password".

- default: `admin-password`

### spec.storage

`KubernetesGrafanaStorage`

Persistent storage for Grafana's embedded state. Empty = ephemeral
(the chart default) — right when everything is provisioned as code
or a `database` carries the state; wrong if people build dashboards
in this UI by hand. Mutually exclusive with `replicas > 1` unless a
`database` is declared (the PVC is ReadWriteOnce).

### spec.storage.size

`string` · optional (explicit presence)

Volume size as a Kubernetes quantity (e.g. "10Gi").

- default: `10Gi`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?\\s?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.storage.storageClass

`string | valueFrom`

Storage class for the PVC. Accepts a literal class name or a
reference to a KubernetesStorageClass resource. Empty = the
cluster's default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.database

`KubernetesGrafanaDatabase`

External database (Postgres or MySQL) for Grafana's state — the HA
path, and the durable path that makes `storage` unnecessary.
The password rides an existing Secret through environment
expansion; it never lands in Grafana's rendered configuration.

### spec.database.engine

`enum` · required

Database engine.

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `kubernetes_grafana_database_engine_unspecified` -- Unspecified. Declare the engine explicitly.
- `postgres` -- PostgreSQL — pairs naturally with a KubernetesPostgres resource.
- `mysql` -- MySQL — pairs naturally with a KubernetesMysql resource.

### spec.database.host

`string | valueFrom` · required

Database host with port (e.g.
"grafana-db-rw.data.svc.cluster.local:5432"). Accepts a literal
endpoint or a reference to a KubernetesPostgres resource (its
read-write endpoint).

- references: KubernetesPostgres (`status.outputs.kube_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.kube_endpoint}} -- a bare string does not parse

### spec.database.name

`string` · required

Database name. The database must exist — Grafana creates its tables
but never the database itself.

- rule: {"required":true}

### spec.database.user

`string` · required

Database user.

- rule: {"required":true}

### spec.database.passwordSecret

`KubernetesGrafanaSecretKeyRef` · required

Existing Secret holding the database password. The modules inject it
as an environment variable the configuration expands at runtime —
the password never appears in Grafana's rendered config.

- rule: {"required":true}

### spec.database.passwordSecret.name

`string` · required

Secret name. The Secret must live in Grafana's namespace.

- rule: {"required":true}

### spec.database.passwordSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.database.sslMode

`string`

Postgres SSL mode ("disable", "require", "verify-full", ...). Empty
= the Grafana default ("disable" in-cluster). Ignored for mysql.

### spec.datasources

`[]KubernetesGrafanaDatasource`

Datasources provisioned as code. Each entry renders into Grafana's
datasource provisioning file — present from first boot, no clicking.

### spec.datasources[].name

`string` · required

Display name of the datasource (e.g. "Prometheus", "Loki").

- rule: {"required":true}

### spec.datasources[].type

`string` · optional (explicit presence)

Datasource type — the Grafana plugin ID: "prometheus", "loki",
"tempo", "postgres", "mysql", "elasticsearch", "cloudwatch", or any
installed plugin's ID. Empty = "prometheus". (Remember: on Grafana
13, elasticsearch and cloudwatch also need the plugin listed in
`plugins`.)

- default: `prometheus`

### spec.datasources[].url

`string | valueFrom` · required

The datasource URL (e.g. "http://monitoring-prometheus.observability.svc.cluster.local:9090").
Accepts a literal URL or a reference to a
KubernetesKubePrometheusStack resource (its Prometheus endpoint) —
the one-line wiring that points this Grafana at the cluster's
metrics.

- references: KubernetesKubePrometheusStack (`status.outputs.prometheus_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKubePrometheusStack, name: <that resource's name>, fieldPath: status.outputs.prometheus_endpoint}} -- a bare string does not parse

### spec.datasources[].isDefault

`bool`

Make this the default datasource new panels start from. Declare on
exactly one entry.

### spec.datasources[].uid

`string`

Stable UID other provisioned dashboards reference. Empty = derived
by Grafana. Set it when dashboards-as-code pin their datasource by
UID.

### spec.datasources[].basicAuth

`KubernetesGrafanaDatasourceBasicAuth`

HTTP basic auth towards the datasource. The password rides an
existing Secret through environment expansion — it never lands in
the rendered provisioning file.

### spec.datasources[].basicAuth.username

`string` · required

Username.

- rule: {"required":true}

### spec.datasources[].basicAuth.passwordSecret

`KubernetesGrafanaSecretKeyRef` · required

Existing Secret holding the password.

- rule: {"required":true}

### spec.datasources[].basicAuth.passwordSecret.name

`string` · required

Secret name. The Secret must live in Grafana's namespace.

- rule: {"required":true}

### spec.datasources[].basicAuth.passwordSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.datasources[].jsonData

`string`

Extra type-specific settings rendered into the datasource's
jsonData (e.g. httpMethod: POST for Prometheus, maxLines for Loki)
as YAML. Never put credentials here — jsonData renders in clear.

### spec.dashboardSidecarEnabled

`bool` · optional (explicit presence)

Discover dashboards from ConfigMaps labeled `grafana_dashboard: "1"`
anywhere in the cluster (the k8s-sidecar watcher). Default true —
this is the composition contract: other components and teams ship
dashboards by creating labeled ConfigMaps, never by editing this
resource.

- default: `true`

### spec.communityDashboards

`[]KubernetesGrafanaCommunityDashboard`

Community dashboards imported by grafana.com ID at install (e.g.
1860 = Node Exporter Full). Pin `revision` for reproducible
installs; a moving latest revision can change the dashboard under
you.

### spec.communityDashboards[].gnetId

`int32`

The grafana.com dashboard ID (the number in
https://grafana.com/grafana/dashboards/<id>).

- rule: {"int32":{"gte":1}}

### spec.communityDashboards[].revision

`int32`

Dashboard revision to pin. Empty = the latest revision at install
time — pin it for reproducible installs.

### spec.communityDashboards[].datasource

`string` · required

Name of the provisioned datasource the dashboard's panels bind to
(must match a `datasources` entry's name).

- rule: {"required":true}

### spec.plugins

`[]string`

Grafana plugins to install at startup, by plugin ID with optional
version (e.g. "grafana-clock-panel",
"grafana-oncall-app 1.3.0"). KNOW THIS: Grafana 13 moved several
once-core datasource plugins (elasticsearch, cloudwatch) out of the
core image — list them here to keep using them; the modules enable
the chart's bundled-plugin shadowing automatically so the install
succeeds on the read-only image directory.

### spec.server

`KubernetesGrafanaServer`

The server identity block of Grafana's configuration.

### spec.server.rootUrl

`string`

The public URL users reach Grafana at (e.g.
"https://grafana.example.com"). OAuth redirect URLs, alert links
and rendered images embed it — set it whenever exposure is composed
in front of this Grafana. Empty = Grafana's localhost default.

### spec.auth

`KubernetesGrafanaAuth`

Who can sign in: anonymous access, the login form, and sign-in
through Google or any OAuth 2.0 / OpenID Connect provider.

### spec.auth.anonymousEnabled

`bool`

Allow anonymous (no-login) access.

### spec.auth.anonymousOrgRole

`string` · optional (explicit presence)

Organization role anonymous visitors get. Empty = "Viewer" — never
hand anonymous users more than Viewer on a reachable endpoint.

- default: `Viewer`

### spec.auth.disableLoginForm

`bool`

Hide the login form (for pure-SSO or pure-anonymous deployments —
make sure another auth path exists, or the UI locks everyone out).
With sign-in declared, keeping the form is the break-glass path: the
admin account still signs in if the identity provider is down.

### spec.auth.google

`KubernetesGrafanaGoogleSignIn`

Sign in with Google. The client is a Web application OAuth client
whose authorized redirect URI is `<server.root_url>/login/google`.

KNOW THIS: who a Google client admits is decided at Google by its
project's consent-screen audience. An Internal audience admits only
the Workspace that owns the project; an External one admits any
Google account (in Testing, only the listed test users). Grafana's
own gate is `allowed_domains`, checked against the signed-in email.
Declaring either sign-in locks every OAuth provider on Grafana's
Administration > Authentication screen, so the manifest stays the only
source of truth for who can sign in (Grafana always leaves LDAP
editable there).

- rule: Google sign-in with allow_sign_up on and no allowed_domains lets any Google account on the internet create a Viewer. List your email domains in allowed_domains (e.g. ["example.com"]), or set allow_sign_up: false so only users that already exist in Grafana can sign in

### spec.auth.google.clientId

`string` · required

The OAuth client's id (`<number>-<hash>.apps.googleusercontent.com`).
A public identifier, never a secret; a `$var/` reference keeps it
with the rest of the client's records.

- rule: {"required":true}

### spec.auth.google.clientSecret

`string | valueFrom` · required · sensitive

The OAuth client's secret. Reference-only: the platform refuses a
literal and resolves the `$secret/` reference at deploy. The modules
write it into the `<name>-sso` Secret and hand it to Grafana as an
environment variable. Grafana reads it only at start, so the pods
carry a checksum of the Secret and a rotated value rolls them on the
next apply.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.auth.google.allowedDomains

`[]string`

Email domains allowed to sign in (e.g. "example.com"). Grafana
refuses any account whose email is outside them. Required while
`allow_sign_up` is on, because a Google client otherwise lets any
Google account create a Viewer.

### spec.auth.google.hostedDomain

`string`

The Google Workspace domain Google's account chooser is limited to
(sent as the `hd` parameter). A hint to Google, not a gate — the gate
is `allowed_domains` and the client's consent-screen audience.

### spec.auth.google.allowSignUp

`bool` · optional (explicit presence)

Create a Grafana user on first sign-in. Off = only users that
already exist in Grafana can sign in. Empty = true.

- default: `true`

### spec.auth.google.autoLogin

`bool`

Skip Grafana's login page and go straight to Google. Leave off while
the login form is the break-glass path people need to find.

### spec.auth.google.roleAttributePath

`string`

JMESPath over the signed-in person's claims that yields their
Grafana role ("Admin", "Editor", "Viewer"), e.g.
`email == 'lead@example.com' && 'Admin' || 'Viewer'`. Empty = every
new user gets Grafana's default role (Viewer) and admins assign roles
in Grafana. Declared = re-evaluated at every sign-in, so a role changed
here takes effect on the next sign-in. KNOW THIS: Grafana ships Google
with role sync off (`skip_org_role_sync = true`), which silently ignores
a role path; the modules switch sync on whenever a path is declared.

### spec.auth.google.roleAttributeStrict

`bool`

Refuse sign-in when `role_attribute_path` yields no valid role,
instead of falling back to the default role.

### spec.auth.genericOauth

`KubernetesGrafanaGenericOAuthSignIn`

Sign in with any OAuth 2.0 / OpenID Connect provider (Okta,
Microsoft Entra ID, Keycloak, Auth0, GitLab, ...). The client's
redirect URI is `<server.root_url>/login/generic_oauth`.

- rule: allowed_groups needs groups_attribute_path, the JMESPath to the groups list in the provider's claims (often "groups"); without it Grafana sees no groups and refuses everyone

### spec.auth.genericOauth.name

`string` · optional (explicit presence)

The label on the sign-in button ("Sign in with <name>"). Empty =
"OAuth".

- default: `OAuth`

### spec.auth.genericOauth.clientId

`string` · required

The OAuth client's id at the provider. A public identifier.

- rule: {"required":true}

### spec.auth.genericOauth.clientSecret

`string | valueFrom` · required · sensitive

The OAuth client's secret. Reference-only, exactly as Google's
`client_secret`: resolved at deploy, written into the `<name>-sso`
Secret, rolled into the pods on the next apply after a rotation.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.auth.genericOauth.authUrl

`string` · required

The provider's authorization endpoint (e.g.
"https://id.example.com/oauth2/v1/authorize"). Grafana does not read
OpenID discovery documents, so the three endpoints are declared.

- rule: {"required":true}

### spec.auth.genericOauth.tokenUrl

`string` · required

The provider's token endpoint.

- rule: {"required":true}

### spec.auth.genericOauth.apiUrl

`string` · required

The provider's userinfo endpoint.

- rule: {"required":true}

### spec.auth.genericOauth.scopes

`[]string`

OAuth scopes to request. Empty = "openid", "email", "profile".

### spec.auth.genericOauth.emailAttributePath

`string`

JMESPath to the email in the provider's claims. Empty = Grafana's
default lookup (the `email` claim).

### spec.auth.genericOauth.loginAttributePath

`string`

JMESPath to the login name. Empty = Grafana's default lookup.

### spec.auth.genericOauth.nameAttributePath

`string`

JMESPath to the display name. Empty = Grafana's default lookup.

### spec.auth.genericOauth.roleAttributePath

`string`

JMESPath over the claims that yields the Grafana role, e.g.
`contains(groups[*], 'platform') && 'Admin' || 'Viewer'`. Empty =
the default role (Viewer).

### spec.auth.genericOauth.roleAttributeStrict

`bool`

Refuse sign-in when `role_attribute_path` yields no valid role.

### spec.auth.genericOauth.groupsAttributePath

`string`

JMESPath to the list of groups in the claims (e.g. "groups"). Needed
by `allowed_groups`.

### spec.auth.genericOauth.allowedGroups

`[]string`

Groups allowed to sign in; a person in none of them is refused.
Requires `groups_attribute_path`.

### spec.auth.genericOauth.allowedDomains

`[]string`

Email domains allowed to sign in.

### spec.auth.genericOauth.allowSignUp

`bool` · optional (explicit presence)

Create a Grafana user on first sign-in. Empty = true.

- default: `true`

### spec.auth.genericOauth.autoLogin

`bool`

Skip Grafana's login page and go straight to the provider.

### spec.auth.genericOauth.usePkce

`bool` · optional (explicit presence)

Use PKCE on the authorization request. Empty = true; turn it off
only for a provider that rejects it.

- default: `true`

### spec.smtp

`KubernetesGrafanaSmtp`

Outbound email (alert notifications, invites, password resets).
Credentials ride an existing Secret.

### spec.smtp.host

`string` · required

SMTP host with port (e.g. "smtp.example.com:587").

- rule: {"required":true}

### spec.smtp.fromAddress

`string`

From address on outgoing mail. Empty = Grafana's default
("admin@grafana.localhost" — set a real one).

### spec.smtp.fromName

`string`

From display name. Empty = "Grafana".

### spec.smtp.credentialsSecretName

`string`

Existing Secret holding the SMTP credentials — keys `user` and
`password` (override the key names via the chart's smtp values in
`helm_values` if the Secret is shaped differently). Empty =
unauthenticated SMTP.

### spec.smtp.skipVerify

`bool`

Skip TLS certificate verification towards the SMTP host — test
fixtures only.

### spec.serviceMonitorEnabled

`bool`

Create a ServiceMonitor for Prometheus scraping of Grafana's own
/metrics (requires the Prometheus Operator CRDs — deploy
KubernetesKubePrometheusStack first). Chart default: false.

### spec.image

`KubernetesGrafanaImage`

Override the Grafana image (air-gap path). Empty = the chart's
official `docker.io/grafana/grafana` at the chart's appVersion.

### spec.image.repository

`string`

Image repository including registry, e.g.
"my.registry.com/grafana/grafana". Empty = "docker.io/grafana/grafana".

### spec.image.tag

`string`

Image tag. Empty = the chart's appVersion for the pinned
chart_version.

### spec.image.pullSecretName

`string`

Name of an existing image-pull Secret in the namespace, for private
mirrors.

### spec.scheduling

`KubernetesGrafanaScheduling`

Scheduling for the Grafana pods.

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for the Grafana pods.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the Grafana pods.

### spec.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.scheduling.priorityClassName

`string`

Priority class name for the Grafana pods.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f` semantics,
identical on both engines). For the chart surface beyond the typed
fields (LDAP in grafana.ini, the image renderer, alerting
provisioning, notifiers, extra mounts/sidecars, ...) — never the
substitute for them. Sign-in has typed fields (`auth.google`,
`auth.generic_oauth`). Do not put secrets here; credential material
belongs in the typed secret references.

## Validation Rules

- `spec.auth.sign_in_requires_root_url`: Google or OAuth sign-in needs server.root_url, the public address people open Grafana at: the provider sends people back to <root_url>/login/google (or /login/generic_oauth), and without it Grafana hands the provider http://localhost:3000, which no browser can reach. Set server.root_url to the URL the exposure layer serves (e.g. https://grafana.example.com)
- `spec.replicas.require_database`: replicas above 1 require the database block — Grafana's embedded SQLite state cannot be shared between pods, so a scaled deployment without an external database splits dashboards and sessions across replicas
- `spec.storage.single_writer`: storage (a ReadWriteOnce volume) cannot back more than one replica — for HA use database for state and leave storage unset

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesGrafana, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace Grafana runs in. |
| `status.outputs.release_name` | `string` | Helm release name (= metadata.name). The modules pin the chart's fullname to it, so the Service and Secret names derive from it. |
| `status.outputs.service` | `string` | name of the Grafana Service (port 80 → container 3000). |
| `status.outputs.endpoint` | `string` | in-cluster endpoint for browsers behind composed exposure and for in-cluster API clients, e.g. http://dashboards.observability.svc.cluster.local |
| `status.outputs.admin_secret_name` | `string` | name of the Secret holding the admin credentials — `<name>`, keys `admin-user` / `admin-password` (the chart generates it once and keeps it stable across upgrades; when spec.admin_secret points at an existing Secret, that name is echoed here instead). |
| `status.outputs.port_forward_command` | `string` | command to port-forward the Grafana UI to a developer laptop, e.g. kubectl port-forward svc/dashboards -n observability 3000:80 |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.storage.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.database.host` | KubernetesPostgres | `status.outputs.kube_endpoint` |
| `spec.datasources[].url` | KubernetesKubePrometheusStack | `status.outputs.prometheus_endpoint` |

## See Also

- [Overview](../README.md)
