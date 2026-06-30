locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = coalesce(var.spec.display_name, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciDbSystem"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  nsg_ids             = [for nsg in var.spec.nsg_ids : nsg.value]
  backup_network_nsg_ids = [for nsg in var.spec.backup_network_nsg_ids : nsg.value]

  database_edition_map = {
    "standard_edition"                       = "STANDARD_EDITION"
    "enterprise_edition"                     = "ENTERPRISE_EDITION"
    "enterprise_edition_high_performance"    = "ENTERPRISE_EDITION_HIGH_PERFORMANCE"
    "enterprise_edition_extreme_performance" = "ENTERPRISE_EDITION_EXTREME_PERFORMANCE"
  }

  license_model_map = {
    "bring_your_own_license" = "BRING_YOUR_OWN_LICENSE"
    "license_included"       = "LICENSE_INCLUDED"
  }

  disk_redundancy_map = {
    "normal" = "NORMAL"
    "high"   = "HIGH"
  }

  storage_volume_performance_mode_map = {
    "balanced"         = "BALANCED"
    "high_performance" = "HIGH_PERFORMANCE"
  }

  storage_management_map = {
    "asm" = "ASM"
    "lvm" = "LVM"
  }

  preference_map = {
    "no_preference"     = "NO_PREFERENCE"
    "custom_preference" = "CUSTOM_PREFERENCE"
  }

  patching_mode_map = {
    "rolling"    = "ROLLING"
    "nonrolling" = "NONROLLING"
  }
}
