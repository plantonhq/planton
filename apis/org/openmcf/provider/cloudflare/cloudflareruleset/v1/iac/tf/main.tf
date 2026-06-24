resource "cloudflare_ruleset" "main" {
  zone_id     = local.zone_id != "" ? local.zone_id : null
  account_id  = local.account_id != "" ? local.account_id : null
  kind        = local.ruleset_kind
  phase       = local.phase
  name        = var.spec.name
  description = var.spec.description

  rules = [
    for r in var.spec.rules : {
      ref         = r.ref != "" ? r.ref : null
      expression  = r.expression
      action      = r.action
      description = r.description != "" ? r.description : null
      enabled     = r.enabled

      action_parameters = r.action_parameters == null ? null : {
        # Origin Rules (route)
        host_header = r.action_parameters.host_header

        origin = r.action_parameters.origin == null ? null : {
          host = r.action_parameters.origin.host
          port = r.action_parameters.origin.port
        }

        sni = r.action_parameters.sni == null ? null : {
          value = r.action_parameters.sni.value
        }

        # Block
        response = r.action_parameters.response == null ? null : {
          status_code  = r.action_parameters.response.status_code
          content      = r.action_parameters.response.content
          content_type = r.action_parameters.response.content_type
        }

        # Rewrite
        uri = r.action_parameters.uri == null ? null : {
          path = r.action_parameters.uri.path == null ? null : {
            value      = r.action_parameters.uri.path.value != "" ? r.action_parameters.uri.path.value : null
            expression = r.action_parameters.uri.path.expression != "" ? r.action_parameters.uri.path.expression : null
          }
          query = r.action_parameters.uri.query == null ? null : {
            value      = r.action_parameters.uri.query.value != "" ? r.action_parameters.uri.query.value : null
            expression = r.action_parameters.uri.query.expression != "" ? r.action_parameters.uri.query.expression : null
          }
        }

        # Header transforms: the v5 map key is the header name.
        headers = length(r.action_parameters.headers) > 0 ? {
          for name, h in r.action_parameters.headers : name => {
            operation  = h.operation
            value      = h.value != "" ? h.value : null
            expression = h.expression != "" ? h.expression : null
          }
        } : null

        # Redirect
        from_value = r.action_parameters.from_value == null ? null : {
          status_code           = r.action_parameters.from_value.status_code
          preserve_query_string = r.action_parameters.from_value.preserve_query_string
          target_url = {
            value      = r.action_parameters.from_value.target_url.value != "" ? r.action_parameters.from_value.target_url.value : null
            expression = r.action_parameters.from_value.target_url.expression != "" ? r.action_parameters.from_value.target_url.expression : null
          }
        }

        # Skip
        phases   = length(r.action_parameters.phases) > 0 ? r.action_parameters.phases : null
        products = length(r.action_parameters.products) > 0 ? r.action_parameters.products : null
        ruleset  = r.action_parameters.ruleset != "" ? r.action_parameters.ruleset : null
        rulesets = length(r.action_parameters.rulesets) > 0 ? r.action_parameters.rulesets : null

        # Execute
        id = r.action_parameters.id != "" ? r.action_parameters.id : null

        overrides = r.action_parameters.overrides == null ? null : {
          action            = r.action_parameters.overrides.action != "" ? r.action_parameters.overrides.action : null
          enabled           = r.action_parameters.overrides.enabled
          sensitivity_level = r.action_parameters.overrides.sensitivity_level != "" ? r.action_parameters.overrides.sensitivity_level : null

          categories = length(r.action_parameters.overrides.categories) > 0 ? [
            for c in r.action_parameters.overrides.categories : {
              category          = c.category
              action            = c.action
              enabled           = c.enabled
              sensitivity_level = c.sensitivity_level != "" ? c.sensitivity_level : null
            }
          ] : null

          rules = length(r.action_parameters.overrides.rules) > 0 ? [
            for ru in r.action_parameters.overrides.rules : {
              id                = ru.id
              action            = ru.action
              enabled           = ru.enabled
              score_threshold   = ru.score_threshold > 0 ? ru.score_threshold : null
              sensitivity_level = ru.sensitivity_level != "" ? ru.sensitivity_level : null
            }
          ] : null
        }

        # Cache
        cache = r.action_parameters.cache

        edge_ttl = r.action_parameters.edge_ttl == null ? null : {
          mode    = r.action_parameters.edge_ttl.mode
          default = r.action_parameters.edge_ttl.default_ttl > 0 ? r.action_parameters.edge_ttl.default_ttl : null

          status_code_ttl = length(r.action_parameters.edge_ttl.status_code_ttls) > 0 ? [
            for s in r.action_parameters.edge_ttl.status_code_ttls : {
              value       = s.value
              status_code = s.status_code
              status_code_range = s.status_code_range == null ? null : {
                from = s.status_code_range.from
                to   = s.status_code_range.to
              }
            }
          ] : null
        }

        browser_ttl = r.action_parameters.browser_ttl == null ? null : {
          mode    = r.action_parameters.browser_ttl.mode
          default = r.action_parameters.browser_ttl.default_ttl > 0 ? r.action_parameters.browser_ttl.default_ttl : null
        }

        serve_stale = r.action_parameters.serve_stale == null ? null : {
          disable_stale_while_updating = r.action_parameters.serve_stale.disable_stale_while_updating
        }
      }
    }
  ]
}
