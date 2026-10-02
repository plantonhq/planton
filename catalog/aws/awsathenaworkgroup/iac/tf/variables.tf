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
  description = "AwsAthenaWorkgroup specification"
  type = object({
    # The AWS region where the Athena workgroup will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Human-readable description of the workgroup shown in the Athena console.
    # Helps teams understand which application or team owns the workgroup.
    # Maximum 1024 characters (enforced by the AWS API).
    description = optional(string, "")

    # Operational state of the workgroup. A DISABLED workgroup rejects new query
    # submissions while keeping its configuration, history, and saved queries
    # intact -- the safe way to pause a team's spend without deleting anything.
    # Defaults to ENABLED.
    state = optional(string)

    # Configuration for customer-managed query result storage in S3: location,
    # encryption, and cross-account access control. When omitted (and
    # managed_query_results is not set), queries must specify their own result
    # location or fall back to the AWS account-level Athena settings.
    #
    # Mutually exclusive with managed_query_results.output_location by AWS's own
    # rule: a workgroup stores results either in your bucket or in AWS-managed
    # storage, never both.
    result_configuration = optional(object({
      # S3 URI where query results are stored, e.g., "s3://my-bucket/athena-results/".
      # Include the trailing slash. When omitted, queries must specify their own
      # result location.
      #
      # This is a plain string (not StringValueOrRef) because it's an S3 URI with a
      # user-defined path prefix, not a direct resource identifier.
      output_location = optional(string, "")

      # Server-side encryption mode for query results written to S3.
      #
      # Valid values:
      # - "SSE_S3"   — Amazon S3-managed encryption keys (no additional cost).
      # - "SSE_KMS"  — AWS KMS-managed key. Requires kms_key_arn. Provides key
      #                rotation control and CloudTrail audit of key usage.
      # - "CSE_KMS"  — Client-side encryption with AWS KMS key. Requires kms_key_arn.
      #                Data is encrypted before leaving the Athena service.
      #
      # When omitted, query results are not encrypted (unless
      # enable_minimum_encryption_configuration is true at the workgroup level,
      # which enforces at least SSE_S3).
      encryption_option = optional(string, "")

      # KMS key ARN for encrypting query results. Required when encryption_option is
      # SSE_KMS or CSE_KMS. Must not be set for SSE_S3 or when encryption is
      # disabled.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_arn = optional(string, "")

      # AWS account ID that owns the S3 bucket used for query results. Set this when
      # the output_location bucket belongs to a different AWS account than the one
      # running the Athena workgroup. Ensures Athena applies the correct bucket
      # ownership controls.
      expected_bucket_owner = optional(string, "")

      # S3 ACL option applied to query result objects. The only valid value is
      # "BUCKET_OWNER_FULL_CONTROL", which grants the bucket owner full control of
      # result objects. Useful for cross-account scenarios where the query executor
      # writes to a bucket owned by another account.
      #
      # When omitted, S3 default ACL behavior applies.
      s3_acl_option = optional(string, "")
    }))

    # AWS-managed query result storage. When this block is present, Athena stores
    # query results in storage that AWS owns and operates -- no S3 bucket to
    # create, secure, or lifecycle. Results are retained for 24 hours and are
    # retrievable only through Athena APIs (GetQueryResults), which is exactly
    # what most programmatic and BI-driven workloads need.
    #
    # Choose this over result_configuration when nothing downstream reads the
    # result files directly from S3. The two are mutually exclusive: AWS rejects
    # a workgroup that sets both managed results and an S3 output_location.
    managed_query_results = optional(object({
      # KMS key that encrypts results in AWS-managed storage. When omitted, AWS
      # encrypts managed results with an AWS-owned key -- already encrypted at
      # rest; supply a key only when your compliance posture requires
      # customer-controlled key rotation and CloudTrail audit of key usage.
      #
      # Unlike the other three KMS fields on this workgroup, this one accepts a
      # full key ARN ONLY -- the provider rejects "alias/..." forms here at plan
      # time.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key = optional(string, "")

      # Whether AWS-managed result storage is on. Unset means on: declaring the
      # block has always meant enabling it, and this switch lets a manifest say
      # the opposite out loud while keeping the block's settings in place.
      enabled = optional(bool)
    }))

    # Maximum number of bytes a single query is allowed to scan. Queries that
    # exceed this limit are cancelled automatically. This is the primary cost
    # control mechanism for Athena workgroups.
    #
    # Must be 0 (no limit, the default) or at least 10485760 (10 MB). AWS
    # enforces this minimum to prevent trivially small limits that would break
    # most queries.
    #
    # Recommended for production: set to a reasonable ceiling based on your
    # dataset sizes (e.g., 10737418240 for 10 GB).
    bytes_scanned_cutoff_per_query = optional(number, 0)

    # When true (the default), workgroup settings override client-side settings
    # for result location, encryption, and other configuration. Individual queries
    # cannot override these values.
    #
    # When false, queries can override workgroup settings. Useful for development
    # workgroups where engineers need flexibility, but not recommended for
    # production where consistent encryption and result locations are required.
    enforce_workgroup_configuration = optional(bool)

    # When true, Athena publishes query execution metrics (data scanned, execution
    # time, etc.) to CloudWatch. Useful for monitoring query performance and cost
    # trends across the workgroup.
    publish_cloudwatch_metrics_enabled = optional(bool)

    # When true, the requester pays for data access charges when querying data in
    # requester-pays S3 buckets. Default is false (the bucket owner pays).
    requester_pays_enabled = optional(bool, false)

    # When true, enforces a minimum encryption level (at least SSE_S3) for all
    # query results written by this workgroup. Queries that do not specify
    # encryption will default to SSE_S3.
    #
    # Useful as a compliance guardrail to ensure no query results are ever written
    # unencrypted, even if result_configuration.encryption_option is not set.
    enable_minimum_encryption_configuration = optional(bool, false)

    # The Athena engine version to use for queries in this workgroup. Leave empty
    # or set to "AUTO" (the default) to use the latest available engine version.
    #
    # Pinning to a specific version (e.g., "Athena engine version 3") is useful
    # when you need to control query behavior across engine upgrades. The actual
    # engine version in use is available in the effective_engine_version output.
    #
    # Valid values depend on the AWS region and change over time as AWS releases
    # new engine versions. AWS validates at apply time.
    selected_engine_version = optional(string, "")

    # IAM role ARN assumed by the workgroup for Apache Spark workloads and for
    # IAM Identity Center-enabled workgroups. Standard Athena SQL workgroups do
    # not need this field.
    #
    # For Spark, the role must have permissions to read from S3 data sources,
    # write results, and access the AWS Glue Data Catalog.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    execution_role = optional(string, "")

    # KMS key that encrypts customer content stored by Athena for this workgroup:
    # Spark notebook cells, session data, and saved calculation results. Relevant
    # only for Spark-enabled workgroups; SQL query RESULTS are encrypted through
    # result_configuration / managed_query_results instead.
    #
    # Accepts a KMS key ARN (the default reference path) or a key alias.
    # When omitted, Athena encrypts notebook content with an AWS-owned key.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    customer_content_encryption_kms_key = optional(string, "")

    # AWS IAM Identity Center integration. Enables trusted identity propagation
    # so queries run as the workforce identity of the console/IDE user instead of
    # a shared IAM role -- per-user auditing in CloudTrail and per-user S3 Access
    # Grants become possible. Create-time setting on the workgroup.
    identity_center = optional(object({
      # Enables IAM Identity Center integration for this workgroup. Queries then
      # execute under the propagated workforce identity of the signed-in user,
      # making per-user auditing and per-user data grants possible.
      enable_identity_center = optional(bool, false)

      # ARN of the IAM Identity Center instance to integrate with, e.g.
      # "arn:aws:sso:::instance/ssoins-...". Find it in the Identity Center
      # console or via `aws sso-admin list-instances`.
      identity_center_instance_arn = optional(string, "")
    }))

    # Amazon S3 Access Grants integration for query results. When enabled, Athena
    # obtains result-bucket credentials from S3 Access Grants (scoped to the
    # calling identity) instead of the workgroup role's static S3 permissions --
    # the fine-grained-access companion to the identity_center block.
    s3_access_grants = optional(object({
      # Turns S3 Access Grants on for query results in this workgroup.
      enable_s3_access_grants = optional(bool, false)

      # Authentication mode used when requesting credentials from S3 Access
      # Grants. The only value AWS currently supports is "DIRECTORY_IDENTITY"
      # (the propagated IAM Identity Center identity), which is why this block
      # pairs with identity_center.
      authentication_type = optional(string, "")

      # When true, Athena creates a per-user prefix under the result location so
      # each user's results are isolated by grant, e.g.
      # s3://bucket/results/${user}/. Recommended with DIRECTORY_IDENTITY so
      # users can only read their own query results.
      create_user_level_prefix = optional(bool, false)
    }))

    # Log delivery for queries and Spark sessions executed in this workgroup.
    # Each destination arm is enabled by its presence: CloudWatch Logs for
    # searchable operational logs, S3 for cheap long-term archive, and
    # Athena-managed storage for zero-setup retention.
    monitoring = optional(object({
      # Delivers logs to a CloudWatch Logs log group -- the destination to pick
      # when logs must be searchable, alarmable, or subscription-filtered.
      cloud_watch_logging = optional(object({
        # Name of the CloudWatch Logs log group to publish to. When omitted, Athena
        # publishes to its service-default log group. 1-512 characters:
        # alphanumeric, periods, underscores, hyphens, and slashes.
        log_group = optional(string, "")

        # Prefix for the log stream names Athena creates inside the log group.
        # Useful when several workgroups share one log group. Same character rules
        # as log_group.
        log_stream_name_prefix = optional(string, "")

        # Which logs to publish, keyed by worker type. Keys are Spark worker types
        # such as "SPARK_DRIVER" and "SPARK_EXECUTOR"; values are the log streams
        # to deliver for that worker, such as "STDOUT" and "STDERR". When omitted,
        # Athena publishes its default set.
        log_types = optional(list(object({
          # Worker type key, e.g. "SPARK_DRIVER" or "SPARK_EXECUTOR".
          key = string

          # Log streams to publish for the worker, e.g. "STDOUT", "STDERR".
          values = list(string)
        })), [])

        # Whether this destination is on. Unset means on: declaring the block has
        # always meant enabling it, and this switch lets a manifest say the opposite
        # out loud while keeping the block's settings in place.
        enabled = optional(bool)
      }))

      # Delivers logs to Athena-managed storage: zero infrastructure to create,
      # retrievable through the Athena console/APIs.
      managed_logging = optional(object({
        # KMS key that encrypts logs in managed storage. Accepts a key ARN (the
        # default reference path) or an alias. When omitted, AWS encrypts the logs
        # with an AWS-owned key.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key = optional(string, "")

        # Whether this destination is on. Unset means on: declaring the block has
        # always meant enabling it, and this switch lets a manifest say the opposite
        # out loud while keeping the block's settings in place.
        enabled = optional(bool)
      }))

      # Delivers logs to an S3 location you own -- the cheap long-term archive
      # destination.
      s3_logging = optional(object({
        # S3 URI where logs are delivered, e.g. "s3://my-log-archive/athena/".
        # Plain string (not a reference) because it is a URI with a user-defined
        # path prefix, not a direct resource identifier.
        log_location = optional(string, "")

        # KMS key that encrypts delivered log objects. Accepts a key ARN (the
        # default reference path) or an alias. When omitted, objects use the
        # bucket's default encryption.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key = optional(string, "")

        # Whether this destination is on. Unset means on: declaring the block has
        # always meant enabling it, and this switch lets a manifest say the opposite
        # out loud while keeping the block's settings in place.
        enabled = optional(bool)
      }))
    }))

    # When true, all named queries and prepared statements associated with the
    # workgroup are deleted when the workgroup is destroyed. When false (the
    # default), destroying a workgroup that contains named queries or prepared
    # statements will fail.
    force_destroy = optional(bool, false)
  })
}
