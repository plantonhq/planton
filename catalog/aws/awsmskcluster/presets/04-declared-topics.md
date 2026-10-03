# Cluster with Declared Contract Topics

An MSK cluster whose topics deploy WITH it: three contract topics (an event stream, its dead-letter queue, and a compacted snapshot topic) declared in the spec and managed through the MSK topic API — no Kafka client, bootstrap connectivity, or credential setup at deploy time, and no init-container topic scripts in your applications.

## When to Use

- Event-driven architectures where topics are part of the platform contract between services, not application-owned scratch space
- Teams that disable `auto.create.topics.enable` (this preset does) so a typo'd topic name fails loudly instead of materializing an accidental single-replica topic
- Infra charts wiring a cluster plus its consumers: the `topic_arns` output exposes each topic for IAM policy scoping

## Configuration Highlights

- **Topics**: keyed by name — adding or removing one topic never churns the others
  - `orders.events`: 6 partitions, 7-day retention, delete cleanup
  - `orders.events.dlq`: 1 partition, 30-day retention (dead letters live longer for forensics)
  - `orders.snapshots`: compacted (`cleanup.policy: compact`) — latest-value-per-key semantics
- **Durability**: replication factor 3 with `min.insync.replicas: 2` on the cluster and the write-path topics
- **Authentication**: SASL/IAM — grant producers/consumers `kafka-cluster:*Topic*` actions scoped by the exported `topic_arns`

## Lifecycle Truths

- Partition counts can only be INCREASED in place (Kafka's contract); replication factor and name changes replace the topic and its data.
- Deleting a declared topic entry deletes the topic. Topic deletion requires `delete.topic.enable=true` (MSK's default — only relevant if your `serverProperties` overrides it).

## Cost Estimate

The cost drivers are the three kafka.m5.large brokers (billed hourly, the dominant line) plus their EBS storage; the declared topics themselves add no charge beyond the data they retain. The verified figure for this preset lives in the kind's generated estimate at `catalog/_pricing/estimates/awsmskcluster.yaml` — computed from the pinned price book, never hand-typed here.

## Customization

- Point `subnetIds`/`securityGroupIds` at Planton-managed nodes with `valueFrom` for full graph composition
- Add per-environment retention by remixing `configs` values — they are plain Kafka topic configs
