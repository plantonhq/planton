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
  description = "DigitalOceanMonitorAlert specification"
  type = object({
    description = string
    metric_type = string
    compare = string
    value = number
    window = string
    enabled = optional(bool)
    droplet_ids = optional(list(string), [])
    load_balancer_ids = optional(list(string), [])
    database_cluster_ids = optional(list(string), [])
    tags = optional(list(string), [])
    alerts = object({
      emails = optional(list(string), [])
      slack = optional(list(object({
        channel = string
        url = string
      })), [])
    })
  })
}
