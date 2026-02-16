locals {
  workgroup_name = var.metadata.name

  tags = {
    "openmcf.org/resource"      = "true"
    "openmcf.org/organization"  = var.metadata.org
    "openmcf.org/environment"   = var.metadata.env
    "openmcf.org/resource-kind" = "AwsAthenaWorkgroup"
    "openmcf.org/resource-id"   = var.metadata.id
  }

  has_result_config  = var.spec.result_configuration != null
  has_encryption     = local.has_result_config && var.spec.result_configuration.encryption_option != ""
  has_acl            = local.has_result_config && var.spec.result_configuration.s3_acl_option != ""
  has_engine_version = var.spec.selected_engine_version != ""
}
