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
  description = "CloudflareNotificationPolicy specification"
  type = object({
    account_id = string
    name = string
    alert_type = string
    description = optional(string, "")
    enabled = optional(bool)
    alert_interval = optional(string, "")
    mechanisms = object({
      emails = optional(list(string), [])
      pagerduty_ids = optional(list(string), [])
      webhook_ids = optional(list(string), [])
    })
    filters = optional(object({
      actions = optional(list(string), [])
      affected_asns = optional(list(string), [])
      affected_components = optional(list(string), [])
      affected_locations = optional(list(string), [])
      airport_code = optional(list(string), [])
      alert_trigger_preferences = optional(list(string), [])
      alert_trigger_preferences_value = optional(list(string), [])
      enabled = optional(list(string), [])
      environment = optional(list(string), [])
      event = optional(list(string), [])
      event_source = optional(list(string), [])
      event_type = optional(list(string), [])
      group_by = optional(list(string), [])
      health_check_id = optional(list(string), [])
      incident_impact = optional(list(string), [])
      input_id = optional(list(string), [])
      insight_class = optional(list(string), [])
      limit = optional(list(string), [])
      logo_tag = optional(list(string), [])
      megabits_per_second = optional(list(string), [])
      new_health = optional(list(string), [])
      new_status = optional(list(string), [])
      packets_per_second = optional(list(string), [])
      pool_id = optional(list(string), [])
      pop_names = optional(list(string), [])
      product = optional(list(string), [])
      project_id = optional(list(string), [])
      protocol = optional(list(string), [])
      query_tag = optional(list(string), [])
      requests_per_second = optional(list(string), [])
      selectors = optional(list(string), [])
      services = optional(list(string), [])
      slo = optional(list(string), [])
      status = optional(list(string), [])
      target_hostname = optional(list(string), [])
      target_ip = optional(list(string), [])
      target_zone_name = optional(list(string), [])
      traffic_exclusions = optional(list(string), [])
      tunnel_id = optional(list(string), [])
      tunnel_name = optional(list(string), [])
      type = optional(list(string), [])
      where = optional(list(string), [])
      zones = optional(list(string), [])
    }))
  })
}