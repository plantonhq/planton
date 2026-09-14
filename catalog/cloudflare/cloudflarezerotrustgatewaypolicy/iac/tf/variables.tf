variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "CloudflareZeroTrustGatewayPolicy specification"
  type = object({
    account_id     = string
    name           = string
    action         = string
    description    = optional(string, "")
    filter         = optional(string, "")
    enabled        = optional(bool)
    precedence     = optional(number)
    traffic        = optional(string, "")
    identity       = optional(string, "")
    device_posture = optional(string, "")
    expiration = optional(object({
      expires_at = string
      duration   = optional(number)
    }))
    schedule = optional(object({
      mon       = optional(string, "")
      tue       = optional(string, "")
      wed       = optional(string, "")
      thu       = optional(string, "")
      fri       = optional(string, "")
      sat       = optional(string, "")
      sun       = optional(string, "")
      time_zone = optional(string, "")
    }))
    rule_settings = optional(object({
      add_headers = optional(map(object({
        values = list(string)
      })), {})
      allow_child_bypass = optional(bool)
      # The API requires the switch inside the block; no literal default, so an
      # unset value can never be mistaken for an explicit false.
      audit_ssh = optional(object({
        command_logging = optional(bool)
      }))
      biso_admin_controls = optional(object({
        version  = optional(string, "")
        copy     = optional(string, "")
        download = optional(string, "")
        paste    = optional(string, "")
        keyboard = optional(string, "")
        printing = optional(string, "")
        upload   = optional(string, "")
        dcp      = optional(bool)
        dd       = optional(bool)
        dk       = optional(bool)
        dp       = optional(bool)
        du       = optional(bool)
        wm_id    = optional(string, "")
      }))
      block_page = optional(object({
        target_uri      = string
        include_context = optional(bool, false)
      }))
      block_page_enabled = optional(bool)
      block_reason       = optional(string, "")
      bypass_parent_rule = optional(bool)
      check_session = optional(object({
        duration = optional(string, "")
        enforce  = optional(bool, false)
      }))
      delete_headers = optional(list(string), [])
      dns_resolvers = optional(object({
        ipv4 = optional(list(object({
          ip                            = string
          port                          = optional(number)
          route_through_private_network = optional(bool, false)
          vnet_id                       = optional(string, "")
        })), [])
        ipv6 = optional(list(object({
          ip                            = string
          port                          = optional(number)
          route_through_private_network = optional(bool, false)
          vnet_id                       = optional(string, "")
        })), [])
      }))
      egress = optional(object({
        ipv4          = optional(string, "")
        ipv4_fallback = optional(string, "")
        ipv6          = optional(string, "")
      }))
      # The API requires the switch inside the block; no literal default, so an
      # unset value can never be mistaken for an explicit false.
      forensic_copy = optional(object({
        enabled = optional(bool)
      }))
      ignore_cname_category_matches      = optional(bool)
      insecure_disable_dnssec_validation = optional(bool)
      ip_categories                      = optional(bool)
      ip_indicator_feeds                 = optional(bool)
      l4override = optional(object({
        ip   = optional(string, "")
        port = optional(number)
      }))
      notification_settings = optional(object({
        enabled         = optional(bool, false)
        include_context = optional(bool, false)
        msg             = optional(string, "")
        support_url     = optional(string, "")
      }))
      override_host = optional(string, "")
      override_ips  = optional(list(string), [])
      # The API requires the switch inside the block; no literal default, so an
      # unset value can never be mistaken for an explicit false.
      payload_log = optional(object({
        enabled = optional(bool)
      }))
      quarantine = optional(object({
        file_types = optional(list(string), [])
      }))
      redirect = optional(object({
        target_uri              = string
        include_context         = optional(bool, false)
        preserve_path_and_query = optional(bool, false)
      }))
      resolve_dns_internally = optional(object({
        fallback = optional(string, "")
        view_id  = optional(string, "")
      }))
      resolve_dns_through_cloudflare = optional(bool)
      set_headers = optional(map(object({
        values = list(string)
      })), {})
      untrusted_cert = optional(object({
        action = optional(string, "")
      }))
    }))
  })
}