variable "metadata" {
  description = "Resource metadata including name and labels"
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
  description = "CloudflareRulesetSpec defines the configuration for a Cloudflare Ruleset"
  type = object({
    # Cloudflare Zone ID (StringValueOrRef flattened to a plain string by the tfvars converter)
    zone_id = optional(string)

    # Cloudflare Account ID (for account-level rulesets)
    account_id = optional(string, "")

    # Ruleset kind: "zone", "custom", "managed", "root"
    ruleset_kind = optional(string, "zone")

    # Processing phase (e.g., "http_request_origin", "http_request_firewall_custom")
    phase = string

    # Human-readable ruleset name
    name = string

    # Informative description
    description = optional(string, "")

    # Ordered list of rules
    rules = list(object({
      ref         = optional(string, "")
      expression  = string
      action      = string
      description = optional(string, "")
      enabled     = optional(bool, true)

      # Rate limiting (http_ratelimit phase)
      ratelimit = optional(object({
        characteristics            = list(string)
        period                     = number
        counting_expression        = optional(string, "")
        mitigation_timeout         = optional(number, 0)
        requests_per_period        = optional(number, 0)
        requests_to_origin         = optional(bool, false)
        score_per_period           = optional(number, 0)
        score_response_header_name = optional(string, "")
      }))

      # Per-rule logging
      logging = optional(object({
        enabled = optional(bool, false)
      }))

      # Exposed-credential detection
      exposed_credential_check = optional(object({
        username_expression = string
        password_expression = string
      }))

      action_parameters = optional(object({
        # Origin Rules (route)
        host_header = optional(string)
        origin = optional(object({
          host = string
          port = number
        }))
        sni = optional(object({
          value = string
        }))

        # Block / serve_error inline response
        response = optional(object({
          status_code  = number
          content      = string
          content_type = string
        }))

        # Rewrite
        uri = optional(object({
          path = optional(object({
            value      = optional(string, "")
            expression = optional(string, "")
          }))
          query = optional(object({
            value      = optional(string, "")
            expression = optional(string, "")
          }))
        }))
        headers = optional(map(object({
          operation  = string
          value      = optional(string, "")
          expression = optional(string, "")
        })), {})

        # Redirect
        from_value = optional(object({
          target_url = object({
            value      = optional(string, "")
            expression = optional(string, "")
          })
          status_code           = number
          preserve_query_string = optional(bool, false)
        }))
        from_list = optional(object({
          name = string
          key  = string
        }))

        # Skip
        phases   = optional(list(string), [])
        products = optional(list(string), [])
        ruleset  = optional(string, "")
        rulesets = optional(list(string), [])

        # Execute
        id = optional(string, "")
        overrides = optional(object({
          action            = optional(string, "")
          enabled           = optional(bool)
          sensitivity_level = optional(string, "")
          categories = optional(list(object({
            category          = string
            action            = string
            enabled           = bool
            sensitivity_level = optional(string, "")
          })), [])
          rules = optional(list(object({
            id                = string
            action            = string
            enabled           = bool
            score_threshold   = optional(number, 0)
            sensitivity_level = optional(string, "")
          })), [])
        }))
        matched_data = optional(object({
          public_key = string
        }))

        # Compress response
        algorithms = optional(list(object({
          name = string
        })), [])

        # Score
        increment = optional(number, 0)

        # Serve error (inline)
        asset_name   = optional(string, "")
        content      = optional(string, "")
        content_type = optional(string, "")
        status_code  = optional(number, 0)

        # Configuration settings (set_config)
        automatic_https_rewrites = optional(bool)
        autominify = optional(object({
          css  = optional(bool)
          html = optional(bool)
          js   = optional(bool)
        }))
        bic                       = optional(bool)
        content_converter         = optional(bool)
        disable_apps              = optional(bool)
        disable_rum               = optional(bool)
        disable_zaraz             = optional(bool)
        email_obfuscation         = optional(bool)
        fonts                     = optional(bool)
        hotlink_protection        = optional(bool)
        mirage                    = optional(bool)
        opportunistic_encryption  = optional(bool)
        polish                    = optional(string, "")
        redirects_for_ai_training = optional(bool)
        request_body_buffering    = optional(string, "")
        response_body_buffering   = optional(string, "")
        rocket_loader             = optional(bool)
        security_level            = optional(string, "")
        server_side_excludes      = optional(bool)
        ssl                       = optional(string, "")
        sxg                       = optional(bool)

        # Cache (set_cache_settings)
        cache                      = optional(bool)
        additional_cacheable_ports = optional(list(number), [])
        edge_ttl = optional(object({
          mode        = string
          default_ttl = optional(number, 0)
          status_code_ttls = optional(list(object({
            value       = number
            status_code = optional(number)
            status_code_range = optional(object({
              from = number
              to   = number
            }))
          })), [])
        }))
        browser_ttl = optional(object({
          mode        = string
          default_ttl = optional(number, 0)
        }))
        serve_stale = optional(object({
          disable_stale_while_updating = optional(bool, false)
        }))
        cache_key = optional(object({
          cache_by_device_type       = optional(bool)
          cache_deception_armor      = optional(bool)
          ignore_query_strings_order = optional(bool)
          custom_key = optional(object({
            cookie = optional(object({
              check_presence = optional(list(string), [])
              include        = optional(list(string), [])
            }))
            header = optional(object({
              check_presence = optional(list(string), [])
              contains       = optional(map(list(string)), {})
              exclude_origin = optional(bool)
              include        = optional(list(string), [])
            }))
            host = optional(object({
              resolved = optional(bool)
            }))
            query_string = optional(object({
              include = optional(object({
                list = optional(list(string), [])
                all  = optional(bool)
              }))
              exclude = optional(object({
                list = optional(list(string), [])
                all  = optional(bool)
              }))
            }))
            user = optional(object({
              device_type = optional(bool)
              geo         = optional(bool)
              lang        = optional(bool)
            }))
          }))
        }))
        cache_reserve = optional(object({
          eligible          = bool
          minimum_file_size = optional(number, 0)
        }))
        origin_cache_control       = optional(bool)
        origin_error_page_passthru = optional(bool)
        read_timeout               = optional(number, 0)
        respect_strong_etags       = optional(bool)
        vary = optional(object({
          default = object({
            action = string
          })
          headers = optional(map(object({
            action      = string
            media_types = optional(list(string), [])
            languages   = optional(list(string), [])
          })), {})
        }))
        strip_etags        = optional(bool)
        strip_last_modified = optional(bool)
        strip_set_cookie   = optional(bool)

        # Log custom fields
        cookie_fields = optional(list(object({ name = string })), [])
        raw_response_fields = optional(list(object({
          name                = string
          preserve_duplicates = optional(bool, false)
        })), [])
        request_fields = optional(list(object({ name = string })), [])
        response_fields = optional(list(object({
          name                = string
          preserve_duplicates = optional(bool, false)
        })), [])
        transformed_request_fields = optional(list(object({ name = string })), [])

        # Set Cache-Control directives
        max_age = optional(object({
          operation       = string
          value           = optional(number, 0)
          cloudflare_only = optional(bool, false)
        }))
        s_maxage = optional(object({
          operation       = string
          value           = optional(number, 0)
          cloudflare_only = optional(bool, false)
        }))
        stale_while_revalidate = optional(object({
          operation       = string
          value           = optional(number, 0)
          cloudflare_only = optional(bool, false)
        }))
        stale_if_error = optional(object({
          operation       = string
          value           = optional(number, 0)
          cloudflare_only = optional(bool, false)
        }))
        private = optional(object({
          operation       = string
          qualifiers      = optional(list(string), [])
          cloudflare_only = optional(bool, false)
        }))
        no_cache = optional(object({
          operation       = string
          qualifiers      = optional(list(string), [])
          cloudflare_only = optional(bool, false)
        }))
        must_revalidate = optional(object({
          operation       = string
          cloudflare_only = optional(bool, false)
        }))
        proxy_revalidate = optional(object({
          operation       = string
          cloudflare_only = optional(bool, false)
        }))
        must_understand = optional(object({
          operation       = string
          cloudflare_only = optional(bool, false)
        }))
        no_transform = optional(object({
          operation       = string
          cloudflare_only = optional(bool, false)
        }))
        immutable = optional(object({
          operation       = string
          cloudflare_only = optional(bool, false)
        }))
        no_store = optional(object({
          operation       = string
          cloudflare_only = optional(bool, false)
        }))
        public = optional(object({
          operation       = string
          cloudflare_only = optional(bool, false)
        }))

        # Set cache tags
        operation  = optional(string, "")
        values     = optional(list(string), [])
        expression = optional(string, "")
      }))
    }))
  })
}
