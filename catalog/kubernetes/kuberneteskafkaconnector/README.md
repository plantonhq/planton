# Kubernetes Kafka Connector

## When NOT to Use This

**A Connect cluster must already exist — in the SAME namespace.**
This component declares one connector (a data pipe);
KubernetesKafkaConnect is the worker fleet that runs it. The
placement contract is strict: a KafkaConnector in another namespace,
or naming a Connect cluster that does not exist there, is accepted by
the API server and then silently never reconciled. Set `namespace` to
the Connect cluster's own namespace.

Also not the right kind when:

- **You want the worker fleet itself** — replicas, plugin delivery,
  the Kafka connection: that is KubernetesKafkaConnect; this
  component is one pipe running on it.
- **The connector class is not on the workers** — the class must
  arrive through the Connect cluster's plugin arm (stock image,
  prebuilt `image`, OCI `plugins`, or `build`). A class the workers
  do not carry fails at reconcile with a "class not found" condition.
- **You plan to manage the connector through the Connect REST API** —
  Connect clusters from this catalog carry the
  `strimzi.io/use-connector-resources` annotation: the operator owns
  connectors and reverts REST-API-made changes. This resource IS the
  management surface.
- **You want cluster-to-cluster replication** — MirrorMaker 2 runs
  its own mirror connectors from a dedicated kind:
  KubernetesKafkaMirrorMaker2, not hand-declared connectors here.

## Overview

**KubernetesKafkaConnector** declares one data pipe on the Strimzi
`kafka.strimzi.io/v1` `KafkaConnector` custom resource. The target
Connect cluster's operator-managed workers run it: a SOURCE connector
streams an external system into Kafka (Debezium CDC from
Postgres/MySQL, file tails, SaaS APIs); a SINK connector streams
Kafka topics out (object stores, search indexes, warehouses). The
`connector_class` decides which one it is.

**The binding contract**: `connect_cluster` (a literal name or a
KubernetesKafkaConnect reference) renders as the
`strimzi.io/cluster` label — how the operator matches the connector
to its cluster, together with the shared namespace.

**Key design points:**

- **`config` is the connector's own key set** — the exact keys its
  documentation defines (connection URLs, table filters, topics,
  converters), values written as strings ("5432", "false").
- **Secrets never ride the config as literals** — Connect supports
  configuration providers: reference values as
  `${secrets:<namespace>/<secret>:<key>}` after enabling the
  KubernetesSecretConfigProvider through the Connect cluster's
  `config` (`config.providers` entries). Database passwords and API
  keys stay in Kubernetes Secrets.
- **`state` is declarative** — `running` (default), `paused` (retains
  tasks, stops moving data), or `stopped` (deallocates tasks — the
  state offset overrides require).
- **`auto_restart` is the production posture** — automatic restarts
  with incremental back-off (default cap 7 attempts); without it a
  transient source-system outage leaves the connector FAILED until a
  human intervenes.
- **Offsets are declared targets, annotated verbs** — `list_offsets`
  names the ConfigMap offsets are written TO when the resource
  carries the `strimzi.io/connector-offsets: list` annotation;
  `alter_offsets` names the ConfigMap offsets are applied FROM under
  the `alter` annotation while the connector is stopped — the replay
  / skip-poison-record / migration-cutover mechanism.
- **`version` replaces the retired config entry** — when workers
  carry several versions of a class, pin the desired one here;
  `connector.plugin.version` in config is rejected at the spec.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: the Connect cluster's OWN namespace (the
  placement contract) — literal or a KubernetesNamespace reference
- **`spec.connect_cluster`**: the Connect cluster to run on — literal
  name or a KubernetesKafkaConnect reference (resolves to its
  `connect_name` output)
- **`spec.connector_class`**: fully-qualified class, e.g.
  `io.debezium.connector.postgresql.PostgresConnector`

### Common

- **`spec.config`**: the connector's documented key set; secrets via
  `${secrets:...}` config-provider references
- **`spec.tasks_max`**: parallel-task ceiling (empty = the Connect
  default, 1); real parallelism is also bounded by the connector's
  own semantics
- **`spec.state`**: `running` / `paused` / `stopped`
- **`spec.auto_restart`**: enable + optional `max_restarts`
- **`spec.list_offsets` / `spec.alter_offsets`**: ConfigMap targets
  for the annotation-triggered offset verbs
- **`spec.version`**: plugin version pin when workers carry several

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the KafkaConnector lives in (the Connect cluster's namespace) |
| `connector_name` | The connector's name inside the Connect cluster (`metadata.name`) — what the Connect REST API and consumer-group names (`connect-<name>`) key off |

## Composing in Infra Charts

- **`spec.connect_cluster`** is a foreign key (default kind
  KubernetesKafkaConnect, field path `status.outputs.connect_name`);
  **`spec.namespace`** references a KubernetesNamespace — set it to
  the Connect cluster's namespace; the offset ConfigMap fields
  reference KubernetesConfigMap resources.
- **Source-system credentials compose through Secrets**: keep the
  database password in a Kubernetes Secret (declared via
  KubernetesSecret or an ExternalSecret) and reference it as
  `${secrets:<namespace>/<secret>:<key>}` — after enabling the
  secrets config provider on the KubernetesKafkaConnect resource.
- **The class arrives through the cluster's plugin arm** — pair a
  Debezium connector with a Connect cluster whose `image`, `plugins`
  or `build` carries the Debezium plugin.

## Examples

### First pipe (stock image, zero plugin machinery)

The stock image carries only the MirrorMaker 2 connector classes, so
the zero-machinery first pipe is a MirrorSource self-mirror — records
produced to `orders` appear on `src.orders`:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafkaConnector
metadata:
  name: first-pipe
spec:
  namespace:
    value: dev-kafka
  connect_cluster:
    value: dev-connect
  connector_class: org.apache.kafka.connect.mirror.MirrorSourceConnector
  tasks_max: 1
  config:
    source.cluster.alias: src
    target.cluster.alias: dev
    source.cluster.bootstrap.servers: dev-kafka-kafka-bootstrap.dev-kafka.svc.cluster.local:9092
    target.cluster.bootstrap.servers: dev-kafka-kafka-bootstrap.dev-kafka.svc.cluster.local:9092
    topics: orders
    replication.factor: "1"
    offset-syncs.topic.replication.factor: "1"
```

### Debezium Postgres CDC (config-provider secret reference)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafkaConnector
metadata:
  name: orders-cdc
spec:
  namespace:
    value: kafka
  connect_cluster:
    value: cdc-connect
  connector_class: io.debezium.connector.postgresql.PostgresConnector
  tasks_max: 1
  config:
    database.hostname: orders-db.postgres.svc.cluster.local
    database.port: "5432"
    database.user: debezium
    database.password: ${secrets:kafka/orders-db-credentials:password}
    database.dbname: orders
    topic.prefix: orders
    plugin.name: pgoutput
    table.include.list: public.orders,public.order_items
  auto_restart:
    enabled: true
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
