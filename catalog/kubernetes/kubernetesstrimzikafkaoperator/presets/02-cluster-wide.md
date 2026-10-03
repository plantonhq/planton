# Cluster-Wide

This preset installs ONE Strimzi cluster operator that manages Kafka
clusters in EVERY namespace — the platform-team posture: the operator
lives in a dedicated control-plane namespace (`strimzi-system`), and
application teams declare KubernetesKafka clusters wherever their
workloads live.

## When to Use

- A shared cluster where multiple teams run their own Kafka clusters
  in their own namespaces
- You want exactly one operator to upgrade, monitor, and size —
  instead of one per namespace

## Key Configuration Choices

- **`watch.any_namespace: true`** — renders the chart's
  `watchAnyNamespace` with cluster-wide RBAC. The trade-off against
  the default: one operator version now governs every Kafka cluster
  on the Kubernetes cluster, so operator upgrades are a
  platform-level event, not a per-team one
- **A dedicated `strimzi-system` namespace** — the operator no longer
  needs to live beside any Kafka cluster, so it gets a control-plane
  namespace of its own
- **Explicit resources** — a control plane reconciling many clusters
  deserves headroom above the chart defaults (requests 200m/384Mi,
  limits 1000m/384Mi); memory is the axis that matters as the number
  of managed clusters grows
- **Everything else at chart defaults** — reconciliation every
  2 minutes, operation timeout 5 minutes, one replica. Raise
  `replicas` to 2 for a leader-elected warm standby if operator
  downtime during node failures is a concern (extra replicas add
  failover, never throughput)

## Placeholders to Replace

None — this preset deploys as-is.

## Related Kinds

- **KubernetesKafka** — declared in ANY namespace; this operator
  reconciles them all
- **KubernetesKafkaTopic / KubernetesKafkaUser** — declared in each
  cluster's namespace, reconciled by that cluster's entity operators
