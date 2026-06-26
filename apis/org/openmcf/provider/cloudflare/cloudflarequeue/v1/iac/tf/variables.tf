variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "CloudflareQueueSpec defines a Cloudflare Queue and its optional consumer"
  type = object({
    # (Required) The Cloudflare account ID that owns this queue.
    account_id = string

    # (Required) The queue name.
    queue_name = string

    # (Optional) Queue-level delivery settings.
    settings = optional(object({
      delivery_delay           = optional(number, 0)
      delivery_paused          = optional(bool, false)
      message_retention_period = optional(number, 0)
    }))

    # (Optional) The queue's single consumer. StringValueOrRef fields (script_name,
    # dead_letter_queue) are flattened to plain strings by the tfvars converter.
    consumer = optional(object({
      type              = string
      script_name       = optional(string, "")
      dead_letter_queue = optional(string, "")
      settings = optional(object({
        batch_size            = optional(number, 0)
        max_concurrency       = optional(number, 0)
        max_retries           = optional(number, 0)
        max_wait_time_ms      = optional(number, 0)
        retry_delay           = optional(number, 0)
        visibility_timeout_ms = optional(number, 0)
      }))
    }))
  })
}
