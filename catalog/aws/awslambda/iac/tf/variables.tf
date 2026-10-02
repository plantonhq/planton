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
  description = "AwsLambda specification"
  type = object({
    # The AWS region the function is created in. Must match the region
    # of the S3 code bucket, any VPC subnets, and the EFS access point
    # it references.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # Free-form description shown in the AWS Console -- operational
    # context for humans browsing the account. Up to 256 characters.
    description = optional(string, "")

    # The IAM execution role the function assumes: it must trust
    # lambda.amazonaws.com and carry the policies the code needs
    # (CloudWatch Logs at minimum -- AWSLambdaBasicExecutionRole; plus
    # AWSLambdaVPCAccessExecutionRole for VPC attachment). Roles own
    # their policies -- this module never attaches policies to a role it
    # merely references. Reference an AwsIamRole role_arn output or pass
    # a literal role ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    role_arn = string

    # The deployment package as a zip archive in S3. Requires runtime
    # and handler. Prefer a bucket in the function's own region --
    # cross-region pulls are slower and billed.
    s3 = optional(object({
      # The S3 bucket holding the deployment package. Must be in the
      # function's region. Reference an AwsS3Bucket bucket_id output or
      # pass a literal bucket name.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      bucket = string

      # The object key (path) of the deployment package zip.
      key = string

      # Pin a specific object version (versioned buckets). Empty deploys
      # the current version; pair with source_code_hash for fully
      # declarative code rolls.
      object_version = optional(string, "")
    }))

    # The function code as a container image in ECR, e.g.
    # "123456789012.dkr.ecr.us-east-1.amazonaws.com/repo:tag". The image
    # defines the runtime and entrypoint, so runtime and handler must
    # stay empty; image_config below can override the entrypoint. Images
    # may be up to 10 GB -- the right choice for heavy dependency trees.
    #
    # Lambda pulls only from a private ECR repository in the same account
    # and Region as the function; there is no registry-login field and AWS
    # accepts none. For an image that lives elsewhere (GHCR, Docker Hub,
    # another Region), push it to ECR or declare a pull-through cache rule on
    # AwsEcrRegistrySettings and point this URI at the cached repository.
    image_uri = optional(string, "")

    # Base64-encoded SHA256 of the deployment package. Set it (usually
    # from your build pipeline) to make code updates declarative: a new
    # hash rolls the function, an unchanged hash is a no-op even when
    # the S3 object is rewritten in place. Leave empty to update only
    # when the S3 key or object version changes.
    source_code_hash = optional(string, "")

    # Base64-encoded SHA256 of the DEPLOYED package as AWS reports it
    # (the digest GetFunction returns). Set it to detect and roll
    # out-of-band code changes from the deployed artifact's own digest —
    # the deploy-side complement to source_code_hash, which hashes the
    # artifact you upload. Most configurations want source_code_hash;
    # use this when the artifact is published by a pipeline you don't
    # control and only the deployed digest is known.
    code_sha256 = optional(string, "")

    # The KMS key that encrypts the deployment package in S3 (bring-
    # your-own-key for the code artifact itself -- distinct from
    # kms_key_arn, which encrypts environment variables). Only
    # meaningful for zip deployments. Reference an AwsKmsKey key_arn
    # output or pass a literal key ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    source_kms_key_arn = optional(string, "")

    # The language runtime for zip deployments, e.g. "nodejs22.x",
    # "python3.13", "java21", "dotnet8", "ruby3.4", "provided.al2023"
    # (custom runtimes / native binaries). AWS retires runtimes on its
    # own schedule and adds new ones frequently, so the accepted set is
    # validated by AWS at deploy time rather than frozen here. Required
    # for zip code; must stay empty for container images (the image
    # carries its own runtime).
    runtime = optional(string, "")

    # The function entrypoint for zip deployments. Format is
    # runtime-specific: "index.handler" (Node.js), "module.function"
    # (Python), "package.Class::method" (Java), "bootstrap" (custom
    # runtimes). Required for zip code; must stay empty for container
    # images (the image CMD/ENTRYPOINT defines it).
    handler = optional(string, "")

    # The instruction-set architecture: "x86_64" or "arm64". Empty keeps
    # the AWS default (x86_64). arm64 (Graviton) is typically ~20%
    # cheaper per GB-second and often faster -- prefer it whenever your
    # runtime and native dependencies support it.
    architecture = optional(string, "")

    # Memory in MB: 128-10240 for standard functions, up to 32768 when
    # the function runs on Lambda Managed Instances (managed_instances)
    # -- AWS enforces the per-platform ceiling at deploy time. CPU and
    # network scale linearly with memory (a full vCPU arrives around
    # 1769 MB), so raising memory is also how you buy CPU -- for
    # CPU-bound code a larger size often costs LESS overall by finishing
    # sooner. 0 keeps the AWS default (128 MB).
    memory_size_mb = optional(number, 0)

    # Maximum execution time per invocation in seconds, 1-900. Size it
    # slightly above the worst expected runtime; API-fronted functions
    # should stay well under their gateway's timeout. 0 keeps the AWS
    # default (3 seconds).
    timeout_seconds = optional(number, 0)

    # Scratch space at /tmp in MB, 512-10240. Sized for workloads that
    # stage files locally (media processing, ML model unpacking). Billed
    # above the free 512 MB. 0 keeps the AWS default (512 MB).
    ephemeral_storage_mb = optional(number, 0)

    # Plain-configuration environment variables available to the code at
    # runtime. Never put secret material here -- environment variables
    # are visible to anyone who can read the function configuration.
    # Give the execution role access to SSM Parameter Store or Secrets
    # Manager and resolve secrets at runtime instead.
    environment = optional(map(string), {})

    # The customer-managed KMS key that encrypts the environment
    # variables at rest (and SnapStart snapshots). Empty uses the
    # AWS-managed key. Reference an AwsKmsKey key_arn output or pass a
    # literal key ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_arn = optional(string, "")

    # Subnets the function's ENIs are created in -- typically private
    # subnets across at least two availability zones. Attaching to a VPC
    # gives the code access to private resources (databases, caches) and
    # removes default internet access (route through a NAT gateway to
    # restore it). Leave empty to run outside any VPC (the default, with
    # direct internet access). Reference AwsSubnet subnet_id outputs or
    # pass literal subnet IDs.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = optional(list(string), [])

    # Security groups attached to the function's ENIs. Outbound rules
    # must allow the services the code reaches (databases, AWS
    # endpoints, the internet via NAT). Required together with
    # subnet_ids when attaching to a VPC. Reference AwsSecurityGroup
    # security_group_id outputs or pass literal security group IDs.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # Allow the VPC-attached function to make outbound IPv6 connections
    # over dual-stack subnets. Only meaningful with a VPC attachment.
    ipv6_allowed_for_dual_stack = optional(bool, false)

    # Where asynchronous invocations that exhaust their retries are
    # sent: an SQS queue or SNS topic ARN. The execution role needs
    # sqs:SendMessage / sns:Publish on the target. Reference an
    # AwsSqsQueue queue_arn output, or pass an SNS topic ARN as a
    # literal (or an explicit-kind reference to AwsSnsTopic topic_arn).
    # For finer-grained routing (separate success/failure destinations,
    # max event age), use async_invoke_config instead.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    dead_letter_target_arn = optional(string, "")

    # AWS X-Ray tracing: "Active" (trace and sample invocations --
    # requires xray:PutTraceSegments on the execution role) or
    # "PassThrough" (only forward upstream trace headers). Empty keeps
    # the AWS default (PassThrough).
    tracing_mode = optional(string, "")

    # Mount an EFS access point into the execution environment --
    # durable shared storage across invocations and functions (ML
    # models, shared caches). Requires a VPC attachment reaching the
    # file system's mount targets.
    file_system_config = optional(object({
      # The EFS ACCESS POINT ARN (not the file system ARN). Reference an
      # AwsEfsAccessPoint resource or pass a literal access point ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      access_point_arn = string

      # Where the file system appears inside the execution environment.
      # Must be under /mnt, e.g. "/mnt/data".
      local_mount_path = string
    }))

    # Container-image entrypoint overrides. Only meaningful with
    # image_uri.
    image_config = optional(object({
      # Override the image ENTRYPOINT.
      entry_point = optional(list(string), [])

      # Override the image CMD (the handler argument for AWS base images).
      command = optional(list(string), [])

      # Override the image working directory.
      working_directory = optional(string, "")
    }))

    # Lambda layer version ARNs merged into the execution environment,
    # in order (later layers shadow earlier ones), up to five. Layers
    # carry shared dependencies and tooling outside the deployment
    # package. Pass literal layer-version ARNs, e.g.
    # "arn:aws:lambda:us-west-2:123456789012:layer:shared-libs:3".
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    layer_arns = optional(list(string), [])

    # Publish a new immutable version on every code or configuration
    # change. Versions are what aliases route to -- enable this when
    # using aliases for traffic shifting or SnapStart (which only
    # applies to published versions).
    publish = optional(bool, false)

    # Maintain the "$LATEST.PUBLISHED" head pointer: "LATEST_PUBLISHED"
    # (the only value AWS currently accepts) keeps a moving qualifier
    # that always resolves to the newest published version -- the
    # addressable target scaling_configs and qualified invocations can
    # pin without naming version numbers. Empty leaves the head pointer
    # unmanaged. Rollout caveat (live-verified us-west-2, 2026-08-11):
    # regions/accounts where the $LATEST.PUBLISHED feature has not rolled
    # out reject the value at CreateFunction with
    # InvalidParameterValueException ("isn't a valid value for this
    # field") -- and AWS creates the function BEFORE rejecting this
    # parameter, so a failed create can leave a live function behind.
    # Mid-rollout is subtler (live-verified us-west-2, 2026-08-13):
    # CreateFunction can ACCEPT the value yet silently not maintain the
    # head -- the "$LATEST.PUBLISHED" qualifier answers
    # ResourceNotFoundException even after versions publish -- while
    # UpdateFunctionCode still rejects the same value. Acceptance at
    # create is NOT availability; the feature is available only where
    # the qualifier actually resolves after a publish. Set this field
    # only where that holds.
    publish_to = optional(string, "")

    # Reserved concurrency for this function. Unset: the function draws
    # from the account's unreserved pool (no dedicated cap). 0: all
    # invocations are throttled -- an operational kill switch. Positive:
    # that many concurrent executions are carved out of the account
    # pool, acting as both a guarantee and a ceiling.
    reserved_concurrent_executions = optional(number)

    # SnapStart: resume new execution environments from a pre-initialized
    # snapshot instead of running init from scratch -- order-of-magnitude
    # cold-start reduction for JVM and other slow-init runtimes. Applies
    # to published versions only (enable publish and invoke through a
    # version or alias to benefit).
    snap_start = optional(bool, false)

    # Run the function on a Lambda Managed Instances capacity provider --
    # dedicated EC2 capacity AWS manages on the function's behalf --
    # instead of the on-demand fleet. The platform for steady
    # high-throughput workloads, memory above 10 GB, and per-tenant
    # isolation.
    managed_instances = optional(object({
      # The Lambda capacity provider supplying the managed EC2 capacity.
      # Pass the capacity provider ARN as a literal, e.g.
      # "arn:aws:lambda:us-west-2:123456789012:capacity-provider:my-cp".
      capacity_provider_arn = string

      # Memory (GiB) provisioned per vCPU in each execution environment --
      # how compute-heavy vs memory-heavy the environments are sized.
      # 0 keeps the AWS default sizing.
      memory_gib_per_vcpu = optional(number, 0)

      # Maximum concurrent invocations one execution environment may serve.
      # 0 keeps the AWS default.
      max_concurrency_per_environment = optional(number, 0)
    }))

    # Durable execution: AWS checkpoints the function's progress so
    # long-running workflows survive interruption and resume where they
    # stopped, far beyond the classic 15-minute cap. ADDING OR REMOVING
    # this block REPLACES the function (an AWS constraint); the values
    # inside update in place.
    durable_config = optional(object({
      # Maximum end-to-end time of one durable invocation in seconds,
      # 1-31622400 (up to 366 days) -- checkpointing is what lets it far
      # exceed the classic 15-minute cap.
      execution_timeout_seconds = optional(number, 0)

      # How long (days, 1-90) AWS retains each durable execution's state
      # and history after it completes. 0 keeps the AWS default (14).
      retention_period_days = optional(number, 0)
    }))

    # Isolate execution environments per tenant: "PER_TENANT" (the only
    # mode AWS currently accepts) dedicates environments to the tenant id
    # callers pass at invoke time -- no cross-tenant reuse of warm
    # state. Create-time immutable: CHANGING it REPLACES the function.
    # Empty keeps the AWS default (environments shared across all
    # invocations).
    tenant_isolation_mode = optional(string, "")

    # CloudWatch Logs delivery. Unset, AWS creates and writes to
    # "/aws/lambda/<function-name>" in plain-text format with the log
    # group owned by AWS (it survives function deletion). Configure to
    # switch to structured JSON, tune log levels, or write into a log
    # group you manage (retention, encryption, subscription filters).
    logging_config = optional(object({
      # "Text" (plain lines, the AWS default) or "JSON" (structured
      # records with built-in level filtering). Empty keeps the AWS
      # default. Level filtering below requires JSON.
      log_format = optional(string, "")

      # Minimum level of application (your code's) log records delivered:
      # "TRACE", "DEBUG", "INFO", "WARN", "ERROR", or "FATAL". JSON format
      # only. Empty delivers everything.
      application_log_level = optional(string, "")

      # Minimum level of system (Lambda platform) log records delivered:
      # "DEBUG", "INFO", or "WARN". JSON format only. Empty delivers
      # everything.
      system_log_level = optional(string, "")

      # Write into a log group you manage -- for retention policy,
      # KMS encryption, or subscription filters -- instead of the
      # AWS-created "/aws/lambda/<function-name>". Reference an
      # AwsCloudwatchLogGroup log_group_name output or pass a literal log
      # group name. The execution role needs logs:CreateLogStream and
      # logs:PutLogEvents on it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      log_group = optional(string, "")
    }))

    # A code-signing configuration ARN: AWS then rejects deployment
    # packages whose signatures don't match the trusted signing
    # profiles. Pass a literal ARN, e.g.
    # "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-...".
    # Only meaningful for zip deployments.
    code_signing_config_arn = optional(string, "")

    # Named pointers to published versions, each optionally splitting
    # traffic between two versions (canary) and pre-warming provisioned
    # concurrency. Aliases are the stable invocation targets clients and
    # event sources should reference -- repointing an alias is how you
    # ship or roll back without touching callers. Each entry
    # materializes as its own alias resource keyed by name, so edits are
    # in-place.
    aliases = optional(list(object({
      # The alias name, e.g. "live", "staging". Unique within the
      # function.
      name = string

      # Free-form description of what this alias routes.
      description = optional(string, "")

      # The published version this alias points to, e.g. "1", or "$LATEST"
      # (unpublished head -- fine for dev aliases, avoid for production).
      # Numbering caveat (live-verified 2026-08-11): AWS never reuses
      # version numbers for a function NAME, even across delete/recreate --
      # a recreated function's first publish continues the old numbering,
      # so a literal pin like "1" that worked on the first deployment 404s
      # at CreateAlias ("Function not found ...:1") on a recreate. Pin
      # literal numbers only against a function whose publish history you
      # know; use "$LATEST" where the alias just needs to exist.
      function_version = string

      # Canary routing: additional version(s) receiving a fraction of this
      # alias's traffic, as version -> weight (0.0-1.0). E.g. {"2": 0.1}
      # sends 10% of traffic to version 2 and 90% to function_version.
      # AWS allows at most one additional version.
      routing_additional_version_weights = optional(map(number), {})

      # Pre-warmed execution environments kept ready for this alias --
      # eliminates cold starts at the cost of paying for idle warmth.
      # Applied as a provisioned-concurrency config keyed by this alias.
      # Unset means no provisioned concurrency. AWS only allows provisioned
      # concurrency on an alias that points at exactly one published version:
      # not on a weighted (canary) alias, and not on $LATEST.
      provisioned_concurrent_executions = optional(number)
    })), [])

    # A built-in HTTPS endpoint for the function -- the zero-
    # infrastructure alternative to an API gateway for simple HTTP
    # services and webhooks.
    function_url = optional(object({
      # Who may invoke the URL: "AWS_IAM" (callers sign requests with
      # SigV4 -- the safe default) or "NONE" (public -- anyone with the
      # URL; AWS still requires an explicit public invoke permission,
      # which the module manages). Choose NONE only for genuinely public
      # webhooks and pair it with your own request validation.
      authorization_type = string

      # "BUFFERED" (the default -- response returned whole, up to 6 MB) or
      # "RESPONSE_STREAM" (stream the response as it is produced, up to
      # 20 MB soft cap -- for large payloads and time-to-first-byte).
      # Empty keeps the AWS default.
      invoke_mode = optional(string, "")

      # Cross-origin resource sharing for browser callers.
      cors = optional(object({
        # Allow credentials (cookies, authorization headers) in cross-origin
        # requests.
        allow_credentials = optional(bool, false)

        # Origins allowed to call the URL, e.g. "https://app.example.com";
        # "*" allows all.
        allow_origins = optional(list(string), [])

        # HTTP methods allowed, e.g. "GET", "POST"; "*" allows all.
        allow_methods = optional(list(string), [])

        # Request headers allowed, e.g. "content-type", "authorization".
        allow_headers = optional(list(string), [])

        # Response headers exposed to browser scripts.
        expose_headers = optional(list(string), [])

        # How long (seconds) browsers may cache the preflight response,
        # up to 86400 (24h). 0 keeps the AWS default (0 -- no caching).
        max_age_seconds = optional(number, 0)
      }))

      # Attach the URL to one of this spec's aliases instead of $LATEST --
      # the URL then serves whatever version the alias routes (including
      # canary weights), so traffic shifting applies to URL callers too.
      # Must name an entry in aliases; empty serves the unpublished head.
      qualifier = optional(string, "")
    }))

    # Resource-policy statements granting other principals and services
    # permission to invoke this function -- how S3 buckets, SNS topics,
    # EventBridge rules, and other accounts are authorized to call it.
    # Each entry materializes as its own permission statement keyed by
    # statement_id, so list edits add/remove statements in place.
    invoke_permissions = optional(list(object({
      # Unique statement identifier within the function's policy, e.g.
      # "allow-s3-uploads-bucket". The per-name key each entry
      # materializes under.
      statement_id = string

      # Who is allowed to invoke: a service principal (e.g.
      # "s3.amazonaws.com", "sns.amazonaws.com", "events.amazonaws.com"),
      # an AWS account ID, or an IAM principal ARN.
      principal = string

      # The Lambda action granted. Empty keeps the sensible default
      # ("lambda:InvokeFunction"); "lambda:InvokeFunctionUrl" grants URL
      # invocation instead.
      action = optional(string, "")

      # Scope a service-principal grant to one source resource, e.g. the
      # S3 bucket ARN or SNS topic ARN that may invoke. Strongly
      # recommended for service principals -- without it ANY resource of
      # that service in ANY account with knowledge of the function name
      # could invoke (the confused-deputy problem).
      source_arn = optional(string, "")

      # Scope a service-principal grant to sources owned by this account
      # ID -- the coarser companion to source_arn.
      source_account = optional(string, "")

      # Grant access to every account in an AWS Organization by org ID,
      # e.g. "o-a1b2c3d4e5".
      principal_org_id = optional(string, "")

      # Required auth type when granting "lambda:InvokeFunctionUrl":
      # "AWS_IAM" or "NONE" (the public-URL grant).
      function_url_auth_type = optional(string, "")

      # Scope the grant to one qualified ARN: a published version number
      # ("3") or an alias name ("live"). The principal may then invoke only
      # that version/alias. Empty grants on the unqualified function.
      qualifier = optional(string, "")

      # Token the caller must present with the invocation -- used by Alexa
      # Skills (the skill id) to pin the grant to one skill.
      event_source_token = optional(string, "")

      # Restrict the grant to invocations that arrive THROUGH the function
      # URL (rejects direct Invoke calls under this statement).
      invoked_via_function_url = optional(bool, false)
    })), [])

    # Delivery and retry behavior for asynchronous invocations (S3
    # events, SNS, EventBridge): retry count, event age, and on-success
    # / on-failure destinations -- the richer successor to the
    # dead-letter queue.
    async_invoke_config = optional(object({
      # Retries after a failed asynchronous invocation, 0-2. Unset keeps
      # the AWS default (2).
      maximum_retry_attempts = optional(number)

      # How long (seconds) an event may wait in the internal queue before
      # Lambda discards it, 60-21600. 0 keeps the AWS default (21600 --
      # 6 hours).
      maximum_event_age_seconds = optional(number, 0)

      # Where successful invocation records are delivered: an SQS queue,
      # SNS topic, EventBridge bus, or Lambda function ARN. Reference an
      # AwsSqsQueue queue_arn output (or an explicit-kind reference /
      # literal ARN for the other targets). The execution role needs send
      # rights on the destination.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      on_success_destination_arn = optional(string, "")

      # Where failed invocation records are delivered after retries are
      # exhausted -- same target types as on_success_destination_arn.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      on_failure_destination_arn = optional(string, "")

      # Apply this config to one qualified scope instead of the whole
      # function: a published version number ("3") or an alias name
      # ("live"). Async invocations through other qualifiers then keep the
      # AWS defaults. Empty applies at function scope.
      qualifier = optional(string, "")
    }))

    # Recursive-loop detection: "Terminate" (the AWS default -- stop
    # runaway self-invocation loops automatically) or "Allow" (opt out
    # for workloads that legitimately self-invoke, e.g. intentional
    # fan-out through SQS). Empty keeps the AWS default.
    recursive_loop = optional(string, "")

    # How runtime patches roll out to the function: automatically, only
    # on function updates, or pinned to a specific runtime version.
    runtime_management = optional(object({
      # "Auto" (the AWS default -- receive runtime patches as AWS releases
      # them), "FunctionUpdate" (patches apply only when the function is
      # next updated -- change-window control), or "Manual" (pin to the
      # exact runtime version in runtime_version_arn -- an emergency
      # rollback tool, not a steady state).
      update_runtime_on = string

      # The exact runtime version ARN to pin, required with (and only
      # meaningful for) "Manual". Obtain it from the function's runtime
      # update events or the Lambda console.
      runtime_version_arn = optional(string, "")

      # Apply the policy to one qualified scope instead of the whole
      # function: a published version number ("3") or an alias name
      # ("live"). Empty applies at function scope ($LATEST).
      qualifier = optional(string, "")
    }))

    # Per-qualifier execution-environment scaling bounds for functions on
    # Lambda Managed Instances: pin a published version's (or the
    # "$LATEST.PUBLISHED" head's) environment fleet between a floor and a
    # ceiling. Each entry materializes as its own scaling-config resource
    # keyed by qualifier, so list edits update in place.
    scaling_configs = optional(list(object({
      # The qualifier the bounds apply to: a published version number
      # ("3") or "$LATEST.PUBLISHED" (the newest published version --
      # maintained by publish_to). Alias names are NOT accepted here (an
      # AWS constraint specific to scaling configs).
      qualifier = string

      # The floor of always-provisioned execution environments. Unset
      # leaves the floor to AWS.
      min_execution_environments = optional(number)

      # The ceiling of execution environments AWS may scale to. Unset
      # leaves the ceiling to AWS.
      max_execution_environments = optional(number)
    })), [])
  })
}
