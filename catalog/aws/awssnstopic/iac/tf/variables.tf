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
  description = "AwsSnsTopic specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Whether to create a FIFO topic. Standard topics are created when false.
    # FIFO topics guarantee strict ordering and exactly-once delivery to SQS FIFO
    # queue subscribers. This setting cannot be changed after topic creation.
    fifo_topic = optional(bool, false)

    # Enable content-based deduplication for FIFO topics. When enabled, SNS uses
    # a SHA-256 hash of the message body as the deduplication ID, removing the
    # need for the publisher to supply an explicit deduplication ID.
    # Only valid when `fifo_topic` is true.
    content_based_deduplication = optional(bool, false)

    # Throughput scope for FIFO topics. Controls whether the throughput quota
    # applies per topic or per message group. "MessageGroup" enables high
    # throughput mode (each message group gets its own quota); it pairs with
    # SQS FIFO subscribers configured with `fifo_throughput_limit:
    # "perMessageGroupId"` for an end-to-end high-throughput FIFO pipeline.
    # Valid values: "Topic", "MessageGroup". Only valid when `fifo_topic` is true.
    fifo_throughput_scope = optional(string, "")

    # Message archive policy for FIFO topics. When set, SNS retains published
    # messages for the configured window so subscriptions can replay them (each
    # AwsSnsSubscription opts into replay via its own `replay_policy`). Expressed
    # as the SNS archive policy JSON document, e.g.
    # {"MessageRetentionPeriod": 30} for a 30-day archive. Only valid when
    # `fifo_topic` is true — AWS does not archive standard topics. The
    # `beginning_archive_time` output reports when the archive became active.
    archive_policy = optional(any)

    # Human-readable display name for the topic. Used as the "from" label in SMS
    # messages and as a readable identifier in the AWS console. Maximum 256
    # characters for Standard topics, 10 characters for SMS display names.
    display_name = optional(string, "")

    # Customer-managed KMS key for server-side encryption. When set, SNS encrypts
    # message bodies using this key. Accepts a direct KMS key ID/ARN or a
    # reference to an AwsKmsKey resource. When not set, SNS does not encrypt
    # messages at rest (unlike SQS, SNS has no "managed SSE" option — encryption
    # requires an explicit KMS key).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # IAM access policy for the topic. Controls which AWS principals can perform
    # actions on this topic (e.g., Publish, Subscribe). Expressed as a standard
    # IAM policy document structure. Common use cases include granting EventBridge
    # permission to publish, allowing cross-account subscriptions, or restricting
    # publishing to specific IAM roles. Note: an SNS topic always carries a
    # policy — when this field is unset AWS applies its default owner-only
    # policy, and removing a previously set policy reverts to that default
    # rather than leaving the topic policy-less.
    policy = optional(any)

    # Data protection policy for the topic. Detects and optionally audits,
    # masks, or blocks sensitive data (PII/PHI such as names, addresses, card
    # numbers) flowing through the topic. Expressed as the SNS data protection
    # policy JSON document (Name/Description/Version/Statement with
    # DataIdentifier selectors and Audit/Deidentify/Deny operations). Only
    # supported on standard topics — AWS rejects data protection policies on
    # FIFO topics.
    data_protection_policy = optional(any)

    # HTTP/HTTPS delivery retry policy for the topic. Expressed as a JSON string
    # matching the SNS delivery policy format. Controls retry backoff, max retries,
    # and throttle behavior for HTTP/S subscriptions. Most users do not need to
    # customize this. When not set, AWS applies its default delivery policy.
    delivery_policy = optional(string, "")

    # Per-protocol delivery status logging. Each configured protocol block makes
    # SNS write delivery success/failure log entries to CloudWatch Logs using the
    # supplied IAM roles. Configure only the protocols this topic actually
    # delivers to — each block is independent.
    delivery_feedback = optional(object({
      # Delivery status logging for mobile platform application endpoints.
      application = optional(object({
        # IAM role SNS assumes to log SUCCESSFUL deliveries for this protocol.
        # Accepts a direct role ARN or a reference to an AwsIamRole resource.
        # Required for success logging; the sample rate controls what fraction of
        # successes are logged.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        success_feedback_role = optional(string, "")

        # IAM role SNS assumes to log FAILED deliveries for this protocol. Accepts
        # a direct role ARN or a reference to an AwsIamRole resource. Failures are
        # always logged when this role is set (there is no failure sample rate).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        failure_feedback_role = optional(string, "")

        # Percentage (0-100) of successful deliveries to log. Leave at 0 to let AWS
        # apply its default. Only meaningful when `success_feedback_role` is set.
        success_feedback_sample_rate = optional(number, 0)
      }))

      # Delivery status logging for Kinesis Data Firehose delivery streams.
      firehose = optional(object({
        # IAM role SNS assumes to log SUCCESSFUL deliveries for this protocol.
        # Accepts a direct role ARN or a reference to an AwsIamRole resource.
        # Required for success logging; the sample rate controls what fraction of
        # successes are logged.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        success_feedback_role = optional(string, "")

        # IAM role SNS assumes to log FAILED deliveries for this protocol. Accepts
        # a direct role ARN or a reference to an AwsIamRole resource. Failures are
        # always logged when this role is set (there is no failure sample rate).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        failure_feedback_role = optional(string, "")

        # Percentage (0-100) of successful deliveries to log. Leave at 0 to let AWS
        # apply its default. Only meaningful when `success_feedback_role` is set.
        success_feedback_sample_rate = optional(number, 0)
      }))

      # Delivery status logging for HTTP/HTTPS endpoints.
      http = optional(object({
        # IAM role SNS assumes to log SUCCESSFUL deliveries for this protocol.
        # Accepts a direct role ARN or a reference to an AwsIamRole resource.
        # Required for success logging; the sample rate controls what fraction of
        # successes are logged.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        success_feedback_role = optional(string, "")

        # IAM role SNS assumes to log FAILED deliveries for this protocol. Accepts
        # a direct role ARN or a reference to an AwsIamRole resource. Failures are
        # always logged when this role is set (there is no failure sample rate).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        failure_feedback_role = optional(string, "")

        # Percentage (0-100) of successful deliveries to log. Leave at 0 to let AWS
        # apply its default. Only meaningful when `success_feedback_role` is set.
        success_feedback_sample_rate = optional(number, 0)
      }))

      # Delivery status logging for Lambda function endpoints.
      lambda = optional(object({
        # IAM role SNS assumes to log SUCCESSFUL deliveries for this protocol.
        # Accepts a direct role ARN or a reference to an AwsIamRole resource.
        # Required for success logging; the sample rate controls what fraction of
        # successes are logged.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        success_feedback_role = optional(string, "")

        # IAM role SNS assumes to log FAILED deliveries for this protocol. Accepts
        # a direct role ARN or a reference to an AwsIamRole resource. Failures are
        # always logged when this role is set (there is no failure sample rate).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        failure_feedback_role = optional(string, "")

        # Percentage (0-100) of successful deliveries to log. Leave at 0 to let AWS
        # apply its default. Only meaningful when `success_feedback_role` is set.
        success_feedback_sample_rate = optional(number, 0)
      }))

      # Delivery status logging for SQS queue endpoints.
      sqs = optional(object({
        # IAM role SNS assumes to log SUCCESSFUL deliveries for this protocol.
        # Accepts a direct role ARN or a reference to an AwsIamRole resource.
        # Required for success logging; the sample rate controls what fraction of
        # successes are logged.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        success_feedback_role = optional(string, "")

        # IAM role SNS assumes to log FAILED deliveries for this protocol. Accepts
        # a direct role ARN or a reference to an AwsIamRole resource. Failures are
        # always logged when this role is set (there is no failure sample rate).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        failure_feedback_role = optional(string, "")

        # Percentage (0-100) of successful deliveries to log. Leave at 0 to let AWS
        # apply its default. Only meaningful when `success_feedback_role` is set.
        success_feedback_sample_rate = optional(number, 0)
      }))
    }))

    # AWS X-Ray tracing configuration. When set to "Active", SNS publishes trace
    # data for messages. When set to "PassThrough", SNS passes through the trace
    # header but does not sample. Leave empty to use the AWS default (PassThrough).
    # Valid values: "Active", "PassThrough".
    tracing_config = optional(string, "")

    # SNS message signature version. Version 1 uses SHA1, version 2 uses SHA256.
    # SHA256 (version 2) is recommended for new topics. Leave at 0 to use the
    # AWS default (version 1).
    signature_version = optional(number, 0)
  })
}
