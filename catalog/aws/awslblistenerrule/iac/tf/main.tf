# A listener rule is the unit of per-service routing on an ALB listener:
# each service deploys its own condition-action pair and removes it when the
# service goes away, while the shared listener stays untouched. The listener
# is create-only (moving a rule replaces it); priority, conditions, actions,
# and transforms update in place -- priority via a dedicated AWS API, so
# re-prioritizing rules never interrupts traffic.
resource "aws_lb_listener_rule" "this" {
  listener_arn = var.spec.listener_arn

  # Unset priority lets AWS append after the current highest -- fine for
  # append-only routing; rules that shadow each other set it explicitly.
  priority = var.spec.priority > 0 ? var.spec.priority : null

  # Condition blocks AND together; the values inside one block OR together,
  # per AWS semantics. Exactly one criterion is set per block (spec
  # validation guarantees it), so each dynamic block fires at most once per
  # condition.
  dynamic "condition" {
    for_each = var.spec.conditions
    content {
      dynamic "host_header" {
        for_each = condition.value.host_header != null ? [condition.value.host_header] : []
        content {
          values       = length(host_header.value.values) > 0 ? host_header.value.values : null
          regex_values = length(host_header.value.regex_values) > 0 ? host_header.value.regex_values : null
        }
      }
      dynamic "path_pattern" {
        for_each = condition.value.path_pattern != null ? [condition.value.path_pattern] : []
        content {
          values       = length(path_pattern.value.values) > 0 ? path_pattern.value.values : null
          regex_values = length(path_pattern.value.regex_values) > 0 ? path_pattern.value.regex_values : null
        }
      }
      dynamic "http_header" {
        for_each = condition.value.http_header != null ? [condition.value.http_header] : []
        content {
          http_header_name = http_header.value.http_header_name
          values           = length(http_header.value.values) > 0 ? http_header.value.values : null
          regex_values     = length(http_header.value.regex_values) > 0 ? http_header.value.regex_values : null
        }
      }
      dynamic "http_request_method" {
        for_each = condition.value.http_request_method != null ? [condition.value.http_request_method] : []
        content {
          values = http_request_method.value.values
        }
      }
      dynamic "query_string" {
        for_each = condition.value.query_string != null ? condition.value.query_string.pairs : []
        content {
          key   = query_string.value.key != "" ? query_string.value.key : null
          value = query_string.value.value
        }
      }
      dynamic "source_ip" {
        for_each = condition.value.source_ip != null ? [condition.value.source_ip] : []
        content {
          values = source_ip.value.values
        }
      }
    }
  }

  # The action chain. Exactly one configuration object matches each action's
  # type (spec validation guarantees it), so each dynamic block below fires
  # for at most one of the six shapes per action.
  dynamic "action" {
    for_each = var.spec.actions
    content {
      type  = action.value.type
      order = action.value.order > 0 ? action.value.order : null

      # A single unweighted target group uses the simple target_group_arn
      # form; AWS treats the weighted forward block and the simple ARN as
      # different configurations, and the simple form avoids spurious diffs
      # on the common case. (Mirrors the Pulumi module exactly.)
      target_group_arn = (
        action.value.forward != null &&
        try(length(action.value.forward.target_groups), 0) == 1 &&
        try(action.value.forward.stickiness, null) == null &&
        try(action.value.forward.target_groups[0].weight, 0) == 0
      ) ? action.value.forward.target_groups[0].arn : null

      dynamic "forward" {
        for_each = (
          action.value.forward != null &&
          !(
            try(length(action.value.forward.target_groups), 0) == 1 &&
            try(action.value.forward.stickiness, null) == null &&
            try(action.value.forward.target_groups[0].weight, 0) == 0
          )
        ) ? [action.value.forward] : []
        content {
          dynamic "target_group" {
            for_each = forward.value.target_groups
            content {
              arn    = target_group.value.arn
              weight = target_group.value.weight > 0 ? target_group.value.weight : null
            }
          }
          # Stickiness is written only when enabled: AWS rejects a duration of
          # 0, and a forward block without stickiness is stickiness off.
          dynamic "stickiness" {
            for_each = try(forward.value.stickiness.enabled, false) ? [forward.value.stickiness] : []
            content {
              enabled  = stickiness.value.enabled
              duration = stickiness.value.duration_seconds
            }
          }
        }
      }

      dynamic "redirect" {
        for_each = action.value.redirect != null ? [action.value.redirect] : []
        content {
          status_code = redirect.value.status_code
          protocol    = redirect.value.protocol != "" ? redirect.value.protocol : null
          port        = redirect.value.port != "" ? redirect.value.port : null
          host        = redirect.value.host != "" ? redirect.value.host : null
          path        = redirect.value.path != "" ? redirect.value.path : null
          query       = redirect.value.query != "" ? redirect.value.query : null
        }
      }

      dynamic "fixed_response" {
        for_each = action.value.fixed_response != null ? [action.value.fixed_response] : []
        content {
          content_type = fixed_response.value.content_type
          status_code  = fixed_response.value.status_code != "" ? fixed_response.value.status_code : null
          message_body = fixed_response.value.message_body != "" ? fixed_response.value.message_body : null
        }
      }

      dynamic "authenticate_cognito" {
        for_each = action.value.authenticate_cognito != null ? [action.value.authenticate_cognito] : []
        content {
          user_pool_arn                       = authenticate_cognito.value.user_pool_arn
          user_pool_client_id                 = authenticate_cognito.value.user_pool_client_id
          user_pool_domain                    = authenticate_cognito.value.user_pool_domain
          authentication_request_extra_params = length(authenticate_cognito.value.authentication_request_extra_params) > 0 ? authenticate_cognito.value.authentication_request_extra_params : null
          on_unauthenticated_request          = authenticate_cognito.value.on_unauthenticated_request != "" ? authenticate_cognito.value.on_unauthenticated_request : null
          scope                               = authenticate_cognito.value.scope != "" ? authenticate_cognito.value.scope : null
          session_cookie_name                 = authenticate_cognito.value.session_cookie_name != "" ? authenticate_cognito.value.session_cookie_name : null
          session_timeout                     = authenticate_cognito.value.session_timeout_seconds > 0 ? authenticate_cognito.value.session_timeout_seconds : null
        }
      }

      dynamic "authenticate_oidc" {
        for_each = action.value.authenticate_oidc != null ? [action.value.authenticate_oidc] : []
        content {
          issuer                              = authenticate_oidc.value.issuer
          authorization_endpoint              = authenticate_oidc.value.authorization_endpoint
          token_endpoint                      = authenticate_oidc.value.token_endpoint
          user_info_endpoint                  = authenticate_oidc.value.user_info_endpoint
          client_id                           = authenticate_oidc.value.client_id
          client_secret                       = authenticate_oidc.value.client_secret
          authentication_request_extra_params = length(authenticate_oidc.value.authentication_request_extra_params) > 0 ? authenticate_oidc.value.authentication_request_extra_params : null
          on_unauthenticated_request          = authenticate_oidc.value.on_unauthenticated_request != "" ? authenticate_oidc.value.on_unauthenticated_request : null
          scope                               = authenticate_oidc.value.scope != "" ? authenticate_oidc.value.scope : null
          session_cookie_name                 = authenticate_oidc.value.session_cookie_name != "" ? authenticate_oidc.value.session_cookie_name : null
          session_timeout                     = authenticate_oidc.value.session_timeout_seconds > 0 ? authenticate_oidc.value.session_timeout_seconds : null
        }
      }

      dynamic "jwt_validation" {
        for_each = action.value.jwt_validation != null ? [action.value.jwt_validation] : []
        content {
          issuer        = jwt_validation.value.issuer
          jwks_endpoint = jwt_validation.value.jwks_endpoint
          dynamic "additional_claim" {
            for_each = jwt_validation.value.additional_claims
            content {
              name   = additional_claim.value.name
              format = additional_claim.value.format
              values = additional_claim.value.values
            }
          }
        }
      }
    }
  }

  # Request rewrites applied before the action runs. Exactly one rewrite
  # config matches each transform's type (spec validation guarantees it).
  dynamic "transform" {
    for_each = var.spec.transforms
    content {
      type = transform.value.type
      dynamic "host_header_rewrite_config" {
        for_each = transform.value.host_header_rewrite != null ? [transform.value.host_header_rewrite] : []
        content {
          rewrite {
            regex   = host_header_rewrite_config.value.regex
            replace = host_header_rewrite_config.value.replace
          }
        }
      }
      dynamic "url_rewrite_config" {
        for_each = transform.value.url_rewrite != null ? [transform.value.url_rewrite] : []
        content {
          rewrite {
            regex   = url_rewrite_config.value.regex
            replace = url_rewrite_config.value.replace
          }
        }
      }
    }
  }

  tags = local.aws_tags
}
