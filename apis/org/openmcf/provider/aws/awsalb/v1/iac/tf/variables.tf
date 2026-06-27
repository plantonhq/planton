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
  description = "AwsAlb specification"
  type = object({
    region = string
    subnets = list(string)
    security_groups = optional(list(string), [])
    internal = optional(bool, false)
    delete_protection_enabled = optional(bool, false)
    idle_timeout_seconds = optional(number, 0)
    dns = optional(object({
      enabled = optional(bool, false)
      route53_zone_id = optional(string, "")
      hostnames = optional(list(string), [])
    }))
    ssl = optional(object({
      enabled = optional(bool, false)
      certificate_arn = optional(string, "")
    }))
  })
}
