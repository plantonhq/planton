locals {
  # Convenience locals
  metadata        = var.metadata
  spec            = var.spec
  is_ssl_enabled  = try(var.spec.ssl.enabled, false)
  certificate_arn = try(var.spec.ssl.certificate_arn, null)

  # resource name and tags
  resource_name = coalesce(try(var.metadata.name, null), "aws-alb")
  tags          = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # Foreign-key types are already flattened to primitive strings by the tofu
  # generator (the orchestrator resolves any value_from before the module runs).
  subnet_ids = try(var.spec.subnets, [])

  security_group_ids = try(var.spec.security_groups, [])

  # dns helpers
  create_dns_records = try(var.spec.dns.enabled, false) && length(try(var.spec.dns.hostnames, [])) > 0
  route53_zone_id    = try(var.spec.dns.route53_zone_id, null)
}


