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
  description = "AwsEventBridgeRule specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Name of the event bus to attach this rule to. Defaults to "default" (the
    # built-in AWS event bus) when not specified. Can reference an
    # AwsEventBridgeBus resource via `valueFrom`.
    #
    # Changing this field forces rule replacement (delete + recreate).
    #
    # Constraint (AWS, not validated here because reference values resolve at
    # deploy time): schedule_expression is only supported on the DEFAULT bus —
    # a scheduled rule on a custom bus is rejected by the AWS API. Custom
    # buses take event-pattern rules only.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    event_bus_name = optional(string, "")

    # Human-readable description of the rule. Maximum 512 characters.
    description = optional(string, "")

    # JSON event pattern that this rule matches against. Events that match the
    # pattern are routed to the rule's targets. Expressed as a structured object
    # in YAML — the IaC module serializes it to JSON. AWS caps the serialized
    # pattern at 4096 characters.
    #
    # Mutually exclusive with `schedule_expression`.
    #
    # Example patterns:
    #   source: ["aws.ec2"]
    #   detail-type: ["EC2 Instance State-change Notification"]
    #   detail:
    #     state: ["running", "stopped"]
    event_pattern = optional(any)

    # Schedule expression for time-based rule triggering. Supports cron and rate
    # expressions.
    #
    # Mutually exclusive with `event_pattern`.
    #
    # Examples:
    #   "rate(5 minutes)"        — fire every 5 minutes
    #   "rate(1 hour)"           — fire every hour
    #   "cron(0 12 * * ? *)"     — fire at noon UTC every day
    #   "cron(0/15 * * * ? *)"   — fire every 15 minutes
    schedule_expression = optional(string, "")

    # Rule state. Controls whether the rule is actively matching events.
    # Valid values:
    # - "ENABLED"  — the rule matches events (IaC default when not set).
    # - "DISABLED" — the rule exists but matches nothing.
    # - "ENABLED_WITH_ALL_CLOUDTRAIL_MANAGEMENT_EVENTS" — additionally matches
    #   read-only CloudTrail management events (AWS API activity like Describe*
    #   calls), which ENABLED rules never receive. Only meaningful for rules
    #   whose pattern matches CloudTrail management events.
    state = optional(string, "")

    # IAM role EventBridge assumes when invoking this rule's targets. Used when
    # a single role should govern the whole rule (targets can also carry their
    # own per-target `role_arn`, which takes precedence for that target).
    # Required for schedule rules whose targets need role-based invocation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    role_arn = optional(string, "")

    # Force rule deletion even when targets are still attached (AWS refuses to
    # delete a rule that has targets unless forced). Keep the default (false)
    # so an unexpectedly shared rule fails loudly instead of vanishing under
    # an out-of-band consumer; the module manages its own targets' teardown
    # ordering either way.
    force_destroy = optional(bool, false)

    # Targets to invoke when the rule matches an event. At least one target is
    # required. Each target specifies a destination (Lambda, SQS, SNS, Step
    # Functions, etc.), optional input transformation, retry policy, dead
    # letter queue, and an optional service-typed parameter block (SQS,
    # Kinesis, HTTP/API-destination, Batch, ECS RunTask, Redshift Data API,
    # SSM Run Command, SageMaker Pipelines, AppSync).
    #
    # AWS limits: maximum 5 targets per rule (enforced here so the quota
    # fails at validate time instead of at PutTargets).
    targets = optional(list(object({
      # User-assigned name for this target. Used as the Pulumi resource name
      # and as the `target_id` in EventBridge. Must be unique within the rule's
      # targets. Maximum 64 characters, alphanumeric plus hyphen, underscore,
      # and period.
      name = string

      # ARN of the target resource. This is the AWS resource that processes
      # matched events. Common targets include Lambda functions, SQS queues,
      # SNS topics, Step Functions state machines, and CloudWatch Log Groups.
      #
      # No `default_kind` is set because the target resource type varies
      # (Lambda, SQS, SNS, etc.). Use `valueFrom` to reference specific
      # Planton resources.
      #
      # Containment-exempt: a target is where matched events are DELIVERED,
      # never where the rule lives -- the rule belongs to its event bus. A
      # target that is itself a container (another event bus, for a
      # cross-bus forward) would otherwise pull the rule inside the bus it
      # forwards to; on a diagram the rule stays on its own bus with a line
      # to each target.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      arn = string

      # IAM role ARN for EventBridge to assume when invoking this target.
      # Required for targets where EventBridge needs to assume a role:
      # Step Functions, ECS, Kinesis, Batch, CodeBuild, CodePipeline, and
      # cross-account event buses.
      #
      # Not needed for targets that use resource-based policies:
      # Lambda (function policy), SQS (queue policy), SNS (topic policy).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = optional(string, "")

      # Constant JSON input to pass to the target instead of the matched event.
      # Maximum 8192 characters. Mutually exclusive with `input_path` and
      # `input_transformer`.
      input = optional(string, "")

      # JSONPath expression to extract a portion of the matched event and pass
      # to the target. Maximum 256 characters. Mutually exclusive with `input`
      # and `input_transformer`.
      #
      # Example: "$.detail" extracts the detail object from the event.
      input_path = optional(string, "")

      # Input transformer to reshape the matched event before passing to the
      # target. Mutually exclusive with `input` and `input_path`.
      input_transformer = optional(object({
        # Map of variable names to JSONPath expressions that extract values from
        # the matched event. Keys become variables available in `input_template`.
        # Maximum 100 entries. Keys must not start with "AWS" (reserved).
        #
        # Example:
        #   instance: "$.detail.instance-id"
        #   state: "$.detail.state"
        input_paths = optional(map(string), {})

        # Template that produces the final input for the target. References
        # variables from `input_paths` using angle brackets: <variable>.
        # Maximum 8192 characters.
        #
        # Example: "Instance <instance> transitioned to <state>"
        input_template = string
      }))

      # Dead letter queue for events that fail delivery to this target. When
      # EventBridge cannot deliver an event after all retry attempts, the event
      # is routed to the specified SQS queue for investigation.
      dead_letter_config = optional(object({
        # ARN of the SQS queue to use as the dead letter queue. The queue must
        # exist in the same AWS account and region as the rule.
        #
        # Accepts a direct ARN or a reference to an AwsSqsQueue resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        arn = string
      }))

      # Retry policy controlling how EventBridge retries failed deliveries to
      # this target. When not set, EventBridge uses the default policy: retry
      # for 24 hours with up to 185 attempts using exponential backoff.
      retry_policy = optional(object({
        # Maximum time in seconds that EventBridge keeps retrying delivery.
        # Range: 60 to 86400 (1 minute to 24 hours). Absent: AWS default of
        # 86400 (24 hours).
        maximum_event_age_in_seconds = optional(number)

        # Maximum number of retry attempts. Range: 0 to 185. Absent: AWS default
        # of 185. Set to 0 to disable retries (event goes to the DLQ immediately
        # on failure).
        maximum_retry_attempts = optional(number)
      }))

      # SQS-specific parameters. Required when targeting a FIFO SQS queue
      # to specify the message group ID.
      sqs_target = optional(object({
        # Message group ID for FIFO SQS queues. Required when the target is a
        # FIFO queue to ensure proper message ordering and deduplication.
        # Ignored for standard queues.
        message_group_id = optional(string, "")
      }))

      # Kinesis-specific parameters. Controls how events map to stream shards.
      kinesis_target = optional(object({
        # JSONPath expression extracting the partition key from the event (e.g.
        # "$.detail.customer_id"), determining which shard receives the record.
        # When unset, EventBridge uses the event ID — an even spread that
        # sacrifices per-entity ordering. Maximum 256 characters.
        partition_key_path = optional(string, "")
      }))

      # HTTP/API-destination parameters. Path, query, and header values applied
      # when the target arn is an EventBridge API destination.
      http_target = optional(object({
        # Values substituted into the API destination's path wildcards ("*"), in
        # order of appearance in the endpoint URL.
        path_parameter_values = optional(list(string), [])

        # Query string parameters appended to the invocation URL.
        query_string_parameters = optional(map(string), {})

        # HTTP headers added to the request. Keys must be valid header names;
        # AWS reserves headers starting with "X-Amz" and "X-Amzn".
        header_parameters = optional(map(string), {})
      }))

      # AWS Batch parameters. Required when the target arn is a Batch job queue.
      batch_target = optional(object({
        # The Batch job definition each matched event submits. Accepts a
        # reference to an AwsBatchJobDefinition resource (its revision-carrying
        # ARN output, so a new revision rolls the rule's submissions) or a
        # literal value -- a bare name or name:revision literal tracks the
        # name's latest ACTIVE revision on the Batch side instead of pinning
        # one.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        job_definition = string

        # Name assigned to the submitted jobs (visible in the Batch console and
        # APIs). Maximum 128 characters.
        job_name = string

        # Size of an array job. Leave at 0 for a regular (non-array) job.
        # Range when set: 2-10000.
        array_size = optional(number, 0)

        # Retry attempts for the submitted job, 1-10. Leave at 0 to use the job
        # definition's own retry strategy.
        job_attempts = optional(number, 0)
      }))

      # ECS RunTask parameters. Required when the target arn is an ECS cluster —
      # the event launches a task from the referenced task definition.
      ecs_target = optional(object({
        # The task definition to launch. Accepts a reference to an
        # AwsEcsTaskDefinition resource (its revision-carrying ARN output, so a
        # new revision rolls the rule's launches) or a literal ARN.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        task_definition_arn = string

        # Number of tasks launched per matched event. Leave at 0 for the AWS
        # default of 1. Maximum 10.
        task_count = optional(number, 0)

        # Launch type for the task. Valid values: "EC2", "FARGATE", "EXTERNAL".
        # Leave empty to use the cluster's default capacity provider strategy, or
        # set `capacity_provider_strategy` for an explicit blend — AWS rejects a
        # launch type combined with a capacity provider strategy.
        launch_type = optional(string, "")

        # Fargate platform version (e.g. "LATEST", "1.4.0"). Only valid with the
        # FARGATE launch type.
        platform_version = optional(string, "")

        # Task group name used for placement decisions (defaults to the family
        # name of the task definition). Maximum 255 characters.
        group = optional(string, "")

        # Capacity provider strategy for the launched tasks. Mutually exclusive
        # with `launch_type` (AWS rejects both together).
        capacity_provider_strategy = optional(list(object({
          # Capacity provider name ("FARGATE", "FARGATE_SPOT", or a cluster-attached
          # EC2 capacity provider name).
          capacity_provider = string

          # Minimum number of tasks guaranteed to this provider before weights
          # apply. Only one strategy entry may carry a non-zero base. Range 0-100000.
          base = optional(number, 0)

          # Relative share of tasks placed on this provider once bases are met.
          # Range 0-1000.
          weight = optional(number, 0)
        })), [])

        # VPC networking for the task. Required for Fargate (awsvpc network mode);
        # used with EC2 launch type only when the task definition uses awsvpc mode.
        network_configuration = optional(object({
          # Subnets the task's elastic network interface is placed in. Accepts
          # direct subnet IDs or references to AwsSubnet resources.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          subnets = list(string)

          # Security groups attached to the task's network interface. When empty,
          # AWS uses the VPC's default security group. Accepts direct IDs or
          # references to AwsSecurityGroup resources.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          security_groups = optional(list(string), [])

          # Assign a public IP to the task's network interface. Only valid for
          # Fargate tasks in public subnets.
          assign_public_ip = optional(bool, false)
        }))

        # Placement strategies ordering candidate instances (EC2 launch type).
        # Evaluated in order; maximum 5 entries.
        ordered_placement_strategy = optional(list(object({
          # Strategy type. Valid values: "random", "spread", "binpack".
          type = string

          # Attribute the strategy applies to — e.g. "instanceId" or
          # "attribute:ecs.availability-zone" for spread, "cpu" or "memory" for
          # binpack. Not used by "random".
          field = optional(string, "")
        })), [])

        # Placement constraints filtering candidate instances (EC2 launch type).
        # Maximum 10 entries.
        placement_constraints = optional(list(object({
          # Constraint type. Valid values: "distinctInstance", "memberOf".
          type = string

          # Cluster query language expression for "memberOf" constraints, e.g.
          # "attribute:ecs.instance-type =~ t2.*". Not used by "distinctInstance".
          expression = optional(string, "")
        })), [])

        # Propagate tags from the task definition to the launched tasks.
        # The only accepted value is "TASK_DEFINITION".
        propagate_tags = optional(string, "")

        # Use Amazon ECS managed tags for the launched tasks (the
        # aws:ecs:clusterName / aws:ecs:serviceName tag pair).
        enable_ecs_managed_tags = optional(bool, false)

        # Enable ECS Exec on the launched tasks for interactive debugging.
        enable_execute_command = optional(bool, false)

        # Tags applied to the ECS TASKS each event launches (cost allocation,
        # ownership) — distinct from the rule's own resource tags, which the
        # module manages from metadata. Merged with the task definition's tags
        # when propagate_tags is set.
        tags = optional(map(string), {})
      }))

      # Redshift Data API parameters. Required when the target arn is a
      # Redshift cluster — each matched event runs a SQL statement against it.
      redshift_target = optional(object({
        # The database the statement runs in. Maximum 64 characters.
        database = string

        # Authenticate as this database user with temporary credentials
        # (GetClusterCredentials). Use this OR secrets_manager_arn — AWS
        # resolves credentials from whichever is provided. Maximum 128
        # characters.
        db_user = optional(string, "")

        # Authenticate with credentials stored in Secrets Manager — the
        # alternative to db_user temporary credentials. Pass the secret ARN
        # as a literal, e.g.
        # "arn:aws:secretsmanager:us-west-2:123456789012:secret:redshift-...".
        secrets_manager_arn = optional(string, "")

        # The SQL statement each matched event runs. Maximum 100,000
        # characters.
        sql = optional(string, "")

        # A name for the statement, visible in the Data API's statement
        # history. Maximum 500 characters.
        statement_name = optional(string, "")

        # Deliver the matched event to the statement as execution context
        # (the Data API's WithEvent flag).
        with_event = optional(bool, false)
      }))

      # SSM Run Command instance selectors. Required when the target arn is a
      # Systems Manager document — each matched event dispatches the command
      # to the instances the selectors match. Up to 5 selectors, combined
      # with AND.
      run_command_targets = optional(list(object({
        # The selector key: "InstanceIds" (select by instance id) or
        # "tag:<tag-key>" (select every instance carrying the tag). Maximum
        # 128 characters.
        key = string

        # The selector values: instance ids for "InstanceIds", tag values for
        # a "tag:" key. 1-50 entries, each 1-256 characters.
        values = list(string)
      })), [])

      # SageMaker Pipelines parameters. Set when the target arn is a
      # SageMaker pipeline — each matched event starts a pipeline execution
      # with these parameters.
      sagemaker_pipeline_target = optional(object({
        # Parameters passed to the pipeline execution, up to 200 — each must
        # be a parameter the pipeline declares.
        pipeline_parameter_list = optional(list(object({
          # The parameter name, as declared by the pipeline definition.
          name = string

          # The value passed for this execution.
          value = string
        })), [])
      }))

      # AppSync parameters. Required when the target arn is an AppSync
      # GraphQL API endpoint — each matched event invokes the operation.
      appsync_target = optional(object({
        # The GraphQL operation (typically a mutation) to invoke, with its
        # selection set — variables are bound from the (transformed) event
        # input. Maximum 1,048,576 characters.
        graphql_operation = string
      }))
    })), [])
  })
}
