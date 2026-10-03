variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "CloudflareWaitingRoomEvent specification"
  type = object({
    waiting_room_id = string
    zone_id = string
    name = string
    event_start_time = string
    event_end_time = string
    prequeue_start_time = optional(string, "")
    shuffle_at_event_start = optional(bool)
    description = optional(string, "")
    suspended = optional(bool)
    custom_page_html = optional(string, "")
    disable_session_renewal = optional(bool)
    new_users_per_minute = optional(number)
    total_active_users = optional(number)
    queueing_method = optional(string)
    session_duration = optional(number)
    turnstile_action = optional(string)
    turnstile_mode = optional(string)
  })
}