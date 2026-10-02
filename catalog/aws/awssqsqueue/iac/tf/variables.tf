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
  description = "AwsSqsQueue specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Whether to create a FIFO queue. Standard queues are created when false.
    # FIFO queues guarantee exactly-once processing and strict ordering within
    # each message group. This setting cannot be changed after queue creation.
    fifo_queue = optional(bool, false)

    # Time in seconds that a received message is hidden from subsequent receive
    # requests. After the timeout expires the message becomes visible again unless
    # it was deleted. Range: 0–43200 (0s to 12h). AWS default: 30.
    visibility_timeout_seconds = optional(number, 0)

    # Duration in seconds that SQS retains a message. After the retention period
    # expires SQS deletes the message regardless of whether it was consumed.
    # Range: 60–1209600 (1 min to 14 days). AWS default: 345600 (4 days).
    # Leave at 0 to use the AWS default.
    message_retention_seconds = optional(number, 0)

    # Maximum size of a message body in bytes. Messages exceeding this limit are
    # rejected by SQS. Range: 1024–1048576 (1 KB to 1 MB). AWS default: 262144 (256 KB).
    # Leave at 0 to use the AWS default.
    max_message_size_bytes = optional(number, 0)

    # Delay in seconds before a newly sent message becomes visible in the queue.
    # Useful for implementing delayed processing patterns.
    # Range: 0–900 (0s to 15 min). AWS default: 0.
    delay_seconds = optional(number, 0)

    # Wait time in seconds for the ReceiveMessage API call. A value greater than
    # 0 enables long polling, which reduces the number of empty responses and
    # lowers cost. Range: 0–20. AWS default: 0 (short polling).
    receive_wait_time_seconds = optional(number, 0)

    # Enable content-based deduplication for FIFO queues. When enabled SQS uses
    # a SHA-256 hash of the message body as the deduplication ID, removing the
    # need for the producer to supply an explicit deduplication ID.
    # Only valid when `fifo_queue` is true.
    content_based_deduplication = optional(bool, false)

    # Deduplication scope for FIFO queues. Controls whether deduplication is
    # applied per message group or across the entire queue.
    # Valid values: "messageGroup", "queue". Only valid when `fifo_queue` is true.
    deduplication_scope = optional(string, "")

    # Throughput limit for FIFO queues. Controls whether throughput quota applies
    # per message group ID or per queue. Set to "perMessageGroupId" to enable
    # high throughput mode for FIFO queues.
    # Valid values: "perMessageGroupId", "perQueue". Only valid when `fifo_queue` is true.
    fifo_throughput_limit = optional(string, "")

    # Dead letter queue configuration. When a message is received more than
    # `max_receive_count` times without being deleted, SQS moves it to the
    # specified target queue for investigation and reprocessing.
    dead_letter_config = optional(object({
      # ARN of the target dead letter queue. Accepts a direct ARN or a reference
      # to another AwsSqsQueue resource. Both queues must be the same type
      # (both Standard or both FIFO) and reside in the same AWS account and region.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      target_arn = string

      # Number of times a message can be received before being moved to the dead
      # letter queue. Must be at least 1. Common values: 3–5 for transient errors,
      # 1 for poison pill detection. Range: 1–1000.
      max_receive_count = optional(number, 0)
    }))

    # Customer-managed KMS key for server-side encryption. When set SQS encrypts
    # message bodies using this key. Accepts a direct KMS key ID/ARN or a
    # reference to an AwsKmsKey resource. Mutually exclusive with
    # `sqs_managed_sse_enabled`.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # Duration in seconds that SQS reuses a data encryption key before calling
    # KMS again. Higher values reduce KMS costs but increase the window for key
    # reuse. Range: 60–86400 (1 min to 24h). AWS default: 300 (5 min).
    # Only relevant when `kms_key_id` is set.
    kms_data_key_reuse_period_seconds = optional(number, 0)

    # SQS-managed server-side encryption (SSE-SQS). SQS manages the encryption
    # key automatically at no additional cost. AWS enables SSE-SQS on newly
    # created queues by default, so this field is PRESENCE-typed:
    # - unset  — keep AWS's default (new queues come up encrypted with SSE-SQS)
    # - true   — pin SSE-SQS on explicitly
    # - false  — explicitly disable server-side encryption (an unencrypted queue)
    #
    # Setting this field at all (either value) conflicts with `kms_key_id`:
    # the provider rejects the combination on config PRESENCE, not value, so
    # an explicit `false` alongside a KMS key is still an error — leave this
    # field unset when using customer-managed KMS encryption.
    sqs_managed_sse_enabled = optional(bool)

    # IAM access policy for the queue. Controls which AWS principals can perform
    # actions on this queue (e.g., SendMessage, ReceiveMessage). Expressed as a
    # standard IAM policy document structure. Common use cases include granting
    # SNS topics permission to publish to this queue or allowing cross-account
    # access.
    policy = optional(any)

    # Controls which source queues are allowed to use THIS queue as their
    # dead-letter queue. This is the permission side of the dead-letter
    # relationship: `dead_letter_config` on a source queue points at a DLQ,
    # while `redrive_allow_policy` on the DLQ itself restricts who may point
    # at it. When unset, AWS allows all source queues in the account (the
    # "allowAll" behavior). Locking a shared DLQ down with "byQueue" prevents
    # unrelated workloads from silently routing their failures into it.
    redrive_allow_policy = optional(object({
      # The redrive permission mode.
      # Valid values:
      # - "allowAll": any source queue in the same account and region may use this
      #   queue as its DLQ (AWS's default behavior when no policy is set).
      # - "denyAll": no source queue may use this queue as a DLQ. Use this to
      #   protect a queue that must never receive redriven messages.
      # - "byQueue": only the queues listed in `source_queue_arns` may use this
      #   queue as their DLQ. The recommended mode for shared/central DLQs.
      redrive_permission = string

      # ARNs of the source queues permitted to use this queue as their dead-letter
      # queue. Each entry accepts a direct ARN or a reference to another
      # AwsSqsQueue resource. Only valid (and required) when `redrive_permission`
      # is "byQueue". AWS caps the list at 10 source queues; to allow more than
      # 10, use "allowAll" instead.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      source_queue_arns = optional(list(string), [])
    }))
  })
}
