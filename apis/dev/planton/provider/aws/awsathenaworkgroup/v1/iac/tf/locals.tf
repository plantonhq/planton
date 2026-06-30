locals {
  workgroup_name = var.metadata.name

  tags = {
    "planton.dev/resource"      = "true"
    "planton.dev/organization"  = var.metadata.org
    "planton.dev/environment"   = var.metadata.env
    "planton.dev/resource-kind" = "AwsAthenaWorkgroup"
    "planton.dev/resource-id"   = var.metadata.id
  }

  has_result_config  = var.spec.result_configuration != null
  has_encryption     = local.has_result_config && var.spec.result_configuration.encryption_option != ""
  has_acl            = local.has_result_config && var.spec.result_configuration.s3_acl_option != ""
  has_engine_version = var.spec.selected_engine_version != ""
}
