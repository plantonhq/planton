variable "metadata" {
  description = "Catalog object metadata"
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
  description = "CloudflareZeroTrustGatewaySettings specification"
  type = object({
    account_id = string
    settings = optional(object({
      activity_log = optional(object({
        enabled = optional(bool)
      }))
      antivirus = optional(object({
        enabled_download_phase = optional(bool)
        enabled_upload_phase   = optional(bool)
        fail_closed            = optional(bool)
        notification_settings = optional(object({
          enabled         = optional(bool)
          msg             = optional(string, "")
          support_url     = optional(string, "")
          include_context = optional(bool)
        }))
      }))
      block_page = optional(object({
        enabled          = optional(bool)
        mode             = optional(string, "")
        target_uri       = optional(string, "")
        include_context  = optional(bool)
        name             = optional(string, "")
        header_text      = optional(string, "")
        footer_text      = optional(string, "")
        suppress_footer  = optional(bool)
        background_color = optional(string, "")
        logo_path        = optional(string, "")
        mailto_address   = optional(string, "")
        mailto_subject   = optional(string, "")
      }))
      body_scanning = optional(object({
        inspection_mode = optional(string, "")
      }))
      browser_isolation = optional(object({
        non_identity_enabled          = optional(bool)
        url_browser_isolation_enabled = optional(bool)
      }))
      certificate = optional(object({
        id = string
      }))
      extended_email_matching = optional(object({
        enabled = optional(bool)
      }))
      fips = optional(object({
        tls = optional(bool)
      }))
      host_selector = optional(object({
        enabled = optional(bool)
      }))
      inspection = optional(object({
        mode = optional(string, "")
      }))
      max_ttl_secs = optional(number)
      protocol_detection = optional(object({
        enabled = optional(bool)
      }))
      sandbox = optional(object({
        enabled         = optional(bool)
        fallback_action = optional(string, "")
      }))
      tls_decrypt = optional(object({
        enabled = optional(bool)
      }))
    }))
    logging = optional(object({
      redact_pii = optional(bool)
      settings_by_rule_type = optional(object({
        dns = optional(object({
          log_all    = optional(bool)
          log_blocks = optional(bool)
        }))
        http = optional(object({
          log_all    = optional(bool)
          log_blocks = optional(bool)
        }))
        l4 = optional(object({
          log_all    = optional(bool)
          log_blocks = optional(bool)
        }))
      }))
    }))
    pac_files = optional(list(object({
      name        = string
      contents    = string
      slug        = optional(string, "")
      description = optional(string, "")
    })), [])
  })
}