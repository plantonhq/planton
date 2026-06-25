# Cloudflare Queue: a guaranteed-delivery message queue for Workers. The optional
# consumer is provisioned as its own resource so toggling it never recreates the queue.
resource "cloudflare_queue" "main" {
  account_id = var.spec.account_id
  queue_name = var.spec.queue_name

  settings = local.queue_settings
}

resource "cloudflare_queue_consumer" "main" {
  count = local.consumer != null ? 1 : 0

  account_id = var.spec.account_id
  queue_id   = cloudflare_queue.main.queue_id
  type       = local.consumer.type

  # script_name applies only to worker (push) consumers.
  script_name = local.consumer.type == "worker" && try(local.consumer.script_name, "") != "" ? local.consumer.script_name : null

  dead_letter_queue = try(local.consumer.dead_letter_queue, "") != "" ? local.consumer.dead_letter_queue : null

  settings = local.consumer_settings
}
