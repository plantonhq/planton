resource "aws_sqs_queue" "this" {
  name      = local.queue_name
  fifo_queue = local.is_fifo

  # Delivery settings — only set when non-zero to let AWS use defaults.
  visibility_timeout_seconds  = local.visibility_timeout_seconds
  message_retention_seconds   = local.message_retention_seconds
  max_message_size            = local.max_message_size
  delay_seconds               = local.delay_seconds
  receive_wait_time_seconds   = local.receive_wait_time_seconds

  # FIFO-specific settings
  content_based_deduplication = local.is_fifo ? coalesce(try(var.spec.content_based_deduplication, null), false) : null
  deduplication_scope         = local.is_fifo ? try(var.spec.deduplication_scope, null) : null
  fifo_throughput_limit       = local.is_fifo ? try(var.spec.fifo_throughput_limit, null) : null

  # Dead letter queue (redrive policy)
  redrive_policy = local.redrive_policy

  # Encryption
  kms_master_key_id                 = local.kms_master_key_id
  kms_data_key_reuse_period_seconds = local.kms_data_key_reuse_period_seconds
  sqs_managed_sse_enabled           = local.sqs_managed_sse_enabled

  # Access policy
  policy = local.policy

  tags = local.tags
}
