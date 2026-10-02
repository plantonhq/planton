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
  description = "AwsStepFunction specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # State machine type. Determines execution semantics and pricing model.
    # - "STANDARD": Long-running, exactly-once, full execution history.
    # - "EXPRESS": High-volume, short-duration, at-most-once.
    # Cannot be changed after creation (forces replacement).
    # When omitted the IaC module defaults to "STANDARD".
    type = optional(string, "")

    # State machine definition in Amazon States Language (ASL). Write the
    # definition as native YAML; the IaC module serializes it to JSON for the
    # AWS API. ASL key casing (StartAt, States, Type, Resource, etc.) is
    # preserved through serialization.
    #
    # Maximum size: 1,048,576 bytes (1 MB) after JSON serialization.
    #
    # Use AWS Step Functions Workflow Studio, CDK, or the ASL specification
    # to author complex workflows, then express the result as YAML here.
    definition = any

    # IAM execution role ARN. The role must have a trust policy for
    # states.amazonaws.com and policies granting access to all services
    # invoked by the workflow (Lambda:InvokeFunction, SQS:SendMessage, etc.).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    role_arn = string

    # Publish a version of the state machine on every create and on every
    # configuration update. Published versions are immutable snapshots
    # (definition + role + logging/tracing/encryption at publish time) addressed
    # by the version ARN exported in stack outputs. Versions are the foundation
    # for alias-based traffic shifting and safe rollbacks: point consumers at a
    # version ARN (or an alias routing between two versions) instead of the
    # mutable state machine ARN. When false (the default), executions always run
    # the latest saved revision.
    publish = optional(bool, false)

    # Named aliases for the state machine. Each alias points at the version
    # published by this deployment (requires publish: true) and exposes a
    # stable ARN consumers can invoke while the underlying version advances
    # on every configuration change. Aliases are keyed by name: renaming an
    # entry replaces that alias without touching its siblings.
    #
    # Weighted canary routing between two specific versions is an imperative
    # deployment-shift operation (the routing weights change DURING a
    # rollout, not in a declarative snapshot) and is deliberately not
    # modeled; each alias here routes 100% of traffic to the version this
    # deployment published.
    aliases = optional(list(object({
      # Alias name. 1-80 characters matching [0-9A-Za-z_-]. The name keys the
      # provider resource: renaming replaces this alias without touching
      # siblings.
      name = optional(string, "")

      # Optional human-readable description of the alias (up to 256 characters).
      description = optional(string, "")
    })), [])

    # Logging configuration for execution history events. When omitted, no
    # logging configuration is sent (new state machines default to level OFF).
    # To explicitly turn logging OFF on a state machine that previously had it
    # on, keep this block with level: "OFF" — removing the block entirely
    # sends nothing and the provider keeps the last applied logging state.
    # Logging is supported for both STANDARD and EXPRESS state machines.
    logging = optional(object({
      # Logging level. Determines which execution history events are logged.
      # - "ALL": Log all event types (recommended for development and debugging).
      # - "ERROR": Log only error events (recommended for production).
      # - "FATAL": Log only fatal errors.
      # - "OFF": Disable logging.
      level = optional(string, "")

      # Whether to include execution input and output data in log entries. When
      # true, the full JSON payloads passed between states are logged. Useful for
      # debugging but may increase log volume and expose sensitive data.
      include_execution_data = optional(bool, false)

      # CloudWatch Logs log group ARN for log delivery. Accepts a direct ARN or a
      # reference to a CloudWatch Log Group resource.
      #
      # Note: AWS requires the ARN to end with ":*". The IaC module automatically
      # appends this suffix if not present, so you can reference a log group ARN
      # directly without worrying about the suffix.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      log_destination = optional(string, "")
    }))

    # Enable AWS X-Ray tracing for the state machine. When true, Step Functions
    # sends trace data to X-Ray for visualizing request flows. Ensure the
    # execution role has xray:PutTraceSegments and xray:PutTelemetryRecords
    # permissions.
    #
    # Tri-state: leave unset to keep the AWS default (tracing off, and no
    # tracing configuration is sent). Set explicitly to true or false to pin
    # the state — an explicit false is what turns tracing OFF on a state
    # machine that previously had it on (simply removing a true value sends
    # nothing, and the provider keeps the last applied state).
    tracing_enabled = optional(bool)

    # Encryption configuration for data at rest. When omitted, AWS uses
    # AWS-owned keys (default, no additional cost). Provide this block to
    # use a customer-managed KMS key for encrypting state machine data,
    # execution history, and input/output payloads.
    #
    # One-way in practice: once a customer-managed key has been applied,
    # REMOVING this block does not revert the state machine to AWS-owned
    # keys — the provider suppresses the block's removal and keeps the last
    # applied key. Reverting requires an out-of-band update today.
    encryption = optional(object({
      # Customer-managed KMS key ARN for encrypting state machine data, execution
      # history, and input/output payloads. The key must be a symmetric encryption
      # key in the same region as the state machine. For cross-account access, use
      # the full key ARN (not alias).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_id = string

      # Duration in seconds for which Step Functions reuses a data encryption key
      # before calling KMS GenerateDataKey again. Higher values reduce KMS API
      # costs but increase the window for key reuse.
      # Range: 60–900 seconds. AWS default: 300 (5 minutes).
      # Leave at 0 to use the AWS default.
      kms_data_key_reuse_period_seconds = optional(number, 0)
    }))
  })
}
