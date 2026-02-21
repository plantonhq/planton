# Terraform Module to Deploy AliCloudRocketmqInstance

This module provisions an Alibaba Cloud RocketMQ 5.x instance with bundled
topics and consumer groups using the `alicloud_rocketmq_*` resources.

## Resources Created

- `alicloud_rocketmq_instance.main` — managed RocketMQ 5.x instance
- `alicloud_rocketmq_topic.topics` — one per topic (via `for_each`)
- `alicloud_rocketmq_consumer_group.consumer_groups` — one per consumer group (via `for_each`)

## Usage

Use the OpenMCF CLI (tofu) with the default local backend:

```shell
openmcf tofu init --manifest hack/manifest.yaml
openmcf tofu plan --manifest hack/manifest.yaml
openmcf tofu apply --manifest hack/manifest.yaml --auto-approve
openmcf tofu destroy --manifest hack/manifest.yaml --auto-approve
```

Credentials are provided via stack input (CLI), not in the manifest `spec`.

## Module Structure

| File | Purpose |
|------|---------|
| `main.tf` | RocketMQ instance resource with dynamic `product_info` block |
| `topics.tf` | Topic resources via `for_each` over `local.topics_map` |
| `consumer_groups.tf` | Consumer group resources with retry policies via `for_each` |
| `locals.tf` | Tag computation, `commodity_code` derivation, internet logic, collection-to-map conversions |
| `variables.tf` | Input variables with validation rules |
| `outputs.tf` | Instance ID, TCP/internet endpoints, topic and consumer group ID maps |
| `provider.tf` | AliCloud provider configuration scoped to `spec.region` |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `instance_id` | `string` | RocketMQ instance ID |
| `tcp_endpoint` | `string` | VPC-internal TCP endpoint |
| `internet_endpoint` | `string` | Public internet TCP endpoint (empty if disabled) |
| `topic_ids` | `map(string)` | Topic name to resource ID mapping |
| `consumer_group_ids` | `map(string)` | Consumer group ID to resource ID mapping |

For more examples, see [`examples.md`](./examples.md) and [`hack/manifest.yaml`](../hack/manifest.yaml).
