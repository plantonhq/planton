locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  deployment_display_name = (
    var.spec.deployment.display_name != ""
    ? var.spec.deployment.display_name
    : "${local.display_name}-deployment"
  )

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciApiGateway"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  endpoint_type_map = {
    "endpoint_type_public"  = "PUBLIC"
    "endpoint_type_private" = "PRIVATE"
  }

  log_level_map = {
    "info"  = "INFO"
    "warn"  = "WARN"
    "error" = "ERROR"
  }

  backend_type_map = {
    "http"             = "HTTP_BACKEND"
    "oracle_functions" = "ORACLE_FUNCTIONS_BACKEND"
    "stock_response"   = "STOCK_RESPONSE_BACKEND"
  }

  public_key_type_map = {
    "remote_jwks" = "REMOTE_JWKS"
    "static_keys" = "STATIC_KEYS"
  }

  key_format_map = {
    "pem"          = "PEM"
    "json_web_key" = "JSON_WEB_KEY"
  }

  rate_key_map = {
    "client_ip" = "CLIENT_IP"
    "total"     = "TOTAL"
  }

  authorization_type_map = {
    "anonymous"           = "ANONYMOUS"
    "any_of"              = "ANY_OF"
    "authentication_only" = "AUTHENTICATION_ONLY"
  }
}
