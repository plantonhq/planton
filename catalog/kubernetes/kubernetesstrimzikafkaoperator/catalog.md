# Strimzi Kafka Operator

Installs the Strimzi cluster operator — the CNCF project that runs Apache Kafka on Kubernetes — from the official `strimzi-kafka-operator` Helm chart. The operator is the ENGINE: it reconciles `Kafka` custom resources (declared with Apache Kafka) into KRaft-mode Kafka clusters, and its per-cluster entity operators reconcile `KafkaTopic` / `KafkaUser` resources into real topics and authenticated users. This component installs and configures the operator itself; Kafka clusters, topics, and users are declared as their own first-class Infra Components.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** — created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Strimzi Helm Release** — the official `strimzi-kafka-operator` chart from the Strimzi chart repository at the pinned `chartVersion`
- **Cluster Operator Deployment** — the Strimzi cluster operator pod(s); extra replicas are leader-elected warm standbys for the operator itself, not added reconciliation throughput
- **Strimzi CRDs** — Kafka, KafkaNodePool, KafkaTopic, KafkaUser, and the rest, shipped in the chart's Helm-native `crds/` directory: installed on first install, never upgraded or deleted by Helm (uninstalling the release never cascade-deletes Kafka clusters)
- **RBAC** — the ClusterRoles, ServiceAccounts, and bindings the operator needs (`createGlobalResources` governs the cluster-scoped set), plus per-namespace RoleBindings for a fenced watch list

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Cluster-admin permissions** — installing the CRDs and ClusterRoles requires them.
- **Watched namespaces must already exist** — with a `watch.namespaces` fence, the chart templates RoleBindings into each listed namespace and the install fails with `namespaces ... not found` otherwise. Create them first (e.g. as Kubernetes Namespace resources the install depends on).
- **A storage class** for persistent volumes — needed by the Kafka clusters this operator will manage (KRaft node pools with persistent storage), not by the operator pod itself.

## Deploy

### Console

Open the deployment store, find **Strimzi Kafka Operator**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Default** preset in the [Presets](#presets) tab for the standard posture.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesStrimziKafkaOperator
metadata:
  name: strimzi
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "kafka"
  createNamespace: true
  chartVersion: "1.1.0"
```

```shell
planton apply -f strimzi-operator.yaml
```

This installs the operator watching its OWN namespace only (the chart default) — Kafka clusters declared in the `kafka` namespace are reconciled; clusters elsewhere are invisible to it. The operator creates no Kafka clusters by itself: declare them with Apache Kafka resources. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the operator to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: kafka-home
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then provisions the operator into it.

## Key Configuration

These are the most important decisions when configuring the operator. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Watch scope — the defining choice.** By default the operator watches ITS OWN namespace only: Kafka clusters live beside their operator, and a `Kafka` resource declared anywhere else is silently never reconciled. Set `watch.anyNamespace` for one platform-owned operator managing clusters everywhere, or `watch.namespaces` for an explicit team fence (the installation namespace is always watched in addition). The two are mutually exclusive — a cluster-wide operator already watches everything.

**CRD lifecycle on upgrades.** The chart ships its CRDs in the Helm-native `crds/` directory — installed on first install, never touched again by Helm. A `chartVersion` upgrade therefore runs new operator code against the EXISTING CRDs; when a Strimzi release's notes call for it (minor upgrades regularly add CRD fields), apply the new CRDs yourself.

**Second operator in the same cluster.** The chart's fixed-name ClusterRoles are owned by the first release. A second install (watching a disjoint namespace set) must set `createGlobalResources: false` or Helm fails trying to create them again.

**Image sourcing for air-gapped clusters.** The `image` block overrides the registry/organization for the operator AND every operand image it deploys (Kafka, entity operators, Cruise Control, exporter). Empty = `quay.io/strimzi` at the chart version.

**Everything else** — reconciliation cadence (`fullReconciliationIntervalMs`), operation timeout (raise on slow storage where broker restarts legitimately exceed 5 minutes), log level, feature gates (the operator's own `+Gate,-Gate` syntax; an unknown gate fails startup), NetworkPolicy/PodDisruptionBudget generation, and the `helmValues` escape hatch (merged last, Helm `-f` semantics) for the chart surface beyond the typed fields.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| Kubernetes Namespace | `spec.namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the operator runs in | Placing Kafka clusters beside their operator (the default watch posture) |
| `release_name` | Helm release name of the operator | Identifying the installation in day-2 tooling |

The operator has no per-cluster surface of its own — Kafka clusters compose against the CRDs it installs, so the outputs identify the installation rather than any workload.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard install** — the pinned chart in its own-namespace posture; Kafka clusters live beside the operator. Start from the **Default** preset.

**Platform-team operator** — one operator in a dedicated control-plane namespace managing Kafka clusters in every namespace. Start from the **Cluster-Wide** preset.

**Fenced teams** — one operator reconciling only an explicit list of team namespaces. Start from the **Fenced Teams** preset (the listed namespaces must already exist).

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — provides the namespace for the operator deployment
- [**Apache Kafka**](/infra-catalog/kubernetes-kafka) — the Kafka clusters this operator reconciles
- [**Kafka Topic**](/infra-catalog/kubernetes-kafka-topic) — topics reconciled by each cluster's entity operator
- [**Kafka User**](/infra-catalog/kubernetes-kafka-user) — authenticated users reconciled by each cluster's entity operator
