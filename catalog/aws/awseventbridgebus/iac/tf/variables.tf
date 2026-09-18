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
  description = "AwsEventBridgeBus specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Human-readable description of the event bus.
    # Maximum length is 512 characters.
    description = optional(string, "")

    # KMS key identifier for encrypting events on this bus. Accepts a KMS key
    # ARN, key ID, key alias, or key alias ARN. When omitted, events are
    # encrypted with an AWS-owned key at no additional cost.
    #
    # Accepts a direct value or a reference to an AwsKmsKey resource.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_identifier = optional(string, "")

    # Partner event source name. Set this only when creating a bus for a SaaS
    # partner integration (e.g., Datadog, Zendesk, PagerDuty).
    #
    # The value must match the pattern: aws.partner/{partner}/{...} and the bus
    # name (`metadata.name`) must match this value exactly.
    #
    # This field is immutable — changing it forces bus replacement.
    event_source_name = optional(string, "")

    # Dead letter queue configuration for the event bus. When set, events that
    # fail delivery to any rule target on this bus are routed to the specified
    # SQS queue for investigation and reprocessing.
    #
    # This is the bus-level DLQ — it catches events that cannot be delivered
    # to ANY target on any rule attached to this bus. Individual rules can also
    # have their own DLQ configuration (configured on AwsEventBridgeRule).
    dead_letter_config = optional(object({
      # ARN of the SQS queue to use as the dead letter queue. The queue must
      # exist in the same AWS account and region as the event bus.
      #
      # Accepts a direct ARN or a reference to an AwsSqsQueue resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      arn = string
    }))

    # Logging configuration for the event bus. When set, EventBridge writes
    # event delivery logs to CloudWatch Logs. Useful for debugging event
    # routing, monitoring delivery failures, and auditing event traffic.
    log_config = optional(object({
      # Logging level. Controls which events are logged.
      #
      # Valid values:
      # - "OFF"   — no logging
      # - "ERROR" — log only delivery failures
      # - "INFO"  — log delivery successes and failures
      # - "TRACE" — log all events including matched/unmatched (most verbose)
      #
      # When the log_config block is provided, this field is required.
      level = string

      # Whether to include the full event detail in log entries.
      #
      # Valid values:
      # - "NONE" — exclude event detail from logs (smaller log volume)
      # - "FULL" — include complete event detail in each log entry
      #
      # Default behavior when omitted: "NONE".
      include_detail = optional(string, "")
    }))

    # Resource-based policy for the event bus. Controls which AWS principals
    # (accounts, organizations, or roles) may put events onto this bus —
    # the mechanism behind cross-account event ingestion. Expressed as a
    # standard IAM policy document structure; one policy per bus (statements
    # express per-account/per-org grants). When unset, only the owning
    # account can put events.
    resource_policy = optional(any)

    # Event archives on this bus. An archive continuously records events
    # delivered to the bus (all events, or the subset matching its event
    # pattern) so they can be replayed later — the disaster-recovery and
    # backfill lever for event-driven systems. Replay itself is an on-demand
    # operation (EventBridge StartReplay), not declarative configuration; the
    # archive defines what is recorded and for how long.
    #
    # Each entry materializes one archive sourced from THIS bus. Archives are
    # bus-scoped children: they cannot outlive the bus and cannot record from
    # any other source. An empty archive costs nothing — storage is billed
    # per GB of archived events.
    archives = optional(any, [])
  })
}
