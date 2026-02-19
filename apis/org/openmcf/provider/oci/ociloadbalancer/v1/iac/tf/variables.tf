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
  description = "OciLoadBalancer specification."
  type = object({
    compartment_id = object({
      value = string
    })
    display_name = optional(string, "")
    shape        = string
    shape_details = optional(object({
      minimum_bandwidth_in_mbps = number
      maximum_bandwidth_in_mbps = number
    }), null)
    subnet_ids = list(object({
      value = string
    }))
    is_private                   = optional(bool, false)
    network_security_group_ids   = optional(list(object({ value = string })), [])
    is_delete_protection_enabled = optional(bool, false)
    ip_mode                      = optional(string, "")
    reserved_ips = optional(list(object({
      id = string
    })), [])
    is_request_id_enabled = optional(bool, false)
    request_id_header     = optional(string, "")

    backend_sets = list(object({
      name   = string
      policy = string
      health_checker = object({
        protocol            = string
        port                = optional(number, 0)
        url_path            = optional(string, "")
        return_code         = optional(number, 0)
        response_body_regex = optional(string, "")
        interval_ms         = optional(number, 0)
        timeout_in_millis   = optional(number, 0)
        retries             = optional(number, 0)
        is_force_plain_text = optional(bool, false)
      })
      backends = optional(list(object({
        ip_address      = string
        port            = number
        weight          = optional(number, 0)
        backup          = optional(bool, false)
        drain           = optional(bool, false)
        offline         = optional(bool, false)
        max_connections = optional(number, 0)
      })), [])
      ssl_configuration = optional(object({
        certificate_ids                  = optional(list(string), [])
        certificate_name                 = optional(string, "")
        cipher_suite_name                = optional(string, "")
        protocols                        = optional(list(string), [])
        server_order_preference          = optional(string, "")
        trusted_certificate_authority_ids = optional(list(string), [])
        verify_depth                     = optional(number, 0)
        verify_peer_certificate          = optional(bool, false)
        has_session_resumption           = optional(bool, false)
      }), null)
      backend_max_connections = optional(number, 0)
      lb_cookie_session_persistence = optional(object({
        cookie_name        = optional(string, "")
        disable_fallback   = optional(bool, false)
        domain             = optional(string, "")
        is_http_only       = optional(bool, false)
        is_secure          = optional(bool, false)
        max_age_in_seconds = optional(number, 0)
        path               = optional(string, "")
      }), null)
      app_cookie_session_persistence = optional(object({
        cookie_name      = string
        disable_fallback = optional(bool, false)
      }), null)
    }))

    listeners = list(object({
      name                      = string
      port                      = number
      protocol                  = string
      default_backend_set_name = string
      ssl_configuration = optional(object({
        certificate_ids                  = optional(list(string), [])
        certificate_name                 = optional(string, "")
        cipher_suite_name                = optional(string, "")
        protocols                        = optional(list(string), [])
        server_order_preference          = optional(string, "")
        trusted_certificate_authority_ids = optional(list(string), [])
        verify_depth                     = optional(number, 0)
        verify_peer_certificate          = optional(bool, false)
        has_session_resumption           = optional(bool, false)
      }), null)
      connection_configuration = optional(object({
        idle_timeout_in_seconds              = optional(number, 0)
        backend_tcp_proxy_protocol_version = optional(number, 0)
      }), null)
      hostname_names       = optional(list(string), [])
      rule_set_names       = optional(list(string), [])
      routing_policy_name = optional(string, "")
    }))

    certificates = optional(list(object({
      certificate_name   = string
      ca_certificate     = optional(string, "")
      public_certificate = optional(string, "")
      private_key        = optional(string, "")
      passphrase         = optional(string, "")
    })), [])

    hostnames = optional(list(object({
      name     = string
      hostname = string
    })), [])

    rule_sets = optional(list(object({
      name = string
      items = list(object({
        action                        = string
        header                        = optional(string, "")
        value                         = optional(string, "")
        prefix                        = optional(string, "")
        suffix                        = optional(string, "")
        redirect_uri = optional(object({
          protocol = optional(string, "")
          host     = optional(string, "")
          port     = optional(number, 0)
          path     = optional(string, "")
          query    = optional(string, "")
        }), null)
        response_code                 = optional(number, 0)
        conditions = optional(list(object({
          attribute_name  = string
          attribute_value = string
          operator        = optional(string, "")
        })), [])
        allowed_methods               = optional(list(string), [])
        status_code                   = optional(number, 0)
        are_invalid_characters_allowed = optional(bool, false)
        http_large_header_size_in_kb  = optional(number, 0)
        default_max_connections       = optional(number, 0)
        ip_max_connections = optional(list(object({
          ip_addresses    = list(string)
          max_connections = number
        })), [])
        description                   = optional(string, "")
      }))
    })), [])
  })
}
