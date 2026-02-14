# Production Kafka Cluster

This preset deploys a 3-broker, 3-ZooKeeper Kafka cluster with production-grade resources, replication, and the Kafka UI. Provides fault tolerance and horizontal throughput scaling.

## When to Use

- Production event streaming and message processing
- High-throughput workloads requiring partitioned topics with replication
- Environments where broker or ZooKeeper node failures must not cause data loss

## Key Configuration Choices

- **3 brokers** with 50Gi disk each -- tolerates 1 broker failure; 150Gi total message storage
- **3 ZooKeeper nodes** -- required for Raft consensus and leader election; tolerates 1 ZK failure
- **Topic with 6 partitions, 3 replicas** -- enables parallel consumption with full replication for durability
- **Higher resources** -- production-appropriate for sustained throughput

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |

## Related Presets

- **01-single-broker** -- Minimal single-broker Kafka for development
- **03-with-schema-registry** -- Adds Schema Registry for schema management
