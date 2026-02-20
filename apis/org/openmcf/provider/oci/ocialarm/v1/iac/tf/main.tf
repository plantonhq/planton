resource "oci_monitoring_alarm" "this" {
  compartment_id       = var.spec.compartment_id.value
  metric_compartment_id = var.spec.metric_compartment_id.value
  namespace            = var.spec.namespace
  query                = var.spec.query
  severity             = local.severity_map[var.spec.severity]
  destinations         = var.spec.destinations
  display_name         = var.metadata.name
  is_enabled           = var.spec.is_enabled
  freeform_tags        = local.freeform_tags

  body                          = var.spec.body != "" ? var.spec.body : null
  alarm_summary                 = var.spec.alarm_summary != "" ? var.spec.alarm_summary : null
  notification_title            = var.spec.notification_title != "" ? var.spec.notification_title : null
  pending_duration              = var.spec.pending_duration != "" ? var.spec.pending_duration : null
  evaluation_slack_duration     = var.spec.evaluation_slack_duration != "" ? var.spec.evaluation_slack_duration : null
  repeat_notification_duration  = var.spec.repeat_notification_duration != "" ? var.spec.repeat_notification_duration : null
  message_format                = var.spec.message_format != "raw" ? local.message_format_map[var.spec.message_format] : null
  metric_compartment_id_in_subtree                = var.spec.metric_compartment_id_in_subtree
  is_notifications_per_metric_dimension_enabled   = var.spec.is_notifications_per_metric_dimension_enabled
  resource_group                = var.spec.resource_group != "" ? var.spec.resource_group : null
  notification_version          = var.spec.notification_version != "" ? var.spec.notification_version : null
  rule_name                     = var.spec.rule_name != "" ? var.spec.rule_name : null

  dynamic "overrides" {
    for_each = var.spec.overrides
    content {
      rule_name        = overrides.value.rule_name
      query            = overrides.value.query != "" ? overrides.value.query : null
      severity         = overrides.value.severity != "" ? local.severity_map[overrides.value.severity] : null
      body             = overrides.value.body != "" ? overrides.value.body : null
      pending_duration = overrides.value.pending_duration != "" ? overrides.value.pending_duration : null
    }
  }
}
