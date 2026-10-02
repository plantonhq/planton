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
  description = "AwsCloudwatchAlarm specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Arithmetic operation to compare the statistic against the threshold.
    #
    # Standard comparison operators:
    # - "GreaterThanOrEqualToThreshold"
    # - "GreaterThanThreshold"
    # - "LessThanThreshold"
    # - "LessThanOrEqualToThreshold"
    #
    # Anomaly detection operators (use with metric_queries and threshold_metric_id):
    # - "LessThanLowerOrGreaterThanUpperThreshold"
    # - "LessThanLowerThreshold"
    # - "GreaterThanUpperThreshold"
    #
    # Required for simple-metric and metric-query alarms. Must NOT be set for
    # PromQL alarms (`evaluation_criteria`) — the query expresses the condition.
    comparison_operator = optional(string, "")

    # Number of consecutive periods over which the metric is compared to the
    # threshold.
    #
    # Combined with `datapoints_to_alarm` this creates an M-of-N evaluation
    # window. For example, evaluation_periods=5 with datapoints_to_alarm=3
    # means "3 out of the last 5 periods must breach to trigger the alarm."
    #
    # Required (must be at least 1) for simple-metric and metric-query alarms.
    # Must NOT be set for PromQL alarms — pending/recovery periods on the
    # PromQL criteria play that role instead.
    evaluation_periods = optional(number, 0)

    # Number of data points within the evaluation window that must breach the
    # threshold to trigger the alarm. Must be less than or equal to
    # `evaluation_periods`. When omitted, defaults to `evaluation_periods`
    # (every period must breach).
    #
    # Use a value lower than `evaluation_periods` for M-of-N evaluation to
    # reduce false positives caused by transient spikes.
    datapoints_to_alarm = optional(number, 0)

    # Static threshold value. The statistic (or metric math result) is compared
    # against this value using `comparison_operator`.
    #
    # Required for static threshold alarms. Must not be set for anomaly
    # detection alarms (use `threshold_metric_id` instead) or PromQL alarms.
    threshold = optional(number, 0)

    # For anomaly detection alarms, set this to the ID of the
    # ANOMALY_DETECTION_BAND function defined in `metric_queries`. The anomaly
    # band serves as a dynamic threshold instead of a static `threshold` value.
    #
    # Mutually exclusive with `threshold`.
    threshold_metric_id = optional(string, "")

    # How the alarm treats missing data points during evaluation.
    #
    # Valid values:
    # - "missing"      — (Default) Missing data is treated as missing; the alarm
    #                     state does not change. Can cause delayed transitions.
    # - "notBreaching" — Missing data is within the threshold. Best for alarms
    #                     on intermittent metrics (e.g., error count on low-traffic
    #                     services) to avoid false alarms during idle periods.
    # - "breaching"    — Missing data is treated as breaching. Use for metrics
    #                     that must always report (e.g., heartbeat checks).
    # - "ignore"       — The current alarm state is maintained regardless of
    #                     missing data.
    treat_missing_data = optional(string, "")

    # Whether actions execute during alarm state transitions. When unset, AWS
    # defaults to true (actions enabled). Set explicitly to false to suppress
    # actions during maintenance windows or alarm tuning — the alarm still
    # evaluates and changes state, it just does not act.
    actions_enabled = optional(bool)

    # Name of the CloudWatch metric to alarm on (e.g., "CPUUtilization",
    # "4xxErrorRate", "ApproximateNumberOfMessagesVisible").
    #
    # When set, `namespace`, `period`, and one of `statistic` or
    # `extended_statistic` must also be set.
    #
    # Mutually exclusive with `metric_queries` and `evaluation_criteria`.
    metric_name = optional(string, "")

    # Namespace of the metric (e.g., "AWS/EC2", "AWS/ECS", "AWS/SQS",
    # "AWS/ApplicationELB"). Custom namespaces are also supported.
    #
    # Required when `metric_name` is set. Must not start with a colon.
    namespace = optional(string, "")

    # Period in seconds over which the statistic is applied. Defines the
    # granularity of the alarm evaluation.
    #
    # Valid values: 10, 20, 30, or any multiple of 60.
    # High-resolution metrics support 10/20/30-second periods.
    # Standard-resolution metrics require multiples of 60.
    #
    # Required when `metric_name` is set.
    period = optional(number, 0)

    # Standard statistic to apply to the metric.
    #
    # Valid values: "SampleCount", "Average", "Sum", "Minimum", "Maximum".
    #
    # Mutually exclusive with `extended_statistic`. Exactly one must be set
    # when using simple metric mode.
    statistic = optional(string, "")

    # Percentile or extended statistic to apply to the metric. Used for
    # percentile-based alarms (e.g., p95 latency, p99 error rate).
    #
    # Examples: "p95", "p99", "p99.9", "IQM", "TM(10%:90%)"
    #
    # Mutually exclusive with `statistic`. Exactly one must be set when
    # using simple metric mode.
    extended_statistic = optional(string, "")

    # Dimensions that identify the specific metric stream to alarm on.
    # Dimensions narrow the metric to a specific resource or subset.
    #
    # Example for EC2 CPU: {"InstanceId": "i-1234567890abcdef0"}
    # Example for ECS service: {"ClusterName": "prod", "ServiceName": "api"}
    #
    # Mutually exclusive with `metric_queries`.
    dimensions = optional(map(string), {})

    # Unit for the metric. When specified, only data points with a matching unit
    # are used for evaluation. Most alarms omit this and use the metric's
    # published unit. Must be a valid CloudWatch StandardUnit when set.
    unit = optional(string, "")

    # Metric math expressions or multi-metric queries. Use this mode for:
    # - Metric math (e.g., error rate = errors / total * 100)
    # - Anomaly detection (ANOMALY_DETECTION_BAND function)
    # - Cross-account metric monitoring
    # - Multi-metric composite evaluations
    #
    # Up to 20 queries are supported. Exactly one query must set
    # `return_data = true` to serve as the alarm's evaluation signal.
    #
    # Mutually exclusive with simple metric fields (`metric_name`,
    # `namespace`, `period`, `statistic`, `extended_statistic`, `dimensions`)
    # and with `evaluation_criteria`.
    metric_queries = optional(list(object({
      # Unique identifier for this query. Used as a variable name in metric math
      # expressions. Must start with a lowercase letter; valid characters are
      # lowercase letters, digits, and underscores.
      id = string

      # Metric math expression or Metrics Insights query. References other
      # queries by their `id` field.
      #
      # Examples:
      #   "m1/m2*100"                           — error rate percentage
      #   "ANOMALY_DETECTION_BAND(m1, 2)"       — anomaly detection with 2 std devs
      #   "METRICS('AWS/EC2')"                   — Metrics Insights query
      #
      # Mutually exclusive with `metric`.
      expression = optional(string, "")

      # Raw metric definition. Use when this query retrieves a metric from
      # CloudWatch rather than computing a value from other queries.
      #
      # Mutually exclusive with `expression`.
      metric = optional(object({
        # Name of the CloudWatch metric (e.g., "CPUUtilization", "5XXError").
        metric_name = string

        # Namespace of the metric (e.g., "AWS/EC2", "AWS/ApplicationELB").
        # Optional in the CloudWatch API, but virtually always needed to address
        # a metric unambiguously. Must not start with a colon.
        namespace = optional(string, "")

        # Period in seconds for the metric data points. Determines the granularity
        # of the data used in expressions.
        #
        # Valid values: 1, 5, 10, 20, 30, or any multiple of 60.
        # High-resolution metrics support sub-minute periods (1, 5, 10, 20, 30).
        period = optional(number, 0)

        # Statistic to apply — either a standard statistic (SampleCount, Average,
        # Sum, Minimum, Maximum) or a percentile/extended statistic (p95, p99.9,
        # IQM, TM(10%:90%), etc.).
        stat = string

        # Dimensions that identify the specific metric stream.
        dimensions = optional(map(string), {})

        # Unit for the metric. Optional; filters data points to those matching
        # the specified unit.
        unit = optional(string, "")
      }))

      # Human-readable label for this query. Displayed in the CloudWatch console
      # and alarm history. Especially useful for expressions to describe what the
      # computed value represents.
      label = optional(string, "")

      # Override period (in seconds) for this query. When omitted, uses the period
      # from the `metric` definition (if present) or inherits the alarm's period.
      #
      # Valid values: 1, 5, 10, 20, 30, or any multiple of 60.
      period = optional(number, 0)

      # Whether this query's result is used as the alarm's evaluation signal.
      # Exactly one query in the `metric_queries` list must set this to true.
      # Queries with `return_data = false` are intermediate values used by
      # expressions but not directly evaluated by the alarm.
      return_data = optional(bool, false)

      # AWS account ID where the metric is located. Use for cross-account alarms
      # that monitor metrics from a different account.
      account_id = optional(string, "")
    })), [])

    # Actions to execute when the alarm transitions to ALARM state. Each action
    # is an ARN — typically an SNS topic ARN, but can also be an Auto Scaling
    # policy, EC2 automation action, Lambda function, or SSM OpsItem.
    #
    # Maximum 5 actions — AWS's quota is 5 actions per alarm state. (The
    # provider caps only ok_actions and insufficient_data_actions; the missing
    # cap on alarm_actions is provider looseness — the spec holds AWS's
    # contract.)
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    alarm_actions = optional(list(string), [])

    # Actions to execute when the alarm transitions to OK state.
    # Maximum 5 actions.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    ok_actions = optional(list(string), [])

    # Actions to execute when the alarm transitions to INSUFFICIENT_DATA state.
    # Maximum 5 actions.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    insufficient_data_actions = optional(list(string), [])

    # Human-readable description of the alarm's purpose. Include details about
    # what the alarm monitors, expected thresholds, and remediation steps.
    # Maximum 1024 characters.
    alarm_description = optional(string, "")

    # Controls alarm behavior during periods with too few data points for
    # statistically significant percentile calculations.
    #
    # Valid values:
    # - "evaluate" — Always evaluate the alarm regardless of sample count.
    # - "ignore"   — Do not change alarm state during low-sample-count periods.
    #
    # Only meaningful when using percentile statistics (extended_statistic or
    # a percentile stat in metric_queries).
    evaluate_low_sample_count_percentiles = optional(string, "")

    # PromQL evaluation criteria for alarming on an Amazon Managed Service for
    # Prometheus workspace. In this mode the alarm evaluates a PromQL query on
    # a schedule (`evaluation_interval`) and transitions based on the query's
    # firing state, so the threshold-based fields (`comparison_operator`,
    # `evaluation_periods`, `threshold`, `datapoints_to_alarm`, and all
    # simple-metric fields) must NOT be set.
    evaluation_criteria = optional(object({
      # The PromQL criteria to evaluate. Required.
      promql_criteria = object({
        # The PromQL query to evaluate, addressing an Amazon Managed Service for
        # Prometheus workspace. The query must produce a boolean-style firing
        # signal, e.g.:
        #   'avg(rate(node_cpu_seconds_total{mode!="idle"}[5m])) > 0.8'
        # Maximum 10000 characters.
        query = string

        # How long (in seconds) the query must continuously return a firing result
        # before the alarm transitions to ALARM — the PromQL analog of "for:" in
        # Prometheus alerting rules. 0 transitions immediately. Maximum 86400
        # (24 hours). When unset, AWS applies its default pending behavior.
        pending_period = optional(number)

        # How long (in seconds) the query must continuously return a non-firing
        # result before the alarm transitions back to OK — dampens flapping.
        # 0 recovers immediately. Maximum 86400 (24 hours). When unset, AWS
        # applies its default recovery behavior.
        recovery_period = optional(number)
      })
    }))

    # How often (in seconds) the PromQL query is evaluated. Only valid together
    # with `evaluation_criteria`. Valid values: 10, 20, 30, or any multiple
    # of 60. When unset, AWS uses its default evaluation interval.
    evaluation_interval = optional(number, 0)
  })
}
