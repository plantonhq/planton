locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-queue")

  # Queue settings: 0 means "use the provider default", so pass those as null.
  queue_settings = var.spec.settings != null ? {
    delivery_delay           = try(var.spec.settings.delivery_delay, 0) > 0 ? var.spec.settings.delivery_delay : null
    delivery_paused          = try(var.spec.settings.delivery_paused, false)
    message_retention_period = try(var.spec.settings.message_retention_period, 0) > 0 ? var.spec.settings.message_retention_period : null
  } : null

  consumer = var.spec.consumer

  # Consumer settings are gated by consumer type: max_concurrency / max_wait_time_ms
  # apply only to worker (push) consumers; visibility_timeout_ms only to http_pull.
  consumer_settings = (local.consumer != null && try(local.consumer.settings, null) != null) ? {
    batch_size            = try(local.consumer.settings.batch_size, 0) > 0 ? local.consumer.settings.batch_size : null
    max_retries           = try(local.consumer.settings.max_retries, 0) > 0 ? local.consumer.settings.max_retries : null
    retry_delay           = try(local.consumer.settings.retry_delay, 0) > 0 ? local.consumer.settings.retry_delay : null
    max_concurrency       = (local.consumer.type == "worker" && try(local.consumer.settings.max_concurrency, 0) > 0) ? local.consumer.settings.max_concurrency : null
    max_wait_time_ms      = (local.consumer.type == "worker" && try(local.consumer.settings.max_wait_time_ms, 0) > 0) ? local.consumer.settings.max_wait_time_ms : null
    visibility_timeout_ms = (local.consumer.type == "http_pull" && try(local.consumer.settings.visibility_timeout_ms, 0) > 0) ? local.consumer.settings.visibility_timeout_ms : null
  } : null
}
