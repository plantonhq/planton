# Pulumi Module to Deploy AliCloudRocketmqInstance

This Pulumi program provisions an Alibaba Cloud RocketMQ 5.x instance with
bundled topics and consumer groups. It targets the modern `rocketmq.*` Pulumi
resources (not the legacy `ons.*` types).

## Resources Created

- `rocketmq.RocketMQInstance` — managed RocketMQ 5.x instance with VPC networking
- `rocketmq.RocketMQTopic` × N — one per entry in `spec.topics[]`
- `rocketmq.ConsumerGroup` × M — one per entry in `spec.consumerGroups[]`

## CLI Usage (OpenMCF Pulumi)

```bash
# Preview
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Update (apply)
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh
openmcf pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Destroy
openmcf pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

## Module Structure

| File | Purpose |
|------|---------|
| `main.go` | Pulumi program entrypoint; loads stack input and delegates to module |
| `Pulumi.yaml` | Project configuration |
| `module/main.go` | Orchestrates instance, topics, and consumer groups; exports outputs |
| `module/locals.go` | Tag computation, defaults (`instanceName`, `paymentType`, `commodityCode`, `internetSpec`, `flowOutType`) |
| `module/outputs.go` | Output constant definitions |
| `module/topics.go` | Creates `rocketmq.RocketMQTopic` resources as children of the instance |
| `module/consumer_groups.go` | Creates `rocketmq.ConsumerGroup` resources with retry policies |

## How It Works

1. The entrypoint (`main.go`) loads the `AliCloudRocketmqInstanceStackInput` from Pulumi config
2. `locals.go` computes tags, resolves the instance name, and derives hidden provider fields
3. `main.go` creates the AliCloud provider (scoped to `spec.region`), then the RocketMQ instance
4. Topics and consumer groups are created in loops, each as a child of the instance resource
5. Stack outputs are exported: instance ID, VPC TCP endpoint, internet endpoint, topic ID map, consumer group ID map

## Debugging

Edit `Pulumi.yaml` and uncomment the `runtime.options.binary` line to run through
a debug helper:

```yaml
name: alicloud-rocketmq-instance
runtime:
  name: go
#  options:
#    binary: ./debug.sh
```

For more details, see the [Pulumi debugging guide](https://github.com/plantonhq/openmcf).
