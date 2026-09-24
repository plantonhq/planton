variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "DigitalOceanUptimeCheck specification"
  type = object({
    check_name = string
    target = string
    type = optional(string, "")
    regions = list(string)
    enabled = optional(bool)
    alerts = optional(list(object({
      alert_name = string
      type = string
      threshold = optional(number)
      comparison = optional(string, "")
      period = string
      notifications = object({
        emails = optional(list(string), [])
        slack = optional(list(object({
          channel = string
          url = string
        })), [])
      })
    })), [])
  })
}
