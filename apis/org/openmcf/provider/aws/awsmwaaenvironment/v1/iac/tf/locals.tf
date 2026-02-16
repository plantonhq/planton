locals {
  tags = {
    "planton.org/resource"      = "true"
    "planton.org/organization"  = var.metadata.org
    "planton.org/environment"   = var.metadata.env
    "planton.org/resource-kind" = "AwsMwaaEnvironment"
    "planton.org/resource-id"   = var.metadata.id
  }

  has_ingress_refs = length(var.spec.security_group_ids) > 0 || length(var.spec.allowed_cidr_blocks) > 0

  # Combine managed SG (if created) with associate_security_group_ids
  effective_security_group_ids = concat(
    var.spec.associate_security_group_ids,
    local.has_ingress_refs ? [aws_security_group.environment[0].id] : []
  )
}
