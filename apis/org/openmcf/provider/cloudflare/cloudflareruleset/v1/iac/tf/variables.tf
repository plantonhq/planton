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

        # Block
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
          status_code          = number
          preserve_query_string = optional(bool, false)
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

        # Cache
        cache = optional(bool)
        edge_ttl = optional(object({
          mode        = string
          default_ttl = optional(number, 0)
          status_code_ttls = optional(list(object({
            value = number
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
      }))
    }))
  })
}
