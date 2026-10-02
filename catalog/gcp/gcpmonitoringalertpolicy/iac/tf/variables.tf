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
  description = "GcpMonitoringAlertPolicy specification"
  type = object({
    # The GCP project that owns the alert policy. Can be a literal project ID
    # or a reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Human-friendly name shown in the console, in incidents, and in
    # notifications. Defaults to metadata.name when left empty (the GCP API
    # requires a display name).
    display_name = optional(string, "")

    # How the results of multiple conditions combine into one incident
    # decision:
    #   AND -- an incident opens only while ALL conditions are met
    #   OR  -- an incident opens while ANY condition is met
    #   AND_WITH_MATCHING_RESOURCE -- AND, but only when the conditions
    #          trigger on the SAME monitored resource (the strictest form)
    # Required by the GCP API even for single-condition policies (use OR or
    # AND — they are equivalent with one condition).
    combiner = string

    # Severity reported on incidents this policy opens: CRITICAL, ERROR, or
    # WARNING. Empty leaves the severity unset in GCP (incidents show no
    # severity level).
    severity = optional(string, "")

    # Whether the policy is evaluated (default true). A disabled policy keeps
    # its configuration but opens no incidents — the safe way to silence a
    # noisy rule while tuning it. Both IaC engines send the value explicitly
    # so behavior is identical regardless of engine.
    enabled = optional(bool)

    # The conditions this policy evaluates (1 to 6 — the GCP API's own
    # bounds), combined by `combiner`.
    conditions = list(object({
      # Name shown for this condition in the console and in incident detail.
      display_name = string

      # Alert when a metric crosses a threshold — the workhorse condition type
      # (CPU above 80%, error rate above 1%, uptime-check failures).
      condition_threshold = optional(object({
        # A monitoring filter (https://cloud.google.com/monitoring/api/v3/filters)
        # selecting the time series to evaluate, e.g.
        #   metric.type="compute.googleapis.com/instance/cpu/utilization" AND
        #   resource.type="gce_instance"
        # The API requires a filter that names a metric.type.
        filter = string

        # How the aggregated value compares to the threshold to trigger:
        # COMPARISON_GT / _GE / _LT / _LE / _EQ / _NE. GT ("above threshold") is
        # by far the most common.
        comparison = string

        # The threshold value compared against the aggregated metric.
        threshold_value = optional(number, 0)

        # How long the comparison must hold before the condition triggers, as a
        # duration in seconds. Must be a multiple of 60s ("0s" triggers on a
        # single violating point; "300s" requires five sustained minutes — the
        # usual flap guard). The GCP API rejects non-minute-aligned values.
        duration = string

        # How raw time series are aligned and combined before comparison. Most
        # metric-type conditions need at least one aggregation (e.g. align
        # ALIGN_MEAN over the alignment_period) — without one, GCP compares every
        # raw point of every series.
        aggregations = optional(list(object({
          # The window each series is aligned over, as a duration in seconds
          # (minute-aligned, e.g. "60s", "300s").
          alignment_period = optional(string, "")

          # How points within the window combine per series: ALIGN_NONE,
          # ALIGN_DELTA, ALIGN_RATE, ALIGN_INTERPOLATE, ALIGN_NEXT_OLDER,
          # ALIGN_MIN, ALIGN_MAX, ALIGN_MEAN, ALIGN_COUNT, ALIGN_SUM, ALIGN_STDDEV,
          # ALIGN_COUNT_TRUE, ALIGN_COUNT_FALSE, ALIGN_FRACTION_TRUE,
          # ALIGN_PERCENTILE_99, ALIGN_PERCENTILE_95, ALIGN_PERCENTILE_50,
          # ALIGN_PERCENTILE_05, ALIGN_PERCENT_CHANGE.
          per_series_aligner = optional(string, "")

          # How aligned series combine ACROSS series: REDUCE_NONE, REDUCE_MEAN,
          # REDUCE_MIN, REDUCE_MAX, REDUCE_SUM, REDUCE_STDDEV, REDUCE_COUNT,
          # REDUCE_COUNT_TRUE, REDUCE_COUNT_FALSE, REDUCE_FRACTION_TRUE,
          # REDUCE_PERCENTILE_99, REDUCE_PERCENTILE_95, REDUCE_PERCENTILE_50,
          # REDUCE_PERCENTILE_05.
          cross_series_reducer = optional(string, "")

          # Resource/metric label keys that survive cross-series reduction — series
          # sharing these label values reduce together (e.g. group error rates by
          # resource.label.zone).
          group_by_fields = optional(list(string), [])
        })), [])

        # For ratio conditions: the filter selecting the DENOMINATOR time series
        # (the condition then evaluates numerator/denominator against the
        # threshold, e.g. error requests / total requests).
        denominator_filter = optional(string, "")

        # Aggregations applied to the denominator series. The API requires the
        # denominator to be aligned identically to the numerator for the ratio to
        # be meaningful.
        denominator_aggregations = optional(list(object({
          # The window each series is aligned over, as a duration in seconds
          # (minute-aligned, e.g. "60s", "300s").
          alignment_period = optional(string, "")

          # How points within the window combine per series: ALIGN_NONE,
          # ALIGN_DELTA, ALIGN_RATE, ALIGN_INTERPOLATE, ALIGN_NEXT_OLDER,
          # ALIGN_MIN, ALIGN_MAX, ALIGN_MEAN, ALIGN_COUNT, ALIGN_SUM, ALIGN_STDDEV,
          # ALIGN_COUNT_TRUE, ALIGN_COUNT_FALSE, ALIGN_FRACTION_TRUE,
          # ALIGN_PERCENTILE_99, ALIGN_PERCENTILE_95, ALIGN_PERCENTILE_50,
          # ALIGN_PERCENTILE_05, ALIGN_PERCENT_CHANGE.
          per_series_aligner = optional(string, "")

          # How aligned series combine ACROSS series: REDUCE_NONE, REDUCE_MEAN,
          # REDUCE_MIN, REDUCE_MAX, REDUCE_SUM, REDUCE_STDDEV, REDUCE_COUNT,
          # REDUCE_COUNT_TRUE, REDUCE_COUNT_FALSE, REDUCE_FRACTION_TRUE,
          # REDUCE_PERCENTILE_99, REDUCE_PERCENTILE_95, REDUCE_PERCENTILE_50,
          # REDUCE_PERCENTILE_05.
          cross_series_reducer = optional(string, "")

          # Resource/metric label keys that survive cross-series reduction — series
          # sharing these label values reduce together (e.g. group error rates by
          # resource.label.zone).
          group_by_fields = optional(list(string), [])
        })), [])

        # Alert on the PREDICTED value instead of the current one: the condition
        # triggers when GCP forecasts the threshold will be crossed within this
        # horizon (a duration, minimum "3600s"). The capacity-planning form of a
        # threshold alert.
        forecast_options = optional(object({
          # How far ahead GCP forecasts, as a duration in seconds (minimum "3600s"
          # — one hour — per the API).
          forecast_horizon = string
        }))

        # How many time series (count) or what fraction of them (percent) must
        # violate before the condition triggers. Default: any single series.
        trigger = optional(object({
          # Absolute number of violating series.
          count = optional(number, 0)

          # Percentage of violating series (0-100).
          percent = optional(number, 0)
        }))

        # How the condition evaluates when data stops arriving:
        # EVALUATION_MISSING_DATA_INACTIVE (missing data closes/keeps the
        # incident closed), _ACTIVE (missing data violates — the paranoid
        # setting), or _NO_OP (missing data changes nothing, the API default).
        evaluation_missing_data = optional(string, "")
      }))

      # Alert when a metric stops reporting for a duration — the "silence is
      # failure" condition (a heartbeat metric going quiet).
      condition_absent = optional(object({
        # A monitoring filter selecting the time series whose ABSENCE triggers
        # the condition.
        filter = string

        # How long the data must be absent before triggering, as a duration in
        # seconds (minute-aligned, maximum 24 hours per the API — e.g. "300s").
        duration = string

        # Alignment applied to the series before absence is judged.
        aggregations = optional(list(object({
          # The window each series is aligned over, as a duration in seconds
          # (minute-aligned, e.g. "60s", "300s").
          alignment_period = optional(string, "")

          # How points within the window combine per series: ALIGN_NONE,
          # ALIGN_DELTA, ALIGN_RATE, ALIGN_INTERPOLATE, ALIGN_NEXT_OLDER,
          # ALIGN_MIN, ALIGN_MAX, ALIGN_MEAN, ALIGN_COUNT, ALIGN_SUM, ALIGN_STDDEV,
          # ALIGN_COUNT_TRUE, ALIGN_COUNT_FALSE, ALIGN_FRACTION_TRUE,
          # ALIGN_PERCENTILE_99, ALIGN_PERCENTILE_95, ALIGN_PERCENTILE_50,
          # ALIGN_PERCENTILE_05, ALIGN_PERCENT_CHANGE.
          per_series_aligner = optional(string, "")

          # How aligned series combine ACROSS series: REDUCE_NONE, REDUCE_MEAN,
          # REDUCE_MIN, REDUCE_MAX, REDUCE_SUM, REDUCE_STDDEV, REDUCE_COUNT,
          # REDUCE_COUNT_TRUE, REDUCE_COUNT_FALSE, REDUCE_FRACTION_TRUE,
          # REDUCE_PERCENTILE_99, REDUCE_PERCENTILE_95, REDUCE_PERCENTILE_50,
          # REDUCE_PERCENTILE_05.
          cross_series_reducer = optional(string, "")

          # Resource/metric label keys that survive cross-series reduction — series
          # sharing these label values reduce together (e.g. group error rates by
          # resource.label.zone).
          group_by_fields = optional(list(string), [])
        })), [])

        # How many series (count) or what fraction (percent) must be absent
        # before triggering.
        trigger = optional(object({
          # Absolute number of violating series.
          count = optional(number, 0)

          # Percentage of violating series (0-100).
          percent = optional(number, 0)
        }))
      }))

      # Alert on every log entry matching a filter. Requires
      # alert_strategy.notification_rate_limit (the API's own pairing for
      # log-based policies).
      condition_matched_log = optional(object({
        # A logging filter (https://cloud.google.com/logging/docs/view/logging-query-language)
        # selecting the entries that trigger notifications, e.g.
        #   resource.type="gce_instance" AND severity>=ERROR
        filter = string

        # Extract values from matched entries into incident labels: map label
        # name -> extractor expression, e.g.
        #   { "vm": "EXTRACT(resource.labels.instance_id)" }
        # Extracted labels appear in notifications and documentation variables.
        label_extractors = optional(map(string), {})
      }))

      # Alert on a Monitoring Query Language (MQL) query — for conditions the
      # structured threshold/absent forms cannot express. MQL is deprecated by
      # Google in favor of PromQL; prefer condition_prometheus_query_language
      # for new policies.
      condition_monitoring_query_language = optional(object({
        # The Monitoring Query Language query whose boolean output drives the
        # condition.
        query = string

        # How long the query output must be true before triggering (duration in
        # seconds, minute-aligned).
        duration = string

        # How many series (count) or what fraction (percent) must violate before
        # triggering.
        trigger = optional(object({
          # Absolute number of violating series.
          count = optional(number, 0)

          # Percentage of violating series (0-100).
          percent = optional(number, 0)
        }))

        # How the condition evaluates when data stops arriving (same values as
        # the threshold condition's field).
        evaluation_missing_data = optional(string, "")
      }))

      # Alert on a PromQL query — the expressive form for rate/ratio/quantile
      # conditions and for teams porting Prometheus alert rules.
      condition_prometheus_query_language = optional(object({
        # The PromQL expression. The condition fires while the expression
        # produces any series (Prometheus alert-rule semantics), e.g.
        #   rate(http_requests_total{code=~"5.."}[5m]) > 0.1
        query = string

        # How long the expression must produce output before triggering
        # (Prometheus "for" semantics). A duration in seconds, e.g. "300s";
        # empty means fire immediately.
        duration = optional(string, "")

        # How often the query is evaluated (default "30s"; must be a multiple of
        # 30 seconds per the API).
        evaluation_interval = optional(string, "")

        # Labels added to every incident this condition opens (Prometheus
        # alert-rule labels; values support PromQL template syntax).
        labels = optional(map(string), {})

        # The rule group name this condition belongs to when imported from a
        # Prometheus rule file — carried for round-trip fidelity, no behavioral
        # effect in GCP.
        rule_group = optional(string, "")

        # The alert rule name when imported from a Prometheus rule file. Becomes
        # the incident's alertname label.
        alert_rule = optional(string, "")

        # Skip GCP's validation that the query only references known metrics
        # (default false). Enable only when alerting on metrics that do not exist
        # yet — a typo in the metric name then goes undetected until it never
        # fires.
        disable_metric_validation = optional(bool, false)
      }))

      # Alert on a SQL query against log analytics data, evaluated on a
      # schedule (minutes/hourly/daily) with a row-count or boolean test on the
      # results.
      condition_sql = optional(object({
        # The SQL query to run against log analytics (GoogleSQL). The query's
        # results feed row_count_test or boolean_test.
        query = string

        # Run the query every N minutes.
        minutes = optional(object({
          # Number of minutes between runs (5 to 1440 per the GCP API).
          periodicity = optional(number, 0)
        }))

        # Run the query every N hours.
        hourly = optional(object({
          # Number of hours between runs (1 to 48 per the GCP API).
          periodicity = optional(number, 0)

          # Minute of the hour the run starts (0-59). Optional; GCP picks one when
          # unset.
          minute_offset = optional(number)
        }))

        # Run the query every N days at a fixed time.
        daily = optional(object({
          # Number of days between runs (1 to 31 per the GCP API).
          periodicity = optional(number, 0)

          # Time of day the run starts (UTC). Omit for GCP's default.
          execution_time = optional(object({
            # Hour (0-23).
            hours = optional(number, 0)

            # Minute (0-59).
            minutes = optional(number, 0)

            # Second (0-59).
            seconds = optional(number, 0)

            # Nanoseconds (0-999999999). Kept for API-shape fidelity; alert schedules
            # realistically use whole seconds.
            nanos = optional(number, 0)
          }))
        }))

        # Trigger when the query's ROW COUNT compares against a threshold.
        row_count_test = optional(object({
          # How the row count compares to the threshold (same comparison values as
          # the threshold condition).
          comparison = string

          # The row-count threshold.
          threshold = optional(number, 0)
        }))

        # Trigger when a BOOLEAN COLUMN of the first result row is true.
        boolean_test = optional(object({
          # The BOOL column of the first result row that decides the condition.
          column = string
        }))
      }))
    }))

    # The notification channels to notify when incidents open, close, or gain
    # new violations. Each entry is a channel resource name
    # (projects/{project}/notificationChannels/{id}) — a literal or a
    # reference to a GcpMonitoringNotificationChannel resource (its
    # channel_name output is exactly this value).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    notification_channels = optional(list(string), [])

    # How notifications behave once an incident is open: auto-close timing,
    # rate limiting (log-based policies), and per-channel re-notification.
    alert_strategy = optional(object({
      # Close an incident automatically after its condition has stopped
      # violating for this long (duration in seconds, minimum "1800s" per the
      # API; GCP's default is 7 days).
      auto_close = optional(string, "")

      # Rate limit for LOG-BASED policies (condition_matched_log): at most one
      # notification per period. The API requires this block for log-match
      # conditions and rejects it for metric conditions.
      notification_rate_limit = optional(object({
        # Not more than one notification per period (duration in seconds, e.g.
        # "300s").
        period = optional(string, "")
      }))

      # Per-channel-subset re-notification cadence for open incidents.
      notification_channel_strategy = optional(list(object({
        # The channels this cadence applies to — full channel resource names,
        # each a literal or a reference to a GcpMonitoringNotificationChannel.
        # Every entry must also appear in the policy's notification_channels.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        notification_channel_names = optional(list(string), [])

        # Send reminder notifications for still-open incidents at this interval
        # (duration in seconds).
        renotify_interval = optional(string, "")
      })), [])

      # When to prompt the configured channels beyond incident open/close:
      # OPENED and/or CLOSED.
      notification_prompts = optional(list(string), [])
    }))

    # Troubleshooting content attached to every notification — the runbook
    # the on-call engineer sees. Supports Markdown and ${variable}
    # substitution (e.g. ${resource.label.instance_id}).
    documentation = optional(object({
      # The runbook body (at most 8192 bytes). Supports the mime_type's syntax
      # and ${variable} substitution — write the steps the on-call engineer
      # should take.
      content = optional(string, "")

      # Content format; "text/markdown" (the default) is the only value the
      # API currently accepts.
      mime_type = optional(string, "")

      # Custom subject line for notifications (at most 255 bytes; not all
      # channel types render it).
      subject = optional(string, "")

      # Reference links (runbooks, dashboards, playbooks) attached to
      # notifications — at most 3 per the API.
      links = optional(list(object({
        # Short display text for the link.
        display_name = optional(string, "")

        # The URL (http/https). Supports ${variable} substitution.
        url = optional(string, "")
      })), [])
    }))

    # User labels attached to the policy for organizing and identifying it
    # (maps to the provider's user_labels), merged with Planton's platform
    # labels (which win on key conflicts). Keys and values may contain only
    # lowercase letters, numerals, underscores, and dashes; keys must begin
    # with a letter.
    labels = optional(map(string), {})

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the policy is deleted; its open incidents close and no
    #                further notifications are sent
    #   "PREVENT" -- destroy FAILS; protects production alerting from
    #                accidental teardown
    #   "ABANDON" -- the policy is removed from management but keeps
    #                evaluating (and paging) in GCP
    deletion_policy = optional(string, "")
  })
}
