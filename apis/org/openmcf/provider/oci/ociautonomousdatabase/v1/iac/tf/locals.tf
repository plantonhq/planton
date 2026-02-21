locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = coalesce(var.spec.display_name, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciAutonomousDatabase"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  nsg_ids = [for nsg in var.spec.nsg_ids : nsg.value]

  db_workload_map = {
    "oltp" = "OLTP"
    "dw"   = "DW"
    "ajd"  = "AJD"
    "apex" = "APEX"
    "lh"   = "LH"
  }

  compute_model_map = {
    "ecpu" = "ECPU"
    "ocpu" = "OCPU"
  }

  database_edition_map = {
    "standard_edition"   = "STANDARD_EDITION"
    "enterprise_edition" = "ENTERPRISE_EDITION"
  }

  license_model_map = {
    "bring_your_own_license" = "BRING_YOUR_OWN_LICENSE"
    "license_included"       = "LICENSE_INCLUDED"
  }

  maintenance_schedule_type_map = {
    "early"   = "EARLY"
    "regular" = "REGULAR"
  }
}
