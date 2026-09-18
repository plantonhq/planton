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
  description = "GcpLogMetric specification"
  type = object({
    # The GCP project whose logs feed the metric. Can be a literal project
    # ID or a reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The metric name — how the metric is addressed from Monitoring as
    # logging.googleapis.com/user/{metric_name}. Defaults to metadata.name.
    # May include forward slashes for namespacing (e.g. "checkout/errors");
    # the API forbids characters that need URL-encoding.
    metric_name = optional(string, "")

    # The logging filter selecting the entries that feed the metric
    # (https://cloud.google.com/logging/docs/view/logging-query-language),
    # e.g. resource.type="cloud_run_revision" AND severity>=ERROR
    filter = string

    # Scope the metric to a specific log BUCKET instead of the project's
    # _Default bucket: the metric then counts entries as they land in that
    # bucket. The full bucket resource name
    # (projects/{p}/locations/{l}/buckets/{b}) — a literal or a reference to
    # a GcpLogBucket resource (its bucket_name output is exactly this
    # value). The bucket must live in the same project as the metric.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    bucket_name = optional(string, "")

    # What this metric measures and how it is meant to be used (shown in
    # the metrics explorer). At most 8000 characters (the provider's own
    # documented cap).
    description = optional(string, "")

    # Pause the metric: a disabled metric keeps its configuration and
    # history but ingests no new points — the safe way to silence a
    # misconfigured extractor while fixing it.
    disabled = optional(bool, false)

    # The shape of the metric this filter produces. Omit for the plain
    # count-of-entries form (GCP then applies its implicit DELTA/INT64
    # descriptor). Required whenever value_extractor, label_extractors, or
    # bucket_options are used.
    metric_descriptor = optional(object({
      # How points accumulate: DELTA (change since the last point — the form
      # log-based counters use), GAUGE (instantaneous value), or CUMULATIVE
      # (running total). Log-based metrics are DELTA except in rare GAUGE
      # cases. Changing the kind REPLACES the metric descriptor.
      metric_kind = string

      # The value type of each point: INT64 for counters, DISTRIBUTION for
      # extracted-value histograms; BOOL, DOUBLE, STRING, and MONEY complete
      # the API's set.
      value_type = string

      # The unit of the values (UCUM syntax, e.g. "ms", "By", "1"). Shown on
      # chart axes.
      unit = optional(string, "")

      # Display name shown in the metrics explorer. Defaults to
      # metadata.name when left empty.
      display_name = optional(string, "")

      # The labels each point carries (filled by label_extractors). Declare
      # every label the extractors populate.
      labels = optional(list(object({
        # The label key. Changing a label's key REPLACES the metric (label
        # schemas are append-only in the API).
        key = string

        # What this label carries.
        description = optional(string, "")

        # The label's value type: BOOL, INT64, or STRING (empty means STRING —
        # the API default). Changing it REPLACES the metric.
        value_type = optional(string, "")
      })), [])
    }))

    # For DISTRIBUTION metrics: the expression extracting the numeric value
    # from each matching entry — EXTRACT(field) or
    # REGEXP_EXTRACT(field, regex), e.g.
    #   REGEXP_EXTRACT(jsonPayload.latency, "(\\d+)ms")
    value_extractor = optional(string, "")

    # Populate metric labels from each matching entry: map of label name
    # (which must be declared in metric_descriptor.labels) to an EXTRACT /
    # REGEXP_EXTRACT expression, e.g.
    #   { "status": "EXTRACT(httpRequest.status)" }
    label_extractors = optional(map(string), {})

    # For DISTRIBUTION metrics: the histogram bucket layout values are
    # recorded into. At least one layout (the provider's own rule when the
    # block is present); layouts may be combined.
    bucket_options = optional(object({
      # Hand-picked bucket boundaries.
      explicit_buckets = optional(object({
        # The boundary values, ascending. N boundaries define N+1 buckets.
        bounds = list(number)
      }))

      # Buckets that grow geometrically — the usual choice for latencies
      # (each bucket `growth_factor` times wider than the last).
      exponential_buckets = optional(object({
        # How many finite buckets (there is always an underflow and an overflow
        # bucket beyond them).
        num_finite_buckets = optional(number, 0)

        # The growth factor between adjacent buckets (must exceed 1).
        growth_factor = optional(number, 0)

        # The scale of the first finite bucket (must be positive).
        scale = optional(number, 0)
      }))

      # Equal-width buckets.
      linear_buckets = optional(object({
        # How many finite buckets.
        num_finite_buckets = optional(number, 0)

        # The starting value of the first finite bucket.
        offset = optional(number, 0)

        # The width of each bucket (must be positive).
        width = optional(number, 0)
      }))
    }))

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the metric is deleted; its historical points expire
    #                with Monitoring retention
    #   "PREVENT" -- destroy FAILS; protects a metric that alert policies
    #                depend on from accidental teardown
    #   "ABANDON" -- the metric is removed from management but keeps
    #                ingesting in GCP
    deletion_policy = optional(string, "")
  })
}
