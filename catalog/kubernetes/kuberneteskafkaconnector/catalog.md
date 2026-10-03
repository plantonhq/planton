# Kafka Connector

Declares ONE data pipe — a connector instance — on a Strimzi `KafkaConnector` custom resource. The target Connect cluster's operator-managed workers run it: a SOURCE connector streams an external system into Kafka (Debezium CDC from Postgres/MySQL, file tails, SaaS APIs); a SINK connector streams Kafka topics out (object stores, search indexes, warehouses). The placement contract is strict: the KafkaConnector must live in the SAME NAMESPACE as its Connect cluster and bind to that cluster through the `strimzi.io/cluster` label (rendered from `connectCluster`). A connector in another namespace, or naming a Connect cluster that does not exist there, is accepted by the API server and then silently never reconciled — no pipe, no error. The connector's CLASS must be available on the Connect cluster's workers — via the stock image, a prebuilt `image`, OCI `plugins`, or a `build` on the KubernetesKafkaConnect resource. A class the workers do not carry fails at reconcile with a "class not found" condition on the resource.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **KafkaConnector** — the Strimzi custom resource, named after this resource, placed in the Connect cluster's own namespace and bound to the cluster through the `strimzi.io/cluster` label (rendered from `connectCluster`)
- **Connector instance** (reconciled by the Connect cluster's operator, not the module) — the running pipe on the Connect workers, keyed by `metadata.name` inside the cluster. Consumer-group names for sinks follow `connect-<name>`

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- A **Kafka Connect** cluster whose workers carry the connector class you declare (the stock Strimzi Connect image ships ONLY the three MirrorMaker 2 connector classes — MirrorSource, MirrorCheckpoint, MirrorHeartbeat; anything else arrives through the Connect cluster's `image`, `plugins`, or `build` arms).
- The namespace you declare must be the Connect cluster's OWN namespace — connectors declared anywhere else are silently never reconciled.
- For secrets in connector config (`database.password`, API keys), enable the **KubernetesSecretConfigProvider** through the Connect cluster's worker `config` (`config.providers` entries) and grant the Connect service account read access to the referenced Secrets.

## Deploy

### Console

Open the deployment store, find **Kafka Connector**, and click **Deploy**. The creation wizard walks you through the Connect cluster + namespace placement pair, the connector class and task parallelism, the connector's own configuration map (with secrets-by-reference teaching), lifecycle state and auto-restart, and offset ConfigMap targets. Start from the **First-pipe preset (stock-image MirrorSource)** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafkaConnector
metadata:
  name: orders-cdc
  org: acme-corp
  env: prod
spec:
  namespace:
    value: kafka # MUST be the Connect cluster's own namespace
  connectCluster:
    value: cdc-connect
  connectorClass: io.debezium.connector.postgresql.PostgresConnector
  tasksMax: 1
  config:
    database.hostname: orders-db.postgres.svc.cluster.local
    database.port: "5432"
    database.user: debezium
    database.password: ${secrets:kafka/orders-db-credentials:password}
    database.dbname: orders
    topic.prefix: orders
    plugin.name: pgoutput
    slot.name: debezium_orders
    table.include.list: public.orders,public.order_items
  autoRestart:
    enabled: true
    maxRestarts: 10
```

```shell
planton apply -f kafka-connector.yaml
```

This declares a Debezium Postgres CDC source: the workers resolve the database password from a Kubernetes Secret at connector start — it never lands in this resource, in IaC state, or in kubectl output. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire both placement facts from the Connect cluster itself so the pair can never drift apart:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesKafkaConnect
      name: cdc-connect
      fieldPath: status.outputs.namespace
  connectCluster:
    valueFrom:
      kind: KubernetesKafkaConnect
      name: cdc-connect
      fieldPath: status.outputs.connect_name
```

The InfraPipeline deploys the Connect cluster first, then declares the connector against it.

## Key Configuration

These are the most important decisions when declaring a connector. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The placement contract** — the KafkaConnector must live in the Connect cluster's OWN namespace. `connectCluster` renders as the `strimzi.io/cluster` label that binds the connector to its cluster. Reference the KubernetesKafkaConnect resource to inherit its `connect_name` output and draw the dependency edge.

**The class decides source vs sink** — a fully-qualified Java class name (e.g. `io.debezium.connector.postgresql.PostgresConnector`). The class must be ON the Connect cluster's workers. The stock image carries ONLY the MirrorMaker 2 connectors; Debezium, S3 sinks, Elasticsearch sinks, and warehouse connectors arrive through the Connect cluster's `image`, `plugins`, or `build` arms.

**Config values are Connect configuration strings** — write numbers and booleans as strings (`"5432"`, `"false"`). The key set belongs to each connector class's own documentation. Do NOT put `connector.plugin.version` in config — it is deprecated upstream; set plugin version on the `version` field instead. Secrets should be config-provider references (`${secrets:<namespace>/<secret>:<key>}`), never literals.

**Task parallelism has two ceilings** — `tasksMax` caps how many parallel tasks the connector may run (empty = Connect default of 1). Real parallelism is also bounded by the connector's own semantics (many CDC connectors run one task per table set regardless). Pin `version` when the workers carry several versions of the class.

**Lifecycle state is the day-2 pause lever** — `running` (default), `paused` (tasks stay allocated, no data moves — cheap to resume), or `stopped` (tasks deallocated — REQUIRED before offsets can be altered). Blank means absent = running.

**Auto-restart is strongly recommended for production** — transient source-system outages otherwise leave the connector FAILED until a human intervenes. `maxRestarts` caps consecutive restarts (empty = operator default of 7 attempts, back-off capped at ~30 minutes).

**Offsets are declared targets, annotated verbs** — `listOffsets.toConfigMap` names where current offsets are written when the resource carries `strimzi.io/connector-offsets: list`. `alterOffsets.fromConfigMap` names where new offsets are read from when the resource carries `strimzi.io/connector-offsets: alter` while the connector is stopped — the replay, skip-poison-record, and migration-cutover mechanism.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesKafkaConnect** | `connectCluster` | `status.outputs.connect_name` |
| **KubernetesConfigMap** | `listOffsets.toConfigMap` | `status.outputs.configmap_name` |
| **KubernetesConfigMap** | `alterOffsets.fromConfigMap` | `status.outputs.configmap_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the KafkaConnector lives in (the Connect cluster's namespace) | Placing related resources beside the Connect cluster |
| `connector_name` | The connector's name inside the Connect cluster (`metadata.name`) — what the Connect REST API and consumer-group names (`connect-<name>`) key off | Monitoring, offset tooling, cross-referencing Connect REST status |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**First-pipe Mirror Source** — the smallest real pipe that runs on the stock Connect image: a MirrorSourceConnector mirroring a topic from the cluster back to itself under an alias. Start from the **First-pipe preset (stock-image MirrorSource)**.

**Debezium Postgres CDC** — a production-shaped change-data-capture source with a config-provider database password and table filters. Pair with a Connect cluster whose workers carry the Debezium Postgres plugin. Start from the **Debezium Postgres CDC preset**.

**Paused pipe with offsets** — operational lifecycle shape: connector declared `paused` with both offset ConfigMap targets wired for day-2 offset surgery. Start from the **Paused pipe with offsets preset**.

## Works With

- [**Kafka Connect**](/infra-catalog/kubernetes-kafka-connect) — the Connect cluster whose workers run this pipe; its `image`, `plugins`, or `build` arms deliver connector classes the stock image lacks
- [**Apache Kafka**](/infra-catalog/kubernetes-kafka) — the event bus this pipe reads from or writes to; bootstrap endpoints come from the Kafka cluster's own outputs
- [**Kafka Topic**](/infra-catalog/kubernetes-kafka-topic) — declare the topics a source emits into or a sink consumes from as code
- [**Kubernetes ConfigMap**](/infra-catalog/kubernetes-config-map) — offset listing and override targets referenced by `listOffsets` and `alterOffsets`
