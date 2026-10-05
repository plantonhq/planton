# Kubernetes Airflow

Apache Airflow 3 from the official `airflow` Helm chart
(https://airflow.apache.org): the API server (UI + REST API),
scheduler, standalone DAG processor, triggerer, and — on Celery
executors — a worker fleet, with the database-migration and
admin-bootstrap Jobs handled at install.

## When NOT to Use This

If tasks are simple containers on a schedule with no dependencies
between them, a `KubernetesCronJob` is the honest tool. If you need a
durable workflow ENGINE (application code that survives crashes and
waits weeks), that is `KubernetesTemporal` — Airflow orchestrates
data pipelines; Temporal executes long-lived business logic.

## Bring your own database

Nothing is bundled: the chart's PostgreSQL subchart is non-production
by upstream's own definition (frozen Bitnami-legacy image) and this
kind never ships it. Declare `database.postgres` (a
`KubernetesPostgres` composes naturally — host defaults to its
read-write Service, credential to its application Secret) or an
external MySQL 8. The database must exist before install (declare it
at the Postgres bootstrap); the migration Job creates the tables.

## The credential path

Airflow consumes its database and broker connections as Secrets
carrying a full connection URI. The module composes those Secrets at
deploy time from the referenced credential Secrets — the password
never appears in this manifest, in rendered chart values, or in a pod
argument. The same discipline covers every key the chart would
otherwise regenerate on each upgrade render (logging out every
session): the Fernet key, API session key, FAB webserver key and JWT
signing secret are module-generated once and stable, exported by
NAME in the outputs. Back up the Fernet key Secret — losing it
orphans every credential Airflow has stored.

## Executors

`executor` defaults to the KubernetesExecutor — every task is its own
ephemeral pod, no broker, no idle workers (the chart's own default is
CeleryExecutor; declare it explicitly for the standing-fleet path and
add `broker`). Airflow 3 also takes a comma-separated multi-executor
list and custom executor class paths (the AWS Batch/ECS and Edge
executors). KEDA scales Celery workers on real queue depth — compose
a `KubernetesKeda` and enable `components.workers.keda`.

## DAG delivery

`dags.git_sync` is the production path: a git-sync sidecar on every
component keeps pipelines current — a merged PR deploys itself.
Private repos ride a credentials Secret (HTTPS token) or SSH key
Secret with pinned `known_hosts`; only Secret NAMES render. The
alternatives: a shared ReadWriteMany volume (`dags.persistence`) or
DAGs baked into a custom image (the default when `dags` is empty).

## Exposure

The API server Service stays ClusterIP; expose it with first-class
kinds (`KubernetesService` for a LoadBalancer, Gateway API kinds for
routes) over the exported `api_server_service` handle. Nothing in
this kind does ingress.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
