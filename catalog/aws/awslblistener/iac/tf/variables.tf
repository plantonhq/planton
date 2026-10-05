variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsLbListener specification"
  type = object({
    region = string
    load_balancer_arn = string
    port = number
    protocol = string
    certificate_arn = optional(string, "")
    additional_certificate_arns = optional(list(string), [])
    ssl_policy = optional(string, "")
    alpn_policy = optional(string, "")
    mutual_authentication = optional(object({
      mode = string
      trust_store_arn = optional(string, "")
      ignore_client_certificate_expiry = optional(bool, false)
      advertise_trust_store_ca_names = optional(string, "")
    }))
    tcp_idle_timeout_seconds = optional(number, 0)
    # An action chain: every entry carries exactly the configuration object
    # matching its type (enforced by spec validation before the module runs).
    default_actions = list(object({
      type = string
      order = optional(number, 0)
      forward = optional(object({
        target_groups = list(object({
          arn = string
          weight = optional(number, 0)
        }))
        stickiness = optional(object({
          enabled = optional(bool, false)
          duration_seconds = optional(number, 0)
        }))
      }))
      redirect = optional(object({
        status_code = string
        protocol = optional(string, "")
        port = optional(string, "")
        host = optional(string, "")
        path = optional(string, "")
        query = optional(string, "")
      }))
      fixed_response = optional(object({
        content_type = string
        status_code = optional(string, "")
        message_body = optional(string, "")
      }))
      authenticate_cognito = optional(object({
        user_pool_arn = string
        user_pool_client_id = string
        user_pool_domain = string
        authentication_request_extra_params = optional(map(string), {})
        on_unauthenticated_request = optional(string, "")
        scope = optional(string, "")
        session_cookie_name = optional(string, "")
        session_timeout_seconds = optional(number, 0)
      }))
      authenticate_oidc = optional(object({
        issuer = string
        authorization_endpoint = string
        token_endpoint = string
        user_info_endpoint = string
        client_id = string
        client_secret = string
        authentication_request_extra_params = optional(map(string), {})
        on_unauthenticated_request = optional(string, "")
        scope = optional(string, "")
        session_cookie_name = optional(string, "")
        session_timeout_seconds = optional(number, 0)
      }))
      jwt_validation = optional(object({
        issuer = string
        jwks_endpoint = string
        additional_claims = optional(list(object({
          name = string
          format = string
          values = list(string)
        })), [])
      }))
    }))
    http_headers = optional(object({
      request = optional(object({
        mtls_clientcert_header_name = optional(string, "")
        mtls_clientcert_issuer_header_name = optional(string, "")
        mtls_clientcert_leaf_header_name = optional(string, "")
        mtls_clientcert_serial_number_header_name = optional(string, "")
        mtls_clientcert_subject_header_name = optional(string, "")
        mtls_clientcert_validity_header_name = optional(string, "")
        tls_cipher_suite_header_name = optional(string, "")
        tls_version_header_name = optional(string, "")
      }))
      response = optional(object({
        access_control_allow_credentials = optional(string, "")
        access_control_allow_headers = optional(string, "")
        access_control_allow_methods = optional(string, "")
        access_control_allow_origin = optional(string, "")
        access_control_expose_headers = optional(string, "")
        access_control_max_age = optional(string, "")
        content_security_policy = optional(string, "")
        server_enabled = optional(bool)
        strict_transport_security = optional(string, "")
        x_content_type_options = optional(string, "")
        x_frame_options = optional(string, "")
      }))
    }))
  })
}
