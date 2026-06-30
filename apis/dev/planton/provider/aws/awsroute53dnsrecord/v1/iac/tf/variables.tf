variable "metadata" {
  description = "Resource metadata including name and labels"
  type = object({
    name = string
  })
}

variable "spec" {
  description = "AWS Route53 DNS Record specification"
  type = object({
    # The AWS region where the resource will be created.
    region = string
    # zone_id as StringValueOrRef - for Terraform we accept the resolved value
    # The Planton CLI resolves value_from references before passing to Terraform
    zone_id = object({
      value = optional(string)
    })
    name = string
    type = string
    ttl  = optional(number, 300)
    values = optional(list(string), [])
    alias_target = optional(object({
      # dns_name as StringValueOrRef - resolved by CLI
      dns_name = object({
        value = optional(string)
      })
      # zone_id as StringValueOrRef - resolved by CLI
      zone_id = object({
        value = optional(string)
      })
      evaluate_target_health = optional(bool, false)
    }))
    routing_policy = optional(object({
      weighted = optional(object({
        weight = number
      }))
      latency = optional(object({
        region = string
      }))
      failover = optional(object({
        failover_type = string
      }))
      geolocation = optional(object({
        continent   = optional(string)
        country     = optional(string)
        subdivision = optional(string)
      }))
    }))
    health_check_id = optional(string)
    set_identifier  = optional(string)
  })
}
