locals {
  cluster_identifier = var.metadata.name

  tags = {
    "planton.dev/resource"      = "true"
    "planton.dev/organization"  = var.metadata.org
    "planton.dev/environment"   = var.metadata.env
    "planton.dev/resource-kind" = "AwsRedshiftCluster"
    "planton.dev/resource-id"   = var.metadata.id
  }

  # Determine cluster type based on node count.
  cluster_type = var.spec.number_of_nodes > 1 ? "multi-node" : "single-node"

  # Networking conditionals.
  create_subnet_group   = length(var.spec.subnet_ids) >= 2
  create_security_group = length(var.spec.security_group_ids) > 0 || length(var.spec.allowed_cidr_blocks) > 0

  # Combine managed SG (if created) with explicitly associated SGs.
  all_security_group_ids = concat(
    local.create_security_group ? [aws_security_group.this[0].id] : [],
    var.spec.associate_security_group_ids
  )

  # Parameter group conditionals.
  create_parameter_group = length(var.spec.parameters) > 0
  effective_parameter_group_name = (
    local.create_parameter_group
    ? aws_redshift_parameter_group.this[0].name
    : var.spec.cluster_parameter_group_name != "" ? var.spec.cluster_parameter_group_name : null
  )

  # Logging conditional.
  create_logging = var.spec.logging != null
}
