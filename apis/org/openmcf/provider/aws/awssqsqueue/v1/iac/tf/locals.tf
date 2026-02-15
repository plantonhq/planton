locals {
  # Name and tags
  resource_name = coalesce(try(var.metadata.name, null), "awssqsqueue")
  is_fifo       = coalesce(try(var.spec.fifo_queue, null), false)

  # FIFO queues must have names ending with ".fifo".
  queue_name = local.is_fifo && !endswith(local.resource_name, ".fifo") ? "${local.resource_name}.fifo" : local.resource_name

  tags = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # Delivery settings — null lets AWS use its defaults.
  visibility_timeout_seconds = try(var.spec.visibility_timeout_seconds, null) != 0 ? try(var.spec.visibility_timeout_seconds, null) : null
  message_retention_seconds  = try(var.spec.message_retention_seconds, null) != 0 ? try(var.spec.message_retention_seconds, null) : null
  max_message_size           = try(var.spec.max_message_size_bytes, null) != 0 ? try(var.spec.max_message_size_bytes, null) : null
  delay_seconds              = try(var.spec.delay_seconds, null) != 0 ? try(var.spec.delay_seconds, null) : null
  receive_wait_time_seconds  = try(var.spec.receive_wait_time_seconds, null) != 0 ? try(var.spec.receive_wait_time_seconds, null) : null

  # Dead letter queue — build redrive policy JSON when configured.
  has_dlq = try(var.spec.dead_letter_config, null) != null
  redrive_policy = local.has_dlq ? jsonencode({
    deadLetterTargetArn = try(var.spec.dead_letter_config.target_arn.value, "")
    maxReceiveCount     = try(var.spec.dead_letter_config.max_receive_count, 5)
  }) : null

  # Encryption
  kms_master_key_id                 = try(var.spec.kms_key_id.value, null)
  kms_data_key_reuse_period_seconds = try(var.spec.kms_data_key_reuse_period_seconds, null) != 0 ? try(var.spec.kms_data_key_reuse_period_seconds, null) : null
  sqs_managed_sse_enabled           = coalesce(try(var.spec.sqs_managed_sse_enabled, null), false) ? true : null

  # Access policy — expect the struct to arrive as a JSON string from the stack input layer.
  policy = try(var.spec.policy, null)
}
