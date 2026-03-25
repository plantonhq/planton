resource "cloudflare_ruleset" "main" {
  zone_id     = local.zone_id != "" ? local.zone_id : null
  account_id  = local.account_id != "" ? local.account_id : null
  kind        = local.ruleset_kind
  phase       = local.phase
  name        = var.spec.name
  description = var.spec.description

  dynamic "rules" {
    for_each = var.spec.rules
    content {
      ref         = rules.value.ref != "" ? rules.value.ref : null
      expression  = rules.value.expression
      action      = rules.value.action
      description = rules.value.description != "" ? rules.value.description : null
      enabled     = rules.value.enabled

      dynamic "action_parameters" {
        for_each = rules.value.action_parameters != null ? [rules.value.action_parameters] : []
        content {
          # Origin Rules (route)
          host_header = action_parameters.value.host_header

          dynamic "origin" {
            for_each = action_parameters.value.origin != null ? [action_parameters.value.origin] : []
            content {
              host = origin.value.host
              port = origin.value.port
            }
          }

          dynamic "sni" {
            for_each = action_parameters.value.sni != null ? [action_parameters.value.sni] : []
            content {
              value = sni.value.value
            }
          }

          # Block
          dynamic "response" {
            for_each = action_parameters.value.response != null ? [action_parameters.value.response] : []
            content {
              status_code  = response.value.status_code
              content      = response.value.content
              content_type = response.value.content_type
            }
          }

          # Rewrite
          dynamic "uri" {
            for_each = action_parameters.value.uri != null ? [action_parameters.value.uri] : []
            content {
              dynamic "path" {
                for_each = uri.value.path != null ? [uri.value.path] : []
                content {
                  value      = path.value.value != "" ? path.value.value : null
                  expression = path.value.expression != "" ? path.value.expression : null
                }
              }
              dynamic "query" {
                for_each = uri.value.query != null ? [uri.value.query] : []
                content {
                  value      = query.value.value != "" ? query.value.value : null
                  expression = query.value.expression != "" ? query.value.expression : null
                }
              }
            }
          }

          dynamic "headers" {
            for_each = action_parameters.value.headers
            content {
              name       = headers.key
              operation  = headers.value.operation
              value      = headers.value.value != "" ? headers.value.value : null
              expression = headers.value.expression != "" ? headers.value.expression : null
            }
          }

          # Redirect
          dynamic "from_value" {
            for_each = action_parameters.value.from_value != null ? [action_parameters.value.from_value] : []
            content {
              status_code          = from_value.value.status_code
              preserve_query_string = from_value.value.preserve_query_string

              dynamic "target_url" {
                for_each = from_value.value.target_url != null ? [from_value.value.target_url] : []
                content {
                  value      = target_url.value.value != "" ? target_url.value.value : null
                  expression = target_url.value.expression != "" ? target_url.value.expression : null
                }
              }
            }
          }

          # Skip
          phases   = length(action_parameters.value.phases) > 0 ? action_parameters.value.phases : null
          products = length(action_parameters.value.products) > 0 ? action_parameters.value.products : null
          ruleset  = action_parameters.value.ruleset != "" ? action_parameters.value.ruleset : null
          rulesets = length(action_parameters.value.rulesets) > 0 ? action_parameters.value.rulesets : null

          # Execute
          id = action_parameters.value.id != "" ? action_parameters.value.id : null

          dynamic "overrides" {
            for_each = action_parameters.value.overrides != null ? [action_parameters.value.overrides] : []
            content {
              action            = overrides.value.action != "" ? overrides.value.action : null
              enabled           = overrides.value.enabled
              sensitivity_level = overrides.value.sensitivity_level != "" ? overrides.value.sensitivity_level : null

              dynamic "categories" {
                for_each = overrides.value.categories
                content {
                  category          = categories.value.category
                  action            = categories.value.action
                  enabled           = categories.value.enabled
                  sensitivity_level = categories.value.sensitivity_level != "" ? categories.value.sensitivity_level : null
                }
              }

              dynamic "rules" {
                for_each = overrides.value.rules
                content {
                  id                = rules.value.id
                  action            = rules.value.action
                  enabled           = rules.value.enabled
                  score_threshold   = rules.value.score_threshold > 0 ? rules.value.score_threshold : null
                  sensitivity_level = rules.value.sensitivity_level != "" ? rules.value.sensitivity_level : null
                }
              }
            }
          }

          # Cache
          cache = action_parameters.value.cache

          dynamic "edge_ttl" {
            for_each = action_parameters.value.edge_ttl != null ? [action_parameters.value.edge_ttl] : []
            content {
              mode    = edge_ttl.value.mode
              default = edge_ttl.value.default_ttl > 0 ? edge_ttl.value.default_ttl : null

              dynamic "status_code_ttl" {
                for_each = edge_ttl.value.status_code_ttls
                content {
                  value       = status_code_ttl.value.value
                  status_code = status_code_ttl.value.status_code > 0 ? status_code_ttl.value.status_code : null

                  dynamic "status_code_range" {
                    for_each = status_code_ttl.value.status_code_range != null ? [status_code_ttl.value.status_code_range] : []
                    content {
                      from = status_code_range.value.from
                      to   = status_code_range.value.to
                    }
                  }
                }
              }
            }
          }

          dynamic "browser_ttl" {
            for_each = action_parameters.value.browser_ttl != null ? [action_parameters.value.browser_ttl] : []
            content {
              mode    = browser_ttl.value.mode
              default = browser_ttl.value.default_ttl > 0 ? browser_ttl.value.default_ttl : null
            }
          }

          dynamic "serve_stale" {
            for_each = action_parameters.value.serve_stale != null ? [action_parameters.value.serve_stale] : []
            content {
              disable_stale_while_updating = serve_stale.value.disable_stale_while_updating
            }
          }
        }
      }
    }
  }
}
