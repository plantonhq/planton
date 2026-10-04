# KubernetesAirflow

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**KubernetesAirflowSpec** deploys Apache Airflow — the
workflow-orchestration platform for data pipelines (ETL, ML
training, analytics scheduling) — from the official `airflow` Helm
chart (https://airflow.apache.org).

WHAT GETS INSTALLED (the Airflow 3 component set): the API server
(UI + REST API, HTTP 8080), the scheduler, the standalone DAG
processor, the triggerer (async deferred tasks, on by default with
its own PVC), Celery workers when a Celery executor is declared,
and the database-migration and admin-user bootstrap Jobs. Those Jobs
ship HOOK-LESS (`useHelmHooks: false`): the chart's post-install
hook defaults deadlock against `wait-for-airflow-migrations` init
containers under any wait-style install (Helm wait, both engines,
Argo CD sync — verified live). This kind models the Airflow 3 line
only (`airflow_version >= 3.0.0`) — the Airflow 2 webserver/flower-
era surfaces are not modeled.

BRING YOUR OWN DATABASE — the chart's bundled PostgreSQL subchart
never ships (upstream marks it non-production and its Bitnami image
line is frozen). Declare the metadata database under `database`:
PostgreSQL (a KubernetesPostgres composes naturally — the
recommended path) or an external MySQL 8+.

CREDENTIALS ARE SECRET-NATIVE END TO END. Airflow consumes its
database and broker connections as Kubernetes Secrets carrying a
full connection URI (`connection` key). The module composes those
Secrets AT APPLY TIME from the referenced credential Secrets —
nothing credential-bearing ever appears in rendered Helm values.
The Fernet key (task-credential encryption), the API-server secret
key, the JWT signing secret, and the admin user's password are all
module-generated into Secrets unless you bring your own — the
chart's publicly-documented defaults (postgres/postgres,
admin/admin) never ship.

DAG DELIVERY is declared under `dags`: sync from a Git repository
(the recommended path — a git-sync sidecar on every component keeps
DAGs current), mount a shared persistent volume, or bake DAGs into
a custom image (the default when `dags` is empty; set
`images.airflow_repository`/`airflow_tag` to your image).

EXPOSURE: the API server Service stays ClusterIP; expose it via
first-class kinds (KubernetesService for a LoadBalancer, Gateway
API kinds for routes) over the exported service handle. The chart's
ingress blocks are never modeled.

The typed fields below cover the chart's meaningful configuration
surface; `helm_values` remains as the escape hatch for chart values
beyond them (merged last, Helm `-f` semantics, identical on both
engines) — Kerberos, per-component env/volumes, the chart's own
OTel collector sidecar, cleanup CronJobs, network policies — a
safety valve, never the primary interface. Never put secret
material in `helm_values`; credentials ride Secrets through the
typed fields and never land in rendered values.

## Example

```yaml
# Full-surface shape: the Celery executor with the bundled Redis broker,
# PgBouncer pooling, git-sync DAG delivery, KEDA worker autoscaling, the
# OpenSearch log read path, logs persistence, sizing, scheduling and an
# escape-hatch entry — the offline plan/preview proof for the widest
# typed rendering.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAirflow
metadata:
  name: airflow-full
spec:
  namespace:
    value: airflow-full
  createNamespace: true
  executor: CeleryExecutor
  database:
    postgres:
      host:
        value: airflow-pg-rw
      port: 5432
      databaseName: airflow
      username: airflow
      passwordSecret:
        secretName:
          value: airflow-pg-app
        secretKey: password
      sslMode: disable
  broker:
    bundledRedis:
      persistenceSize: 1Gi
      resources:
        requests:
          cpu: 50m
          memory: 64Mi
  dags:
    gitSync:
      repo: https://github.com/example/pipelines.git
      ref: main
      subPath: dags
      periodSeconds: 30
      credentialsSecret: pipelines-git-token
  components:
    apiServer:
      replicas: 2
      resources:
        requests:
          cpu: 250m
          memory: 512Mi
    scheduler:
      replicas: 2
      resources:
        requests:
          cpu: 250m
          memory: 512Mi
    dagProcessor:
      replicas: 1
    triggerer:
      persistenceSize: 2Gi
    workers:
      persistenceSize: 2Gi
      resources:
        requests:
          cpu: 500m
          memory: 1Gi
      keda:
        enabled: true
        minReplicas: 0
        maxReplicas: 5
        pollingIntervalSeconds: 10
  pgbouncer:
    enabled: true
    metadataPoolSize: 20
    resultBackendPoolSize: 5
    maxClientConnections: 150
  logging:
    persistence:
      enabled: true
      size: 5Gi
    opensearch:
      host:
        value: logs-opensearch
      port: 9200
      username: airflow
      passwordSecret:
        secretName:
          value: logs-opensearch-airflow-auth
  adminUser:
    username: platform-admin
    email: platform@example.com
  statsdEnabled: true
  scheduling:
    nodeSelector:
      workload-tier: data
  helmValues: |
    workers:
      terminationGracePeriodSeconds: 300
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `1.22.0` |  |
| `spec.airflowVersion` | `string` |  | `3.2.2` |  |
| `spec.executor` | `string` |  | `KubernetesExecutor` |  |
| `spec.database` | `KubernetesAirflowDatabase` | yes |  |  |
| `spec.database.postgres` | `KubernetesAirflowPostgres` |  |  |  |
| `spec.database.postgres.host` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.rw_service`) |
| `spec.database.postgres.port` | `int32` |  | `5432` |  |
| `spec.database.postgres.databaseName` | `string` |  | `airflow` |  |
| `spec.database.postgres.username` | `string` |  | `airflow` |  |
| `spec.database.postgres.passwordSecret` | `KubernetesAirflowPasswordSecret` | yes |  |  |
| `spec.database.postgres.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.password_secret.name`) |
| `spec.database.postgres.passwordSecret.secretKey` | `string` |  | `password` |  |
| `spec.database.postgres.sslMode` | `string` |  | `disable` |  |
| `spec.database.mysql` | `KubernetesAirflowMysql` |  |  |  |
| `spec.database.mysql.host` | `string \| valueFrom` | yes |  | KubernetesMysql (`status.outputs.primary_service`) |
| `spec.database.mysql.port` | `int32` |  | `3306` |  |
| `spec.database.mysql.databaseName` | `string` |  | `airflow` |  |
| `spec.database.mysql.username` | `string` |  | `airflow` |  |
| `spec.database.mysql.passwordSecret` | `KubernetesAirflowMysqlPasswordSecret` | yes |  |  |
| `spec.database.mysql.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesMysql (`status.outputs.root_password_secret.name`) |
| `spec.database.mysql.passwordSecret.secretKey` | `string` |  | `password` |  |
| `spec.broker` | `KubernetesAirflowBroker` |  |  |  |
| `spec.broker.bundledRedis` | `KubernetesAirflowBundledRedis` |  |  |  |
| `spec.broker.bundledRedis.persistenceSize` | `string` |  | `1Gi` |  |
| `spec.broker.bundledRedis.storageClass` | `string` |  |  |  |
| `spec.broker.bundledRedis.resources` | `ContainerResources` |  |  |  |
| `spec.broker.bundledRedis.resources.limits` | `CpuMemory` |  |  |  |
| `spec.broker.bundledRedis.resources.limits.cpu` | `string` |  |  |  |
| `spec.broker.bundledRedis.resources.limits.memory` | `string` |  |  |  |
| `spec.broker.bundledRedis.resources.requests` | `CpuMemory` |  |  |  |
| `spec.broker.bundledRedis.resources.requests.cpu` | `string` |  |  |  |
| `spec.broker.bundledRedis.resources.requests.memory` | `string` |  |  |  |
| `spec.broker.valkey` | `KubernetesAirflowValkeyBroker` |  |  |  |
| `spec.broker.valkey.host` | `string \| valueFrom` | yes |  | KubernetesValkey (`status.outputs.service`) |
| `spec.broker.valkey.port` | `int32` |  | `6379` |  |
| `spec.broker.valkey.username` | `string` |  |  |  |
| `spec.broker.valkey.passwordSecret` | `KubernetesAirflowBrokerPasswordSecret` |  |  |  |
| `spec.broker.valkey.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesValkey (`status.outputs.password_secret.name`) |
| `spec.broker.valkey.passwordSecret.secretKey` | `string` |  | `password` |  |
| `spec.broker.valkey.databaseNumber` | `int32` |  | `0` |  |
| `spec.broker.existingBrokerUrlSecret` | `KubernetesAirflowBrokerUrlSecret` |  |  |  |
| `spec.broker.existingBrokerUrlSecret.secretName` | `string` | yes |  |  |
| `spec.dags` | `KubernetesAirflowDags` |  |  |  |
| `spec.dags.gitSync` | `KubernetesAirflowGitSync` |  |  |  |
| `spec.dags.gitSync.repo` | `string` | yes |  |  |
| `spec.dags.gitSync.ref` | `string` |  |  |  |
| `spec.dags.gitSync.subPath` | `string` |  |  |  |
| `spec.dags.gitSync.periodSeconds` | `int32` |  | `5` |  |
| `spec.dags.gitSync.depth` | `int32` |  | `1` |  |
| `spec.dags.gitSync.credentialsSecret` | `string` |  |  |  |
| `spec.dags.gitSync.sshKeySecret` | `string` |  |  |  |
| `spec.dags.gitSync.knownHosts` | `string` |  |  |  |
| `spec.dags.gitSync.resources` | `ContainerResources` |  |  |  |
| `spec.dags.gitSync.resources.limits` | `CpuMemory` |  |  |  |
| `spec.dags.gitSync.resources.limits.cpu` | `string` |  |  |  |
| `spec.dags.gitSync.resources.limits.memory` | `string` |  |  |  |
| `spec.dags.gitSync.resources.requests` | `CpuMemory` |  |  |  |
| `spec.dags.gitSync.resources.requests.cpu` | `string` |  |  |  |
| `spec.dags.gitSync.resources.requests.memory` | `string` |  |  |  |
| `spec.dags.persistence` | `KubernetesAirflowDagsPersistence` |  |  |  |
| `spec.dags.persistence.size` | `string` |  | `1Gi` |  |
| `spec.dags.persistence.storageClass` | `string` |  |  |  |
| `spec.dags.persistence.existingClaim` | `string` |  |  |  |
| `spec.components` | `KubernetesAirflowComponents` |  |  |  |
| `spec.components.apiServer` | `KubernetesAirflowApiServer` |  |  |  |
| `spec.components.apiServer.replicas` | `int32` |  | `1` |  |
| `spec.components.apiServer.resources` | `ContainerResources` |  |  |  |
| `spec.components.apiServer.resources.limits` | `CpuMemory` |  |  |  |
| `spec.components.apiServer.resources.limits.cpu` | `string` |  |  |  |
| `spec.components.apiServer.resources.limits.memory` | `string` |  |  |  |
| `spec.components.apiServer.resources.requests` | `CpuMemory` |  |  |  |
| `spec.components.apiServer.resources.requests.cpu` | `string` |  |  |  |
| `spec.components.apiServer.resources.requests.memory` | `string` |  |  |  |
| `spec.components.scheduler` | `KubernetesAirflowScheduler` |  |  |  |
| `spec.components.scheduler.replicas` | `int32` |  | `1` |  |
| `spec.components.scheduler.resources` | `ContainerResources` |  |  |  |
| `spec.components.scheduler.resources.limits` | `CpuMemory` |  |  |  |
| `spec.components.scheduler.resources.limits.cpu` | `string` |  |  |  |
| `spec.components.scheduler.resources.limits.memory` | `string` |  |  |  |
| `spec.components.scheduler.resources.requests` | `CpuMemory` |  |  |  |
| `spec.components.scheduler.resources.requests.cpu` | `string` |  |  |  |
| `spec.components.scheduler.resources.requests.memory` | `string` |  |  |  |
| `spec.components.dagProcessor` | `KubernetesAirflowDagProcessor` |  |  |  |
| `spec.components.dagProcessor.replicas` | `int32` |  | `1` |  |
| `spec.components.dagProcessor.resources` | `ContainerResources` |  |  |  |
| `spec.components.dagProcessor.resources.limits` | `CpuMemory` |  |  |  |
| `spec.components.dagProcessor.resources.limits.cpu` | `string` |  |  |  |
| `spec.components.dagProcessor.resources.limits.memory` | `string` |  |  |  |
| `spec.components.dagProcessor.resources.requests` | `CpuMemory` |  |  |  |
| `spec.components.dagProcessor.resources.requests.cpu` | `string` |  |  |  |
| `spec.components.dagProcessor.resources.requests.memory` | `string` |  |  |  |
| `spec.components.triggerer` | `KubernetesAirflowTriggerer` |  |  |  |
| `spec.components.triggerer.enabled` | `bool` |  | `true` |  |
| `spec.components.triggerer.replicas` | `int32` |  | `1` |  |
| `spec.components.triggerer.persistenceSize` | `string` |  | `100Gi` |  |
| `spec.components.triggerer.resources` | `ContainerResources` |  |  |  |
| `spec.components.triggerer.resources.limits` | `CpuMemory` |  |  |  |
| `spec.components.triggerer.resources.limits.cpu` | `string` |  |  |  |
| `spec.components.triggerer.resources.limits.memory` | `string` |  |  |  |
| `spec.components.triggerer.resources.requests` | `CpuMemory` |  |  |  |
| `spec.components.triggerer.resources.requests.cpu` | `string` |  |  |  |
| `spec.components.triggerer.resources.requests.memory` | `string` |  |  |  |
| `spec.components.workers` | `KubernetesAirflowWorkers` |  |  |  |
| `spec.components.workers.replicas` | `int32` |  | `1` |  |
| `spec.components.workers.resources` | `ContainerResources` |  |  |  |
| `spec.components.workers.resources.limits` | `CpuMemory` |  |  |  |
| `spec.components.workers.resources.limits.cpu` | `string` |  |  |  |
| `spec.components.workers.resources.limits.memory` | `string` |  |  |  |
| `spec.components.workers.resources.requests` | `CpuMemory` |  |  |  |
| `spec.components.workers.resources.requests.cpu` | `string` |  |  |  |
| `spec.components.workers.resources.requests.memory` | `string` |  |  |  |
| `spec.components.workers.persistenceEnabled` | `bool` |  | `true` |  |
| `spec.components.workers.persistenceSize` | `string` |  | `100Gi` |  |
| `spec.components.workers.keda` | `KubernetesAirflowWorkersKeda` |  |  |  |
| `spec.components.workers.keda.enabled` | `bool` |  |  |  |
| `spec.components.workers.keda.minReplicas` | `int32` |  | `0` |  |
| `spec.components.workers.keda.maxReplicas` | `int32` |  | `10` |  |
| `spec.components.workers.keda.pollingIntervalSeconds` | `int32` |  | `5` |  |
| `spec.components.workers.keda.cooldownPeriodSeconds` | `int32` |  | `30` |  |
| `spec.pgbouncer` | `KubernetesAirflowPgBouncer` |  |  |  |
| `spec.pgbouncer.enabled` | `bool` |  |  |  |
| `spec.pgbouncer.metadataPoolSize` | `int32` |  | `10` |  |
| `spec.pgbouncer.resultBackendPoolSize` | `int32` |  | `5` |  |
| `spec.pgbouncer.maxClientConnections` | `int32` |  | `100` |  |
| `spec.pgbouncer.resources` | `ContainerResources` |  |  |  |
| `spec.pgbouncer.resources.limits` | `CpuMemory` |  |  |  |
| `spec.pgbouncer.resources.limits.cpu` | `string` |  |  |  |
| `spec.pgbouncer.resources.limits.memory` | `string` |  |  |  |
| `spec.pgbouncer.resources.requests` | `CpuMemory` |  |  |  |
| `spec.pgbouncer.resources.requests.cpu` | `string` |  |  |  |
| `spec.pgbouncer.resources.requests.memory` | `string` |  |  |  |
| `spec.logging` | `KubernetesAirflowLogging` |  |  |  |
| `spec.logging.persistence` | `KubernetesAirflowLogsPersistence` |  |  |  |
| `spec.logging.persistence.enabled` | `bool` |  |  |  |
| `spec.logging.persistence.size` | `string` |  | `100Gi` |  |
| `spec.logging.persistence.storageClass` | `string` |  |  |  |
| `spec.logging.elasticsearch` | `KubernetesAirflowLogSearchBackend` |  |  |  |
| `spec.logging.elasticsearch.host` | `string \| valueFrom` | yes |  | KubernetesOpenSearch (`status.outputs.service_name`) |
| `spec.logging.elasticsearch.port` | `int32` |  | `9200` |  |
| `spec.logging.elasticsearch.scheme` | `string` |  | `http` |  |
| `spec.logging.elasticsearch.username` | `string` |  |  |  |
| `spec.logging.elasticsearch.passwordSecret` | `KubernetesAirflowLogBackendPasswordSecret` |  |  |  |
| `spec.logging.elasticsearch.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesOpenSearch (`status.outputs.admin_credentials_secret_name`) |
| `spec.logging.elasticsearch.passwordSecret.secretKey` | `string` |  | `password` |  |
| `spec.logging.opensearch` | `KubernetesAirflowLogSearchBackend` |  |  |  |
| `spec.logging.opensearch.host` | `string \| valueFrom` | yes |  | KubernetesOpenSearch (`status.outputs.service_name`) |
| `spec.logging.opensearch.port` | `int32` |  | `9200` |  |
| `spec.logging.opensearch.scheme` | `string` |  | `http` |  |
| `spec.logging.opensearch.username` | `string` |  |  |  |
| `spec.logging.opensearch.passwordSecret` | `KubernetesAirflowLogBackendPasswordSecret` |  |  |  |
| `spec.logging.opensearch.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesOpenSearch (`status.outputs.admin_credentials_secret_name`) |
| `spec.logging.opensearch.passwordSecret.secretKey` | `string` |  | `password` |  |
| `spec.adminUser` | `KubernetesAirflowAdminUser` |  |  |  |
| `spec.adminUser.create` | `bool` |  | `true` |  |
| `spec.adminUser.username` | `string` |  | `admin` |  |
| `spec.adminUser.email` | `string` |  | `admin@example.com` |  |
| `spec.adminUser.passwordSecret` | `KubernetesAirflowExistingSecretRef` |  |  |  |
| `spec.adminUser.passwordSecret.secretName` | `string` | yes |  |  |
| `spec.adminUser.passwordSecret.secretKey` | `string` |  | `password` |  |
| `spec.security` | `KubernetesAirflowSecurity` |  |  |  |
| `spec.security.fernetKeySecretName` | `string` |  |  |  |
| `spec.security.apiSecretKeySecretName` | `string` |  |  |  |
| `spec.security.jwtSecretName` | `string` |  |  |  |
| `spec.statsdEnabled` | `bool` |  | `true` |  |
| `spec.loadExamples` | `bool` |  |  |  |
| `spec.scheduling` | `KubernetesAirflowScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.images` | `KubernetesAirflowImages` |  |  |  |
| `spec.images.airflowRepository` | `string` |  |  |  |
| `spec.images.airflowTag` | `string` |  |  |  |
| `spec.images.airflowDigest` | `string` |  |  |  |
| `spec.images.statsdRepository` | `string` |  |  |  |
| `spec.images.redisRepository` | `string` |  |  |  |
| `spec.images.pgbouncerRepository` | `string` |  |  |  |
| `spec.images.gitSyncRepository` | `string` |  |  |  |
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
resource. When false, the namespace must already exist. KNOW
THIS: the module reads the database (and broker) credential
Secrets to compose Airflow's connection Secrets, and a Secret can
only be read from the workload's OWN namespace — co-locate
Airflow with its database (the default composition) or replicate
the credential Secret into this namespace.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "1.22.0" — chart 1.22.0 ships
Airflow 3.2.2). Versions must exist as SERVED charts in the
repository index (https://airflow.apache.org). Chart versions
from 1.22.0 require Helm 3.19+ template functions — supported by
both deployment engines here.

- default: `1.22.0`

### spec.airflowVersion

`string` · optional (explicit presence)

The Airflow application version being deployed — the chart uses
it to gate version-specific rendering (which components exist,
how auth is wired). Only set this when `images` points at a
custom image built from a different Airflow release; it must
match that image's Airflow version or the chart renders the
wrong component set. This kind models the Airflow 3 line —
versions below 3.0.0 are rejected (the Airflow 2 webserver-era
surfaces are deliberately not modeled).

- default: `3.2.2`

### spec.executor

`string` · optional (explicit presence)

The task executor — HOW Airflow runs your tasks.

  - "KubernetesExecutor" (this kind's default): every task runs
    as its own ephemeral pod. No broker, no idle workers;
    per-task pod startup cost. The natural zero-dependency
    default on Kubernetes (the chart's own default is
    CeleryExecutor for historical reasons — declare it explicitly
    to get the Celery path).
  - "CeleryExecutor": a fleet of long-running worker pods
    consumes tasks from a message broker. Low task latency,
    needs `broker`.
  - "LocalExecutor": tasks run inside the scheduler pod — dev and
    small installs only.

Airflow 3 supports MULTIPLE executors as a comma-separated list
(e.g. "CeleryExecutor,KubernetesExecutor" — per-task routing via
the DAG's `executor` attribute), and custom executor class paths
(e.g. the AWS Batch/ECS executors from the amazon provider). Any
Celery-family member requires `broker`.

- default: `KubernetesExecutor`
- rule: {"string":{"pattern":"^(([a-zA-Z_][a-zA-Z0-9_]*\\.)*[A-Z][a-zA-Z0-9]+Executor)(,(([a-zA-Z_][a-zA-Z0-9_]*\\.)*[A-Z][a-zA-Z0-9]+Executor))*$"}}

### spec.database

`KubernetesAirflowDatabase` · required

The metadata database — Airflow's single source of truth (DAG
runs, task state, connections, users). Required — nothing is
bundled.

- rule: {"required":true}

### spec.database.postgres

`KubernetesAirflowPostgres`

PostgreSQL — the recommended backend. Defaults compose a
KubernetesPostgres resource; any reachable PostgreSQL ≥ 13
(RDS, Cloud SQL, Aurora) works with literal values.

### spec.database.postgres.host

`string | valueFrom` · required

Database server host. Defaults compose a KubernetesPostgres
resource's read-write Service; any reachable hostname works as a
literal value.

- references: KubernetesPostgres (`status.outputs.rw_service`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.rw_service}} -- a bare string does not parse

### spec.database.postgres.port

`int32` · optional (explicit presence)

Database server port. Empty = 5432.

- default: `5432`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.database.postgres.databaseName

`string` · optional (explicit presence)

Database name holding Airflow's metadata. Must EXIST before
install (the migration Job creates tables, never the database) —
on a KubernetesPostgres, declare it at bootstrap
(`initdb.database`). Empty = "airflow".

- default: `airflow`
- rule: {"string":{"pattern":"^[a-zA-Z_][a-zA-Z0-9_$]*$"}}

### spec.database.postgres.username

`string` · optional (explicit presence)

Database user. Needs full rights inside `database_name`
(ownership is simplest — declare the same user as the database
owner at bootstrap). Empty = "airflow".

- default: `airflow`

### spec.database.postgres.passwordSecret

`KubernetesAirflowPasswordSecret` · required

The Secret holding the user's password (wired into the composed
connection Secret at apply time — it never lands in rendered
values). Defaults compose a KubernetesPostgres resource's
application-user Secret. Same-namespace constraint applies.

- rule: {"required":true}

### spec.database.postgres.passwordSecret.secretName

`string | valueFrom` · required

Name of the Secret. Defaults compose a KubernetesPostgres
resource's application-user Secret (`<cluster>-app`); set the
matching Secret for other database kinds or external stores.
Same-namespace constraint (a Kubernetes rule, not a chart one):
the Secret must live in the namespace Airflow installs into.

- references: KubernetesPostgres (`status.outputs.password_secret.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.password_secret.name}} -- a bare string does not parse

### spec.database.postgres.passwordSecret.secretKey

`string` · optional (explicit presence)

Key inside the Secret holding the password. Empty = "password"
(the KubernetesPostgres application-Secret convention).

- default: `password`

### spec.database.postgres.sslMode

`string` · optional (explicit presence)

PostgreSQL SSL mode for the connection (`sslmode` query
parameter). Empty = "disable" (the chart default — in-cluster
traffic). Set "require" or stricter for managed databases over
untrusted networks.

- default: `disable`
- rule: {"string":{"in":["disable","allow","prefer","require","verify-ca","verify-full"]}}

### spec.database.mysql

`KubernetesAirflowMysql`

MySQL 8+. External only — point at a KubernetesMysql resource
or any reachable MySQL 8 server.

### spec.database.mysql.host

`string | valueFrom` · required

Database server host. Defaults compose a KubernetesMysql
resource's client Service; any reachable MySQL 8 host works as a
literal value.

- references: KubernetesMysql (`status.outputs.primary_service`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesMysql, name: <that resource's name>, fieldPath: status.outputs.primary_service}} -- a bare string does not parse

### spec.database.mysql.port

`int32` · optional (explicit presence)

Database server port. Empty = 3306.

- default: `3306`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.database.mysql.databaseName

`string` · optional (explicit presence)

Database name holding Airflow's metadata. Must EXIST before
install. Empty = "airflow".

- default: `airflow`
- rule: {"string":{"pattern":"^[a-zA-Z_][a-zA-Z0-9_$]*$"}}

### spec.database.mysql.username

`string` · optional (explicit presence)

Database user with full rights inside `database_name`. Empty =
"airflow".

- default: `airflow`

### spec.database.mysql.passwordSecret

`KubernetesAirflowMysqlPasswordSecret` · required

The Secret holding the user's password (wired into the composed
connection Secret at apply time). Same-namespace constraint
applies.

- rule: {"required":true}

### spec.database.mysql.passwordSecret.secretName

`string | valueFrom` · required

Name of the Secret. Defaults compose a KubernetesMysql
resource's root credential Secret; for a dedicated Airflow user
(recommended), point at the Secret holding that user's password.
Same-namespace constraint applies.

- references: KubernetesMysql (`status.outputs.root_password_secret.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesMysql, name: <that resource's name>, fieldPath: status.outputs.root_password_secret.name}} -- a bare string does not parse

### spec.database.mysql.passwordSecret.secretKey

`string` · optional (explicit presence)

Key inside the Secret holding the password. Empty = "password".

- default: `password`

### spec.broker

`KubernetesAirflowBroker`

The Celery message broker. REQUIRED when `executor` contains a
Celery family member ("CeleryExecutor" or
"CeleryKubernetesExecutor"); must be empty otherwise — the
KubernetesExecutor and LocalExecutor paths have no broker.

### spec.broker.bundledRedis

`KubernetesAirflowBundledRedis`

The chart's bundled single-instance Redis StatefulSet. Upstream
pins the image to `redis:7.2-bookworm` — the last
BSD-3-licensed Redis line (deliberate upstream choice; never
bump it past 7.2 via image overrides without checking the
license). Fine for dev and small installs; the composed Valkey
arm is the licensing-clean production path.

### spec.broker.bundledRedis.persistenceSize

`string` · optional (explicit presence)

Persistent volume size for the Redis StatefulSet. Empty = "1Gi"
(the chart default). The broker holds only in-flight task
messages — small volumes are fine.

- default: `1Gi`
- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?(Ei|Pi|Ti|Gi|Mi|Ki|E|P|T|G|M|k|m)?$"}}

### spec.broker.bundledRedis.storageClass

`string`

Storage class for the Redis PVC. Empty = the cluster default.

### spec.broker.bundledRedis.resources

`ContainerResources`

CPU/memory for the Redis container. Empty = the chart's defaults
(no requests).

### spec.broker.bundledRedis.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.broker.bundledRedis.resources.limits.cpu

`string`

### spec.broker.bundledRedis.resources.limits.memory

`string`

### spec.broker.bundledRedis.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.broker.bundledRedis.resources.requests.cpu

`string`

### spec.broker.bundledRedis.resources.requests.memory

`string`

### spec.broker.valkey

`KubernetesAirflowValkeyBroker`

A Redis-protocol broker you operate — a KubernetesValkey
composes naturally (the recommended production path).

### spec.broker.valkey.host

`string | valueFrom` · required

Broker host. Defaults compose a KubernetesValkey resource's
client Service; any reachable Redis-protocol host works as a
literal value.

- references: KubernetesValkey (`status.outputs.service`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesValkey, name: <that resource's name>, fieldPath: status.outputs.service}} -- a bare string does not parse

### spec.broker.valkey.port

`int32` · optional (explicit presence)

Broker port. Empty = 6379.

- default: `6379`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.broker.valkey.username

`string`

Redis-protocol username for ACL-authenticated brokers (a
KubernetesValkey ACL user's name). Empty = password-only AUTH
(the URL renders as `redis://:pass@host`).

### spec.broker.valkey.passwordSecret

`KubernetesAirflowBrokerPasswordSecret`

The Secret holding the broker password (composed into the broker
URL at apply time — never rendered into values). Defaults
compose a KubernetesValkey resource's auth Secret (one key per
username — set `secret_key` to the username). Empty message =
an unauthenticated broker (never production).

### spec.broker.valkey.passwordSecret.secretName

`string | valueFrom` · required

Name of the Secret. Defaults compose a KubernetesValkey
resource's auth Secret (`<name>-auth`). Same-namespace
constraint applies.

- references: KubernetesValkey (`status.outputs.password_secret.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesValkey, name: <that resource's name>, fieldPath: status.outputs.password_secret.name}} -- a bare string does not parse

### spec.broker.valkey.passwordSecret.secretKey

`string` · optional (explicit presence)

Key inside the Secret holding the password. For a
KubernetesValkey auth Secret this is the ACL username. Empty =
"password".

- default: `password`

### spec.broker.valkey.databaseNumber

`int32` · optional (explicit presence)

Redis database number for the Celery broker. Empty = 0.

- default: `0`
- rule: {"int32":{"lte":15,"gte":0}}

### spec.broker.existingBrokerUrlSecret

`KubernetesAirflowBrokerUrlSecret`

An existing Secret already holding the complete broker URL
under the `connection` key (e.g.
`redis://:password@my-broker:6379/0`) — the escape arm for
brokers the typed arms do not model (SQS, external Redis with
TLS). The module passes the Secret NAME to the chart; the URL
itself never renders.

### spec.broker.existingBrokerUrlSecret.secretName

`string` · required

Name of the Secret. It must carry the full broker URL under the
`connection` key (the chart's contract). Same-namespace
constraint applies.

- rule: {"required":true}

### spec.dags

`KubernetesAirflowDags`

DAG delivery — how your pipeline code reaches every Airflow
component. Empty = DAGs are baked into the container image (set
`images.airflow_repository`/`airflow_tag` to your image). The
recommended production path is `git_sync`.

### spec.dags.gitSync

`KubernetesAirflowGitSync`

A git-sync sidecar on every component clones your repository
and keeps it current — push to deploy pipelines.

- rule: ssh_key_secret pairs with an SSH clone URL (git@host:org/repo.git or ssh://...) — git-sync ignores SSH keys for https:// repositories. Switch repo to the SSH form or use credentials_secret.
- rule: credentials_secret (GITSYNC_USERNAME/GITSYNC_PASSWORD) pairs with an https:// clone URL — SSH repositories authenticate with ssh_key_secret instead.
- rule: Declare at most one of credentials_secret and ssh_key_secret — a repository is cloned over HTTPS or SSH, never both.

### spec.dags.gitSync.repo

`string` · required

Git repository clone URL. HTTPS form
(`https://github.com/org/repo.git`) pairs with
`credentials_secret` (or none for public repos); SSH form
(`git@github.com:org/repo.git` or `ssh://...`) pairs with
`ssh_key_secret`.

- rule: {"required":true}

### spec.dags.gitSync.ref

`string`

The git revision to sync — a branch name, tag, or commit hash
(git-sync v4 `ref`). Empty = the repository's default branch
(HEAD).

RENDERING: the chart emits BOTH env generations unconditionally —
`GITSYNC_REF` from `ref` (v4) and `GIT_SYNC_BRANCH` from `branch`
(legacy) — and git-sync v4 translates the deprecated `--branch`
OVER `--ref`. A ref-only render therefore silently syncs the
chart's default branch (`v2-2-stable`, verified live). The module
always writes BOTH keys to this field's value — including the
empty string, which neutralizes the chart defaults so Empty=HEAD
holds (git-sync treats empty ref/branch as HEAD).

### spec.dags.gitSync.subPath

`string`

Path within the repository where DAGs live. Empty = the
repository root.

### spec.dags.gitSync.periodSeconds

`int32` · optional (explicit presence)

Seconds between sync attempts. Low values deploy DAG changes
faster but poll the remote more; high values risk components
briefly disagreeing about DAG versions. Empty = 5.

- default: `5`
- rule: {"int32":{"lte":3600,"gte":1}}

### spec.dags.gitSync.depth

`int32` · optional (explicit presence)

Shallow-clone depth. Empty = 1 (the chart default — fetches only
the synced revision; raise it only if your DAGs read git
history).

- default: `1`
- rule: {"int32":{"gte":0}}

### spec.dags.gitSync.credentialsSecret

`string`

Secret holding HTTPS credentials for private repositories, with
the git-sync v4 keys `GITSYNC_USERNAME` and `GITSYNC_PASSWORD`
(a token works as the password). Only the Secret NAME renders.
HTTPS repos only.

### spec.dags.gitSync.sshKeySecret

`string`

Secret holding the SSH private key for private repositories,
under the key `gitSshKey` (the chart's contract). Only the
Secret NAME renders. SSH repos only.

### spec.dags.gitSync.knownHosts

`string`

known_hosts content for SSH cloning — pin your Git host's keys
so the sidecar refuses man-in-the-middle answers. Strongly
recommended with `ssh_key_secret`.

### spec.dags.gitSync.resources

`ContainerResources`

CPU/memory for the git-sync sidecar containers. Empty = the
chart's defaults (no requests).

### spec.dags.gitSync.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.dags.gitSync.resources.limits.cpu

`string`

### spec.dags.gitSync.resources.limits.memory

`string`

### spec.dags.gitSync.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.dags.gitSync.resources.requests.cpu

`string`

### spec.dags.gitSync.resources.requests.memory

`string`

### spec.dags.persistence

`KubernetesAirflowDagsPersistence`

A shared ReadWriteMany volume holding DAG files — for teams
that publish DAGs via CI into a volume. Requires a storage
class capable of ReadWriteMany when any component runs more
than one replica.

### spec.dags.persistence.size

`string` · optional (explicit presence)

Volume size. Empty = "1Gi" (the chart default — DAG files are
small).

- default: `1Gi`
- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?(Ei|Pi|Ti|Gi|Mi|Ki|E|P|T|G|M|k|m)?$"}}

### spec.dags.persistence.storageClass

`string`

Storage class. Multi-replica components need a ReadWriteMany
class here (the chart mounts one PVC on every component). Empty
= the cluster default.

### spec.dags.persistence.existingClaim

`string`

Use an existing PVC (your CI publishes DAGs into it) instead of
chart-created storage. Empty = the chart creates the PVC.

### spec.components

`KubernetesAirflowComponents`

Per-component sizing and behavior for the Airflow 3 component
set. Empty = one replica of each with the chart's defaults (no
resource requests — fine for dev, size them for production; the
triggerer ships a 100Gi PVC by default — set
`triggerer.persistence_size` for small clusters).

### spec.components.apiServer

`KubernetesAirflowApiServer`

The API server (UI + REST API).

### spec.components.apiServer.replicas

`int32` · optional (explicit presence)

Replicas. Empty = 1. Safe to scale — state lives in the
database; sessions ride the shared API secret key.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.components.apiServer.resources

`ContainerResources`

CPU/memory for the API server container. Empty = the chart's
defaults.

### spec.components.apiServer.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.components.apiServer.resources.limits.cpu

`string`

### spec.components.apiServer.resources.limits.memory

`string`

### spec.components.apiServer.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.components.apiServer.resources.requests.cpu

`string`

### spec.components.apiServer.resources.requests.memory

`string`

### spec.components.scheduler

`KubernetesAirflowScheduler`

The scheduler — the heart of Airflow; it decides what runs when.

### spec.components.scheduler.replicas

`int32` · optional (explicit presence)

Replicas. Empty = 1. Multiple schedulers are an Airflow HA
feature (they coordinate through database row locks — PostgreSQL
recommended when running more than one).

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.components.scheduler.resources

`ContainerResources`

CPU/memory for the scheduler container. Empty = the chart's
defaults.

### spec.components.scheduler.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.components.scheduler.resources.limits.cpu

`string`

### spec.components.scheduler.resources.limits.memory

`string`

### spec.components.scheduler.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.components.scheduler.resources.requests.cpu

`string`

### spec.components.scheduler.resources.requests.memory

`string`

### spec.components.dagProcessor

`KubernetesAirflowDagProcessor`

The standalone DAG processor — parses DAG files continuously
(always standalone on Airflow 3).

### spec.components.dagProcessor.replicas

`int32` · optional (explicit presence)

Replicas. Empty = 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.components.dagProcessor.resources

`ContainerResources`

CPU/memory for the DAG processor container. Empty = the chart's
defaults.

### spec.components.dagProcessor.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.components.dagProcessor.resources.limits.cpu

`string`

### spec.components.dagProcessor.resources.limits.memory

`string`

### spec.components.dagProcessor.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.components.dagProcessor.resources.requests.cpu

`string`

### spec.components.dagProcessor.resources.requests.memory

`string`

### spec.components.triggerer

`KubernetesAirflowTriggerer`

The triggerer — runs deferred/async operators (sensors that wait
without holding a worker slot).

### spec.components.triggerer.enabled

`bool` · optional (explicit presence)

Run the triggerer. Empty = true (the chart default) — deferred
operators silently never fire without it, so disable only when
you are certain no DAG uses deferrable operators.

- default: `true`

### spec.components.triggerer.replicas

`int32` · optional (explicit presence)

Replicas. Empty = 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.components.triggerer.persistenceSize

`string` · optional (explicit presence)

Persistent volume size for the triggerer StatefulSet. KNOW THIS:
the chart default is 100Gi — far more than most installs need;
"1Gi" is plenty for the triggerer's local state on dev clusters.

- default: `100Gi`
- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?(Ei|Pi|Ti|Gi|Mi|Ki|E|P|T|G|M|k|m)?$"}}

### spec.components.triggerer.resources

`ContainerResources`

CPU/memory for the triggerer container. Empty = the chart's
defaults.

### spec.components.triggerer.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.components.triggerer.resources.limits.cpu

`string`

### spec.components.triggerer.resources.limits.memory

`string`

### spec.components.triggerer.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.components.triggerer.resources.requests.cpu

`string`

### spec.components.triggerer.resources.requests.memory

`string`

### spec.components.workers

`KubernetesAirflowWorkers`

Celery workers. Only meaningful with a Celery-family executor —
KubernetesExecutor tasks run as ephemeral pods sized via the
worker pod template instead.

### spec.components.workers.replicas

`int32` · optional (explicit presence)

Worker replicas. Empty = 1. Ignored while KEDA autoscaling is
enabled (KEDA owns the replica count).

- default: `1`
- rule: {"int32":{"gte":0}}

### spec.components.workers.resources

`ContainerResources`

CPU/memory for the worker containers. Empty = the chart's
defaults (no requests — size these for production; every Celery
worker runs up to 16 tasks concurrently by default).

### spec.components.workers.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.components.workers.resources.limits.cpu

`string`

### spec.components.workers.resources.limits.memory

`string`

### spec.components.workers.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.components.workers.resources.requests.cpu

`string`

### spec.components.workers.resources.requests.memory

`string`

### spec.components.workers.persistenceEnabled

`bool` · optional (explicit presence)

Persist worker-local task logs on a PVC per worker (the workers
run as a StatefulSet). Empty = true (the chart default, 100Gi
volumes — set `persistence_size` down for dev clusters). When
false, workers run as a Deployment with ephemeral logs — pair
with a remote logging backend.

- default: `true`

### spec.components.workers.persistenceSize

`string` · optional (explicit presence)

Persistent volume size per worker. Empty = "100Gi" (the chart
default; "1Gi" is fine for dev).

- default: `100Gi`
- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?(Ei|Pi|Ti|Gi|Mi|Ki|E|P|T|G|M|k|m)?$"}}

### spec.components.workers.keda

`KubernetesAirflowWorkersKeda`

KEDA autoscaling — scale workers on the REAL queue depth (KEDA
polls the metadata database for queued/running tasks). Requires
the KEDA operator on the cluster — a KubernetesKeda composes
naturally. The chart disables the plain HPA path when KEDA is
on.

- rule: KEDA min_replicas cannot exceed max_replicas — the autoscaler needs a valid range to scale within.

### spec.components.workers.keda.enabled

`bool`

Enable KEDA autoscaling for the Celery workers.

### spec.components.workers.keda.minReplicas

`int32` · optional (explicit presence)

Minimum worker replicas. Empty = 0 (scale to zero between
pipeline runs — the whole point of queue-driven scaling).

- default: `0`
- rule: {"int32":{"gte":0}}

### spec.components.workers.keda.maxReplicas

`int32` · optional (explicit presence)

Maximum worker replicas. Empty = 10.

- default: `10`
- rule: {"int32":{"gte":1}}

### spec.components.workers.keda.pollingIntervalSeconds

`int32` · optional (explicit presence)

Seconds between KEDA's queue-depth polls of the metadata
database. Empty = 5.

- default: `5`
- rule: {"int32":{"gte":1}}

### spec.components.workers.keda.cooldownPeriodSeconds

`int32` · optional (explicit presence)

Seconds KEDA waits before scaling to zero. Empty = 30.

- default: `30`
- rule: {"int32":{"gte":0}}

### spec.pgbouncer

`KubernetesAirflowPgBouncer`

PgBouncer connection pooling between Airflow and the metadata
database. Airflow opens MANY short-lived database connections
(every task heartbeat is one) — production installs on PostgreSQL
should enable this. The module composes the PgBouncer
configuration (including its credential file) into a Secret from
the declared database credentials — the chart's own rendering
path would put the password in Helm values and is never used.
PostgreSQL only — rejected when the database is MySQL.

### spec.pgbouncer.enabled

`bool`

Enable PgBouncer.

### spec.pgbouncer.metadataPoolSize

`int32` · optional (explicit presence)

Pool size for the metadata database. Empty = 10 (the chart
default).

- default: `10`
- rule: {"int32":{"gte":1}}

### spec.pgbouncer.resultBackendPoolSize

`int32` · optional (explicit presence)

Pool size for the Celery result-backend connection. Empty = 5
(the chart default).

- default: `5`
- rule: {"int32":{"gte":1}}

### spec.pgbouncer.maxClientConnections

`int32` · optional (explicit presence)

Maximum client connections PgBouncer accepts. Empty = 100 (the
chart default).

- default: `100`
- rule: {"int32":{"gte":1}}

### spec.pgbouncer.resources

`ContainerResources`

CPU/memory for the PgBouncer container. Empty = the chart's
defaults.

### spec.pgbouncer.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.pgbouncer.resources.limits.cpu

`string`

### spec.pgbouncer.resources.limits.memory

`string`

### spec.pgbouncer.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.pgbouncer.resources.requests.cpu

`string`

### spec.pgbouncer.resources.requests.memory

`string`

### spec.logging

`KubernetesAirflowLogging`

Where task logs live and how the UI reads them. Empty = logs on
pod-local storage (lost on pod replacement — fine for dev only).
Enable `persistence` for a shared logs volume, or point the UI's
log READ path at an Elasticsearch/OpenSearch cluster your tasks
already ship logs to.

### spec.logging.persistence

`KubernetesAirflowLogsPersistence`

Keep task logs on a shared persistent volume so the UI can read
them after pods rotate. Requires ReadWriteMany storage when
components scale beyond one replica.

### spec.logging.persistence.enabled

`bool`

Enable the shared logs PVC.

### spec.logging.persistence.size

`string` · optional (explicit presence)

Volume size. Empty = "100Gi" (the chart default; size down for
dev).

- default: `100Gi`
- rule: {"string":{"pattern":"^[0-9]+(\\.[0-9]+)?(Ei|Pi|Ti|Gi|Mi|Ki|E|P|T|G|M|k|m)?$"}}

### spec.logging.persistence.storageClass

`string`

Storage class. Needs ReadWriteMany capability when any
log-writing component runs multiple replicas. Empty = the
cluster default.

### spec.logging.elasticsearch

`KubernetesAirflowLogSearchBackend`

Elasticsearch log read path.

- rule: A username without a password_secret cannot authenticate — declare the Secret holding this user's password, or clear username for an unauthenticated dev backend.

### spec.logging.elasticsearch.host

`string | valueFrom` · required

Backend host. For OpenSearch, defaults compose a
KubernetesOpenSearch resource's client Service.

- references: KubernetesOpenSearch (`status.outputs.service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesOpenSearch, name: <that resource's name>, fieldPath: status.outputs.service_name}} -- a bare string does not parse

### spec.logging.elasticsearch.port

`int32` · optional (explicit presence)

Backend port. Empty = 9200.

- default: `9200`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.logging.elasticsearch.scheme

`string` · optional (explicit presence)

Connection scheme. Empty = "http".

- default: `http`
- rule: {"string":{"in":["http","https"]}}

### spec.logging.elasticsearch.username

`string`

Username for the backend. Empty = unauthenticated (dev only).

### spec.logging.elasticsearch.passwordSecret

`KubernetesAirflowLogBackendPasswordSecret`

The Secret holding the user's password (composed into the
connection Secret at apply time). Required when `username` is
set. Same-namespace constraint applies.

### spec.logging.elasticsearch.passwordSecret.secretName

`string | valueFrom` · required

Name of the Secret. Defaults compose a KubernetesOpenSearch
resource's operator-generated admin-credentials Secret (keys
`username`/`password`); that output is EMPTY when a custom
security config replaces the operator bootstrap — point at your
custom credential Secret then. The elasticsearch arm has no
in-catalog kind: set an explicit Secret name for an external
Elasticsearch. Same-namespace constraint (a Kubernetes rule, not
a chart one): the Secret must live in the namespace Airflow
installs into.

- references: KubernetesOpenSearch (`status.outputs.admin_credentials_secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesOpenSearch, name: <that resource's name>, fieldPath: status.outputs.admin_credentials_secret_name}} -- a bare string does not parse

### spec.logging.elasticsearch.passwordSecret.secretKey

`string` · optional (explicit presence)

Key inside the Secret holding the password. Empty = "password"
(the KubernetesOpenSearch admin-Secret convention).

- default: `password`

### spec.logging.opensearch

`KubernetesAirflowLogSearchBackend`

OpenSearch log read path — a KubernetesOpenSearch composes
naturally.

- rule: A username without a password_secret cannot authenticate — declare the Secret holding this user's password, or clear username for an unauthenticated dev backend.

### spec.logging.opensearch.host

`string | valueFrom` · required

Backend host. For OpenSearch, defaults compose a
KubernetesOpenSearch resource's client Service.

- references: KubernetesOpenSearch (`status.outputs.service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesOpenSearch, name: <that resource's name>, fieldPath: status.outputs.service_name}} -- a bare string does not parse

### spec.logging.opensearch.port

`int32` · optional (explicit presence)

Backend port. Empty = 9200.

- default: `9200`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.logging.opensearch.scheme

`string` · optional (explicit presence)

Connection scheme. Empty = "http".

- default: `http`
- rule: {"string":{"in":["http","https"]}}

### spec.logging.opensearch.username

`string`

Username for the backend. Empty = unauthenticated (dev only).

### spec.logging.opensearch.passwordSecret

`KubernetesAirflowLogBackendPasswordSecret`

The Secret holding the user's password (composed into the
connection Secret at apply time). Required when `username` is
set. Same-namespace constraint applies.

### spec.logging.opensearch.passwordSecret.secretName

`string | valueFrom` · required

Name of the Secret. Defaults compose a KubernetesOpenSearch
resource's operator-generated admin-credentials Secret (keys
`username`/`password`); that output is EMPTY when a custom
security config replaces the operator bootstrap — point at your
custom credential Secret then. The elasticsearch arm has no
in-catalog kind: set an explicit Secret name for an external
Elasticsearch. Same-namespace constraint (a Kubernetes rule, not
a chart one): the Secret must live in the namespace Airflow
installs into.

- references: KubernetesOpenSearch (`status.outputs.admin_credentials_secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesOpenSearch, name: <that resource's name>, fieldPath: status.outputs.admin_credentials_secret_name}} -- a bare string does not parse

### spec.logging.opensearch.passwordSecret.secretKey

`string` · optional (explicit presence)

Key inside the Secret holding the password. Empty = "password"
(the KubernetesOpenSearch admin-Secret convention).

- default: `password`

### spec.adminUser

`KubernetesAirflowAdminUser`

The bootstrap admin user for the Airflow UI and API (the chart's
FAB auth manager posture). The module generates the password into
a Secret (`<name>-admin-auth`, key `password`) unless
`password_secret` points at an existing one — the chart's
admin/admin default never ships.

### spec.adminUser.create

`bool` · optional (explicit presence)

Create the admin user at install (the chart's bootstrap Job).
Empty = true. Disable only when an external identity story
(declared via `helm_values`) owns user management.

- default: `true`

### spec.adminUser.username

`string` · optional (explicit presence)

Admin username. Empty = "admin".

- default: `admin`

### spec.adminUser.email

`string` · optional (explicit presence)

Admin email. Empty = "admin@example.com".

- default: `admin@example.com`

### spec.adminUser.passwordSecret

`KubernetesAirflowExistingSecretRef`

Existing Secret holding the admin password. Empty = the module
GENERATES the password into `<name>-admin-auth` (key `password`)
— the chart's admin/admin default never ships. The password
reaches the bootstrap Job as an environment variable from the
Secret, never as a rendered pod argument.

### spec.adminUser.passwordSecret.secretName

`string` · required

Name of the Secret. Same-namespace constraint applies.

- rule: {"required":true}

### spec.adminUser.passwordSecret.secretKey

`string` · optional (explicit presence)

Key inside the Secret. Empty = "password".

- default: `password`

### spec.security

`KubernetesAirflowSecurity`

Bring-your-own security Secrets. Empty = the module generates
each one (the correct default for a fresh install). Set these to
compose externally-managed keys (e.g. ExternalSecrets) or to
share a Fernet key across a DR pair. Each Secret must carry the
exact key the chart expects (fernet-key: `fernet-key`, API secret
key: `webserver-secret-key`, JWT: `jwt-secret`).

### spec.security.fernetKeySecretName

`string`

Secret holding the Fernet key under the `fernet-key` key.
Airflow encrypts every stored connection password and variable
with it — LOSING IT ORPHANS ALL STORED SECRETS, so share it
across DR replicas and rotate it only with Airflow's documented
rotation procedure. Empty = module-generated
(`<name>-fernet-key`).

### spec.security.apiSecretKeySecretName

`string`

Secret holding the API server's session/signing secret under the
`api-secret-key` key. All API server replicas must share it.
Empty = module-generated (`<name>-api-secret-key`). The module
also always generates the FAB webserver session key
(`<name>-webserver-secret-key`) — the chart would otherwise
regenerate it on every upgrade render, logging out every UI
session.

### spec.security.jwtSecretName

`string`

Secret holding the JWT signing secret under the `jwt-secret` key
— Airflow 3 components authenticate to the API server with
tokens signed by it. Empty = module-generated
(`<name>-jwt-secret`).

### spec.statsdEnabled

`bool` · optional (explicit presence)

Emit StatsD metrics through the chart's statsd-exporter sidecar
Deployment (Prometheus scrape port 9102, pod annotations set).
Enabled by default, matching the chart. The chart ships no
ServiceMonitor — compose scraping via annotations or a
KubernetesKubePrometheusStack additionalScrapeConfig.

- default: `true`

### spec.loadExamples

`bool`

Load Airflow's bundled example DAGs at startup. Dev/evaluation
only — never enable in production (the examples poll external
sites and clutter the DAG list). Matches the chart default
(false).

WIRING: the module sets this via the `AIRFLOW__CORE__LOAD_EXAMPLES`
env var, NOT `config.core.load_examples`. The official image bakes
`AIRFLOW__CORE__LOAD_EXAMPLES=False` as a container env, and
Airflow's precedence puts env above airflow.cfg — a cfg-only True
is silently defeated (verified live: examples never parsed; the
chart's own docs prescribe the env route).

### spec.scheduling

`KubernetesAirflowScheduling`

Pod scheduling applied to all Airflow pods (components, Jobs,
sidecars).

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for all Airflow pods.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for all Airflow pods.

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

`KubernetesAirflowImages`

Container image overrides for air-gapped clusters and private
mirrors. Empty = the chart's pinned upstream images.

### spec.images.airflowRepository

`string`

The Airflow image repository used by every Airflow component
(e.g. "my-registry.example.com/mirror/airflow" — or your custom
DAG-baked image). Empty = "apache/airflow".

### spec.images.airflowTag

`string`

The Airflow image tag. Empty = the chart's pinned tag for
`airflow_version`. When this tag ships a different Airflow
release, set `airflow_version` to match it.

### spec.images.airflowDigest

`string`

Image digest (overrides the tag when set) — for
supply-chain-pinned deploys.

### spec.images.statsdRepository

`string`

StatsD-exporter image (combined form). Empty =
"quay.io/prometheus/statsd-exporter".

### spec.images.redisRepository

`string`

Bundled-Redis image (combined form). Empty = "redis" (tag pinned
7.2-bookworm by the chart — the last BSD-licensed line; check
licensing before overriding the tag upward).

### spec.images.pgbouncerRepository

`string`

PgBouncer image (combined form). Empty = "apache/airflow" (the
chart's pgbouncer build).

### spec.images.gitSyncRepository

`string`

git-sync sidecar image (combined form). Empty =
"registry.k8s.io/git-sync/git-sync".

### spec.helmValues

`string`

Additional Helm values merged LAST (Helm `-f` semantics,
identical on both engines) — the escape hatch for chart values
the typed fields do not model: Kerberos, per-component
env/volumes/lifecycle hooks, the chart's OTel collector sidecar,
cleanup/database-cleanup CronJobs, network policies, priority
classes, Airflow config overrides under `config`. YAML document
as a string. Never put secret material here; credentials belong
in the typed secret-reference fields, which keep them out of
rendered values. Never re-enable the `postgresql` subchart — the
bundled database is non-production by upstream's own definition
and this module always disables it.

## Validation Rules

- `broker.required_for_celery`: A Celery-family executor needs a message broker: set spec.broker (bundled Redis, a composed KubernetesValkey, or an existing broker-URL Secret) when executor contains CeleryExecutor or CeleryKubernetesExecutor.
- `broker.forbidden_without_celery`: spec.broker only applies to Celery-family executors — remove it, or set spec.executor to include CeleryExecutor explicitly (unset executor defaults to KubernetesExecutor, which runs tasks without a message broker).
- `airflow_version.v3_line_only`: This kind models the Airflow 3 line — set airflow_version to 3.0.0 or newer. Airflow 2 installs use the chart's webserver-era surfaces this spec deliberately does not model.
- `pgbouncer.postgres_only`: PgBouncer pools PostgreSQL connections — it cannot front a MySQL metadata database. Remove spec.pgbouncer or switch the database to the postgres arm.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesAirflow, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace Airflow runs in. |
| `status.outputs.api_server_service` | `string` | Name of the API server Service (`<name>-api-server`) — the UI + REST API front door. The handle exposure kinds route to. |
| `status.outputs.api_server_endpoint` | `string` | In-cluster API server endpoint, `http://<name>-api-server.<namespace>.svc.cluster.local:8080` — what in-cluster clients (and the REST API's callers) use. |
| `status.outputs.admin_password_secret` | `KubernetesSecretKey` | The admin user's credential: the Secret and key holding the bootstrap admin password (module-generated `<name>-admin-auth` unless an existing Secret was declared). Username lives in the spec (`admin_user.username`, default "admin"). Empty when `admin_user.create` is false — no bootstrap credential exists to point at. |
| `status.outputs.admin_password_secret.name` | `string` | The name of the Kubernetes Secret. |
| `status.outputs.admin_password_secret.key` | `string` | The key within the Kubernetes Secret. |
| `status.outputs.metadata_connection_secret_name` | `string` | Name of the module-owned Secret holding the metadata database connection URI (key `connection`) — the credential handle other tools (e.g. a data-lineage sidecar) can mount to reach the same database. |
| `status.outputs.broker_url_secret_name` | `string` | Name of the module-owned Secret holding the Celery broker URL (key `connection`); empty when no Celery executor is declared. |
| `status.outputs.fernet_key_secret_name` | `string` | Name of the Secret holding the Fernet key (`fernet-key` key) — back it up: losing the Fernet key orphans every credential Airflow has stored. |
| `status.outputs.port_forward_command` | `string` | Port-forward command for reaching the Airflow UI from a workstation when no exposure is composed (`kubectl port-forward svc/<name>-api-server -n <namespace> 8080:8080`). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.database.postgres.host` | KubernetesPostgres | `status.outputs.rw_service` |
| `spec.database.postgres.passwordSecret.secretName` | KubernetesPostgres | `status.outputs.password_secret.name` |
| `spec.database.mysql.host` | KubernetesMysql | `status.outputs.primary_service` |
| `spec.database.mysql.passwordSecret.secretName` | KubernetesMysql | `status.outputs.root_password_secret.name` |
| `spec.broker.valkey.host` | KubernetesValkey | `status.outputs.service` |
| `spec.broker.valkey.passwordSecret.secretName` | KubernetesValkey | `status.outputs.password_secret.name` |
| `spec.logging.elasticsearch.host` | KubernetesOpenSearch | `status.outputs.service_name` |
| `spec.logging.elasticsearch.passwordSecret.secretName` | KubernetesOpenSearch | `status.outputs.admin_credentials_secret_name` |
| `spec.logging.opensearch.host` | KubernetesOpenSearch | `status.outputs.service_name` |
| `spec.logging.opensearch.passwordSecret.secretName` | KubernetesOpenSearch | `status.outputs.admin_credentials_secret_name` |

## See Also

- [Overview](../README.md)
