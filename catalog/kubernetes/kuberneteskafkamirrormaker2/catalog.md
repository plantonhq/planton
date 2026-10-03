# Kafka MirrorMaker 2

Deploys a MirrorMaker 2 replication engine on Kubernetes as a Strimzi `KafkaMirrorMaker2` custom resource — continuous, offset-aware mirroring of topics, records, and consumer-group positions from one or more source Kafka clusters into one target. This is the migration on-ramp: point the workers at a running MSK, Confluent Cloud, or datacenter cluster, replicate into your Strimzi-managed Kafka, then cut consumers over with their offsets intact (checkpointing translates source offsets to target offsets). Each `mirrors` entry is one source; the target connection and the engine's group identity live once on the resource.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Strimzi `KafkaMirrorMaker2` resource** — the engine declaration (target connection, one entry per source, per-mirror connector tuning), reconciled by the watching Strimzi operator into Connect-protocol mirror workers. Under the hood each mirror runs a MirrorSourceConnector (records + topic configuration) and a MirrorCheckpointConnector (consumer-group offset translation).
- **Kubernetes Namespace** — created only when `createNamespace` is `true`; the natural home is the target Kafka cluster's own namespace. A Strimzi operator must watch it or the resource is accepted and silently never reconciled.
- **JMX Prometheus metrics ConfigMap** — created only when `metrics.enabled` is `true`; the canonical Strimzi Connect rule set wired as the deployment's `metricsConfig` — mirroring lag rides the standard Connect/MirrorMaker metric families.
- **Internal Connect storage topics** — derived on the target cluster by the engine itself (config / status / offset); default names derive from `metadata.name` and must stay unique among Connect-protocol workloads sharing the target.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster. The connection decides WHICH cluster the workers run in; the namespace is the placement unit inside it.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Strimzi Kafka Operator watching the workers' namespace** — the declared prerequisite; without it the resource never reconciles.
- **A target Kafka cluster** — usually an Apache Kafka resource whose bootstrap endpoint and cluster CA resolve from its outputs.
- **Network reachability to every source bootstrap address** — for migrations the sources are usually EXTERNAL (MSK, Confluent Cloud, a datacenter cluster); the workers must reach them from inside the cluster.
- **Matching TLS trust and credentials per connection** — target and each source carry their own `tls` / `authentication` blocks, and each must match what that cluster's listener accepts.

## Deploy

### Console

Open the deployment store, find **Kafka MirrorMaker 2**, and click **Deploy**. The creation wizard walks you through where the workers run, the target connection, the engine identity, the mirrors repeater (the star step), worker count, sizing, scheduling, and metrics. Start from the **Migrate from MSK preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafkaMirrorMaker2
metadata:
  name: msk-migration
  org: acme-corp
  env: prod
spec:
  namespace:
    value: kafka
  createNamespace: false
  target:
    bootstrapServers:
      valueFrom:
        kind: KubernetesKafka
        name: event-bus
        fieldPath: status.outputs.internal_bootstrap_endpoint
  mirrors:
    - source:
        alias: msk
        bootstrapServers:
          value: b-1.mycluster.abc123.c2.kafka.us-east-1.amazonaws.com:9094
      topicsPattern: ".*"
```

```shell
planton apply -f msk-migration.yaml
```

This starts continuous mirroring of every topic (and, by default, every consumer group's offsets) from the MSK cluster into the `event-bus` cluster; watch mirror connector status through the exported REST endpoint and cut consumers over only after checkpoint lag is acceptable. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the target bootstrap and trust from the sibling Kafka cluster so the migration graph cannot drift:

```yaml
spec:
  namespace:
    value: kafka
  target:
    bootstrapServers:
      valueFrom:
        kind: KubernetesKafka
        name: event-bus
        fieldPath: status.outputs.internal_bootstrap_endpoint
    tls:
      trustedCertificates:
        - secretName:
            valueFrom:
              kind: KubernetesKafka
              name: event-bus
              fieldPath: status.outputs.cluster_ca_cert_secret_name
          certificate: ca.crt
  mirrors:
    - source:
        alias: msk
        bootstrapServers:
          value: b-1.mycluster.abc123.c2.kafka.us-east-1.amazonaws.com:9094
```

The InfraPipeline deploys the target Kafka cluster first, then starts the mirror against its resolved endpoint and CA.

## Key Configuration

These are the most important decisions when configuring a MirrorMaker 2 deployment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Replication policy is the cutover decision.** By default, mirrored topics land alias-prefixed (`msk.orders`) — safe for coexistence, wrong for a clean migration where consumers expect original names. Set `replication.policy.class` to `org.apache.kafka.connect.mirror.IdentityReplicationPolicy` in the mirror's `sourceConnector.config` AND `checkpointConnector.config` (they must stay in lockstep) to keep original names. Choose before mirroring starts: switching policy after topics exist strands the already-mirrored copies under the old names.

**Aliases are identities, not labels.** Each source `alias` must be unique and must differ from the target's alias (which defaults to `target` when empty — the spec enforces this). Under the default policy the alias prefixes every mirrored topic name permanently, so pick one you can live with in topic lists.

**Group identity must be unique on the target.** `target.groupId` and the three storage topics default from `metadata.name` and share the Connect protocol with Kafka Connect clusters — a collision with any Connect-protocol workload on the same target corrupts both engines' state.

**Checkpointing is what makes cutover safe.** `groupsPattern` (default `.*`) decides which consumer groups' offsets are translated into the target. Narrow it deliberately — a group that is not checkpointed reprocesses or skips data at cutover. The alternative people reach for (snapshot-and-replay) loses consumer positions entirely; preserving them is this component's reason to exist.

**Match each connection to its listener.** Target and every source carry independent `tls` / `authentication` blocks: Confluent Cloud sources use `plain` with the API key/secret, MSK SCRAM sources use `scram-sha-512`, a Strimzi-managed target trusts its cluster CA by referencing the Apache Kafka resource. A mismatched type fails at connect time, not at apply time.

**Auto-restart for anything long-running.** Migrations run for days; without `autoRestart` on the mirror connectors, a transient source outage leaves the mirror FAILED until a human intervenes (operator default when enabled: 7 attempts, back-off capped at ~30 minutes). Scale `tasksMax` on the source connector with the source's partition volume, and set worker `resources` plus equal `jvm.xms`/`jvm.xmx` in production.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesKafka** | `target.bootstrapServers` (and `mirrors[].source.bootstrapServers` for in-cluster sources) | `status.outputs.internal_bootstrap_endpoint` |
| **KubernetesKafka** | `target.tls.trustedCertificates[].secretName` (and per-source) | `status.outputs.cluster_ca_cert_secret_name` |
| **KubernetesKafkaUser** | `authentication.certificateAndKey.secretName` / `authentication.passwordSecret.secretName` (target and per-source) | `status.outputs.secret_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the mirror workers run in | Placing related resources beside the engine |
| `mirrormaker_name` | The deployment's name (`metadata.name`) | Monitoring and cross-referencing the engine's Connect-protocol identity |
| `rest_api_endpoint` | In-cluster Connect REST endpoint (`http://<name>-mirrormaker2-api.<namespace>.svc.cluster.local:8083`) | Read-only inspection of mirror connector status during cutover |

## Common Patterns

**Migrate off MSK** — source = the running MSK cluster (SCRAM over TLS), target = the Apache Kafka destination, identity replication policy so topics keep their names, checkpointing on for offset-intact consumer cutover. Start from the **Migrate from MSK preset**.

**Migrate off Confluent Cloud** — the same posture with `plain` authentication carrying the Confluent API key/secret. Start from the **Migrate from Confluent Cloud preset**.

**Active-passive DR** — a permanent mirror into a standby cluster: alias-prefixed names (both clusters stay live, so identity policy would collide), auto-restart on, metrics enabled so replication lag is visible before you need the standby. Start from the **Active-passive DR preset**.

## Works With

- [**Apache Kafka**](/infra-catalog/kubernetes-kafka) — the usual target; its outputs resolve the bootstrap endpoint and cluster CA
- [**Strimzi Kafka Operator**](/infra-catalog/kubernetes-strimzi-kafka-operator) — the declared prerequisite: it must watch the workers' namespace
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — the placement unit, naturally shared with the target cluster
- [**Kafka Connect**](/infra-catalog/kubernetes-kafka-connect) — general-purpose Connect for non-mirror integrations; shares the Connect protocol, so group identities must not collide
- [**Kafka User**](/infra-catalog/kubernetes-kafka-user) — the authenticated principal for a Strimzi-managed target or source
- [**Kafka UI**](/infra-catalog/kubernetes-kafka-ui) — observe topics and consumer lag on both sides during cutover
