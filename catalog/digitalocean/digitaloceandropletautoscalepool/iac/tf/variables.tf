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
  description = "DigitalOceanDropletAutoscalePool specification"
  type = object({
    pool_name = string
    static = optional(object({
      target_instances = number
    }))
    dynamic = optional(object({
      min_instances = number
      max_instances = number
      target_cpu_utilization = optional(number)
      target_memory_utilization = optional(number)
      cooldown_minutes = optional(number)
    }))
    droplet_template = object({
      size = string
      region = string
      image = string
      ssh_keys = list(string)
      vpc = optional(string, "")
      project_id = optional(string, "")
      tags = optional(list(string), [])
      with_droplet_agent = optional(bool, false)
      ipv6 = optional(bool, false)
      user_data = optional(string, "")
      public_networking = optional(bool)
    })
  })
}
