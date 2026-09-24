# GcpManagedKafkaTopic Guide

The judgment this guide protects: a topic's partition count and replication are promises to every producer and consumer, and two of them can never be taken back -- replication is fixed at creation and partitions only grow. Decide them for the traffic you expect in a year, not the traffic you have today.

## Partitions

Partitions are Kafka's unit of parallelism: a consumer group reads each partition with one consumer, so the partition count caps how many consumers work in parallel. `partitionCount` grows in place but never shrinks. Growing it re-hashes keys, so messages with the same key land on a different partition after the change -- per-key ordering holds within the old messages and within the new ones, not across the boundary.

## Replication

`replicationFactor` is how many brokers keep a copy of each partition. Google runs brokers in three zones and recommends 3, which survives a zone outage. It is immutable: changing it replaces the topic and deletes its messages.

## Topic configuration

`configs` overrides the cluster's defaults for this topic, keyed by Kafka property. The common ones are `retention.ms` (how long messages are kept), `cleanup.policy: compact` (keep the latest value per key instead of expiring by time -- the changelog pattern), `min.compaction.lag.ms`, and `compression.type`. Retention and replication together decide how much storage the topic holds, and storage is billed.

## Access

Creating a topic does not let anyone use it. Grant producers WRITE and consumers READ with a `GcpManagedKafkaAcl` on `topic/{name}` -- and READ on the consumer group with a second ACL on `consumerGroup/{name}`.
