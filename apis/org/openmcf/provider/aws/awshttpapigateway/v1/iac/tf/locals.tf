locals {
  # Resource name from metadata
  api_name = var.metadata.name

  # AWS tags from metadata
  tags = merge({
    Name = local.api_name
  }, try(var.metadata.labels, {}))

  # Stage configuration with defaults
  stage_config = var.spec.stage != null ? var.spec.stage : {}
  stage_name   = coalesce(try(local.stage_config.name, null), "$default")
  auto_deploy  = coalesce(try(local.stage_config.auto_deploy, null), local.stage_name == "$default" ? true : false)

  # Integration deduplication: create a map of unique integration keys to integration config
  # Integration key format: "{integration_type}:{integration_uri}:{payload_format_version}"
  # Routes with the same integration_type, integration_uri, and payload_format_version
  # share one underlying API Gateway integration resource.
  integration_map = {
    for route in var.spec.routes : format("%s:%s:%s",
      route.integration.integration_type,
      route.integration.integration_uri.value,
      coalesce(route.integration.payload_format_version, "2.0")
    ) => route.integration
  }

  # Route-to-integration mapping: map each route index (as string) to its deduplicated integration key
  # This helps us reference the correct integration when creating routes
  route_integration_keys = {
    for idx, route in var.spec.routes : tostring(idx) => format("%s:%s:%s",
      route.integration.integration_type,
      route.integration.integration_uri.value,
      coalesce(route.integration.payload_format_version, "2.0")
    )
  }

  # Authorizer map: create a map of authorizer name to authorizer config
  authorizer_map = {
    for auth in var.spec.authorizers : auth.name => auth
  }
}

