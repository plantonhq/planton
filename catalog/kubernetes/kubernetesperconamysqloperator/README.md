# Kubernetes Percona MySQL Operator

## When NOT to Use This

**This component installs the ENGINE, not a database.** It deploys the
Percona Operator for MySQL (based on Percona XtraDB Cluster) — the
controller that reconciles `PerconaXtraDBCluster` custom resources into
running Galera clusters. To get an actual MySQL database, deploy this
first, then declare a KubernetesMysql.

Also not the right kind when:

- **You want a MySQL database** — that is KubernetesMysql; this component
  is the operator it requires.
- **You want a managed cloud database** — use the host cloud's managed
  MySQL kinds; this operator is for running MySQL ON the cluster.
- **You need a second widened-watch operator on the same cluster** — the
  operator's CR-validation webhook is a single fixed-name cluster-scoped
  object, so one cluster carries at most ONE operator with a widened
  watch scope (own-namespace installs can coexist freely).

## What It Deploys

One Helm release of the official `pxc-operator` chart (pinned 1.20.0 —
chart and operator versions move together), named after `metadata.name`.
The release renders the operator Deployment, its ServiceAccount, and
watch-scoped RBAC.

### Watch scope decides almost everything

- **Default (no `watch` block):** the operator watches its OWN namespace
  only, with a namespaced Role. KubernetesMysql databases must be
  declared in that same namespace.
- **`watch.cluster_wide: true`:** one operator reconciles databases in
  every namespace, with a ClusterRole.
- **`watch.namespaces: [...]`:** an explicit namespace fence (the listed
  namespaces must already exist).

The two widened arms are mutually exclusive with each other (validated at
the spec), and they change the RBAC grain the chart renders — which in
turn changes the validation-webhook posture below.

### CRD lifecycle

The chart ships the `PerconaXtraDBCluster`, `PerconaXtraDBClusterBackup`,
and `PerconaXtraDBClusterRestore` CRDs in its Helm-native `crds/`
directory: installed on first install, never upgraded or deleted by Helm.
Uninstalling the operator therefore NEVER cascade-deletes the database
clusters — they simply stop being reconciled until an operator returns.
Operator upgrades via `chart_version` run new operator code against the
existing CRDs.

### The validation webhook is module-owned (widened watch)

With cluster-scoped RBAC, the upstream operator registers ONE fixed-name,
cluster-scoped `ValidatingWebhookConfiguration`
(`percona-xtradbcluster-webhook`, failurePolicy Fail) pointing at a
Service in its own namespace — and nothing upstream ever removes it: the
object cannot ride the Deployment's ownerReference (Kubernetes never
garbage-collects a cluster-scoped dependent of a namespaced owner), and
an operator that finds it already present refreshes only the CA bundle,
never the service pointer. Left to the operator, uninstalling it strands
a Fail-closed webhook whose service no longer exists — bricking every
future `PerconaXtraDBCluster` admission in the cluster.

This module therefore renders the webhook itself in the widened-watch
arms: created before the operator (whose startup registration then merely
refreshes the CA bundle), deleted with the resource. Destroying the
operator leaves the cluster clean. Own-namespace installs render nothing
— there the chart's namespaced Role denies the registration and the
webhook never exists, the upstream posture.

## Configuration Surface

Control-plane sizing (`replicas`, `resources`), reconcile concurrency
(`max_concurrent_reconciles`, `s3_workers_limit`), logging
(`log.structured`, `log.level`), telemetry opt-out (`disable_telemetry`),
leader-election timing, the XtraBackup-sidecar feature gate, scheduling
(`node_selector`, `tolerations`), private-registry images (`image`,
`image_pull_secrets`), and the `helm_values` escape hatch (merged last,
Helm `-f` semantics) on top of the typed surface.

## Outputs

| Output | Meaning |
|---|---|
| `namespace` | Namespace the operator runs in |
| `release_name` | Helm release name (`metadata.name`) |

Databases compose against the CRDs the operator installs — declare a
[KubernetesMysql](../kubernetesmysql/README.md) with its namespace
inside the operator's watch scope.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
