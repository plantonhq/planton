variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for the DigitalOcean Load Balancer"
  type = object({
    load_balancer_name = string
    region             = string
    vpc                = object({
      value = optional(string)
      ref   = optional(string)
    })
    forwarding_rules = list(object({
      entry_port       = number
      entry_protocol   = string
      target_port      = number
      target_protocol  = string
      certificate_name = optional(string)
    }))
    health_check = optional(object({
      port              = number
      protocol          = string
      path              = optional(string)
      check_interval_sec = optional(number, 10)
    }))
    droplet_ids           = optional(list(object({
      value = optional(string)
      ref   = optional(string)
    })))
    droplet_tag            = optional(string)
    enable_sticky_sessions = optional(bool, false)
  })
}