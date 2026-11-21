# MirrorMaker 2 Migration

## Use Case

Workers for moving topics from one Managed Kafka cluster to another with MirrorMaker 2: the target cluster hosts the workers, and the source cluster's DNS domain is made resolvable so the MirrorMaker source connector can reach it.

## When to Use

- Migrating to a new cluster
- Replicating topics across clusters or regions

## What This Creates

- A 6-vCPU, 24 GiB Connect cluster on `events` that can resolve `legacy-events`' private DNS domain

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `projectId`, `kafkaCluster`, `primarySubnet` | `my-gcp-project` paths | Stand-ins: name your project, reference the GcpManagedKafkaCluster and a GcpSubnetwork (`valueFrom`) or write their paths. The Kafka cluster is fixed at creation; the subnet is private, at least a /22, and in this region. |
| `dnsDomainNames` | `legacy-events` domain | The source cluster's bootstrap address without `bootstrap.` and the port. |
| `capacityConfig` | 6 vCPU, 24 GiB | Replication throughput scales with workers. |
