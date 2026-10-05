# KubernetesStatefulSet Guide

The judgment this guide carries: before hand-rolling a stateful system on
this kind, check whether the catalog already runs that system for you —
the most common StatefulSet mistake on this platform is rebuilding what an
operator-backed kind does better.

## Check the catalog before building

For well-known stateful systems, the catalog carries first-class,
operator-backed kinds: KubernetesPostgres, KubernetesMysql,
KubernetesMongodb, KubernetesKafka, KubernetesClickHouse,
KubernetesOpenSearch, KubernetesRabbitMq, KubernetesValkey and more (the
[provider index](../reference-index.md) is the authoritative list, and
the [catalog guide](../../_docs/GUIDE.md) covers compatible substitutes).
Those kinds encode replication, failover, backups and upgrades that a
plain StatefulSet leaves entirely to you. Reach for KubernetesStatefulSet
when the stateful application is your own — or genuinely absent from the
catalog.

## The storage class edge is a name, not a reference

`volumeClaimTemplates[].storageClass` is a plain string. Unlike a
`valueFrom` reference, naming a class created by a KubernetesStorageClass
resource creates no dependency edge: nothing orders the two deploys and
nothing draws the relationship on the diagram. Keep the StorageClass in
the shared-cluster chart, deployed before any application environment
that names it.

## Namespace ownership

`spec.namespace` is a required foreign key targeting KubernetesNamespace.
`createNamespace: true` makes this workload the namespace's owner in IaC
state — safe only when it is the sole tenant. Any shared namespace wants a
dedicated KubernetesNamespace component instead; the failure story and the
`valueFrom` wiring: [namespace-ownership pattern](../../_patterns/namespace-ownership.md).

## On the diagram

An operator-backed database renders as a first-class node other resources
reference into (connection endpoints, credential secrets); a hand-rolled
StatefulSet renders as a workload whose stateful internals — volume
claims, per-replica identity — are invisible to the graph. When the
system matters to the architecture, prefer the kind that shows it.

## Pairs well with

- KubernetesNamespace — the namespace owner (pattern above).
- KubernetesStorageClass — pins volume performance characteristics
  (deploy-ordering caveat above).
