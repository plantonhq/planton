variable "metadata" {
  type = object({
    name   = string
    id     = optional(string, "")
    org    = optional(string, "")
    env    = optional(string, "")
    labels = optional(map(string), {})
  })
}

variable "spec" {
  type = object({
    compartment_id = object({
      value = string
    })
    custom_encryption_key_id = optional(object({
      value = string
    }))
    dead_letter_queue_delivery_count = optional(number)
    retention_in_seconds             = optional(number)
    timeout_in_seconds               = optional(number)
    visibility_in_seconds            = optional(number)
    channel_consumption_limit        = optional(number)
    is_large_messages_enabled        = optional(bool, false)
    consumer_group_config = optional(object({
      is_primary_enabled                       = optional(bool)
      primary_dead_letter_queue_delivery_count = optional(number)
      primary_display_name                     = optional(string, "")
    }))
  })
}
