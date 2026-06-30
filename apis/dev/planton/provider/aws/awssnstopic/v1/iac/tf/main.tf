resource "aws_sns_topic" "this" {
  name      = local.topic_name
  fifo_topic = local.is_fifo

  # FIFO-specific settings
  content_based_deduplication = local.is_fifo ? coalesce(try(var.spec.content_based_deduplication, null), false) : null
  fifo_throughput_scope       = local.is_fifo ? try(var.spec.fifo_throughput_scope, null) : null

  # Display name
  display_name = try(var.spec.display_name, null)

  # Encryption
  kms_master_key_id = local.kms_master_key_id

  # Access policy
  policy = local.policy

  # Delivery policy
  delivery_policy = local.delivery_policy

  # Observability
  tracing_config    = local.tracing_config
  signature_version = local.signature_version

  tags = local.tags
}

resource "aws_sns_topic_subscription" "this" {
  for_each = local.subscriptions_map

  topic_arn = aws_sns_topic.this.arn
  protocol  = each.value.protocol
  endpoint  = try(each.value.endpoint.value, "")

  raw_message_delivery = coalesce(try(each.value.raw_message_delivery, null), false)

  # Filter policy
  filter_policy       = try(jsonencode(each.value.filter_policy), null)
  filter_policy_scope = try(each.value.filter_policy_scope, null)

  # Redrive policy (subscription DLQ)
  redrive_policy = local.subscription_redrive_policies[each.key]

  # Firehose role
  subscription_role_arn = try(each.value.subscription_role_arn.value, null)
}
