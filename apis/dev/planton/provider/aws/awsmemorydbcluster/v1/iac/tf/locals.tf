locals {
  tags = {
    "planton.dev/resource"      = "true"
    "planton.dev/organization"  = var.metadata.org
    "planton.dev/environment"   = var.metadata.env
    "planton.dev/resource-kind" = "AwsMemorydbCluster"
    "planton.dev/resource-id"   = var.metadata.id
  }

  create_subnet_group    = length(var.spec.subnet_ids) > 0
  create_parameter_group = length(var.spec.parameters) > 0 && var.spec.parameter_group_family != null
}
