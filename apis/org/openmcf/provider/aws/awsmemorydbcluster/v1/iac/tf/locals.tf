locals {
  tags = {
    "openmcf.org/resource"      = "true"
    "openmcf.org/organization"  = var.metadata.org
    "openmcf.org/environment"   = var.metadata.env
    "openmcf.org/resource-kind" = "AwsMemorydbCluster"
    "openmcf.org/resource-id"   = var.metadata.id
  }

  create_subnet_group    = length(var.spec.subnet_ids) > 0
  create_parameter_group = length(var.spec.parameters) > 0 && var.spec.parameter_group_family != null
}
