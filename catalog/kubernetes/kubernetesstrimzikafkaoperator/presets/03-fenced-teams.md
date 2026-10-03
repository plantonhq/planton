# Fenced Teams

This preset installs one Strimzi cluster operator with an EXPLICIT
namespace fence: it reconciles Kafka clusters only in the listed team
namespaces (plus its own installation namespace, which is always
watched in addition). The middle ground between the own-namespace
default and the cluster-wide posture — one operator to run, but its
blast radius is enumerated rather than unbounded.

## When to Use

- A shared cluster where SOME namespaces run Kafka and the platform
  team wants the operator's reach spelled out in review-able YAML
- Compliance postures where cluster-wide RBAC for a workload operator
  is not acceptable

## Key Configuration Choices

- **`watch.namespaces: [kafka-team-a, kafka-team-b]`** — renders the
  chart's `watchNamespaces` list. The listed namespaces must already
  exist; adding a team later means updating this resource. Mutually
  exclusive with `any_namespace` (a cluster-wide operator already
  watches everything)
- **`full_reconciliation_interval_ms: 300000`** — the periodic
  full-repair pass every 5 minutes instead of the 2-minute chart
  default: less reconciliation churn across a fleet, at the cost of
  drift repaired more slowly (event-driven reconciliation still fires
  immediately on spec changes)
- **`operation_timeout_ms: 600000`** — 10 minutes instead of the
  5-minute chart default for internal operations like rolling a
  broker pod: on slow storage a legitimate restart can exceed the
  default, and an expired timeout fails the whole reconciliation
- **Explicit resources** — same reasoning as the cluster-wide preset:
  a control plane reconciling several teams' clusters deserves
  headroom over the chart defaults
- **A second fenced operator?** — set `create_global_resources: false`
  on it: the chart's fixed-name ClusterRoles are owned by the first
  release, and a second install conflicts on them otherwise

## Placeholders to Replace

- `kafka-team-a`, `kafka-team-b` — the real team namespaces (they must
  exist before the operator installs)

## Related Kinds

- **KubernetesKafka** — declared inside the fenced namespaces only;
  clusters declared elsewhere are silently never reconciled
- **KubernetesKafkaTopic / KubernetesKafkaUser** — declared in each
  cluster's namespace, reconciled by that cluster's entity operators
