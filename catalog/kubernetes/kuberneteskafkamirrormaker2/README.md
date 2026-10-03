# Kubernetes Kafka MirrorMaker 2

## When NOT to Use This

**The operator must already be on the cluster.** This component
declares a MirrorMaker 2 deployment; KubernetesStrimziKafkaOperator
installs the ENGINE that reconciles it, and must watch this
namespace.

Also not the right kind when:

- **You want a general-purpose Connect cluster** — MirrorMaker 2 IS
  Connect under the hood, but it is a dedicated replication engine
  with its own kind; arbitrary connectors belong on
  KubernetesKafkaConnect + KubernetesKafkaConnector.
- **You want to mirror INTO several clusters from one resource** —
  the Strimzi model on this line is ONE target per resource (the old
  multi-cluster list is gone upstream): one KafkaMirrorMaker2 per
  target cluster, each with as many sources as needed.
- **You expect application traffic to move by itself** — MirrorMaker
  2 replicates topics, records, topic configuration and consumer
  offsets; repointing producers and consumers at the target cluster
  is your cutover step (the checkpoint connector makes it seamless,
  not automatic).
- **You need a Strimzi MM2 surface the spec deliberately leaves
  out** — tracing, remote JMX, custom logging, per-connector pause
  states (see the research doc). Reachable by declaring the raw
  custom resource through KubernetesManifest.

## Overview

**KubernetesKafkaMirrorMaker2** declares a MirrorMaker 2 replication
engine on the Strimzi `kafka.strimzi.io/v1` `KafkaMirrorMaker2`
custom resource — continuous, offset-aware mirroring of topics and
consumer groups from one or more SOURCE Kafka clusters into one
TARGET cluster. This is the migration on-ramp: point a mirror at a
running Confluent / MSK / self-hosted cluster, let it replicate
topics, records and consumer positions into your Strimzi-managed
cluster, then cut consumers over with their offsets intact.

**The shape**: one `target` (where mirrored data lands and where the
engine keeps its Connect-style state) plus one `mirrors` entry per
source cluster. Under the hood each mirror runs a
MirrorSourceConnector (records + topic configuration) and a
MirrorCheckpointConnector (consumer-group offset translation — what
makes consumer cutover seamless).

**Alias semantics**: every cluster carries an `alias` naming it in
the replication flow. Source aliases must be unique and must differ
from the target's alias (default `target`) — both spec-validated.
Under the DEFAULT replication policy, mirrored topics arrive PREFIXED
with the source alias (`prod-msk.orders`); for migrations you usually
want ORIGINAL names — set `replication.policy.class` to
`org.apache.kafka.connect.mirror.IdentityReplicationPolicy` on BOTH
the source and checkpoint connectors of a mirror.

**Key design points:**

- **The connection blocks are the shared Strimzi client shapes** —
  target and every source read identically: bootstrap address
  (literal or a KubernetesKafka reference), TLS trust (a CA Secret —
  a KubernetesKafka's cluster CA by reference), and authentication
  (mutual TLS, scram-sha-512/-256, plain, custom) with credentials
  from Secrets, never the manifest. For migrations the source
  bootstrap is usually an EXTERNAL address.
- **Group identity defaults from metadata.name** — the engine's
  group.id and three storage topics on the TARGET
  (`<name>-mirrormaker2-{configs,status,offsets}`) must be unique
  among Connect-protocol workloads (Connect clusters included)
  sharing the target cluster.
- **Scope is pattern-based** — `topics_pattern` / `groups_pattern`
  (default `.*`) with exclude patterns; the engine's built-in
  exclusions already skip internal topics.
- **Storage replication factors must fit the target** — the engine's
  internal-topic default of 3 wedges on a single-broker dev target;
  set the three `*.storage.replication.factor` entries to `"-1"`
  (broker default) there via `target.config`, `"3"` in production.
- **`auto_restart` per connector** — long-running migrations survive
  transient source outages instead of parking FAILED.
- **Metrics are module-owned** — `metrics.enabled` renders the
  canonical Strimzi rules ConfigMap (`<name>-mm2-metrics`) and wires
  it as `metricsConfig`; mirroring lag rides the standard
  Connect/MirrorMaker metric families.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace for the deployment — literal or a
  KubernetesNamespace reference; the operator must watch it
- **`spec.target`**: the target cluster — `bootstrap_servers`
  (literal or KubernetesKafka reference — the migrate-INTO-Planton
  wiring), optional `tls`/`authentication`, optional group-identity
  overrides, `config` for engine client tuning and the storage
  replication factors
- **`spec.mirrors`**: at least one — each with `source` (alias,
  bootstrap, tls, authentication, client config), scope patterns, and
  per-mirror `source_connector` / `checkpoint_connector` tuning
  (tasks_max, config, auto_restart)

### Common

- **`spec.replicas`**: worker count (default 1) — scale with source
  partition counts; `source_connector.tasks_max` bounds per-mirror
  parallelism
- **`spec.version`**: Kafka version the workers run; empty = the
  operator's default
- **`spec.resources` / `spec.jvm`**: worker sizing (set xms = xmx in
  production)
- **`spec.rack` / `spec.metrics` / `spec.node_selector` /
  `spec.tolerations`**

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the deployment runs in |
| `mirrormaker_name` | The deployment's name (`metadata.name`) |
| `rest_api_endpoint` | In-cluster Connect REST endpoint of the underlying engine (`http://<name>-mirrormaker2-api.<namespace>.svc.cluster.local:8083`) — read-only inspection of mirror connector status |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace); **`target.bootstrap_servers`** references a
  KubernetesKafka (`status.outputs.internal_bootstrap_endpoint`);
  TLS **`trusted_certificates[].secret_name`** references a
  KubernetesKafka (`status.outputs.cluster_ca_cert_secret_name`);
  authentication Secret names reference a KubernetesKafkaUser
  (`status.outputs.secret_name`).
- **Migration wiring**: the target is the KubernetesKafka sibling by
  reference; sources are literal external bootstraps (Confluent, MSK)
  with their credentials in Secrets you declare (KubernetesSecret /
  ExternalSecret).
- **Verify replication before cutover**: the `rest_api_endpoint`
  exposes mirror connector status; consumer-group offset translation
  needs `checkpoint_connector` running (with
  `sync.group.offsets.enabled: "true"` where you want automatic
  group-offset sync).

## Examples

### Migrate from an external cluster (original topic names)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafkaMirrorMaker2
metadata:
  name: msk-migration
spec:
  namespace:
    value: kafka
  target:
    bootstrap_servers:
      value: events-kafka-bootstrap.kafka.svc.cluster.local:9092
    config:
      config.storage.replication.factor: "3"
      offset.storage.replication.factor: "3"
      status.storage.replication.factor: "3"
  mirrors:
    - source:
        alias: prod-msk
        bootstrap_servers:
          value: b-1.prod-msk.abc123.kafka.us-west-2.amazonaws.com:9096
        tls:
          trusted_certificates:
            - secret_name:
                value: prod-msk-ca
              pattern: "*.crt"
        authentication:
          type: scram-sha-512
          username: mm2-reader
          password_secret:
            secret_name:
              value: prod-msk-credentials
            password: password
      topics_pattern: "orders.*,payments.*"
      groups_pattern: ".*"
      source_connector:
        tasks_max: 8
        config:
          replication.policy.class: org.apache.kafka.connect.mirror.IdentityReplicationPolicy
        auto_restart:
          enabled: true
      checkpoint_connector:
        config:
          replication.policy.class: org.apache.kafka.connect.mirror.IdentityReplicationPolicy
          sync.group.offsets.enabled: "true"
        auto_restart:
          enabled: true
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
