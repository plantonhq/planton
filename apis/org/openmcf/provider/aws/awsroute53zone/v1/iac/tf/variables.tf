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
  description = "AwsRoute53Zone specification"
  type = object({
    region = string
    is_private = optional(bool, false)
    vpc_associations = optional(list(object({
      vpc_id = string
      vpc_region = string
    })), [])
    enable_query_logging = optional(bool, false)
    query_log_group_name = optional(string, "")
    enable_dnssec = optional(bool, false)
    records = optional(list(object({
      record_type = string
      name = string
      ttl_seconds = optional(number, 0)
      values = optional(list(string), [])
      alias_target = optional(object({
        dns_name = string
        hosted_zone_id = string
        evaluate_target_health = optional(bool, false)
      }))
      routing_policy = optional(object({
        weighted = optional(object({
          weight = optional(number, 0)
        }))
        latency = optional(object({
          region = string
        }))
        failover = optional(object({
          type = string
        }))
        geolocation = optional(object({
          continent = optional(string, "")
          country = optional(string, "")
          subdivision = optional(string, "")
        }))
      }))
      health_check_id = optional(string, "")
      set_identifier = optional(string, "")
    })), [])
  })
}
