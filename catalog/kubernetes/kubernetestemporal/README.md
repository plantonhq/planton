# Kubernetes Temporal

## When NOT to Use This

**One resource is ONE Temporal cluster** — the durable workflow engine
(long-running business logic, human-in-the-loop flows, saga
orchestration, AI-agent pipelines) from the official `temporal` Helm
chart, on a database you bring.

Not the right kind when:

- **You want managed Temporal** — that is Temporal Cloud; point your
  workers at it and deploy nothing here.
- **You want to declare the workflows it runs** — workflows are CODE
  registered by your workers through the SDKs. This kind installs the
  engine; `temporal_namespaces` declares the logical namespaces
  workflows run in, never the workflows themselves.
- **Container-step pipelines, not code-first durability** — DAGs of
  containers (CI, data pipelines) are `KubernetesArgoWorkflows`
  territory.
- **You have no database** — nothing is bundled, deliberately. Declare
  a PostgreSQL (a `KubernetesPostgres` composes by reference — the
  recommended path), a MySQL 8 (`KubernetesMysql`), or an external
  Cassandra you operate yourself.

## The two-database contract

Temporal keeps workflow state in the default store (`temporal`) and
its search index in the visibility store (`temporal_visibility`). Both
must exist before install unless `create_databases` is set (which
needs create-database privileges). On a `KubernetesPostgres`: declare
`temporal` at bootstrap and add the visibility database with one line
of `post_init_sql`, both owned by the application user — the schema
Jobs handle everything inside them. Cassandra serves the DEFAULT store
only; a SQL `visibility` block is required with it (and enforced on
the spec).

## The credential path

The database password rides the chart's `existingSecret` contract: the
server and schema-Job pods read it through a secretKeyRef and it never
lands in rendered values. The reference defaults compose a
`KubernetesPostgres`'s operator-maintained application Secret (key
`password`), which survives failovers. One Kubernetes constraint to
respect: a secretKeyRef cannot cross namespaces — co-locate Temporal
with its database or replicate the Secret.

## The one immutable number

`num_history_shards` (default 512) is baked into the default store's
schema at FIRST install and cannot be changed without a full cluster
migration. Pick for the cluster you will grow into.

## Exposure

The frontend (gRPC 7233) and Web UI (HTTP 8080) stay ClusterIP;
workers connect in-cluster through the exported `frontend_endpoint`,
and external exposure composes from first-class kinds over the
exported service handles. Nothing in this kind does ingress.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
