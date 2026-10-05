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
  description = "AwsRestApiGateway specification"
  type = object({
    region                       = string
    description                  = optional(string, "")
    api_key_source               = optional(string, "")
    binary_media_types           = optional(list(string), [])
    minimum_compression_size     = optional(number)
    disable_execute_api_endpoint = optional(bool, false)
    endpoint_configuration = optional(object({
      type             = optional(string, "")
      ip_address_type  = optional(string, "")
      vpc_endpoint_ids = optional(list(string), [])
    }))
    endpoint_access_mode = optional(string, "")
    security_policy      = optional(string, "")
    policy               = optional(any)
    routes = optional(list(object({
      path                   = optional(string, "")
      method                 = optional(string, "")
      authorization          = optional(string, "")
      authorizer_name        = optional(string, "")
      authorization_scopes   = optional(list(string), [])
      api_key_required       = optional(bool, false)
      operation_name         = optional(string, "")
      request_parameters     = optional(map(bool), {})
      request_models         = optional(map(string), {})
      request_validator_name = optional(string, "")
      integration = object({
        type                           = optional(string, "")
        uri                            = optional(string, "")
        http_method                    = optional(string, "")
        credentials_arn                = optional(string, "")
        connection_type                = optional(string, "")
        vpc_link_id                    = optional(string, "")
        passthrough_behavior           = optional(string, "")
        content_handling               = optional(string, "")
        cache_key_parameters           = optional(list(string), [])
        cache_namespace                = optional(string, "")
        request_parameters             = optional(map(string), {})
        request_templates              = optional(map(string), {})
        timeout_milliseconds           = optional(number, 0)
        response_transfer_mode         = optional(string, "")
        tls_insecure_skip_verification = optional(bool, false)
      })
      responses = optional(list(object({
        status_code                     = optional(string, "")
        response_models                 = optional(map(string), {})
        response_parameters             = optional(map(bool), {})
        selection_pattern               = optional(string, "")
        integration_response_parameters = optional(map(string), {})
        integration_response_templates  = optional(map(string), {})
        content_handling                = optional(string, "")
      })), [])
    })), [])
    openapi = optional(object({
      body             = string
      fail_on_warnings = optional(bool, false)
      parameters       = optional(map(string), {})
      mode             = optional(string, "")
    }))
    models = optional(list(object({
      name         = string
      content_type = string
      description  = optional(string, "")
      schema       = optional(string, "")
    })), [])
    request_validators = optional(list(object({
      name                        = string
      validate_request_body       = optional(bool, false)
      validate_request_parameters = optional(bool, false)
    })), [])
    authorizers = optional(list(object({
      name                           = string
      type                           = optional(string, "")
      lambda_invoke_uri              = optional(string, "")
      credentials_arn                = optional(string, "")
      provider_arns                  = optional(list(string), [])
      identity_source                = optional(string, "")
      identity_validation_expression = optional(string, "")
      result_ttl_seconds             = optional(number)
    })), [])
    gateway_responses = optional(list(object({
      response_type       = optional(string, "")
      status_code         = optional(string, "")
      response_parameters = optional(map(string), {})
      response_templates  = optional(map(string), {})
    })), [])
    stage = optional(object({
      name                 = optional(string, "")
      description          = optional(string, "")
      stage_variables      = optional(map(string), {})
      xray_tracing_enabled = optional(bool, false)
      cache_cluster = optional(object({
        enabled = optional(bool, false)
        size    = optional(string, "")
      }))
      client_certificate = optional(object({
        generate                = optional(bool, false)
        existing_certificate_id = optional(string, "")
        description             = optional(string, "")
      }))
      access_log = optional(object({
        destination_arn = string
        format          = string
      }))
      documentation_version = optional(string, "")
      method_settings = optional(list(object({
        method_path                                = string
        metrics_enabled                            = optional(bool)
        logging_level                              = optional(string, "")
        data_trace_enabled                         = optional(bool)
        throttling_burst_limit                     = optional(number)
        throttling_rate_limit                      = optional(number)
        caching_enabled                            = optional(bool)
        cache_ttl_in_seconds                       = optional(number)
        cache_data_encrypted                       = optional(bool)
        require_authorization_for_cache_control    = optional(bool)
        unauthorized_cache_control_header_strategy = optional(string, "")
      })), [])
    }))
    documentation = optional(object({
      parts = optional(list(object({
        location = object({
          type        = optional(string, "")
          path        = optional(string, "")
          method      = optional(string, "")
          name        = optional(string, "")
          status_code = optional(string, "")
        })
        properties = string
      })), [])
      published_version = optional(object({
        version     = string
        description = optional(string, "")
      }))
    }))
  })
}