# Kubernetes Strimzi Kafka Operator

## When NOT to Use This

**This component installs the ENGINE, not a Kafka cluster.** It deploys
the Strimzi cluster operator — the CNCF project that runs Apache Kafka
on Kubernetes — the controller that reconciles `Kafka` and
`KafkaNodePool` custom resources into KRaft-mode Kafka clusters. To get
an actual Kafka cluster, deploy this first, then declare a
KubernetesKafka.

Also not the right kind when:

- **You want a Kafka cluster** — that is KubernetesKafka; this
  component is the operator it requires.
- **You want topics or users** — those are KubernetesKafkaTopic and
  KubernetesKafkaUser, reconciled by the per-cluster entity operators
  that each KubernetesKafka deploys.
- **You want a managed cloud Kafka** — use the host cloud's managed
  streaming kinds; this operator is for running Kafka ON the cluster.
- **You expect ZooKeeper-based clusters** — Strimzi removed ZooKeeper
  support in 0.46; this operator line is KRaft-only (Kafka's built-in
  Raft metadata quorum). Clusters still on ZooKeeper must migrate
  before they can be managed here.
- **You need a second operator without RBAC coordination** — a second
  install in the same cluster (watching a disjoint namespace set) must
  set `create_global_resources: false`: the chart's fixed-name
  ClusterRoles are owned by the first release, and a second install
  conflicts on them.

## What It Deploys

One Helm release of the official `strimzi-kafka-operator` chart
(pinned 1.1.0 — chart and operator versions move together; the SERVED
index at https://strimzi.io/charts/ governs the version, since the
Chart.yaml inside the Strimzi source tree carries a build-time
placeholder), named after `metadata.name`. The release renders the
operator Deployment, its ServiceAccount, watch-scoped RBAC, and the
Strimzi CRDs.

### Watch scope decides the deployment topology

- **Default (no `watch` block):** the operator watches its OWN
  namespace only (the chart default — Kafka clusters live beside their
  operator). KubernetesKafka clusters must be declared in that same
  namespace.
- **`watch.any_namespace: true`:** one operator reconciles Kafka
  clusters in every namespace, with cluster-wide RBAC.
- **`watch.namespaces: [...]`:** an explicit namespace fence (the
  installation namespace is always watched in addition).

The two widened arms are mutually exclusive (validated at the spec) —
a cluster-wide operator already watches everything.

### CRD lifecycle

The chart ships the Strimzi CRDs (Kafka, KafkaNodePool, KafkaTopic,
KafkaUser, and the rest of the family) in its Helm-native `crds/`
directory: installed on first install, never upgraded or deleted by
Helm. Uninstalling the operator therefore NEVER cascade-deletes the
Kafka clusters — they simply stop being reconciled until an operator
returns. The same posture cuts the other way on upgrades: a
`chart_version` upgrade runs new operator code against the EXISTING
CRDs — apply the new release's CRDs yourself when an upgrade's release
notes call for it (Strimzi minor releases regularly add CRD fields).

The operator registers no admission webhooks and creates no
cluster-scoped singletons at runtime — uninstall leaves nothing
stranded beyond the deliberately-kept CRDs.

## Configuration Surface

Control-plane sizing (`replicas` — extra replicas are leader-elected
warm standbys, not reconciliation throughput; `resources`),
reconciliation timing (`full_reconciliation_interval_ms`,
`operation_timeout_ms`), logging (`log_level`), Strimzi `feature_gates`,
the cluster DNS domain (`kubernetes_service_dns_domain`),
leader election (`leader_election_enabled`), operand policy generation
(`generate_network_policy`, `generate_pod_disruption_budget`),
cluster-scoped RBAC ownership (`create_global_resources`), scheduling
(`node_selector`, `tolerations`), private-registry images (`image` —
which steers the operator AND every operand image it deploys — and
`image_pull_secrets`), and the `helm_values` escape hatch (merged last,
Helm `-f` semantics) on top of the typed surface.

## Outputs

| Output | Meaning |
|---|---|
| `namespace` | Namespace the operator runs in |
| `release_name` | Helm release name (`metadata.name`) |

Kafka clusters compose against the CRDs the operator installs — declare
a [KubernetesKafka](../kuberneteskafka/README.md) with its namespace
inside the operator's watch scope.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
