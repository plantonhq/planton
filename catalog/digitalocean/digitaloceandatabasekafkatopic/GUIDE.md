# DigitalOcean Database Kafka Topic -- Operational Guide

What experience with this kind teaches that the field reference cannot.

## Partitions are a one-way door

Kafka only ever ADDS partitions: raising `partition_count` applies in place, lowering it is an API error, and the only way down is destroying the topic (and its messages). Adding partitions also reshuffles key-to-partition mapping for keyed producers -- consumers relying on key ordering see keys move. Size partitions for the topic's mature throughput on day one; grow only deliberately.

## The live partition count is invisible

DigitalOcean applies partition changes asynchronously and the provider deliberately never reads the count back -- drift detection cannot see a partition change made out-of-band, and an imported topic lands at the schema default. Treat the manifest as the single source of truth for partitions.

## Renaming is deletion

`topic_name` is the topic's identity: changing it destroys the topic -- messages, consumer offsets, everything -- and creates an empty one. Treat renames as migrations with a consumer cutover plan, never as edits.

## Set cleanup_policy explicitly when you touch config

The provider seeds `cleanup_policy: compact_delete` the moment ANY config block is present, which is not Kafka's own `delete` default. If you add a config block just to tune retention, state the cleanup policy you actually want alongside it.

## min_insync_replicas pairs with your producers' acks

The floor for durable writes is `min_insync_replicas: 2` WITH producers sending `acks=all` -- either alone protects nothing. And note the provider never reads this leaf back from the server (it defaults it locally to 1): an out-of-band change is invisible to drift detection.

## Retention: -1 means forever, and forever fills disks

`retention_ms: -1` / `retention_bytes: -1` disable time/size limits. On a compacted topic that is the normal shape (compaction bounds growth per key); on a `delete` topic it grows without bound against the CLUSTER's paid storage. Set a real retention on every delete-policy topic.

## What is deliberately NOT here

Consumer groups, ACLs, and credentials (they live on the cluster's users -- the DigitalOceanDatabaseUser kind's `kafka_acls` pair topics with permissions); cluster-level Kafka settings (the cluster kind's engine config); and the schema registry (its own kind, DigitalOceanDatabaseKafkaSchema).
