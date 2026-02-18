variable "metadata" {
  description = "Resource metadata (name, org, env, id)"
  type = object({
    name = string
    org  = optional(string, "")
    env  = optional(string, "")
    id   = optional(string, "")
  })
}

variable "spec" {
  description = "AwsNetworkLoadBalancerSpec configuration"
  type = object({
    # The AWS region where the Network Load Balancer will be created.
    region = string

    subnet_mappings = list(object({
      subnet_id            = string
      allocation_id        = optional(string)
      private_ipv4_address = optional(string)
    }))
    security_groups                  = optional(list(string), [])
    internal                         = optional(bool, false)
    delete_protection_enabled        = optional(bool, false)
    cross_zone_load_balancing_enabled = optional(bool, false)
    ip_address_type                  = optional(string, "ipv4")
    dns_record_client_routing_policy = optional(string, "")
    listeners = list(object({
      name     = string
      port     = number
      protocol = string
      tls = optional(object({
        certificate_arn = string
        ssl_policy      = optional(string)
      }))
      tcp_idle_timeout_seconds = optional(number, 0)
      alpn_policy              = optional(string, "")
      target_group = object({
        port                         = number
        protocol                     = string
        target_type                  = optional(string, "instance")
        deregistration_delay_seconds = optional(number)
        preserve_client_ip           = optional(bool, false)
        proxy_protocol_v2            = optional(bool, false)
        connection_termination       = optional(bool, false)
        stickiness_enabled           = optional(bool, false)
        health_check = optional(object({
          protocol            = optional(string)
          port                = optional(string)
          path                = optional(string)
          healthy_threshold   = optional(number)
          unhealthy_threshold = optional(number)
          interval_seconds    = optional(number)
          timeout_seconds     = optional(number)
          matcher             = optional(string)
        }))
      })
    }))
    dns = optional(object({
      enabled        = optional(bool, false)
      route53_zone_id = optional(string)
      hostnames      = optional(list(string), [])
    }))
  })
}
