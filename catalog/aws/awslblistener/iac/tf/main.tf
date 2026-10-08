# A listener is the port/protocol entry point on a load balancer and the
# anchor both TLS material and per-service listener rules hang off. The load
# balancer is create-only (moving a listener replaces it); port, protocol,
# certificates, and actions update in place.
resource "aws_lb_listener" "this" {
  load_balancer_arn = var.spec.load_balancer_arn
  port              = var.spec.port
  protocol          = var.spec.protocol

  # TLS material only applies to HTTPS/TLS listeners; the certificate is the
  # DEFAULT one (clients that match no SNI certificate get it).
  certificate_arn = local.is_tls_protocol ? var.spec.certificate_arn : null
  ssl_policy      = var.spec.ssl_policy != "" ? var.spec.ssl_policy : null
  alpn_policy     = var.spec.alpn_policy != "" ? var.spec.alpn_policy : null

  # NLB TCP listeners only; 0 keeps the AWS default (350s).
  tcp_idle_timeout_seconds = var.spec.tcp_idle_timeout_seconds > 0 ? var.spec.tcp_idle_timeout_seconds : null

  dynamic "mutual_authentication" {
    for_each = var.spec.mutual_authentication != null ? [var.spec.mutual_authentication] : []
    content {
      mode                             = mutual_authentication.value.mode
      trust_store_arn                  = mutual_authentication.value.trust_store_arn != "" ? mutual_authentication.value.trust_store_arn : null
      ignore_client_certificate_expiry = mutual_authentication.value.ignore_client_certificate_expiry ? true : null
      advertise_trust_store_ca_names   = mutual_authentication.value.advertise_trust_store_ca_names != "" ? mutual_authentication.value.advertise_trust_store_ca_names : null
    }
  }

  # The action chain. Exactly one configuration object matches each action's
  # type (spec validation guarantees it), so each dynamic block below fires
  # for at most one of the six shapes per action.
  dynamic "default_action" {
    for_each = var.spec.default_actions
    content {
      type  = default_action.value.type
      order = default_action.value.order > 0 ? default_action.value.order : null

      # A single unweighted target group uses the simple target_group_arn
      # form; AWS treats the weighted forward block and the simple ARN as
      # different configurations, and the simple form avoids spurious diffs
      # on the common case. (Mirrors the Pulumi module exactly.)
      target_group_arn = (
        default_action.value.forward != null &&
        try(length(default_action.value.forward.target_groups), 0) == 1 &&
        try(default_action.value.forward.stickiness, null) == null &&
        try(default_action.value.forward.target_groups[0].weight, 0) == 0
      ) ? default_action.value.forward.target_groups[0].arn : null

      dynamic "forward" {
        for_each = (
          default_action.value.forward != null &&
          !(
            try(length(default_action.value.forward.target_groups), 0) == 1 &&
            try(default_action.value.forward.stickiness, null) == null &&
            try(default_action.value.forward.target_groups[0].weight, 0) == 0
          )
        ) ? [default_action.value.forward] : []
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
        for_each = default_action.value.redirect != null ? [default_action.value.redirect] : []
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
        for_each = default_action.value.fixed_response != null ? [default_action.value.fixed_response] : []
        content {
          content_type = fixed_response.value.content_type
          status_code  = fixed_response.value.status_code != "" ? fixed_response.value.status_code : null
          message_body = fixed_response.value.message_body != "" ? fixed_response.value.message_body : null
        }
      }

      dynamic "authenticate_cognito" {
        for_each = default_action.value.authenticate_cognito != null ? [default_action.value.authenticate_cognito] : []
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
        for_each = default_action.value.authenticate_oidc != null ? [default_action.value.authenticate_oidc] : []
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
        for_each = default_action.value.jwt_validation != null ? [default_action.value.jwt_validation] : []
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

  # HTTPS-only request-header injection (TLS/mTLS connection details toward
  # targets). Unset fields inject nothing.
  routing_http_request_x_amzn_mtls_clientcert_header_name               = local.request_headers != null && try(local.request_headers.mtls_clientcert_header_name, "") != "" ? local.request_headers.mtls_clientcert_header_name : null
  routing_http_request_x_amzn_mtls_clientcert_issuer_header_name        = local.request_headers != null && try(local.request_headers.mtls_clientcert_issuer_header_name, "") != "" ? local.request_headers.mtls_clientcert_issuer_header_name : null
  routing_http_request_x_amzn_mtls_clientcert_leaf_header_name          = local.request_headers != null && try(local.request_headers.mtls_clientcert_leaf_header_name, "") != "" ? local.request_headers.mtls_clientcert_leaf_header_name : null
  routing_http_request_x_amzn_mtls_clientcert_serial_number_header_name = local.request_headers != null && try(local.request_headers.mtls_clientcert_serial_number_header_name, "") != "" ? local.request_headers.mtls_clientcert_serial_number_header_name : null
  routing_http_request_x_amzn_mtls_clientcert_subject_header_name       = local.request_headers != null && try(local.request_headers.mtls_clientcert_subject_header_name, "") != "" ? local.request_headers.mtls_clientcert_subject_header_name : null
  routing_http_request_x_amzn_mtls_clientcert_validity_header_name      = local.request_headers != null && try(local.request_headers.mtls_clientcert_validity_header_name, "") != "" ? local.request_headers.mtls_clientcert_validity_header_name : null
  routing_http_request_x_amzn_tls_cipher_suite_header_name              = local.request_headers != null && try(local.request_headers.tls_cipher_suite_header_name, "") != "" ? local.request_headers.tls_cipher_suite_header_name : null
  routing_http_request_x_amzn_tls_version_header_name                   = local.request_headers != null && try(local.request_headers.tls_version_header_name, "") != "" ? local.request_headers.tls_version_header_name : null

  # HTTP/HTTPS response-header overrides (CORS and browser security headers
  # served uniformly at the edge). Unset fields leave application headers
  # untouched.
  routing_http_response_access_control_allow_credentials_header_value = local.response_headers != null && try(local.response_headers.access_control_allow_credentials, "") != "" ? local.response_headers.access_control_allow_credentials : null
  routing_http_response_access_control_allow_headers_header_value     = local.response_headers != null && try(local.response_headers.access_control_allow_headers, "") != "" ? local.response_headers.access_control_allow_headers : null
  routing_http_response_access_control_allow_methods_header_value     = local.response_headers != null && try(local.response_headers.access_control_allow_methods, "") != "" ? local.response_headers.access_control_allow_methods : null
  routing_http_response_access_control_allow_origin_header_value      = local.response_headers != null && try(local.response_headers.access_control_allow_origin, "") != "" ? local.response_headers.access_control_allow_origin : null
  routing_http_response_access_control_expose_headers_header_value    = local.response_headers != null && try(local.response_headers.access_control_expose_headers, "") != "" ? local.response_headers.access_control_expose_headers : null
  routing_http_response_access_control_max_age_header_value           = local.response_headers != null && try(local.response_headers.access_control_max_age, "") != "" ? local.response_headers.access_control_max_age : null
  routing_http_response_content_security_policy_header_value          = local.response_headers != null && try(local.response_headers.content_security_policy, "") != "" ? local.response_headers.content_security_policy : null
  routing_http_response_server_enabled                                = local.response_headers != null ? local.response_headers.server_enabled : null
  routing_http_response_strict_transport_security_header_value        = local.response_headers != null && try(local.response_headers.strict_transport_security, "") != "" ? local.response_headers.strict_transport_security : null
  routing_http_response_x_content_type_options_header_value           = local.response_headers != null && try(local.response_headers.x_content_type_options, "") != "" ? local.response_headers.x_content_type_options : null
  routing_http_response_x_frame_options_header_value                  = local.response_headers != null && try(local.response_headers.x_frame_options, "") != "" ? local.response_headers.x_frame_options : null

  tags = local.aws_tags

  lifecycle {
    # The spec cannot enforce certificate requiredness per protocol (proto
    # validation cannot inspect reference fields), so the module is the
    # enforcement point: fail at plan time with a clear message instead of
    # letting AWS reject the apply.
    precondition {
      condition     = !local.is_tls_protocol || var.spec.certificate_arn != ""
      error_message = "certificate_arn is required when protocol is 'HTTPS' or 'TLS'."
    }
  }
}

# Additional SNI certificates: the client's SNI hostname picks the matching
# certificate and the default certificate serves the rest. AWS models each as
# a separate listener-certificate attachment; they fold into this module
# because an attachment is pure glue with no referenceable identity. Keyed by
# index because a certificate may be a reference resolved only at apply time.
resource "aws_lb_listener_certificate" "this" {
  count = length(var.spec.additional_certificate_arns)

  listener_arn    = aws_lb_listener.this.arn
  certificate_arn = var.spec.additional_certificate_arns[count.index]
}
