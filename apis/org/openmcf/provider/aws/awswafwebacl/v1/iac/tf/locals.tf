locals {
  resource_name = coalesce(try(var.metadata.name, null), "awswafwebacl")

  tags = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # Scope: REGIONAL or CLOUDFRONT.
  scope = var.spec.scope

  # Default action type: allow or block.
  default_action_type = var.spec.default_action.type

  # Rules as a list from the spec.
  rules = try(var.spec.rules, [])

  # Custom response bodies indexed by key.
  custom_response_bodies = {
    for body in try(var.spec.custom_response_bodies, []) :
    body.key => body
  }

  # Logging configuration (optional).
  logging_enabled = try(var.spec.logging, null) != null

  # Visibility config with smart defaults.
  acl_metric_name        = try(var.spec.visibility_config.metric_name, local.resource_name)
  acl_metrics_enabled    = try(var.spec.visibility_config.cloudwatch_metrics_enabled, true)
  acl_sampled_enabled    = try(var.spec.visibility_config.sampled_requests_enabled, true)
}
