variable "metadata" {
  description = "Resource metadata."
  type = object({
    name   = string
    id     = optional(string, "")
    org    = optional(string, "")
    env    = optional(string, "")
    labels = optional(map(string), {})
  })
}

variable "spec" {
  description = "OciNetworkLoadBalancer specification."
  type = object({
    compartment_id = object({
      value = string
    })
    display_name = optional(string, "")
    subnet_id = object({
      value = string
    })
    is_private                      = optional(bool, false)
    is_preserve_source_destination  = optional(bool, false)
    is_symmetric_hash_enabled       = optional(bool, false)
    network_security_group_ids      = optional(list(object({ value = string })), [])
    nlb_ip_version                  = optional(string, "")
    reserved_ips = optional(list(object({
      id = string
    })), [])
    assigned_ipv6         = optional(string, "")
    assigned_private_ipv4 = optional(string, "")
    subnet_ipv6cidr       = optional(string, "")

    backend_sets = list(object({
      name   = string
      policy = string
      health_checker = object({
        protocol            = string
        port                = optional(number, 0)
        url_path            = optional(string, "")
        return_code         = optional(number, 0)
        response_body_regex = optional(string, "")
        interval_in_millis  = optional(number, 0)
        timeout_in_millis   = optional(number, 0)
        retries             = optional(number, 0)
        request_data        = optional(string, "")
        response_data       = optional(string, "")
        dns_health_check = optional(object({
          domain_name        = string
          query_class        = optional(string, "")
          query_type         = optional(string, "")
          rcodes             = optional(list(string), [])
          transport_protocol = optional(string, "")
        }), null)
      })
      backends = optional(list(object({
        port       = number
        ip_address = optional(string, "")
        target_id  = optional(string, "")
        weight     = optional(number, 0)
        is_backup  = optional(bool, false)
        is_drain   = optional(bool, false)
        is_offline = optional(bool, false)
        name       = optional(string, "")
      })), [])
      is_preserve_source                          = optional(bool, false)
      is_fail_open                                = optional(bool, false)
      is_instant_failover_enabled                 = optional(bool, false)
      is_instant_failover_tcp_reset_enabled       = optional(bool, false)
      are_operationally_active_backends_preferred = optional(bool, false)
      ip_version                                  = optional(string, "")
    }))

    listeners = list(object({
      name                      = string
      port                      = number
      protocol                  = string
      default_backend_set_name  = string
      ip_version                = optional(string, "")
      is_ppv2_enabled           = optional(bool, false)
      tcp_idle_timeout          = optional(number, 0)
      udp_idle_timeout          = optional(number, 0)
      l3ip_idle_timeout         = optional(number, 0)
    }))
  })
}
