resource "aws_cloudwatch_metric_alarm" "this" {
  alarm_name          = local.resource_name
  comparison_operator = local.comparison_operator
  evaluation_periods  = local.evaluation_periods

  # Simple metric fields (used when NOT using metric_queries).
  metric_name        = local.metric_name
  namespace          = local.namespace
  period             = local.period
  statistic          = local.statistic
  extended_statistic = local.extended_statistic
  dimensions         = local.dimensions
  unit               = local.unit

  # Evaluation.
  datapoints_to_alarm = local.datapoints_to_alarm
  threshold           = local.threshold
  threshold_metric_id = local.threshold_metric_id
  treat_missing_data  = local.treat_missing_data

  # Actions.
  actions_enabled           = local.actions_enabled
  alarm_actions             = local.alarm_actions
  ok_actions                = local.ok_actions
  insufficient_data_actions = local.insufficient_data_actions

  # Description.
  alarm_description = local.alarm_description

  # Percentile.
  evaluate_low_sample_count_percentiles = local.evaluate_low_sample_count_percentiles

  # Metric queries (used for math expressions / multi-metric alarms).
  dynamic "metric_query" {
    for_each = local.metric_queries
    content {
      id          = metric_query.value.id
      expression  = try(metric_query.value.expression, null)
      label       = try(metric_query.value.label, null)
      period      = try(metric_query.value.period, null)
      return_data = try(metric_query.value.return_data, null)
      account_id  = try(metric_query.value.account_id, null)

      dynamic "metric" {
        for_each = try(metric_query.value.metric, null) != null ? [metric_query.value.metric] : []
        content {
          metric_name = metric.value.metric_name
          namespace   = metric.value.namespace
          period      = metric.value.period
          stat        = metric.value.stat
          dimensions  = try(metric.value.dimensions, null)
          unit        = try(metric.value.unit, null)
        }
      }
    }
  }

  tags = local.tags
}
