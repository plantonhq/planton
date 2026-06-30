resource "oci_apigateway_deployment" "this" {
  compartment_id = var.spec.compartment_id.value
  gateway_id     = oci_apigateway_gateway.this.id
  path_prefix    = var.spec.deployment.path_prefix
  display_name   = local.deployment_display_name
  freeform_tags  = local.freeform_tags

  specification {

    # ── Logging Policies ─────────────────────────────────────────────

    dynamic "logging_policies" {
      for_each = var.spec.deployment.logging_policies != null ? [var.spec.deployment.logging_policies] : []
      content {
        dynamic "access_log" {
          for_each = logging_policies.value.access_log != null ? [logging_policies.value.access_log] : []
          content {
            is_enabled = access_log.value.is_enabled
          }
        }
        dynamic "execution_log" {
          for_each = logging_policies.value.execution_log != null ? [logging_policies.value.execution_log] : []
          content {
            is_enabled = execution_log.value.is_enabled
            log_level  = execution_log.value.log_level != "" ? lookup(local.log_level_map, execution_log.value.log_level, null) : null
          }
        }
      }
    }

    # ── Request Policies ─────────────────────────────────────────────

    dynamic "request_policies" {
      for_each = var.spec.deployment.request_policies != null ? [var.spec.deployment.request_policies] : []
      content {

        # ── JWT Authentication ─────────────────────────────────────

        dynamic "authentication" {
          for_each = request_policies.value.authentication != null ? [request_policies.value.authentication] : []
          content {
            type                         = "JWT_AUTHENTICATION"
            issuers                      = length(authentication.value.issuers) > 0 ? authentication.value.issuers : null
            audiences                    = length(authentication.value.audiences) > 0 ? authentication.value.audiences : null
            token_header                 = authentication.value.token_header != "" ? authentication.value.token_header : null
            token_query_param            = authentication.value.token_query_param != "" ? authentication.value.token_query_param : null
            token_auth_scheme            = authentication.value.token_auth_scheme != "" ? authentication.value.token_auth_scheme : null
            max_clock_skew_in_seconds    = authentication.value.max_clock_skew_in_seconds
            is_anonymous_access_allowed  = authentication.value.is_anonymous_access_allowed

            public_keys {
              type                        = lookup(local.public_key_type_map, authentication.value.public_keys.type, null)
              uri                         = authentication.value.public_keys.uri != "" ? authentication.value.public_keys.uri : null
              is_ssl_verify_disabled      = authentication.value.public_keys.is_ssl_verify_disabled
              max_cache_duration_in_hours = authentication.value.public_keys.max_cache_duration_in_hours

              dynamic "keys" {
                for_each = authentication.value.public_keys.keys
                content {
                  kid    = keys.value.kid
                  format = lookup(local.key_format_map, keys.value.format, null)
                  key    = keys.value.key != "" ? keys.value.key : null
                  kty    = keys.value.kty != "" ? keys.value.kty : null
                  alg    = keys.value.alg != "" ? keys.value.alg : null
                  n      = keys.value.n != "" ? keys.value.n : null
                  e      = keys.value.e != "" ? keys.value.e : null
                  use    = keys.value.use != "" ? keys.value.use : null
                }
              }
            }

            dynamic "verify_claims" {
              for_each = authentication.value.verify_claims
              content {
                key         = verify_claims.value.key != "" ? verify_claims.value.key : null
                values      = length(verify_claims.value.values) > 0 ? verify_claims.value.values : null
                is_required = verify_claims.value.is_required
              }
            }
          }
        }

        # ── CORS ──────────────────────────────────────────────────

        dynamic "cors" {
          for_each = request_policies.value.cors != null ? [request_policies.value.cors] : []
          content {
            allowed_origins              = cors.value.allowed_origins
            allowed_methods              = length(cors.value.allowed_methods) > 0 ? cors.value.allowed_methods : null
            allowed_headers              = length(cors.value.allowed_headers) > 0 ? cors.value.allowed_headers : null
            exposed_headers              = length(cors.value.exposed_headers) > 0 ? cors.value.exposed_headers : null
            is_allow_credentials_enabled = cors.value.is_allow_credentials_enabled
            max_age_in_seconds           = cors.value.max_age_in_seconds
          }
        }

        # ── Rate Limiting ─────────────────────────────────────────

        dynamic "rate_limiting" {
          for_each = request_policies.value.rate_limiting != null ? [request_policies.value.rate_limiting] : []
          content {
            rate_in_requests_per_second = rate_limiting.value.rate_in_requests_per_second
            rate_key                    = lookup(local.rate_key_map, rate_limiting.value.rate_key, null)
          }
        }
      }
    }

    # ── Routes ───────────────────────────────────────────────────────

    dynamic "routes" {
      for_each = var.spec.deployment.routes
      content {
        path    = routes.value.path
        methods = length(routes.value.methods) > 0 ? routes.value.methods : null

        backend {
          type                       = lookup(local.backend_type_map, routes.value.backend.type, null)
          url                        = routes.value.backend.url != "" ? routes.value.backend.url : null
          function_id                = routes.value.backend.function_id != "" ? routes.value.backend.function_id : null
          status                     = routes.value.backend.status != 0 ? routes.value.backend.status : null
          body                       = routes.value.backend.body != "" ? routes.value.backend.body : null
          connect_timeout_in_seconds = routes.value.backend.connect_timeout_in_seconds
          read_timeout_in_seconds    = routes.value.backend.read_timeout_in_seconds
          send_timeout_in_seconds    = routes.value.backend.send_timeout_in_seconds
          is_ssl_verify_disabled     = routes.value.backend.is_ssl_verify_disabled

          dynamic "headers" {
            for_each = routes.value.backend.headers
            content {
              name  = headers.value.name
              value = headers.value.value
            }
          }
        }

        dynamic "request_policies" {
          for_each = routes.value.authorization != null ? [routes.value.authorization] : []
          content {
            authorization {
              type          = lookup(local.authorization_type_map, request_policies.value.type, null)
              allowed_scope = length(request_policies.value.allowed_scope) > 0 ? request_policies.value.allowed_scope : null
            }
          }
        }

        dynamic "logging_policies" {
          for_each = routes.value.logging_policies != null ? [routes.value.logging_policies] : []
          content {
            dynamic "access_log" {
              for_each = logging_policies.value.access_log != null ? [logging_policies.value.access_log] : []
              content {
                is_enabled = access_log.value.is_enabled
              }
            }
            dynamic "execution_log" {
              for_each = logging_policies.value.execution_log != null ? [logging_policies.value.execution_log] : []
              content {
                is_enabled = execution_log.value.is_enabled
                log_level  = execution_log.value.log_level != "" ? lookup(local.log_level_map, execution_log.value.log_level, null) : null
              }
            }
          }
        }
      }
    }
  }
}
