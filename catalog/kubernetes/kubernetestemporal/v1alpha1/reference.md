# KubernetesTemporal

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**KubernetesTemporalSpec** deploys Temporal — the durable workflow
engine (long-running business logic, human-in-the-loop flows, saga
orchestration, AI-agent pipelines) — from the official `temporal`
Helm chart (https://go.temporal.io/helm-charts).

WHAT GETS INSTALLED: the four Temporal server services (frontend,
history, matching, worker) as separate Deployments, the Web UI (on
by default), an admin-tools pod for operational commands, and the
schema-setup Jobs that prepare the databases before the server
starts.

BRING YOUR OWN DATABASE — nothing is bundled. Temporal stores every
workflow's state in a database you declare under `database`:
PostgreSQL (a KubernetesPostgres composes naturally — the
recommended path), MySQL (a KubernetesMysql composes), or an
external Cassandra cluster you operate yourself. The chart's old
bundled Cassandra/Elasticsearch/Prometheus/Grafana subcharts were
removed upstream; declaring their legacy keys through `helm_values`
makes the chart itself fail rendering.

TWO DATABASES, ONE SERVER: Temporal keeps workflow state in the
default store (`database_name`, default "temporal") and its search
index in the visibility store (`visibility_database_name`, default
"temporal_visibility"). Both must EXIST before install unless
`create_databases` is set (which needs a database user with
create-database privileges). On a KubernetesPostgres, declare the
first database at bootstrap and create the second with one line of
`post_init_sql` — both owned by the same application user.

EXPOSURE: the frontend (gRPC 7233) and Web UI (HTTP 8080) Services
stay ClusterIP; expose them via first-class kinds (KubernetesService
for a LoadBalancer, Gateway API kinds for routes) over the exported
service handles. Workers connect in-cluster through
`frontend_endpoint`.

The typed fields below cover the chart's meaningful configuration
surface; `helm_values` remains as the escape hatch for chart values
beyond them (merged last, Helm `-f` semantics, identical on both
engines) — mTLS certificate mounts, JWT authorization, multi-cluster
replication (clusterMetadata), extra dynamic-config keys — a safety
valve, never the primary interface. Never put secret material in
`helm_values`; database passwords ride existing Secrets through the
typed fields and never land in rendered values.

## Example

```yaml
# Full-surface shape: the PostgreSQL arm with per-service sizing, web UI,
# declarative Temporal namespaces, dynamic-config limits, S3 archival,
# ServiceMonitor and an escape-hatch entry — the offline plan/preview
# proof for the widest typed rendering.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTemporal
metadata:
  name: temporal-full
spec:
  namespace:
    value: temporal-full
  createNamespace: true
  database:
    postgres:
      host:
        value: temporal-pg-rw
      username: temporal
      passwordSecret:
        secretName:
          value: temporal-pg-app
      maxConns: 30
      tls:
        enabled: true
        hostVerification: true
        serverName: temporal-pg.example.internal
    createDatabases: false
  numHistoryShards: 1024
  services:
    frontend:
      replicas: 2
      resources:
        requests:
          cpu: 200m
          memory: 256Mi
    history:
      replicas: 2
      resources:
        requests:
          cpu: 500m
          memory: 1Gi
        limits:
          memory: 2Gi
    matching:
      replicas: 2
    worker:
      replicas: 1
  internalFrontendEnabled: true
  webUi:
    replicas: 2
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
  temporalNamespaces:
    - name: default
    - name: payments
      retention: 30d
  dynamicConfig:
    blobSizeLimitError: 10485760
    blobSizeLimitWarn: 5242880
    historySizeLimitError: 104857600
  archival:
    s3:
      region: us-east-1
    historyUri: s3://temporal-archival/history
    visibilityUri: s3://temporal-archival/visibility
  serviceMonitorEnabled: true
  logLevel: info
  scheduling:
    nodeSelector:
      workload-tier: platform
  helmValues: |
    server:
      podLabels:
        team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `1.6.0` |  |
| `spec.database` | `KubernetesTemporalDatabase` | yes |  |  |
| `spec.database.postgres` | `KubernetesTemporalPostgres` |  |  |  |
| `spec.database.postgres.host` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.rw_service`) |
| `spec.database.postgres.port` | `int32` |  | `5432` |  |
| `spec.database.postgres.username` | `string` | yes |  |  |
| `spec.database.postgres.passwordSecret` | `KubernetesTemporalPasswordSecret` | yes |  |  |
| `spec.database.postgres.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.password_secret.name`) |
| `spec.database.postgres.passwordSecret.secretKey` | `string` |  | `password` |  |
| `spec.database.postgres.maxConns` | `int32` |  | `20` |  |
| `spec.database.postgres.maxIdleConns` | `int32` |  | `20` |  |
| `spec.database.postgres.maxConnLifetime` | `string` |  | `1h` |  |
| `spec.database.postgres.tls` | `KubernetesTemporalDatabaseTls` |  |  |  |
| `spec.database.postgres.tls.enabled` | `bool` |  |  |  |
| `spec.database.postgres.tls.hostVerification` | `bool` |  |  |  |
| `spec.database.postgres.tls.serverName` | `string` |  |  |  |
| `spec.database.mysql` | `KubernetesTemporalMysql` |  |  |  |
| `spec.database.mysql.host` | `string \| valueFrom` | yes |  | KubernetesMysql (`status.outputs.primary_service`) |
| `spec.database.mysql.port` | `int32` |  | `3306` |  |
| `spec.database.mysql.username` | `string` | yes |  |  |
| `spec.database.mysql.passwordSecret` | `KubernetesTemporalMysqlPasswordSecret` | yes |  |  |
| `spec.database.mysql.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesMysql (`status.outputs.root_password_secret.name`) |
| `spec.database.mysql.passwordSecret.secretKey` | `string` | yes |  |  |
| `spec.database.mysql.maxConns` | `int32` |  | `20` |  |
| `spec.database.mysql.maxIdleConns` | `int32` |  | `20` |  |
| `spec.database.mysql.maxConnLifetime` | `string` |  | `1h` |  |
| `spec.database.mysql.tls` | `KubernetesTemporalDatabaseTls` |  |  |  |
| `spec.database.mysql.tls.enabled` | `bool` |  |  |  |
| `spec.database.mysql.tls.hostVerification` | `bool` |  |  |  |
| `spec.database.mysql.tls.serverName` | `string` |  |  |  |
| `spec.database.cassandra` | `KubernetesTemporalCassandra` |  |  |  |
| `spec.database.cassandra.hosts` | `[]string` | yes |  |  |
| `spec.database.cassandra.port` | `int32` |  | `9042` |  |
| `spec.database.cassandra.username` | `string` | yes |  |  |
| `spec.database.cassandra.passwordSecret` | `KubernetesTemporalPasswordSecret` | yes |  |  |
| `spec.database.cassandra.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.password_secret.name`) |
| `spec.database.cassandra.passwordSecret.secretKey` | `string` |  | `password` |  |
| `spec.database.cassandra.replicationFactor` | `int32` |  | `3` |  |
| `spec.database.cassandra.datacenter` | `string` |  |  |  |
| `spec.database.cassandra.tls` | `KubernetesTemporalDatabaseTls` |  |  |  |
| `spec.database.cassandra.tls.enabled` | `bool` |  |  |  |
| `spec.database.cassandra.tls.hostVerification` | `bool` |  |  |  |
| `spec.database.cassandra.tls.serverName` | `string` |  |  |  |
| `spec.database.databaseName` | `string` |  | `temporal` |  |
| `spec.database.visibilityDatabaseName` | `string` |  | `temporal_visibility` |  |
| `spec.database.visibility` | `KubernetesTemporalVisibility` |  |  |  |
| `spec.database.visibility.postgres` | `KubernetesTemporalPostgres` |  |  |  |
| `spec.database.visibility.postgres.host` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.rw_service`) |
| `spec.database.visibility.postgres.port` | `int32` |  | `5432` |  |
| `spec.database.visibility.postgres.username` | `string` | yes |  |  |
| `spec.database.visibility.postgres.passwordSecret` | `KubernetesTemporalPasswordSecret` | yes |  |  |
| `spec.database.visibility.postgres.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.password_secret.name`) |
| `spec.database.visibility.postgres.passwordSecret.secretKey` | `string` |  | `password` |  |
| `spec.database.visibility.postgres.maxConns` | `int32` |  | `20` |  |
| `spec.database.visibility.postgres.maxIdleConns` | `int32` |  | `20` |  |
| `spec.database.visibility.postgres.maxConnLifetime` | `string` |  | `1h` |  |
| `spec.database.visibility.postgres.tls` | `KubernetesTemporalDatabaseTls` |  |  |  |
| `spec.database.visibility.postgres.tls.enabled` | `bool` |  |  |  |
| `spec.database.visibility.postgres.tls.hostVerification` | `bool` |  |  |  |
| `spec.database.visibility.postgres.tls.serverName` | `string` |  |  |  |
| `spec.database.visibility.mysql` | `KubernetesTemporalMysql` |  |  |  |
| `spec.database.visibility.mysql.host` | `string \| valueFrom` | yes |  | KubernetesMysql (`status.outputs.primary_service`) |
| `spec.database.visibility.mysql.port` | `int32` |  | `3306` |  |
| `spec.database.visibility.mysql.username` | `string` | yes |  |  |
| `spec.database.visibility.mysql.passwordSecret` | `KubernetesTemporalMysqlPasswordSecret` | yes |  |  |
| `spec.database.visibility.mysql.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesMysql (`status.outputs.root_password_secret.name`) |
| `spec.database.visibility.mysql.passwordSecret.secretKey` | `string` | yes |  |  |
| `spec.database.visibility.mysql.maxConns` | `int32` |  | `20` |  |
| `spec.database.visibility.mysql.maxIdleConns` | `int32` |  | `20` |  |
| `spec.database.visibility.mysql.maxConnLifetime` | `string` |  | `1h` |  |
| `spec.database.visibility.mysql.tls` | `KubernetesTemporalDatabaseTls` |  |  |  |
| `spec.database.visibility.mysql.tls.enabled` | `bool` |  |  |  |
| `spec.database.visibility.mysql.tls.hostVerification` | `bool` |  |  |  |
| `spec.database.visibility.mysql.tls.serverName` | `string` |  |  |  |
| `spec.database.visibility.databaseName` | `string` |  |  |  |
| `spec.database.createDatabases` | `bool` |  |  |  |
| `spec.database.skipSchemaSetup` | `bool` |  |  |  |
| `spec.numHistoryShards` | `int32` |  | `512` |  |
| `spec.services` | `KubernetesTemporalServices` |  |  |  |
| `spec.services.frontend` | `KubernetesTemporalServiceConfig` |  |  |  |
| `spec.services.frontend.replicas` | `int32` |  |  |  |
| `spec.services.frontend.resources` | `ContainerResources` |  |  |  |
| `spec.services.frontend.resources.limits` | `CpuMemory` |  |  |  |
| `spec.services.frontend.resources.limits.cpu` | `string` |  |  |  |
| `spec.services.frontend.resources.limits.memory` | `string` |  |  |  |
| `spec.services.frontend.resources.requests` | `CpuMemory` |  |  |  |
| `spec.services.frontend.resources.requests.cpu` | `string` |  |  |  |
| `spec.services.frontend.resources.requests.memory` | `string` |  |  |  |
| `spec.services.history` | `KubernetesTemporalServiceConfig` |  |  |  |
| `spec.services.history.replicas` | `int32` |  |  |  |
| `spec.services.history.resources` | `ContainerResources` |  |  |  |
| `spec.services.history.resources.limits` | `CpuMemory` |  |  |  |
| `spec.services.history.resources.limits.cpu` | `string` |  |  |  |
| `spec.services.history.resources.limits.memory` | `string` |  |  |  |
| `spec.services.history.resources.requests` | `CpuMemory` |  |  |  |
| `spec.services.history.resources.requests.cpu` | `string` |  |  |  |
| `spec.services.history.resources.requests.memory` | `string` |  |  |  |
| `spec.services.matching` | `KubernetesTemporalServiceConfig` |  |  |  |
| `spec.services.matching.replicas` | `int32` |  |  |  |
| `spec.services.matching.resources` | `ContainerResources` |  |  |  |
| `spec.services.matching.resources.limits` | `CpuMemory` |  |  |  |
| `spec.services.matching.resources.limits.cpu` | `string` |  |  |  |
| `spec.services.matching.resources.limits.memory` | `string` |  |  |  |
| `spec.services.matching.resources.requests` | `CpuMemory` |  |  |  |
| `spec.services.matching.resources.requests.cpu` | `string` |  |  |  |
| `spec.services.matching.resources.requests.memory` | `string` |  |  |  |
| `spec.services.worker` | `KubernetesTemporalServiceConfig` |  |  |  |
| `spec.services.worker.replicas` | `int32` |  |  |  |
| `spec.services.worker.resources` | `ContainerResources` |  |  |  |
| `spec.services.worker.resources.limits` | `CpuMemory` |  |  |  |
| `spec.services.worker.resources.limits.cpu` | `string` |  |  |  |
| `spec.services.worker.resources.limits.memory` | `string` |  |  |  |
| `spec.services.worker.resources.requests` | `CpuMemory` |  |  |  |
| `spec.services.worker.resources.requests.cpu` | `string` |  |  |  |
| `spec.services.worker.resources.requests.memory` | `string` |  |  |  |
| `spec.internalFrontendEnabled` | `bool` |  |  |  |
| `spec.webUi` | `KubernetesTemporalWebUi` |  |  |  |
| `spec.webUi.enabled` | `bool` |  | `true` |  |
| `spec.webUi.replicas` | `int32` |  |  |  |
| `spec.webUi.resources` | `ContainerResources` |  |  |  |
| `spec.webUi.resources.limits` | `CpuMemory` |  |  |  |
| `spec.webUi.resources.limits.cpu` | `string` |  |  |  |
| `spec.webUi.resources.limits.memory` | `string` |  |  |  |
| `spec.webUi.resources.requests` | `CpuMemory` |  |  |  |
| `spec.webUi.resources.requests.cpu` | `string` |  |  |  |
| `spec.webUi.resources.requests.memory` | `string` |  |  |  |
| `spec.adminToolsEnabled` | `bool` |  | `true` |  |
| `spec.temporalNamespaces` | `[]KubernetesTemporalNamespace` |  |  |  |
| `spec.temporalNamespaces[].name` | `string` | yes |  |  |
| `spec.temporalNamespaces[].retention` | `string` |  | `3d` |  |
| `spec.dynamicConfig` | `KubernetesTemporalDynamicConfig` |  |  |  |
| `spec.dynamicConfig.historySizeLimitError` | `int64` |  |  |  |
| `spec.dynamicConfig.historySizeLimitWarn` | `int64` |  |  |  |
| `spec.dynamicConfig.historyCountLimitError` | `int64` |  |  |  |
| `spec.dynamicConfig.historyCountLimitWarn` | `int64` |  |  |  |
| `spec.dynamicConfig.blobSizeLimitError` | `int64` |  |  |  |
| `spec.dynamicConfig.blobSizeLimitWarn` | `int64` |  |  |  |
| `spec.archival` | `KubernetesTemporalArchival` |  |  |  |
| `spec.archival.s3` | `KubernetesTemporalArchivalS3` |  |  |  |
| `spec.archival.s3.region` | `string` | yes |  |  |
| `spec.archival.gcs` | `KubernetesTemporalArchivalGcs` |  |  |  |
| `spec.archival.filestore` | `KubernetesTemporalArchivalFilestore` |  |  |  |
| `spec.archival.historyUri` | `string` | yes |  |  |
| `spec.archival.visibilityUri` | `string` | yes |  |  |
| `spec.serviceMonitorEnabled` | `bool` |  |  |  |
| `spec.logLevel` | `string` |  | `info` |  |
| `spec.scheduling` | `KubernetesTemporalScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.images` | `KubernetesTemporalImages` |  |  |  |
| `spec.images.server` | `ContainerImage` |  |  |  |
| `spec.images.server.repo` | `string` |  |  |  |
| `spec.images.server.tag` | `string` |  |  |  |
| `spec.images.server.pullSecretName` | `string` |  |  |  |
| `spec.images.webUi` | `ContainerImage` |  |  |  |
| `spec.images.webUi.repo` | `string` |  |  |  |
| `spec.images.webUi.tag` | `string` |  |  |  |
| `spec.images.webUi.pullSecretName` | `string` |  |  |  |
| `spec.images.adminTools` | `ContainerImage` |  |  |  |
| `spec.images.adminTools.repo` | `string` |  |  |  |
| `spec.images.adminTools.tag` | `string` |  |  |  |
| `spec.images.adminTools.pullSecretName` | `string` |  |  |  |
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
governance labels) before installing and deleted with the
resource. When false, the namespace must already exist. KNOW THIS:
the database password is read through a secretKeyRef, and a
secretKeyRef can only read Secrets in the workload's OWN namespace
— co-locate Temporal with its database (the default composition)
or replicate the credential Secret into this namespace.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "1.6.0" — chart 1.6.0 ships
Temporal 1.31.2). Versions must exist as SERVED charts in the
repository index (https://go.temporal.io/helm-charts).

- default: `1.6.0`

### spec.database

`KubernetesTemporalDatabase` · required

The database Temporal persists every workflow's state in.
Required — nothing is bundled.

- rule: {"required":true}
- rule: the cassandra backend requires a `visibility` SQL block — Temporal removed Cassandra visibility support in v1.21, so the search index must live in PostgreSQL or MySQL

### spec.database.postgres

`KubernetesTemporalPostgres`

PostgreSQL — the recommended backend. Defaults compose a
KubernetesPostgres resource; any reachable PostgreSQL ≥ 12
(RDS, Cloud SQL, Aurora) works with literal values.

### spec.database.postgres.host

`string | valueFrom` · required

PostgreSQL host — a Service name (same namespace) or a full FQDN
(cross-namespace or external, e.g.
"pg-main-rw.data.svc.cluster.local" or an RDS endpoint). Accepts
a literal or a reference to a KubernetesPostgres resource (its
read-write Service — always the current primary). The port is
declared separately — do not include one here.

- references: KubernetesPostgres (`status.outputs.rw_service`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.rw_service}} -- a bare string does not parse

### spec.database.postgres.port

`int32` · optional (explicit presence)

PostgreSQL port. Empty = 5432.

- default: `5432`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.database.postgres.username

`string` · required

Database username. On a KubernetesPostgres this is the bootstrap
`owner` role (default: same as the bootstrap database name) —
ownership of both Temporal databases gives the schema Jobs
everything they need.

- rule: {"required":true}

### spec.database.postgres.passwordSecret

`KubernetesTemporalPasswordSecret` · required

The user's password, read from an existing Secret (the chart
wires it as a secretKeyRef into the server and schema-Job pods —
it never lands in rendered values).

- rule: {"required":true}

### spec.database.postgres.passwordSecret.secretName

`string | valueFrom` · required

Secret name. Accepts a literal or a reference to a
KubernetesPostgres resource (its `<cluster>-app` credential
Secret, maintained by the operator across failovers). KNOW THIS
(a Kubernetes constraint, not a chart one): a secretKeyRef can
only read Secrets in the workload's OWN namespace — co-locate
Temporal with its database or replicate the Secret.

- references: KubernetesPostgres (`status.outputs.password_secret.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.password_secret.name}} -- a bare string does not parse

### spec.database.postgres.passwordSecret.secretKey

`string` · optional (explicit presence)

Key within the Secret holding the password. Empty = "password"
(the key a KubernetesPostgres application Secret uses).

- default: `password`

### spec.database.postgres.maxConns

`int32` · optional (explicit presence)

Maximum open connections per service to this store. Empty = 20
(the chart preset's value). The effective total is per Temporal
service instance — four services × replicas all hold their own
pools; size the database's max_connections accordingly.

- default: `20`
- rule: {"int32":{"gt":0}}

### spec.database.postgres.maxIdleConns

`int32` · optional (explicit presence)

Maximum idle connections kept per pool. Empty = 20.

- default: `20`
- rule: {"int32":{"gt":0}}

### spec.database.postgres.maxConnLifetime

`string` · optional (explicit presence)

Maximum lifetime of a pooled connection (Go duration, e.g. "1h").
Empty = "1h".

- default: `1h`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?(ms|s|m|h)$"}}

### spec.database.postgres.tls

`KubernetesTemporalDatabaseTls`

TLS towards the database.

- rule: host_verification is meaningless without enabled: true — hostname verification only applies to a TLS connection

### spec.database.postgres.tls.enabled

`bool`

Connect with TLS.

### spec.database.postgres.tls.hostVerification

`bool`

Verify the server hostname against its certificate. Only
meaningful with `enabled: true`.

### spec.database.postgres.tls.serverName

`string`

Expected server name (SNI) when it differs from the connect host
— required by some serverless/proxied database offerings.

### spec.database.mysql

`KubernetesTemporalMysql`

MySQL 8. Defaults compose a KubernetesMysql resource; any
reachable MySQL 8 works with literal values.

### spec.database.mysql.host

`string | valueFrom` · required

MySQL host — a Service name (same namespace) or a full FQDN
(cross-namespace or external). Accepts a literal or a reference
to a KubernetesMysql resource (its primary Service — writes
always land on the current primary). The port is declared
separately — do not include one here.

- references: KubernetesMysql (`status.outputs.primary_service`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesMysql, name: <that resource's name>, fieldPath: status.outputs.primary_service}} -- a bare string does not parse

### spec.database.mysql.port

`int32` · optional (explicit presence)

MySQL port. Empty = 3306.

- default: `3306`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.database.mysql.username

`string` · required

Database username. Needs full rights on both Temporal databases
(plus CREATE on the server when `create_databases` is on). On a
KubernetesMysql, either use a declared user with grants on the
temporal databases or root.

- rule: {"required":true}

### spec.database.mysql.passwordSecret

`KubernetesTemporalMysqlPasswordSecret` · required

The user's password, read from an existing Secret (wired as a
secretKeyRef — it never lands in rendered values). The default
reference composes a KubernetesMysql's root credential Secret;
set `secret_key` to match the referenced user (e.g. "root").

- rule: {"required":true}

### spec.database.mysql.passwordSecret.secretName

`string | valueFrom` · required

Secret name. Accepts a literal or a reference to a
KubernetesMysql resource (its operator-maintained credential
Secret). Same-namespace constraint applies (secretKeyRef).

- references: KubernetesMysql (`status.outputs.root_password_secret.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesMysql, name: <that resource's name>, fieldPath: status.outputs.root_password_secret.name}} -- a bare string does not parse

### spec.database.mysql.passwordSecret.secretKey

`string` · required

Key within the Secret holding the declared user's password (on a
KubernetesMysql credential Secret the keys are per-user, e.g.
"root").

- rule: {"required":true}

### spec.database.mysql.maxConns

`int32` · optional (explicit presence)

Maximum open connections per service to this store. Empty = 20.
Four services × replicas each hold their own pools.

- default: `20`
- rule: {"int32":{"gt":0}}

### spec.database.mysql.maxIdleConns

`int32` · optional (explicit presence)

Maximum idle connections kept per pool. Empty = 20.

- default: `20`
- rule: {"int32":{"gt":0}}

### spec.database.mysql.maxConnLifetime

`string` · optional (explicit presence)

Maximum lifetime of a pooled connection (Go duration, e.g. "1h").
Empty = "1h".

- default: `1h`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?(ms|s|m|h)$"}}

### spec.database.mysql.tls

`KubernetesTemporalDatabaseTls`

TLS towards the database.

- rule: host_verification is meaningless without enabled: true — hostname verification only applies to a TLS connection

### spec.database.mysql.tls.enabled

`bool`

Connect with TLS.

### spec.database.mysql.tls.hostVerification

`bool`

Verify the server hostname against its certificate. Only
meaningful with `enabled: true`.

### spec.database.mysql.tls.serverName

`string`

Expected server name (SNI) when it differs from the connect host
— required by some serverless/proxied database offerings.

### spec.database.cassandra

`KubernetesTemporalCassandra`

An EXTERNAL Cassandra cluster you operate outside Planton (the
catalog has no Cassandra kind). Serves the default store ONLY —
Temporal removed Cassandra visibility support in v1.21, so a
`visibility` SQL block is REQUIRED with this backend.

### spec.database.cassandra.hosts

`[]string` · required

Cassandra contact points (hostnames or IPs; the driver discovers
the ring from them). At least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.database.cassandra.port

`int32` · optional (explicit presence)

CQL port. Empty = 9042.

- default: `9042`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.database.cassandra.username

`string` · required

Cassandra username. Needs full rights on the Temporal keyspace
(plus keyspace-creation rights when `create_databases` is on).

- rule: {"required":true}

### spec.database.cassandra.passwordSecret

`KubernetesTemporalPasswordSecret` · required

The user's password, read from an existing Secret you create in
the install namespace (wired as a secretKeyRef — it never lands
in rendered values).

- rule: {"required":true}

### spec.database.cassandra.passwordSecret.secretName

`string | valueFrom` · required

Secret name. Accepts a literal or a reference to a
KubernetesPostgres resource (its `<cluster>-app` credential
Secret, maintained by the operator across failovers). KNOW THIS
(a Kubernetes constraint, not a chart one): a secretKeyRef can
only read Secrets in the workload's OWN namespace — co-locate
Temporal with its database or replicate the Secret.

- references: KubernetesPostgres (`status.outputs.password_secret.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.password_secret.name}} -- a bare string does not parse

### spec.database.cassandra.passwordSecret.secretKey

`string` · optional (explicit presence)

Key within the Secret holding the password. Empty = "password"
(the key a KubernetesPostgres application Secret uses).

- default: `password`

### spec.database.cassandra.replicationFactor

`int32` · optional (explicit presence)

Replication factor used ONLY when `create_databases` creates the
keyspace (SimpleStrategy). Empty = 3. Existing keyspaces keep
their own replication settings.

- default: `3`
- rule: {"int32":{"gt":0}}

### spec.database.cassandra.datacenter

`string`

Local datacenter name for keyspace creation and driver locality.
Empty = the driver default.

### spec.database.cassandra.tls

`KubernetesTemporalDatabaseTls`

TLS towards Cassandra.

- rule: host_verification is meaningless without enabled: true — hostname verification only applies to a TLS connection

### spec.database.cassandra.tls.enabled

`bool`

Connect with TLS.

### spec.database.cassandra.tls.hostVerification

`bool`

Verify the server hostname against its certificate. Only
meaningful with `enabled: true`.

### spec.database.cassandra.tls.serverName

`string`

Expected server name (SNI) when it differs from the connect host
— required by some serverless/proxied database offerings.

### spec.database.databaseName

`string` · optional (explicit presence)

Name of the default store database (the Cassandra KEYSPACE when
the backend is cassandra). Empty = "temporal".

- default: `temporal`
- rule: {"string":{"pattern":"^[a-zA-Z_][a-zA-Z0-9_]*$"}}

### spec.database.visibilityDatabaseName

`string` · optional (explicit presence)

Name of the visibility store database. Empty =
"temporal_visibility". Ignored when a `visibility` block declares
its own database name.

- default: `temporal_visibility`
- rule: {"string":{"pattern":"^[a-zA-Z_][a-zA-Z0-9_]*$"}}

### spec.database.visibility

`KubernetesTemporalVisibility`

A SEPARATE SQL connection for the visibility store. REQUIRED when
the backend is cassandra (Cassandra cannot serve visibility since
Temporal v1.21). Optional for SQL backends — declare it to place
visibility on a different server than workflow state; empty = the
same server and credentials as the default store, database
`visibility_database_name`.

### spec.database.visibility.postgres

`KubernetesTemporalPostgres`

PostgreSQL visibility store.

### spec.database.visibility.postgres.host

`string | valueFrom` · required

PostgreSQL host — a Service name (same namespace) or a full FQDN
(cross-namespace or external, e.g.
"pg-main-rw.data.svc.cluster.local" or an RDS endpoint). Accepts
a literal or a reference to a KubernetesPostgres resource (its
read-write Service — always the current primary). The port is
declared separately — do not include one here.

- references: KubernetesPostgres (`status.outputs.rw_service`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.rw_service}} -- a bare string does not parse

### spec.database.visibility.postgres.port

`int32` · optional (explicit presence)

PostgreSQL port. Empty = 5432.

- default: `5432`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.database.visibility.postgres.username

`string` · required

Database username. On a KubernetesPostgres this is the bootstrap
`owner` role (default: same as the bootstrap database name) —
ownership of both Temporal databases gives the schema Jobs
everything they need.

- rule: {"required":true}

### spec.database.visibility.postgres.passwordSecret

`KubernetesTemporalPasswordSecret` · required

The user's password, read from an existing Secret (the chart
wires it as a secretKeyRef into the server and schema-Job pods —
it never lands in rendered values).

- rule: {"required":true}

### spec.database.visibility.postgres.passwordSecret.secretName

`string | valueFrom` · required

Secret name. Accepts a literal or a reference to a
KubernetesPostgres resource (its `<cluster>-app` credential
Secret, maintained by the operator across failovers). KNOW THIS
(a Kubernetes constraint, not a chart one): a secretKeyRef can
only read Secrets in the workload's OWN namespace — co-locate
Temporal with its database or replicate the Secret.

- references: KubernetesPostgres (`status.outputs.password_secret.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.password_secret.name}} -- a bare string does not parse

### spec.database.visibility.postgres.passwordSecret.secretKey

`string` · optional (explicit presence)

Key within the Secret holding the password. Empty = "password"
(the key a KubernetesPostgres application Secret uses).

- default: `password`

### spec.database.visibility.postgres.maxConns

`int32` · optional (explicit presence)

Maximum open connections per service to this store. Empty = 20
(the chart preset's value). The effective total is per Temporal
service instance — four services × replicas all hold their own
pools; size the database's max_connections accordingly.

- default: `20`
- rule: {"int32":{"gt":0}}

### spec.database.visibility.postgres.maxIdleConns

`int32` · optional (explicit presence)

Maximum idle connections kept per pool. Empty = 20.

- default: `20`
- rule: {"int32":{"gt":0}}

### spec.database.visibility.postgres.maxConnLifetime

`string` · optional (explicit presence)

Maximum lifetime of a pooled connection (Go duration, e.g. "1h").
Empty = "1h".

- default: `1h`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?(ms|s|m|h)$"}}

### spec.database.visibility.postgres.tls

`KubernetesTemporalDatabaseTls`

TLS towards the database.

- rule: host_verification is meaningless without enabled: true — hostname verification only applies to a TLS connection

### spec.database.visibility.postgres.tls.enabled

`bool`

Connect with TLS.

### spec.database.visibility.postgres.tls.hostVerification

`bool`

Verify the server hostname against its certificate. Only
meaningful with `enabled: true`.

### spec.database.visibility.postgres.tls.serverName

`string`

Expected server name (SNI) when it differs from the connect host
— required by some serverless/proxied database offerings.

### spec.database.visibility.mysql

`KubernetesTemporalMysql`

MySQL 8 visibility store.

### spec.database.visibility.mysql.host

`string | valueFrom` · required

MySQL host — a Service name (same namespace) or a full FQDN
(cross-namespace or external). Accepts a literal or a reference
to a KubernetesMysql resource (its primary Service — writes
always land on the current primary). The port is declared
separately — do not include one here.

- references: KubernetesMysql (`status.outputs.primary_service`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesMysql, name: <that resource's name>, fieldPath: status.outputs.primary_service}} -- a bare string does not parse

### spec.database.visibility.mysql.port

`int32` · optional (explicit presence)

MySQL port. Empty = 3306.

- default: `3306`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.database.visibility.mysql.username

`string` · required

Database username. Needs full rights on both Temporal databases
(plus CREATE on the server when `create_databases` is on). On a
KubernetesMysql, either use a declared user with grants on the
temporal databases or root.

- rule: {"required":true}

### spec.database.visibility.mysql.passwordSecret

`KubernetesTemporalMysqlPasswordSecret` · required

The user's password, read from an existing Secret (wired as a
secretKeyRef — it never lands in rendered values). The default
reference composes a KubernetesMysql's root credential Secret;
set `secret_key` to match the referenced user (e.g. "root").

- rule: {"required":true}

### spec.database.visibility.mysql.passwordSecret.secretName

`string | valueFrom` · required

Secret name. Accepts a literal or a reference to a
KubernetesMysql resource (its operator-maintained credential
Secret). Same-namespace constraint applies (secretKeyRef).

- references: KubernetesMysql (`status.outputs.root_password_secret.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesMysql, name: <that resource's name>, fieldPath: status.outputs.root_password_secret.name}} -- a bare string does not parse

### spec.database.visibility.mysql.passwordSecret.secretKey

`string` · required

Key within the Secret holding the declared user's password (on a
KubernetesMysql credential Secret the keys are per-user, e.g.
"root").

- rule: {"required":true}

### spec.database.visibility.mysql.maxConns

`int32` · optional (explicit presence)

Maximum open connections per service to this store. Empty = 20.
Four services × replicas each hold their own pools.

- default: `20`
- rule: {"int32":{"gt":0}}

### spec.database.visibility.mysql.maxIdleConns

`int32` · optional (explicit presence)

Maximum idle connections kept per pool. Empty = 20.

- default: `20`
- rule: {"int32":{"gt":0}}

### spec.database.visibility.mysql.maxConnLifetime

`string` · optional (explicit presence)

Maximum lifetime of a pooled connection (Go duration, e.g. "1h").
Empty = "1h".

- default: `1h`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?(ms|s|m|h)$"}}

### spec.database.visibility.mysql.tls

`KubernetesTemporalDatabaseTls`

TLS towards the database.

- rule: host_verification is meaningless without enabled: true — hostname verification only applies to a TLS connection

### spec.database.visibility.mysql.tls.enabled

`bool`

Connect with TLS.

### spec.database.visibility.mysql.tls.hostVerification

`bool`

Verify the server hostname against its certificate. Only
meaningful with `enabled: true`.

### spec.database.visibility.mysql.tls.serverName

`string`

Expected server name (SNI) when it differs from the connect host
— required by some serverless/proxied database offerings.

### spec.database.visibility.databaseName

`string` · optional (explicit presence)

Visibility database name on THIS server. Empty = the parent
database block's `visibility_database_name` (default
"temporal_visibility").

- rule: {"string":{"pattern":"^[a-zA-Z_][a-zA-Z0-9_]*$"}}

### spec.database.createDatabases

`bool`

Have the schema Jobs CREATE the databases/keyspace before setting
up schemas. Off by default — the databases are expected to exist
(on a KubernetesPostgres: declare the default database at
bootstrap and add the visibility database via `post_init_sql`).
Turning this on requires create-database privileges (CREATEDB on
PostgreSQL) for the declared user.

### spec.database.skipSchemaSetup

`bool`

Skip the schema-setup Jobs entirely — you manage Temporal's
schemas yourself with temporal-sql-tool/temporal-cassandra-tool.
Leave off unless you have a dedicated schema pipeline: the server
crash-loops against an empty or outdated schema.

### spec.numHistoryShards

`int32` · optional (explicit presence)

Number of history shards for the cluster. IMMUTABLE — the value is
baked into the default store's schema at first install and CANNOT
be changed afterwards without a full cluster migration; pick for
the cluster you will grow into, not the one you start with. Higher
values raise parallelism and throughput ceilings at the cost of
per-shard overhead. Empty = 512 (the upstream default, safe for
most production workloads).

- default: `512`
- rule: {"int32":{"lte":16384,"gte":1}}

### spec.services

`KubernetesTemporalServices`

Per-service sizing for the four Temporal server services. Empty =
one replica of each with the chart's defaults (no resource
requests — fine for dev, size them for production).

### spec.services.frontend

`KubernetesTemporalServiceConfig`

The frontend service — the gRPC/HTTP API gateway clients and workers connect to.

### spec.services.frontend.replicas

`int32` · optional (explicit presence)

Number of replicas. Empty = 1. All four services scale
horizontally; history is the stateful heavy-lifter (shard
ownership redistributes across its replicas).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.services.frontend.resources

`ContainerResources`

Container resources. Typical production starting points: history
is the most demanding (workflow state caches), frontend and
matching moderate, worker light.

### spec.services.frontend.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.services.frontend.resources.limits.cpu

`string`

### spec.services.frontend.resources.limits.memory

`string`

### spec.services.frontend.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.services.frontend.resources.requests.cpu

`string`

### spec.services.frontend.resources.requests.memory

`string`

### spec.services.history

`KubernetesTemporalServiceConfig`

The history service — owns workflow state and the history shards; the most resource-intensive.

### spec.services.history.replicas

`int32` · optional (explicit presence)

Number of replicas. Empty = 1. All four services scale
horizontally; history is the stateful heavy-lifter (shard
ownership redistributes across its replicas).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.services.history.resources

`ContainerResources`

Container resources. Typical production starting points: history
is the most demanding (workflow state caches), frontend and
matching moderate, worker light.

### spec.services.history.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.services.history.resources.limits.cpu

`string`

### spec.services.history.resources.limits.memory

`string`

### spec.services.history.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.services.history.resources.requests.cpu

`string`

### spec.services.history.resources.requests.memory

`string`

### spec.services.matching

`KubernetesTemporalServiceConfig`

The matching service — task-queue management and dispatch to workers.

### spec.services.matching.replicas

`int32` · optional (explicit presence)

Number of replicas. Empty = 1. All four services scale
horizontally; history is the stateful heavy-lifter (shard
ownership redistributes across its replicas).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.services.matching.resources

`ContainerResources`

Container resources. Typical production starting points: history
is the most demanding (workflow state caches), frontend and
matching moderate, worker light.

### spec.services.matching.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.services.matching.resources.limits.cpu

`string`

### spec.services.matching.resources.limits.memory

`string`

### spec.services.matching.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.services.matching.resources.requests.cpu

`string`

### spec.services.matching.resources.requests.memory

`string`

### spec.services.worker

`KubernetesTemporalServiceConfig`

The worker service — Temporal's internal system workflows.

### spec.services.worker.replicas

`int32` · optional (explicit presence)

Number of replicas. Empty = 1. All four services scale
horizontally; history is the stateful heavy-lifter (shard
ownership redistributes across its replicas).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.services.worker.resources

`ContainerResources`

Container resources. Typical production starting points: history
is the most demanding (workflow state caches), frontend and
matching moderate, worker light.

### spec.services.worker.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.services.worker.resources.limits.cpu

`string`

### spec.services.worker.resources.limits.memory

`string`

### spec.services.worker.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.services.worker.resources.requests.cpu

`string`

### spec.services.worker.resources.requests.memory

`string`

### spec.internalFrontendEnabled

`bool`

Run the internal-frontend service. Temporal's system workers
connect through it instead of the public frontend — REQUIRED when
you enable authorization on the public frontend via `helm_values`
(otherwise the server's own workers would need JWTs). Off by
default, matching the chart.

### spec.webUi

`KubernetesTemporalWebUi`

The Temporal Web UI.

### spec.webUi.enabled

`bool` · optional (explicit presence)

Deploy the Web UI. Empty = true (the chart default). The UI is
read-mostly; workflows are driven through the frontend gRPC API.

- default: `true`

### spec.webUi.replicas

`int32` · optional (explicit presence)

UI replicas. Empty = 1 (the UI is stateless — scale freely).

- rule: {"int32":{"lte":20,"gte":1}}

### spec.webUi.resources

`ContainerResources`

Container resources for the UI.

### spec.webUi.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.webUi.resources.limits.cpu

`string`

### spec.webUi.resources.limits.memory

`string`

### spec.webUi.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.webUi.resources.requests.cpu

`string`

### spec.webUi.resources.requests.memory

`string`

### spec.adminToolsEnabled

`bool` · optional (explicit presence)

Deploy the admin-tools pod — a shell with `temporal` and the
schema tools pre-installed, for operational commands
(`kubectl exec` into it). The schema Jobs use this image either
way. Enabled by default, matching the chart.

- default: `true`

### spec.temporalNamespaces

`[]KubernetesTemporalNamespace`

Temporal namespaces to create declaratively after install (a Job
runs `temporal operator namespace create` for each, idempotently).
Workflows execute inside a Temporal namespace — declare at least
one, or create them later with the CLI/UI. Note these are
TEMPORAL namespaces (a logical isolation unit inside the server),
not Kubernetes namespaces.

### spec.temporalNamespaces[].name

`string` · required

Namespace name (e.g. "default", "payments").

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z][a-zA-Z0-9._-]{1,231}$"}}

### spec.temporalNamespaces[].retention

`string` · optional (explicit presence)

Workflow-execution retention period — how long CLOSED workflow
histories stay queryable before deletion (archival, when
configured, keeps them beyond this). Duration with a d/h/m
suffix. Empty = "3d".

- default: `3d`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?(m|h|d)$"}}

### spec.dynamicConfig

`KubernetesTemporalDynamicConfig`

Runtime limits pushed through Temporal's dynamic-config file.
These control workflow history and payload size ceilings and
apply without a server restart. Empty = Temporal's defaults.
Other dynamic-config keys ride `helm_values` under
`server.dynamicConfig`.

### spec.dynamicConfig.historySizeLimitError

`int64` · optional (explicit presence)

Maximum workflow history size in bytes before termination.
Upstream default: 50 MB. Long-running workflows with large
payloads should raise this or use ContinueAsNew.

- rule: {"int64":{"gte":"1048576"}}

### spec.dynamicConfig.historySizeLimitWarn

`int64` · optional (explicit presence)

Warning threshold for history size in bytes. Upstream default:
10 MB.

- rule: {"int64":{"gte":"524288"}}

### spec.dynamicConfig.historyCountLimitError

`int64` · optional (explicit presence)

Maximum workflow history event count before termination.
Upstream default: 51200.

- rule: {"int64":{"gte":"1000"}}

### spec.dynamicConfig.historyCountLimitWarn

`int64` · optional (explicit presence)

Warning threshold for history event count. Upstream default:
10240.

- rule: {"int64":{"gte":"500"}}

### spec.dynamicConfig.blobSizeLimitError

`int64` · optional (explicit presence)

Maximum size in bytes of a single payload (activity input/output,
signal, marker). Upstream default: 2 MB. Raise for workflows
passing large blobs — better yet, pass references to object
storage.

- rule: {"int64":{"gte":"1048576"}}

### spec.dynamicConfig.blobSizeLimitWarn

`int64` · optional (explicit presence)

Warning threshold for payload size in bytes. Upstream default:
512 KB.

- rule: {"int64":{"gte":"262144"}}

### spec.archival

`KubernetesTemporalArchival`

Archive closed workflow histories and visibility records to
long-term storage (S3, GCS, or a mounted filesystem) so they
survive retention-driven deletion from the database. Empty =
archival disabled (the upstream default).

- rule: with the s3 provider, history_uri and visibility_uri must use the s3:// scheme
- rule: with the gcs provider, history_uri and visibility_uri must use the gs:// scheme
- rule: with the filestore provider, history_uri and visibility_uri must use the file:// scheme

### spec.archival.s3

`KubernetesTemporalArchivalS3`

Amazon S3 (or S3-compatible endpoints configured via helm_values).

### spec.archival.s3.region

`string` · required

AWS region of the archival bucket(s).

- rule: {"required":true}

### spec.archival.gcs

`KubernetesTemporalArchivalGcs`

Google Cloud Storage.

### spec.archival.filestore

`KubernetesTemporalArchivalFilestore`

A filesystem path INSIDE the history pods — dev/test only (the
path is a pod-local mount unless you add a shared volume via
helm_values); histories archived here do not survive pod loss.

### spec.archival.historyUri

`string` · required

Archival URI for workflow HISTORIES, scheme matching the provider
(s3://bucket[/prefix], gs://bucket[/prefix], file:///path).
Becomes the namespace default — every Temporal namespace archives
here unless overridden per namespace.

- rule: {"required":true}

### spec.archival.visibilityUri

`string` · required

Archival URI for VISIBILITY records, scheme matching the provider.

- rule: {"required":true}

### spec.serviceMonitorEnabled

`bool`

Emit Prometheus metrics via ServiceMonitor resources (one per
server service). Requires the Prometheus Operator CRDs on the
cluster — a KubernetesKubePrometheusStack composes naturally.
When false (default), pods still carry prometheus.io scrape
annotations for annotation-based collection. These monitors keep
every series: Temporal serves its latencies as histograms per
operation and per task queue, tens of thousands of series per
installation even when idle. Where storage is budgeted, leave this
off and declare a KubernetesServiceMonitor on the per-service
headless Services' `metrics` port with a keep list (service and
persistence requests and errors, approximate_backlog_age_seconds,
task_schedule_to_start_latency with a few bounds).

### spec.logLevel

`string` · optional (explicit presence)

Server log level. Empty = "info". Accepts a single level or
Temporal's comma form (e.g. "debug,info").

- default: `info`
- rule: {"string":{"pattern":"^(debug|info|warn|error)(,(debug|info|warn|error))*$"}}

### spec.scheduling

`KubernetesTemporalScheduling`

Pod scheduling for the server services, schema Jobs, admin-tools
and Web UI pods.

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for all Temporal pods.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for all Temporal pods.

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

### spec.images

`KubernetesTemporalImages`

Container image overrides for air-gapped clusters and private
mirrors. Empty = the chart's pinned upstream images
(temporalio/server, temporalio/ui, temporalio/admin-tools).

### spec.images.server

`ContainerImage`

The Temporal server image (upstream: temporalio/server). Used by
all four services.

### spec.images.server.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.images.server.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.images.server.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.images.webUi

`ContainerImage`

The Web UI image (upstream: temporalio/ui).

### spec.images.webUi.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.images.webUi.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.images.webUi.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.images.adminTools

`ContainerImage`

The admin-tools image (upstream: temporalio/admin-tools). Also
runs the schema and namespace Jobs — needed even with
`admin_tools_enabled: false`.

### spec.images.adminTools.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.images.adminTools.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.images.adminTools.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.helmValues

`string`

Additional Helm values merged LAST (Helm `-f` semantics, identical
on both engines) — the escape hatch for chart values the typed
fields do not model: mTLS mounts (`server.config.tls` +
`server.additionalVolumes`), JWT authorization
(`server.config.authorization`), multi-cluster replication
(`server.config.clusterMetadata`), extra dynamic-config keys, and
per-service scheduling. YAML document as a string. Never put
secret material here; passwords belong in the typed
secret-reference fields, which keep them out of rendered values.
The legacy bundled-subchart keys (cassandra, elasticsearch,
prometheus, grafana, mysql, postgresql) were REMOVED upstream —
setting them makes the chart itself fail rendering.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesTemporal, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the cluster runs in. |
| `status.outputs.frontend_service` | `string` | Name of the frontend Service (`<name>-frontend`) — the gRPC/HTTP API gateway. The handle exposure kinds route to. |
| `status.outputs.frontend_endpoint` | `string` | In-cluster frontend gRPC endpoint, `<name>-frontend.<namespace>.svc.cluster.local:7233` — what Temporal SDK workers and clients set as their server address. |
| `status.outputs.frontend_http_endpoint` | `string` | In-cluster frontend HTTP API endpoint (port 7243) — Temporal's HTTP/JSON API for clients that cannot speak gRPC. |
| `status.outputs.web_ui_service` | `string` | Name of the Web UI Service (`<name>-web`); empty when the UI is disabled. |
| `status.outputs.web_ui_endpoint` | `string` | In-cluster Web UI endpoint, `http://<name>-web.<namespace>.svc.cluster.local:8080`; empty when the UI is disabled. |
| `status.outputs.port_forward_frontend_command` | `string` | Port-forward command for reaching the frontend from a workstation when no exposure is composed (`kubectl port-forward svc/<name>-frontend -n <namespace> 7233:7233`). |
| `status.outputs.port_forward_web_ui_command` | `string` | Port-forward command for reaching the Web UI from a workstation (`kubectl port-forward svc/<name>-web -n <namespace> 8080:8080`); empty when the UI is disabled. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.database.postgres.host` | KubernetesPostgres | `status.outputs.rw_service` |
| `spec.database.postgres.passwordSecret.secretName` | KubernetesPostgres | `status.outputs.password_secret.name` |
| `spec.database.mysql.host` | KubernetesMysql | `status.outputs.primary_service` |
| `spec.database.mysql.passwordSecret.secretName` | KubernetesMysql | `status.outputs.root_password_secret.name` |
| `spec.database.cassandra.passwordSecret.secretName` | KubernetesPostgres | `status.outputs.password_secret.name` |
| `spec.database.visibility.postgres.host` | KubernetesPostgres | `status.outputs.rw_service` |
| `spec.database.visibility.postgres.passwordSecret.secretName` | KubernetesPostgres | `status.outputs.password_secret.name` |
| `spec.database.visibility.mysql.host` | KubernetesMysql | `status.outputs.primary_service` |
| `spec.database.visibility.mysql.passwordSecret.secretName` | KubernetesMysql | `status.outputs.root_password_secret.name` |

## See Also

- [Overview](../README.md)
