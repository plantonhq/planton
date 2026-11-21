# MirrorMaker 2 Source

## Use Case

Replicate topics from another Kafka cluster with MirrorMaker 2 -- the migration and disaster-recovery pattern. Read both bootstrap addresses from `gcloud managed-kafka clusters describe`; the Connect cluster must resolve the source cluster's DNS domain.

## When to Use

- Migrating topics to a new cluster
- Keeping a standby cluster in step

## What This Creates

- A MirrorMaker source connector replicating every `orders.*` topic from `legacy` into `events`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `projectId`, `connectCluster`, both `bootstrap.servers` | `my-gcp-project`, `migration-connect`, stand-in addresses | Name your project, reference the GcpManagedKafkaConnectCluster (`valueFrom`), and write each cluster's real bootstrap address (read with `gcloud managed-kafka clusters describe`) -- a value inside `configs` takes no reference. |
| `configs.source.cluster.bootstrap.servers` | legacy address | Read it from the source cluster; its format can differ per cluster. |
| `configs.topics` | `orders.*` | A regular expression of topics to replicate. |
