locals {
  resource_name = coalesce(try(var.metadata.name, null), "awscloudwatchalarm")

  tags = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # ─── Required fields ────────────────────────────────────────────────

  # Comparison operator — required (e.g. GreaterThanOrEqualToThreshold).
  comparison_operator = var.spec.comparison_operator

  # Evaluation periods — required.
  evaluation_periods = var.spec.evaluation_periods

  # ─── Simple metric fields (used when NOT using metric_queries) ─────

  # Metric name — null when empty (metric_queries mode).
  metric_name = try(var.spec.metric_name, null) != "" ? try(var.spec.metric_name, null) : null

  # Namespace — null when empty.
  namespace = try(var.spec.namespace, null) != "" ? try(var.spec.namespace, null) : null

  # Period — null when 0 or unset.
  period = try(var.spec.period, 0) != 0 ? try(var.spec.period, null) : null

  # Statistic — null when empty (e.g. Average, Sum, Maximum, Minimum, SampleCount).
  statistic = try(var.spec.statistic, null) != "" ? try(var.spec.statistic, null) : null

  # Extended statistic — null when empty (e.g. p99).
  extended_statistic = try(var.spec.extended_statistic, null) != "" ? try(var.spec.extended_statistic, null) : null

  # Dimensions — null when not configured.
  dimensions = try(var.spec.dimensions, null) != null ? try(var.spec.dimensions, {}) : null

  # Unit — null when empty.
  unit = try(var.spec.unit, null) != "" ? try(var.spec.unit, null) : null

  # ─── Evaluation fields ─────────────────────────────────────────────

  # Datapoints to alarm — null when 0 or unset (defaults to evaluation_periods).
  datapoints_to_alarm = try(var.spec.datapoints_to_alarm, 0) != 0 ? try(var.spec.datapoints_to_alarm, null) : null

  # Threshold — null when using anomaly detection (threshold_metric_id set).
  threshold = try(var.spec.threshold_metric_id, "") != "" ? null : try(var.spec.threshold, null)

  # Threshold metric ID — null when empty (used for anomaly detection).
  threshold_metric_id = try(var.spec.threshold_metric_id, null) != "" ? try(var.spec.threshold_metric_id, null) : null

  # Treat missing data — null when empty (defaults to "missing").
  treat_missing_data = try(var.spec.treat_missing_data, null) != "" ? try(var.spec.treat_missing_data, null) : null

  # ─── Actions ───────────────────────────────────────────────────────

  # Actions enabled — defaults to true.
  actions_enabled = try(var.spec.actions_enabled, true)

  # Alarm actions — list of ARNs (SNS topics, Auto Scaling policies, etc.).
  alarm_actions = try(var.spec.alarm_actions, [])

  # OK actions — list of ARNs.
  ok_actions = try(var.spec.ok_actions, [])

  # Insufficient data actions — list of ARNs.
  insufficient_data_actions = try(var.spec.insufficient_data_actions, [])

  # ─── Description ───────────────────────────────────────────────────

  # Alarm description — null when empty.
  alarm_description = try(var.spec.alarm_description, null) != "" ? try(var.spec.alarm_description, null) : null

  # ─── Percentile ────────────────────────────────────────────────────

  # Evaluate low sample count percentiles — null when empty.
  evaluate_low_sample_count_percentiles = try(var.spec.evaluate_low_sample_count_percentiles, null) != "" ? try(var.spec.evaluate_low_sample_count_percentiles, null) : null

  # ─── Metric queries (used for math expressions / multi-metric) ─────

  metric_queries = try(var.spec.metric_queries, [])
}
