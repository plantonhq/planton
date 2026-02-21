locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = coalesce(var.spec.display_name, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciComputeInstance"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  source_type_map = {
    "image"       = "image"
    "boot_volume" = "bootVolume"
  }

  nsg_ids = [for nsg in var.spec.create_vnic_details.nsg_ids : nsg.value]

  firmware_map = {
    "bios"    = "BIOS"
    "uefi_64" = "UEFI_64"
  }

  recovery_action_map = {
    "restore_instance" = "RESTORE_INSTANCE"
    "stop_instance"    = "STOP_INSTANCE"
  }

  platform_type_map = {
    "amd_milan_bm"     = "AMD_MILAN_BM"
    "amd_milan_bm_gpu" = "AMD_MILAN_BM_GPU"
    "amd_rome_bm"      = "AMD_ROME_BM"
    "amd_rome_bm_gpu"  = "AMD_ROME_BM_GPU"
    "amd_vm"           = "AMD_VM"
    "generic_bm"       = "GENERIC_BM"
    "intel_icelake_bm" = "INTEL_ICELAKE_BM"
    "intel_skylake_bm" = "INTEL_SKYLAKE_BM"
    "intel_vm"         = "INTEL_VM"
  }
}
