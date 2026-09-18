variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsKinesisStreamConsumer specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # ARN of the Kinesis Data Stream to register this consumer with. The consumer
    # receives a dedicated 2 MB/s read throughput pipe for every shard in the
    # stream, independent of other consumers.
    #
    # ForceNew: changing the stream ARN forces consumer replacement (deregister
    # + re-register). A consumer can only be registered with one stream at a time.
    #
    # Accepts a direct ARN string or a reference to an AwsKinesisStream resource
    # via valueFrom.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    stream_arn = string

    # Resource-based access policy for the consumer, as a standard IAM policy
    # document. The primary use is cross-account enhanced fan-out: granting
    # another account's principals SubscribeToShard/DescribeStreamConsumer on
    # this consumer without role assumption. AWS models this as a separate
    # resource-policy API keyed by the consumer ARN; it is folded here because
    # the policy has no identity of its own and follows the consumer's
    # lifecycle. (The stream itself carries its own resource_policy for
    # stream-level grants.)
    resource_policy = optional(any)
  })
}
