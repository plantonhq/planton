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
  description = "CloudflareWaitingRoom specification"
  type = object({
    zone_id = string
    name = string
    host = string
    path = optional(string)
    new_users_per_minute = optional(number, 0)
    total_active_users = optional(number, 0)
    session_duration = optional(number)
    suspended = optional(bool)
    queue_all = optional(bool)
    queueing_method = optional(string)
    queueing_status_code = optional(number)
    cookie_attributes = optional(object({
      samesite = optional(string)
      secure = optional(string)
    }))
    cookie_suffix = optional(string, "")
    custom_page_html = optional(string, "")
    default_template_language = optional(string)
    description = optional(string, "")
    disable_session_renewal = optional(bool)
    json_response_enabled = optional(bool)
    additional_routes = optional(list(object({
      host = string
      path = optional(string)
    })), [])
    enabled_origin_commands = optional(list(string), [])
    turnstile_action = optional(string)
    turnstile_mode = optional(string)
    bypass_rules = optional(list(object({
      expression = string
      description = optional(string, "")
      enabled = optional(bool)
    })), [])
  })
}