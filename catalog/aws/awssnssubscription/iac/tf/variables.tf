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
  description = "AwsSnsSubscription specification"
  type = object({
    # The AWS region of the subscription — must be the topic's region.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # The SNS topic to subscribe to. Accepts a reference to an AwsSnsTopic
    # resource or a literal topic ARN (including a cross-account topic that has
    # granted this account sns:Subscribe). Create-time immutable — a different
    # topic is a different subscription.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    topic_arn = string

    # Protocol for message delivery. Determines how SNS delivers messages and
    # what `endpoint` must contain. Create-time immutable.
    # - "sqs": SQS queue ARN
    # - "lambda": Lambda function ARN
    # - "http" / "https": URL endpoint (requires endpoint-side confirmation)
    # - "email" / "email-json": email address (always requires manual confirmation)
    # - "sms": phone number in E.164 format
    # - "firehose": Kinesis Data Firehose delivery stream ARN (requires
    #   subscription_role_arn)
    # - "application": mobile platform endpoint ARN
    protocol = string

    # Endpoint to deliver messages to. The format depends on the protocol (see
    # `protocol` field documentation). Accepts a direct value or a reference to
    # another resource's output — e.g. an SQS subscription references an
    # AwsSqsQueue's queue_arn, a Lambda subscription references an AwsLambda's
    # function_arn. Create-time immutable.
    #
    # No `default_kind` is set because the target resource varies by protocol
    # (SQS queue, Lambda function, Firehose stream, plain URL, ...).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    endpoint = string

    # Filter policy to select which messages this subscription receives.
    # Messages that match no filter are not delivered. The filter can be applied
    # to message attributes (default) or the message body (see
    # `filter_policy_scope`). Expressed as a JSON structure in YAML.
    filter_policy = optional(any)

    # Scope for the filter policy. Controls whether the filter is evaluated
    # against message attributes or the message body.
    # Valid values: "MessageAttributes" (default), "MessageBody".
    # Only relevant when `filter_policy` is set.
    filter_policy_scope = optional(string, "")

    # When true, messages are delivered as-is without JSON wrapping. Supported
    # for SQS, HTTP/S, and Firehose protocols. When false (default), SNS wraps
    # the message in a JSON envelope containing metadata (MessageId, TopicArn,
    # Timestamp, etc.).
    raw_message_delivery = optional(bool, false)

    # Dead letter queue for this subscription's delivery failures. When SNS
    # cannot deliver a message after all retry attempts, the message is routed
    # to the specified SQS queue. This is separate from any DLQ the endpoint
    # itself has — it catches SNS-to-subscriber delivery failures.
    dead_letter_config = optional(object({
      # ARN of the SQS dead letter queue for failed message deliveries. Accepts a
      # direct ARN or a reference to an AwsSqsQueue resource. The queue must
      # reside in the same AWS account and region as the SNS topic, and its
      # resource policy must allow sns.amazonaws.com to SendMessage.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      dead_letter_target_arn = string
    }))

    # HTTP/HTTPS delivery retry policy override for this subscription. Expressed
    # as a JSON string matching the SNS delivery policy format. Overrides the
    # topic-level delivery policy for this subscription only. Most users do not
    # need to customize this.
    delivery_policy = optional(string, "")

    # Replay policy for archived messages. When the topic (FIFO with
    # `archive_policy`) retains messages, a new subscription can replay the
    # archive from a starting point before receiving live traffic — the
    # mechanism for backfilling a new consumer. Expressed as the SNS replay
    # policy JSON document, e.g. {"PointType": "Timestamp",
    # "StartingPoint": "2026-07-01T00:00:00Z"}.
    replay_policy = optional(any)

    # IAM role ARN for Firehose delivery. Required when protocol is "firehose".
    # The role must grant SNS permission to write to the Firehose delivery
    # stream. Accepts a direct ARN or a reference to an AwsIamRole resource.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subscription_role_arn = optional(string, "")

    # Set to true when the HTTP/S endpoint confirms subscriptions on its own
    # (responds to the SubscriptionConfirmation callback without a human).
    # Deployment then waits for the confirmation to complete instead of leaving
    # the subscription pending. Only meaningful for "http"/"https" protocols —
    # SQS, Lambda, Firehose, and application subscriptions confirm automatically,
    # and email subscriptions always require a manual click.
    endpoint_auto_confirms = optional(bool, false)

    # Minutes to wait for an HTTP/S endpoint to confirm the subscription before
    # deployment fails. Leave at 0 for the AWS provider default (1 minute).
    # Only meaningful for "http"/"https" protocols.
    confirmation_timeout_minutes = optional(number, 0)
  })
}
