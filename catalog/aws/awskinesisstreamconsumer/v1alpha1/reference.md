# AwsKinesisStreamConsumer

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsKinesisStreamConsumerSpec defines the desired configuration for an Amazon
Kinesis Data Stream enhanced fan-out consumer.

An enhanced fan-out consumer (also known as a registered consumer) provides
a dedicated 2 MB/s read throughput pipe from a Kinesis stream, independent
of all other consumers. This contrasts with standard consumers (GetRecords
API), which share the shard's 2 MB/s read capacity across all readers.

Enhanced fan-out uses the SubscribeToShard API with HTTP/2 push delivery,
achieving ~70ms propagation delay compared to ~200ms for standard polling.

Key characteristics:
- Each consumer gets dedicated 2 MB/s per shard (not shared).
- Push-based delivery via HTTP/2 (SubscribeToShard).
- Up to 20 consumers per stream (soft limit, can be increased).
- Consumer name is derived from metadata.name (ForceNew — cannot be renamed).
- Changing the stream_arn forces consumer replacement (ForceNew).
- No configuration beyond the stream reference — AWS manages all internals.

Common use cases:
- Multiple independent applications reading from the same stream without
  contention (e.g., analytics pipeline + real-time dashboard + audit trail).
- Lambda event source mappings with enhanced fan-out for low-latency triggers.
- High-throughput consumers that need guaranteed 2 MB/s per shard.

Notes:
- The consumer name (from metadata.name) is immutable after creation.
- Credentials, region, and deployment workflow live outside this spec in
  IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsKinesisStreamConsumer
metadata:
  name: test-consumer
spec:
  region: us-west-2
  streamArn:
    value: arn:aws:kinesis:us-west-2:123456789012:stream/test-stream
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.streamArn` | `string \| valueFrom` | yes |  | AwsKinesisStream (`status.outputs.stream_arn`) |
| `spec.resourcePolicy` | `object` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.streamArn

`string | valueFrom` · required

ARN of the Kinesis Data Stream to register this consumer with. The consumer
receives a dedicated 2 MB/s read throughput pipe for every shard in the
stream, independent of other consumers.

ForceNew: changing the stream ARN forces consumer replacement (deregister
+ re-register). A consumer can only be registered with one stream at a time.

Accepts a direct ARN string or a reference to an AwsKinesisStream resource
via valueFrom.

- references: AwsKinesisStream (`status.outputs.stream_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKinesisStream, name: <that resource's name>, fieldPath: status.outputs.stream_arn}} -- a bare string does not parse

### spec.resourcePolicy

`object`

Resource-based access policy for the consumer, as a standard IAM policy
document. The primary use is cross-account enhanced fan-out: granting
another account's principals SubscribeToShard/DescribeStreamConsumer on
this consumer without role assumption. AWS models this as a separate
resource-policy API keyed by the consumer ARN; it is folded here because
the policy has no identity of its own and follows the consumer's
lifecycle. (The stream itself carries its own resource_policy for
stream-level grants.)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsKinesisStreamConsumer, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.consumer_arn` | `string` | The Amazon Resource Name (ARN) of the registered stream consumer. This is the primary identifier used for Lambda event source mappings configured with enhanced fan-out (StartingPosition + ConsumerARN), IAM policies, and API calls (SubscribeToShard, DescribeStreamConsumer). Format: arn:aws:kinesis:{region}:{account}:stream/{stream-name}/consumer/{consumer-name}:{creation-timestamp} |
| `status.outputs.consumer_name` | `string` | The name of the registered stream consumer. Matches metadata.name from the resource spec. Used for human-readable identification and as the consumer name in Kinesis API calls (ListStreamConsumers filtering). |
| `status.outputs.stream_arn` | `string` | The ARN of the parent Kinesis Data Stream this consumer is registered with. Echoed back for convenience — enables downstream resources to discover the stream without a separate lookup. |
| `status.outputs.creation_timestamp` | `string` | RFC3339 timestamp of when the consumer was registered with the stream. Useful for operational visibility and debugging registration order. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.streamArn` | AwsKinesisStream | `status.outputs.stream_arn` |

## See Also

- [Overview](../README.md)
