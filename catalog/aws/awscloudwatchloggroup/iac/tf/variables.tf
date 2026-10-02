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
  description = "AwsCloudwatchLogGroup specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Number of days to retain log events. After this period, log events are
    # automatically deleted. Set to 0 (the default) to retain log events
    # indefinitely — they will never expire.
    #
    # AWS only accepts specific values for this field. Any value not in the
    # allowed set will be rejected by the AWS API.
    #
    # Allowed values: 0 (never expire), 1, 3, 5, 7, 14, 30, 60, 90, 120, 150,
    # 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653.
    #
    # Recommended: Set an explicit retention for cost control. Indefinite retention
    # (0) accumulates storage costs over time.
    retention_in_days = optional(number, 0)

    # ARN of the KMS key to use for encrypting log data at rest. When omitted,
    # CloudWatch Logs uses its default server-side encryption (SSE-CWL).
    #
    # Customer-managed KMS keys provide:
    # - Key rotation control
    # - Cross-account access via key policy
    # - CloudTrail audit trail of log data access
    # - Compliance with regulations requiring customer-controlled encryption keys
    #
    # The KMS key must be in the same region as the log group, and its key policy
    # must allow the CloudWatch Logs service principal
    # (`logs.<region>.amazonaws.com`) to use the key. Associating or
    # disassociating the key updates the log group in place (no replacement).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # The class of the log group, which determines pricing, feature availability,
    # and data durability characteristics.
    #
    # Valid values:
    # - "STANDARD" — Full feature set including metric filters, subscription
    #   filters, Logs Insights, and Contributor Insights. Default when omitted.
    # - "INFREQUENT_ACCESS" — Reduced cost (~50% cheaper storage) with a subset
    #   of features. Supports Logs Insights and managed ingestion to S3, but does
    #   NOT support metric filters, subscription filters, or Contributor Insights.
    #   Best for high-volume logs accessed infrequently (VPC flow logs, CDN access
    #   logs, compliance archives).
    # - "DELIVERY" — Purpose-built for AWS service log delivery (VPC Flow Logs,
    #   CloudTrail, Route53 Resolver). Lowest cost. Retention is managed by AWS
    #   and retention_in_days must not be set.
    #
    # This field is ForceNew: changing it requires replacing the log group.
    log_group_class = optional(string, "")

    # When true, the log group is protected from deletion. Any attempt to delete
    # the log group (including via IaC destroy) will fail until this flag is set
    # to false. Useful for protecting production log groups from accidental
    # deletion.
    #
    # Three states, and the difference matters for turning protection OFF: the
    # provider attribute is Optional+Computed, so an OMITTED value keeps whatever
    # the log group already has — only an EXPLICIT false disables protection.
    # - unset — new groups get AWS's default (unprotected); existing groups keep
    #   their current protection state.
    # - true  — protection enabled.
    # - false — protection explicitly disabled (the only way back off).
    deletion_protection_enabled = optional(bool)

    # Metric filters that extract CloudWatch metrics from log events flowing
    # through this group. Each filter matches events against a pattern and
    # publishes a metric value for every match — the standard way to alarm on
    # "ERROR" counts, parsed latencies, or any signal that lives only in logs.
    #
    # Filters are keyed by `name` within the group; names must be unique.
    # Not supported on INFREQUENT_ACCESS log groups (AWS rejects the call).
    metric_filters = optional(list(object({
      # Name of the metric filter, unique within the log group. Changing the
      # name replaces the filter. Must not contain ':' or '*' (the
      # PutMetricFilter contract).
      name = string

      # Filter pattern that selects which log events produce metric values.
      # An empty pattern matches ALL log events.
      #
      # Pattern syntax supports plain terms ("ERROR"), JSON field matching
      # ({ $.statusCode = 500 }), and space-delimited column matching
      # ([ip, user, ts, request, status=5*, size]).
      # Maximum 1024 characters.
      pattern = optional(string, "")

      # When true, this filter matches against log events AFTER they pass through
      # the log group's transformer (or an account-level transformer managed
      # outside this resource). When false or unset, the filter matches raw
      # ingested events.
      #
      # The provider attribute is Optional+Computed, so an omitted value keeps the
      # filter's existing setting — only an explicit false switches an existing
      # filter back to matching raw events.
      apply_on_transformed_logs = optional(bool)

      # The metric to publish for each matching log event. Required — a filter
      # without a transformation does nothing.
      transformation = object({
        # Name of the CloudWatch metric to publish (e.g. "ErrorCount").
        # Must not contain ':', '*', or '$' (the MetricTransformation contract).
        metric_name = string

        # Namespace for the metric (e.g. "MyApp/Errors"). Custom namespaces keep
        # log-derived metrics separate from AWS service namespaces.
        # Must not contain ':', '*', or '$' (the MetricTransformation contract).
        metric_namespace = string

        # Value to publish for each matching event. Either a literal number
        # ("1" to count occurrences) or a field reference that extracts a numeric
        # value from the matched event (e.g. "$.latencyMs" for JSON patterns, or
        # "$size" for a named column in space-delimited patterns).
        # Maximum 100 characters.
        metric_value = string

        # Value to publish for periods when NO log events match the pattern.
        # Typically "0" so count metrics report zero instead of missing data —
        # which lets alarms use standard missing-data handling. When unset, no
        # value is published for non-matching periods.
        #
        # AWS does not allow a default value on filters that publish dimensions.
        default_value = optional(number)

        # Dimensions to publish with the metric, mapping dimension names to field
        # references in the matched event (e.g. {"ErrorCode": "$.errorCode"}).
        # Maximum 3 dimensions per AWS limit.
        #
        # Every unique dimension-value combination creates a distinct custom
        # metric (billed separately) — keep dimension cardinality low.
        dimensions = optional(map(string), {})

        # Unit for the metric (e.g. "Count", "Milliseconds", "Bytes"). Defaults to
        # "None" when unset. Must be a valid CloudWatch StandardUnit.
        unit = optional(string, "")
      })
    })), [])

    # Real-time subscription filters that stream matching log events to a
    # destination: a Kinesis data stream, a Kinesis Data Firehose delivery
    # stream, or a Lambda function. Use these to fan logs out to analytics
    # pipelines, SIEM tooling, or custom processing.
    #
    # AWS allows at most TWO subscription filters per log group. Names must be
    # unique within the group. Not supported on INFREQUENT_ACCESS log groups.
    subscription_filters = optional(list(object({
      # Name of the subscription filter, unique within the log group. Changing
      # the name replaces the filter. Must not contain ':' or '*' (the
      # PutSubscriptionFilter contract; the provider validates length only —
      # the character rule is AWS's own).
      name = string

      # ARN of the destination that receives matching log events. Supported
      # destinations:
      # - Kinesis data stream (same-account) — reference an AwsKinesisStream
      # - Kinesis Data Firehose delivery stream — reference an AwsKinesisFirehose
      # - Lambda function — reference an AwsLambda
      # - Cross-account CloudWatch Logs destination (by literal ARN)
      #
      # No default kind is set because all destination types are equally common —
      # reference the specific resource kind's ARN output, or provide a literal ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      destination_arn = string

      # Filter pattern that selects which log events are delivered. An empty
      # pattern delivers ALL log events. Same syntax as metric filter patterns.
      # Maximum 1024 characters.
      filter_pattern = optional(string, "")

      # ARN of the IAM role that grants CloudWatch Logs permission to put records
      # to the destination. REQUIRED for Kinesis stream and Firehose destinations
      # (the role must trust `logs.amazonaws.com` and allow `kinesis:PutRecord` /
      # `firehose:PutRecord` on the destination). NOT used for Lambda destinations —
      # those authorize via a Lambda resource-based permission instead.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = optional(string, "")

      # How log data is distributed to Kinesis stream destinations:
      # - "ByLogStream" — (default) events from the same log stream go to the
      #   same shard, preserving per-stream ordering.
      # - "Random" — events are spread across shards for maximum throughput,
      #   at the cost of ordering.
      #
      # Only meaningful for Kinesis stream destinations.
      distribution = optional(string, "")

      # System fields to include with each delivered log event. Supported values:
      # "@aws.account" (the source account ID), "@aws.region" (the source
      # region), and "@source.log" (the source log group and stream). Useful when
      # a central destination aggregates logs from many accounts/regions/groups.
      emit_system_fields = optional(list(string), [])

      # When true, this filter matches and delivers log events AFTER they pass
      # through the log group's transformer (or an account-level transformer
      # managed outside this resource). When false or unset, the filter operates
      # on raw ingested events.
      #
      # The provider attribute is Optional+Computed, so an omitted value keeps the
      # filter's existing setting — only an explicit false switches an existing
      # filter back to raw events.
      apply_on_transformed_logs = optional(bool)
    })), [])

    # CloudWatch Logs data protection policy for this log group, as a JSON
    # policy document. Data protection audits and masks sensitive data (PII
    # such as email addresses, credit card numbers, or custom identifiers) in
    # log events as they are ingested.
    #
    # The document follows the CloudWatch Logs data protection policy schema:
    # a `Name`, `Version: "2021-06-01"`, and a `Statement` list carrying one
    # "Audit" operation statement and one "Deidentify" (masking) statement over
    # the same data identifiers. Audit findings destinations (CloudWatch Logs,
    # S3, Firehose) are optional — masking works without them.
    #
    # One policy per log group. Masked data is visible only to principals with
    # the `logs:Unmask` permission.
    data_protection_policy = optional(any)

    # CloudWatch Logs field index policy for this log group, as a JSON policy
    # document with a `Fields` array of log-event field names to index
    # (e.g. {"Fields": ["requestId", "userId"]}).
    #
    # Field indexes make Logs Insights queries that filter on the indexed
    # fields faster and cheaper: matching scans skip events where the indexed
    # value cannot match. Index up to 20 fields per policy. One policy per log
    # group; an account-level policy (managed outside this resource) can also
    # apply — the two merge at query time.
    field_index_policy = optional(any)

    # Names of log streams to pre-create inside this log group. Most log streams
    # are created at runtime by the writing agent or AWS service and should NOT
    # be listed here — pre-create a stream only when something depends on the
    # stream existing before the first write (an agent configured with a fixed
    # stream name, IAM policies scoped to specific stream ARNs, or tooling that
    # writes with `sequenceToken` semantics).
    #
    # Stream names must be 1-512 characters and must not contain ':' or '*'
    # (the CreateLogStream contract). Streams are deleted with the group.
    log_streams = optional(list(string), [])

    # Log transformer for this log group: an ordered pipeline of 1-20 processors
    # that parses and reshapes every log event at ingestion time. The transformed
    # form is what Logs Insights queries, metric filters, and subscription
    # filters see (the latter two only when their `apply_on_transformed_logs` is
    # set). One transformer per log group; supported ONLY on STANDARD class log
    # groups (the PutTransformer contract).
    #
    # The first processor must be a parser (parse_json, grok, csv,
    # parse_key_value, or one of the vended-log parsers). An account-level
    # transformer (managed outside this resource) can also exist; when both
    # apply, the log-group-level transformer wins and the account-level one is
    # ignored.
    transformer = optional(object({
      # The processor pipeline, applied in order. Each entry configures exactly
      # one processor.
      processors = list(object({
        # Adds new key-value pairs to the log event. Single-use per transformer.
        add_keys = optional(object({
          # Keys to add. 1-5 entries.
          entries = list(object({
            # Key of the new entry. 1-128 characters. Use dot notation for nested
            # fields (e.g. "metadata.environment").
            key = string

            # Value of the new entry. 1-256 characters.
            value = string

            # When true, overwrites the value if the key already exists in the log
            # event. Defaults to false (existing values win).
            overwrite_if_exists = optional(bool, false)
          }))
        }))

        # Copies values from existing keys to new keys. Single-use per transformer.
        copy_value = optional(object({
          # Values to copy. 1-5 entries.
          entries = list(object({
            # Key to copy from. 1-128 characters.
            source = string

            # Key to copy the value to. 1-128 characters.
            target = string

            # When true, overwrites the value if the target key already exists.
            # Defaults to false.
            overwrite_if_exists = optional(bool, false)
          }))
        }))

        # Parses comma-separated values into columns. Counts as a parser (may be
        # the pipeline's first processor). Up to 5 per transformer.
        csv = optional(object({
          # Names for the parsed columns. When omitted, default names
          # (column_1, column_2, ...) are used. Up to 100 names, each 1-128
          # characters.
          columns = optional(list(string), [])

          # Character separating columns in the source value. Defaults to ",".
          # 1-2 characters.
          delimiter = optional(string, "")

          # Character used as a text qualifier for a single column of data.
          # Defaults to '"'. Exactly 1 character.
          quote_character = optional(string, "")

          # Path to the field to parse. When omitted, the whole @message is
          # processed. 1-128 characters.
          source = optional(string, "")
        }))

        # Converts a datetime string into a target format.
        date_time_converter = optional(object({
          # Key holding the datetime string to convert. 1-128 characters.
          source = string

          # Key to store the converted result in. 1-128 characters.
          target = string

          # Patterns to match against the source value (Java DateTimeFormatter
          # syntax, e.g. "dd/MMM/yyyy:HH:mm:ss" — or "epoch" for epoch timestamps).
          # 1-5 patterns.
          match_patterns = list(string)

          # Locale of the source field (e.g. "en-US"). Defaults to locale.ROOT.
          locale = optional(string, "")

          # Time zone of the source field (e.g. "America/Los_Angeles"). Defaults
          # to UTC.
          source_timezone = optional(string, "")

          # Datetime format for the converted value in the target field. Defaults to
          # "yyyy-MM-dd'T'HH:mm:ss.SSS'Z". 1-64 characters.
          target_format = optional(string, "")

          # Time zone of the target field. Defaults to UTC.
          target_timezone = optional(string, "")
        }))

        # Deletes keys from the log event.
        delete_keys = optional(object({
          # Keys to delete. 1-5 keys, each non-empty.
          with_keys = list(string)
        }))

        # Parses unstructured text with a grok pattern. Counts as a parser.
        # Single-use per transformer.
        grok = optional(object({
          # Grok pattern to match against the log event (e.g.
          # "%{COMMONAPACHELOG}" or a composition of named grok patterns).
          # 1-512 characters.
          match = string

          # Path to the field to parse. When omitted, the whole @message is
          # processed. 1-128 characters.
          source = optional(string, "")
        }))

        # Converts a list of key-value objects into a map.
        list_to_map = optional(object({
          # Key of the field holding the list of objects to convert.
          # 1-128 characters.
          source = string

          # Field within each source object whose value becomes a key in the
          # generated map. 1-128 characters.
          key = string

          # Field within each source object whose value is placed into the map's
          # values. When omitted, the whole source object becomes the value.
          # 1-128 characters.
          value_key = optional(string, "")

          # Key of the field that will hold the generated map. When omitted, the
          # map is placed under the root node. 1-128 characters.
          target = optional(string, "")

          # When true, lists of values in the generated map are flattened into
          # single items using flattened_element.
          flatten = optional(bool, false)

          # Which element to keep when flattening: "first" or "last". Required when
          # flatten is true.
          flattened_element = optional(string, "")
        }))

        # Converts the values of the given keys to lowercase.
        lower_case_string = optional(object({
          # Keys whose values the operation applies to. 1-10 keys, each non-empty.
          with_keys = list(string)
        }))

        # Moves values from one key to another.
        move_keys = optional(object({
          # Keys to move. 1-5 entries.
          entries = list(object({
            # Key to move. 1-128 characters.
            source = string

            # Key to move the value to. 1-128 characters.
            target = string

            # When true, overwrites the value if the target key already exists.
            # Defaults to false.
            overwrite_if_exists = optional(bool, false)
          }))
        }))

        # Parses CloudFront vended access logs into JSON fields. Must be the first
        # processor when present; single-use per transformer.
        parse_cloudfront = optional(object({
          # Source field to parse. The only allowed value is "@message"; when
          # omitted, the whole log message is processed.
          source = optional(string, "")
        }))

        # Parses JSON-format log events. Counts as a parser. Up to 5 per
        # transformer.
        parse_json = optional(object({
          # Path to the field to parse. Defaults to "@message". 1-128 characters.
          source = optional(string, "")

          # Location to put the parsed key-value pairs. When omitted, they are
          # placed under the root node. 1-128 characters.
          destination = optional(string, "")
        }))

        # Parses a field into key-value pairs. Counts as a parser. Up to 5 per
        # transformer.
        parse_key_value = optional(object({
          # Path to the field to parse. Defaults to "@message". 1-128 characters.
          source = optional(string, "")

          # Destination field for the extracted key-value pairs. 1-128 characters.
          destination = optional(string, "")

          # Delimiter between key-value pairs in the source (e.g. ";"). Defaults
          # to "&". 1-128 characters.
          field_delimiter = optional(string, "")

          # Delimiter between the key and the value within a pair (e.g. ":").
          # Defaults to "=". 1-128 characters.
          key_value_delimiter = optional(string, "")

          # Prefix added to all transformed keys. 1-128 characters.
          key_prefix = optional(string, "")

          # Value inserted when a pair cannot be split successfully.
          # 1-128 characters.
          non_match_value = optional(string, "")

          # When true, overwrites the value if the destination key already exists.
          # Defaults to false.
          overwrite_if_exists = optional(bool, false)
        }))

        # Parses RDS for PostgreSQL vended logs into JSON fields. Must be the first
        # processor when present; single-use per transformer.
        parse_postgres = optional(object({
          # Source field to parse. The only allowed value is "@message"; when
          # omitted, the whole log message is processed.
          source = optional(string, "")
        }))

        # Parses Route 53 vended logs into JSON fields. Must be the first processor
        # when present; single-use per transformer.
        parse_route53 = optional(object({
          # Source field to parse. The only allowed value is "@message"; when
          # omitted, the whole log message is processed.
          source = optional(string, "")
        }))

        # Converts log events into Open Cybersecurity Schema Framework (OCSF)
        # events. Must be the first processor when present; single-use per
        # transformer.
        parse_to_ocsf = optional(object({
          # Service or process producing the log events. Valid values:
          # "CloudTrail", "Route53Resolver", "VPCFlow", "EKSAudit", "AWSWAF".
          event_source = string

          # OCSF schema version for the transformed events. Valid values: "V1.1",
          # "V1.5".
          ocsf_version = string

          # Source field to parse. The only allowed value is "@message"; when
          # omitted, the whole log message is processed.
          source = optional(string, "")
        }))

        # Parses Amazon VPC flow logs into JSON fields. Must be the first processor
        # when present; single-use per transformer.
        parse_vpc = optional(object({
          # Source field to parse. The only allowed value is "@message"; when
          # omitted, the whole log message is processed.
          source = optional(string, "")
        }))

        # Parses AWS WAF vended logs into JSON fields. Must be the first processor
        # when present; single-use per transformer.
        parse_waf = optional(object({
          # Source field to parse. The only allowed value is "@message"; when
          # omitted, the whole log message is processed.
          source = optional(string, "")
        }))

        # Renames keys in the log event.
        rename_keys = optional(object({
          # Keys to rename. 1-5 entries.
          entries = list(object({
            # Key to rename. 1-128 characters.
            key = string

            # New name for the key. 1-128 characters.
            rename_to = string

            # When true, overwrites the value if the new key name already exists.
            # Defaults to false.
            overwrite_if_exists = optional(bool, false)
          }))
        }))

        # Splits field values into arrays using a delimiter.
        split_string = optional(object({
          # Fields to split. 1-10 entries.
          entries = list(object({
            # Key of the field to split. 1-128 characters.
            source = string

            # Separator characters to split the string on. 1-128 characters.
            delimiter = string
          }))
        }))

        # Replaces regex matches in field values with a replacement string.
        substitute_string = optional(object({
          # Fields to substitute. 1-10 entries.
          entries = list(object({
            # Key of the field to modify. 1-128 characters.
            source = string

            # Regular expression whose matches are replaced. 1-128 characters.
            from = string

            # Replacement string for each match. 1-128 characters.
            to = string
          }))
        }))

        # Trims leading and trailing whitespace from the given keys' values.
        trim_string = optional(object({
          # Keys whose values the operation applies to. 1-10 keys, each non-empty.
          with_keys = list(string)
        }))

        # Converts field values to a different data type.
        type_converter = optional(object({
          # Fields to convert. 1-5 entries.
          entries = list(object({
            # Key whose value is converted. 1-128 characters.
            key = string

            # Type to convert the value to. Valid values: "boolean", "integer",
            # "double", "string".
            type = string
          }))
        }))

        # Converts the values of the given keys to uppercase.
        upper_case_string = optional(object({
          # Keys whose values the operation applies to. 1-10 keys, each non-empty.
          with_keys = list(string)
        }))
      }))
    }))
  })
}
