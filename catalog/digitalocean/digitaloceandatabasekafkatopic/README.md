# DigitalOcean Database Kafka Topic

Built for 100% parity with the Terraform DigitalOcean provider's `digitalocean_database_kafka_topic` resource at the pinned provider version.

## What this component models

A topic on a DigitalOcean managed Kafka cluster, with the complete per-topic configuration block -- partitions, replication, cleanup/compaction policy, retention, segment tuning, and message-format controls.

The component covers the provider's full argument surface:

- `cluster` -- the owning Kafka cluster, wired by reference (or a literal cluster UUID)
- `topic_name` -- the topic's name (create-only: renaming replaces the topic and drops its messages)
- `partition_count` -- 3 to 2048; partitions can only be ADDED after creation (Kafka semantics, API-enforced)
- `replication_factor` -- at least 2; ceiling is the cluster's node count (API-enforced)
- `config` -- all 23 per-topic tunables the provider models, leaf for leaf

## Quick start

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseKafkaTopic
metadata:
  name: orders-events
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: my-kafka-cluster
      fieldPath: status.outputs.cluster_id
  topicName: orders-events
```

Deploy with either provisioner; both produce identical resources and outputs.

## Outputs

| Output | Description |
|---|---|
| `cluster_id` | UUID of the Kafka cluster the topic lives in |
| `topic_name` | The topic's name (its API identity within the cluster) |

The topic's provisioning state is not an output: creation is asynchronous and a state captured at apply time goes stale, so read it live from DigitalOcean (the control panel's Topics tab, `doctl databases topics get`, or `GET /v2/databases/{cluster_id}/topics/{name}`).

## Behavior worth knowing

- **Partitions only grow.** Raising `partition_count` applies in place; lowering it is rejected by the API. DigitalOcean also never reports the live count back (changes apply asynchronously), so the configured value is authoritative.
- **The config block's numeric tunables are 64-bit.** The spec models them as real integers; the modules render them to the strings the provider's wire format requires.
- **An empty config block is not a no-op.** When any config is present, the provider seeds `cleanup_policy` to `compact_delete` unless set explicitly -- set it deliberately.
- **`retention_bytes` / `retention_ms` accept -1** for unlimited; every other numeric tunable is unsigned.
- **`min_insync_replicas` is locally defaulted (to 1), never read from the server** -- leaving it unset always writes 1, even if the server was tuned out-of-band.

## Module layout

- `iac/tf/` -- OpenTofu/Terraform module (provider pinned `~> 2.99`)
- `iac/pulumi/` -- Pulumi module (Go, pulumi-digitalocean SDK)
- Both engines wire the same spec fields and export the same outputs; behavioral parity is the contract.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
