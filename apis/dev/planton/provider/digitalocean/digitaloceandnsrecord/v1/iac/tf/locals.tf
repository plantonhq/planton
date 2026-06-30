locals {
  # Extract domain from StringValueOrRef (use direct value)
  domain = var.spec.domain.value

  # DNS record configuration
  name  = var.spec.name
  type  = var.spec.type
  value = var.spec.value.value

  ttl_seconds = coalesce(var.spec.ttl_seconds, 1800)

  # Type-specific fields
  priority = var.spec.priority
  weight   = var.spec.weight
  port     = var.spec.port
  flags    = var.spec.flags
  tag      = var.spec.tag

  # Construct hostname for output
  hostname = local.name == "@" ? local.domain : "${local.name}.${local.domain}"
}
