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

      ratelimit = r.ratelimit == null ? null : {
        characteristics            = r.ratelimit.characteristics
        period                     = r.ratelimit.period
        counting_expression        = r.ratelimit.counting_expression != "" ? r.ratelimit.counting_expression : null
        mitigation_timeout         = r.ratelimit.mitigation_timeout > 0 ? r.ratelimit.mitigation_timeout : null
        requests_per_period        = r.ratelimit.requests_per_period > 0 ? r.ratelimit.requests_per_period : null
        requests_to_origin         = r.ratelimit.requests_to_origin
        score_per_period           = r.ratelimit.score_per_period > 0 ? r.ratelimit.score_per_period : null
        score_response_header_name = r.ratelimit.score_response_header_name != "" ? r.ratelimit.score_response_header_name : null
      }

      logging = r.logging == null ? null : {
        enabled = r.logging.enabled
      }

      exposed_credential_check = r.exposed_credential_check == null ? null : {
        username_expression = r.exposed_credential_check.username_expression
        password_expression = r.exposed_credential_check.password_expression
      }

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

        # Block / serve_error inline response
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
        from_list = r.action_parameters.from_list == null ? null : {
          name = r.action_parameters.from_list.name
          key  = r.action_parameters.from_list.key
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

        matched_data = r.action_parameters.matched_data == null ? null : {
          public_key = r.action_parameters.matched_data.public_key
        }

        # Compress response
        algorithms = length(r.action_parameters.algorithms) > 0 ? [
          for a in r.action_parameters.algorithms : { name = a.name }
        ] : null

        # Score
        increment = r.action_parameters.increment > 0 ? r.action_parameters.increment : null

        # Serve error (inline)
        asset_name   = r.action_parameters.asset_name != "" ? r.action_parameters.asset_name : null
        content      = r.action_parameters.content != "" ? r.action_parameters.content : null
        content_type = r.action_parameters.content_type != "" ? r.action_parameters.content_type : null
        status_code  = r.action_parameters.status_code > 0 ? r.action_parameters.status_code : null

        # Configuration settings (set_config)
        automatic_https_rewrites = r.action_parameters.automatic_https_rewrites
        autominify = r.action_parameters.autominify == null ? null : {
          css  = r.action_parameters.autominify.css
          html = r.action_parameters.autominify.html
          js   = r.action_parameters.autominify.js
        }
        bic                       = r.action_parameters.bic
        content_converter         = r.action_parameters.content_converter
        disable_apps              = r.action_parameters.disable_apps
        disable_rum               = r.action_parameters.disable_rum
        disable_zaraz             = r.action_parameters.disable_zaraz
        email_obfuscation         = r.action_parameters.email_obfuscation
        fonts                     = r.action_parameters.fonts
        hotlink_protection        = r.action_parameters.hotlink_protection
        mirage                    = r.action_parameters.mirage
        opportunistic_encryption  = r.action_parameters.opportunistic_encryption
        polish                    = r.action_parameters.polish != "" ? r.action_parameters.polish : null
        redirects_for_ai_training = r.action_parameters.redirects_for_ai_training
        request_body_buffering    = r.action_parameters.request_body_buffering != "" ? r.action_parameters.request_body_buffering : null
        response_body_buffering   = r.action_parameters.response_body_buffering != "" ? r.action_parameters.response_body_buffering : null
        rocket_loader             = r.action_parameters.rocket_loader
        security_level            = r.action_parameters.security_level != "" ? r.action_parameters.security_level : null
        server_side_excludes      = r.action_parameters.server_side_excludes
        ssl                       = r.action_parameters.ssl != "" ? r.action_parameters.ssl : null
        sxg                       = r.action_parameters.sxg

        # Cache (set_cache_settings)
        cache                      = r.action_parameters.cache
        additional_cacheable_ports = length(r.action_parameters.additional_cacheable_ports) > 0 ? r.action_parameters.additional_cacheable_ports : null

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

        cache_key = r.action_parameters.cache_key == null ? null : {
          cache_by_device_type       = r.action_parameters.cache_key.cache_by_device_type
          cache_deception_armor      = r.action_parameters.cache_key.cache_deception_armor
          ignore_query_strings_order = r.action_parameters.cache_key.ignore_query_strings_order
          custom_key = r.action_parameters.cache_key.custom_key == null ? null : {
            cookie = r.action_parameters.cache_key.custom_key.cookie == null ? null : {
              check_presence = length(r.action_parameters.cache_key.custom_key.cookie.check_presence) > 0 ? r.action_parameters.cache_key.custom_key.cookie.check_presence : null
              include        = length(r.action_parameters.cache_key.custom_key.cookie.include) > 0 ? r.action_parameters.cache_key.custom_key.cookie.include : null
            }
            header = r.action_parameters.cache_key.custom_key.header == null ? null : {
              check_presence = length(r.action_parameters.cache_key.custom_key.header.check_presence) > 0 ? r.action_parameters.cache_key.custom_key.header.check_presence : null
              contains       = length(r.action_parameters.cache_key.custom_key.header.contains) > 0 ? r.action_parameters.cache_key.custom_key.header.contains : null
              exclude_origin = r.action_parameters.cache_key.custom_key.header.exclude_origin
              include        = length(r.action_parameters.cache_key.custom_key.header.include) > 0 ? r.action_parameters.cache_key.custom_key.header.include : null
            }
            host = r.action_parameters.cache_key.custom_key.host == null ? null : {
              resolved = r.action_parameters.cache_key.custom_key.host.resolved
            }
            query_string = r.action_parameters.cache_key.custom_key.query_string == null ? null : {
              include = r.action_parameters.cache_key.custom_key.query_string.include == null ? null : {
                list = length(r.action_parameters.cache_key.custom_key.query_string.include.list) > 0 ? r.action_parameters.cache_key.custom_key.query_string.include.list : null
                all  = r.action_parameters.cache_key.custom_key.query_string.include.all
              }
              exclude = r.action_parameters.cache_key.custom_key.query_string.exclude == null ? null : {
                list = length(r.action_parameters.cache_key.custom_key.query_string.exclude.list) > 0 ? r.action_parameters.cache_key.custom_key.query_string.exclude.list : null
                all  = r.action_parameters.cache_key.custom_key.query_string.exclude.all
              }
            }
            user = r.action_parameters.cache_key.custom_key.user == null ? null : {
              device_type = r.action_parameters.cache_key.custom_key.user.device_type
              geo         = r.action_parameters.cache_key.custom_key.user.geo
              lang        = r.action_parameters.cache_key.custom_key.user.lang
            }
          }
        }

        cache_reserve = r.action_parameters.cache_reserve == null ? null : {
          eligible          = r.action_parameters.cache_reserve.eligible
          minimum_file_size = r.action_parameters.cache_reserve.minimum_file_size > 0 ? r.action_parameters.cache_reserve.minimum_file_size : null
        }

        origin_cache_control       = r.action_parameters.origin_cache_control
        origin_error_page_passthru = r.action_parameters.origin_error_page_passthru
        read_timeout               = r.action_parameters.read_timeout > 0 ? r.action_parameters.read_timeout : null
        respect_strong_etags       = r.action_parameters.respect_strong_etags

        vary = r.action_parameters.vary == null ? null : {
          default = {
            action = r.action_parameters.vary.default.action
          }
          headers = length(r.action_parameters.vary.headers) > 0 ? {
            for name, h in r.action_parameters.vary.headers : name => {
              action      = h.action
              media_types = length(h.media_types) > 0 ? h.media_types : null
              languages   = length(h.languages) > 0 ? h.languages : null
            }
          } : null
        }

        strip_etags         = r.action_parameters.strip_etags
        strip_last_modified = r.action_parameters.strip_last_modified
        strip_set_cookie    = r.action_parameters.strip_set_cookie

        # Log custom fields
        cookie_fields = length(r.action_parameters.cookie_fields) > 0 ? [
          for f in r.action_parameters.cookie_fields : { name = f.name }
        ] : null
        raw_response_fields = length(r.action_parameters.raw_response_fields) > 0 ? [
          for f in r.action_parameters.raw_response_fields : { name = f.name, preserve_duplicates = f.preserve_duplicates }
        ] : null
        request_fields = length(r.action_parameters.request_fields) > 0 ? [
          for f in r.action_parameters.request_fields : { name = f.name }
        ] : null
        response_fields = length(r.action_parameters.response_fields) > 0 ? [
          for f in r.action_parameters.response_fields : { name = f.name, preserve_duplicates = f.preserve_duplicates }
        ] : null
        transformed_request_fields = length(r.action_parameters.transformed_request_fields) > 0 ? [
          for f in r.action_parameters.transformed_request_fields : { name = f.name }
        ] : null

        # Set Cache-Control directives
        max_age                = r.action_parameters.max_age == null ? null : { operation = r.action_parameters.max_age.operation, value = r.action_parameters.max_age.value > 0 ? r.action_parameters.max_age.value : null, cloudflare_only = r.action_parameters.max_age.cloudflare_only }
        s_maxage               = r.action_parameters.s_maxage == null ? null : { operation = r.action_parameters.s_maxage.operation, value = r.action_parameters.s_maxage.value > 0 ? r.action_parameters.s_maxage.value : null, cloudflare_only = r.action_parameters.s_maxage.cloudflare_only }
        stale_while_revalidate = r.action_parameters.stale_while_revalidate == null ? null : { operation = r.action_parameters.stale_while_revalidate.operation, value = r.action_parameters.stale_while_revalidate.value > 0 ? r.action_parameters.stale_while_revalidate.value : null, cloudflare_only = r.action_parameters.stale_while_revalidate.cloudflare_only }
        stale_if_error         = r.action_parameters.stale_if_error == null ? null : { operation = r.action_parameters.stale_if_error.operation, value = r.action_parameters.stale_if_error.value > 0 ? r.action_parameters.stale_if_error.value : null, cloudflare_only = r.action_parameters.stale_if_error.cloudflare_only }
        private                = r.action_parameters.private == null ? null : { operation = r.action_parameters.private.operation, qualifiers = length(r.action_parameters.private.qualifiers) > 0 ? r.action_parameters.private.qualifiers : null, cloudflare_only = r.action_parameters.private.cloudflare_only }
        no_cache               = r.action_parameters.no_cache == null ? null : { operation = r.action_parameters.no_cache.operation, qualifiers = length(r.action_parameters.no_cache.qualifiers) > 0 ? r.action_parameters.no_cache.qualifiers : null, cloudflare_only = r.action_parameters.no_cache.cloudflare_only }
        must_revalidate        = r.action_parameters.must_revalidate == null ? null : { operation = r.action_parameters.must_revalidate.operation, cloudflare_only = r.action_parameters.must_revalidate.cloudflare_only }
        proxy_revalidate       = r.action_parameters.proxy_revalidate == null ? null : { operation = r.action_parameters.proxy_revalidate.operation, cloudflare_only = r.action_parameters.proxy_revalidate.cloudflare_only }
        must_understand        = r.action_parameters.must_understand == null ? null : { operation = r.action_parameters.must_understand.operation, cloudflare_only = r.action_parameters.must_understand.cloudflare_only }
        no_transform           = r.action_parameters.no_transform == null ? null : { operation = r.action_parameters.no_transform.operation, cloudflare_only = r.action_parameters.no_transform.cloudflare_only }
        immutable              = r.action_parameters.immutable == null ? null : { operation = r.action_parameters.immutable.operation, cloudflare_only = r.action_parameters.immutable.cloudflare_only }
        no_store               = r.action_parameters.no_store == null ? null : { operation = r.action_parameters.no_store.operation, cloudflare_only = r.action_parameters.no_store.cloudflare_only }
        public                 = r.action_parameters.public == null ? null : { operation = r.action_parameters.public.operation, cloudflare_only = r.action_parameters.public.cloudflare_only }

        # Set cache tags
        operation  = r.action_parameters.operation != "" ? r.action_parameters.operation : null
        values     = length(r.action_parameters.values) > 0 ? r.action_parameters.values : null
        expression = r.action_parameters.expression != "" ? r.action_parameters.expression : null
      }
    }
  ]
}
