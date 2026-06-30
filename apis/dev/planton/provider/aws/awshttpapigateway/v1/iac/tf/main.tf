# ---------------------------------------------------------------------------
# 1. Create the HTTP API
# ---------------------------------------------------------------------------

resource "aws_apigatewayv2_api" "this" {
  name          = local.api_name
  protocol_type = "HTTP"
  description   = try(var.spec.description, null) != "" ? var.spec.description : null

  # Disable default execute-api endpoint
  disable_execute_api_endpoint = var.spec.disable_execute_api_endpoint

  # CORS configuration
  dynamic "cors_configuration" {
    for_each = var.spec.cors_configuration != null ? [var.spec.cors_configuration] : []
    content {
      allow_origins     = cors_configuration.value.allow_origins
      allow_methods     = cors_configuration.value.allow_methods
      allow_headers     = cors_configuration.value.allow_headers
      expose_headers    = cors_configuration.value.expose_headers
      max_age           = cors_configuration.value.max_age_seconds > 0 ? cors_configuration.value.max_age_seconds : null
      allow_credentials = cors_configuration.value.allow_credentials
    }
  }

  tags = local.tags
}

# ---------------------------------------------------------------------------
# 2. Create the stage
# ---------------------------------------------------------------------------

resource "aws_apigatewayv2_stage" "this" {
  api_id      = aws_apigatewayv2_api.this.id
  name        = local.stage_name
  auto_deploy = local.auto_deploy

  # Access logging
  dynamic "access_log_settings" {
    for_each = try(local.stage_config.access_log, null) != null ? [local.stage_config.access_log] : []
    content {
      destination_arn = access_log_settings.value.destination_arn.value
      format          = access_log_settings.value.format
    }
  }

  # Default throttling
  dynamic "default_route_settings" {
    for_each = try(local.stage_config.default_throttle, null) != null && (
      try(local.stage_config.default_throttle.burst_limit, 0) > 0 ||
      try(local.stage_config.default_throttle.rate_limit, 0) > 0
    ) ? [local.stage_config.default_throttle] : []
    content {
      throttling_burst_limit = default_route_settings.value.burst_limit > 0 ? default_route_settings.value.burst_limit : null
      throttling_rate_limit  = default_route_settings.value.rate_limit > 0 ? default_route_settings.value.rate_limit : null
    }
  }

  # Stage variables
  stage_variables = try(local.stage_config.stage_variables, {})

  tags = local.tags
}

# ---------------------------------------------------------------------------
# 3. Create integrations (deduplicated)
# ---------------------------------------------------------------------------

resource "aws_apigatewayv2_integration" "this" {
  for_each = local.integration_map

  api_id                 = aws_apigatewayv2_api.this.id
  integration_type       = each.value.integration_type
  integration_uri        = each.value.integration_uri.value
  payload_format_version = coalesce(each.value.payload_format_version, "2.0")

  # Integration method (defaults to POST for Lambda, route method for HTTP)
  integration_method = try(each.value.integration_method, null) != "" ? each.value.integration_method : null

  # Timeout (50-30000 milliseconds)
  timeout_milliseconds = try(each.value.timeout_milliseconds, 0) > 0 ? each.value.timeout_milliseconds : null
}

# ---------------------------------------------------------------------------
# 4. Create authorizers (if any)
# ---------------------------------------------------------------------------

resource "aws_apigatewayv2_authorizer" "this" {
  for_each = local.authorizer_map

  api_id          = aws_apigatewayv2_api.this.id
  name            = each.value.name
  authorizer_type = each.value.authorizer_type

  # Identity sources
  identity_sources = length(each.value.identity_sources) > 0 ? each.value.identity_sources : null

  # JWT configuration
  dynamic "jwt_configuration" {
    for_each = each.value.authorizer_type == "JWT" && each.value.jwt_configuration != null ? [each.value.jwt_configuration] : []
    content {
      issuer    = jwt_configuration.value.issuer
      audiences = length(jwt_configuration.value.audiences) > 0 ? jwt_configuration.value.audiences : null
    }
  }

  # Lambda authorizer (REQUEST type)
  authorizer_uri                    = each.value.authorizer_type == "REQUEST" && try(each.value.authorizer_uri.value, null) != null ? each.value.authorizer_uri.value : null
  authorizer_credentials_arn        = each.value.authorizer_type == "REQUEST" && try(each.value.authorizer_credentials_arn.value, null) != null ? each.value.authorizer_credentials_arn.value : null
  enable_simple_responses           = each.value.authorizer_type == "REQUEST" && try(each.value.enable_simple_responses, false) ? true : null
  authorizer_payload_format_version = each.value.authorizer_type == "REQUEST" && try(each.value.authorizer_payload_format_version, null) != "" ? each.value.authorizer_payload_format_version : null

  # Cache TTL (0-3600 seconds)
  authorizer_result_ttl_in_seconds = try(each.value.result_ttl_seconds, 0) > 0 ? each.value.result_ttl_seconds : null
}

# ---------------------------------------------------------------------------
# 5. Create routes
# ---------------------------------------------------------------------------

resource "aws_apigatewayv2_route" "this" {
  for_each = {
    for idx, route in var.spec.routes : tostring(idx) => route
  }

  api_id    = aws_apigatewayv2_api.this.id
  route_key = each.value.route_key

  # Target format: "integrations/{integrationId}"
  target = "integrations/${aws_apigatewayv2_integration.this[local.route_integration_keys[each.key]].id}"

  # Authorization
  authorization_type = try(each.value.authorization_type, "") != "" && each.value.authorization_type != "NONE" ? each.value.authorization_type : null

  # Authorizer reference
  authorizer_id = try(each.value.authorization_type, "") == "JWT" && try(each.value.authorizer_name, "") != "" ? aws_apigatewayv2_authorizer.this[each.value.authorizer_name].id : null

  # Authorization scopes (JWT)
  authorization_scopes = try(each.value.authorization_type, "") == "JWT" && length(try(each.value.authorization_scopes, [])) > 0 ? each.value.authorization_scopes : null
}
