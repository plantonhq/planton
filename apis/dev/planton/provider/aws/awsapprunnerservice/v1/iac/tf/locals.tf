locals {
  service_name = var.metadata.name

  tags = {
    "planton.org/resource"      = "true"
    "planton.org/organization"  = var.metadata.org
    "planton.org/environment"   = var.metadata.env
    "planton.org/resource-kind" = "AwsAppRunnerService"
    "planton.org/resource-id"   = var.metadata.id
    "Name"                      = var.metadata.name
  }

  # VPC connector logic
  create_inline_vpc_connector = length(var.spec.subnet_ids) > 0
  use_external_vpc_connector  = var.spec.vpc_connector_arn != ""

  # Egress type: VPC when any connector is in use, DEFAULT otherwise
  egress_type = local.create_inline_vpc_connector || local.use_external_vpc_connector ? "VPC" : "DEFAULT"

  # Resolved VPC connector ARN (inline-created takes precedence over external)
  effective_vpc_connector_arn = (
    local.create_inline_vpc_connector
    ? aws_apprunner_vpc_connector.this[0].arn
    : (local.use_external_vpc_connector ? var.spec.vpc_connector_arn : null)
  )

  # Auth configuration is needed for private ECR images or code source connections
  needs_auth_config = (
    (var.spec.image_source != null && var.spec.image_source.access_role_arn != "") ||
    var.spec.code_source != null
  )
}
