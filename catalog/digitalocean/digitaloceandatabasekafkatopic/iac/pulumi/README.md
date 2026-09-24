# Pulumi Module: DigitalOcean Database Kafka Topic

Provisions a topic on a DigitalOcean managed Kafka cluster -- the complete `digitalocean_database_kafka_topic` resource surface including the full config block. Behavioral parity with the Terraform module is the contract.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean.DatabaseKafkaTopic` | The topic: identity, partitions, replication, and the 23-leaf config block |

## Inputs

`DigitalOceanDatabaseKafkaTopicStackInput`: the target `DigitalOceanDatabaseKafkaTopic` resource and the DigitalOcean provider config (API token).

## Outputs

Exactly the `DigitalOceanDatabaseKafkaTopicStackOutputs` contract: `cluster_id`, `topic_name`. The SDK's `State` property is deliberately not exported (an apply-time snapshot of an asynchronous create goes stale; live state belongs to whoever reads the API).

## Behavior notes

- The bridge exposes the provider's config block as a `Configs` ARRAY (the upstream schema declares an unbounded list); this module always sends exactly one element.
- The 64-bit numeric tunables are `*string` in the SDK; present spec values are rendered with `strconv`, absent leaves stay nil and are never sent.
- All spec leaves are presence-gated -- proto zero values never reach the API.
- `partition_count` is never read back by the provider (async application) -- the configured value is authoritative.
