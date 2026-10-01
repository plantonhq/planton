# Terraform Module: DigitalOcean Database Kafka Topic

Provisions a topic on a DigitalOcean managed Kafka cluster -- the complete `digitalocean_database_kafka_topic` resource surface including the full config block.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean_database_kafka_topic.topic` | The topic: identity, partitions, replication, and the 23-leaf config block |

## Inputs

Generated `variables.tf` mirrors the `DigitalOceanDatabaseKafkaTopicSpec` proto: `cluster` (flattened reference string), `topic_name`, optional `partition_count` / `replication_factor`, and the optional `config` object. Authentication uses `digitalocean_token` (sensitive).

## Outputs

Exactly the `DigitalOceanDatabaseKafkaTopicStackOutputs` contract: `cluster_id`, `topic_name`. The provider's `state` attribute is deliberately not exported (an apply-time snapshot of an asynchronous create goes stale; live state belongs to whoever reads the API).

## Behavior notes

- The config block's 64-bit numeric tunables are rendered to strings in `locals.tf` (the provider carries them as strings because Terraform numbers are not 64-bit safe); absent leaves stay null and are never sent.
- Enum-walled string leaves normalize `""` to null so the provider's value validation never sees an empty string.
- When a config block is present the provider seeds `cleanup_policy` to `compact_delete` unless set explicitly.
- `partition_count` is never read back by the provider (async application) -- the configured value is authoritative.
- Import: `terraform import ... <cluster_id>,<topic_name>` (see `iac/import-map.yaml`).
