# Kubernetes Percona Mongo Operator

## When NOT to Use This

**This component installs the ENGINE, not a database.** It deploys the
Percona Operator for MongoDB — the controller that reconciles
`PerconaServerMongoDB` custom resources into running replica sets and
sharded clusters. To get an actual MongoDB database, deploy this first,
then declare a KubernetesMongodb.

Also not the right kind when:

- **You want a MongoDB database** — that is KubernetesMongodb; this
  component is the operator it requires.
- **You want managed MongoDB** — use AtlasMongodb or the host cloud's
  managed document-store kinds; this operator is for running MongoDB ON
  the cluster.

## What It Deploys

One Helm release of the official `psmdb-operator` chart (pinned 1.22.0 —
chart and operator versions move together), named after `metadata.name`.
The release renders the operator Deployment, its ServiceAccount, and
watch-scoped RBAC.

### Watch scope decides where databases can live

- **Default (no `watch` block):** the operator watches its OWN namespace
  only, with a namespaced Role. KubernetesMongodb databases must be
  declared in that same namespace.
- **`watch.cluster_wide: true`:** one operator reconciles databases in
  every namespace, with a ClusterRole.
- **`watch.namespaces: [...]`:** an explicit namespace fence (the listed
  namespaces must already exist).

The two widened arms are mutually exclusive (validated at the spec).

### CRD lifecycle

The chart ships the `PerconaServerMongoDB`,
`PerconaServerMongoDBBackup`, and `PerconaServerMongoDBRestore` CRDs in
its Helm-native `crds/` directory: installed on first install, never
upgraded or deleted by Helm. Uninstalling the operator therefore NEVER
cascade-deletes the database clusters — they simply stop being
reconciled until an operator returns. Operator upgrades via
`chart_version` run new operator code against the existing CRDs.

## Configuration Surface

Control-plane sizing (`replicas`, `resources`), reconcile concurrency
(`max_concurrent_reconciles`), logging (`log.structured`, `log.level`),
telemetry opt-out (`disable_telemetry`), scheduling (`node_selector`,
`tolerations`), private-registry images (`image`, `image_pull_secrets`),
and the `helm_values` escape hatch (merged last, Helm `-f` semantics) on
top of the typed surface.

## Outputs

| Output | Meaning |
|---|---|
| `namespace` | Namespace the operator runs in |
| `release_name` | Helm release name (`metadata.name`) |

Databases compose against the CRDs the operator installs — declare a
[KubernetesMongodb](../kubernetesmongodb/README.md) with its namespace
inside the operator's watch scope.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
