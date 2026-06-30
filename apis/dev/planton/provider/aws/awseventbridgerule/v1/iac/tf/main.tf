# ---- EventBridge Rule ----

resource "aws_cloudwatch_event_rule" "this" {
  name                = local.resource_name
  description         = try(var.spec.description, null)
  event_bus_name      = local.event_bus_name
  event_pattern       = local.event_pattern
  schedule_expression = local.schedule_expression
  state               = local.state

  tags = local.tags
}

# ---- EventBridge Targets ----

resource "aws_cloudwatch_event_target" "this" {
  for_each = local.targets

  rule           = aws_cloudwatch_event_rule.this.name
  event_bus_name = local.event_bus_name
  target_id      = each.key
  arn            = try(each.value.arn.value, each.value.arn)

  role_arn = try(each.value.role_arn.value, null) != "" ? try(each.value.role_arn.value, null) : null

  # Input transformation (mutually exclusive)
  input      = try(each.value.input, null) != "" ? try(each.value.input, null) : null
  input_path = try(each.value.input_path, null) != "" ? try(each.value.input_path, null) : null

  dynamic "input_transformer" {
    for_each = try(each.value.input_transformer, null) != null ? [each.value.input_transformer] : []
    content {
      input_paths    = try(input_transformer.value.input_paths, null)
      input_template = input_transformer.value.input_template
    }
  }

  # Dead letter config
  dynamic "dead_letter_config" {
    for_each = try(each.value.dead_letter_config, null) != null ? [each.value.dead_letter_config] : []
    content {
      arn = try(dead_letter_config.value.arn.value, dead_letter_config.value.arn)
    }
  }

  # Retry policy
  dynamic "retry_policy" {
    for_each = try(each.value.retry_policy, null) != null ? [each.value.retry_policy] : []
    content {
      maximum_event_age_in_seconds = try(retry_policy.value.maximum_event_age_in_seconds, null)
      maximum_retry_attempts       = try(retry_policy.value.maximum_retry_attempts, null)
    }
  }

  # SQS target (message_group_id for FIFO queues)
  dynamic "sqs_target" {
    for_each = try(each.value.sqs_config.message_group_id, "") != "" ? [each.value.sqs_config] : []
    content {
      message_group_id = sqs_target.value.message_group_id
    }
  }
}
