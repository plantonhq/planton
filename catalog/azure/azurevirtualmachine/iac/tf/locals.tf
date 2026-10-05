locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurevirtualmachine"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  id_tag = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "resource_id" = var.metadata.id } : {}

  # Metadata-derived tags first, then the user's spec tags merged over them:
  # user tags deliberately win so an org's governance conventions (cost
  # center, owner) can override the derived values where they collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # The explicit OS discriminator -- which of the two ARM surfaces this
  # VM deploys through.
  is_linux = var.spec.os_profile.linux != null

  linux   = var.spec.os_profile.linux
  windows = var.spec.os_profile.windows

  # Map the spec enums' name strings to ARM values. Only explicit
  # choices are ever sent (null otherwise), so an unspecified spec and
  # Azure's defaults deploy identically on both engines.
  caching_map = {
    "NONE"       = "None"
    "READ_ONLY"  = "ReadOnly"
    "READ_WRITE" = "ReadWrite"
  }

  os_disk_storage_map = {
    "STANDARD_LRS"     = "Standard_LRS"
    "STANDARD_SSD_LRS" = "StandardSSD_LRS"
    "PREMIUM_LRS"      = "Premium_LRS"
    "STANDARD_SSD_ZRS" = "StandardSSD_ZRS"
    "PREMIUM_ZRS"      = "Premium_ZRS"
  }

  diff_disk_placement = (
    var.spec.os_disk.diff_disk_settings == null ? null :
    var.spec.os_disk.diff_disk_settings.placement == "CACHE_DISK" ? "CacheDisk" :
    var.spec.os_disk.diff_disk_settings.placement == "RESOURCE_DISK" ? "ResourceDisk" :
    var.spec.os_disk.diff_disk_settings.placement == "NVME_DISK" ? "NvmeDisk" : null
  )

  security_encryption_type = (
    var.spec.os_disk.security_encryption_type == "VM_GUEST_STATE_ONLY" ? "VMGuestStateOnly" :
    var.spec.os_disk.security_encryption_type == "DISK_WITH_VM_GUEST_STATE" ? "DiskWithVMGuestState" : null
  )

  identity_type = (
    var.spec.identity == null ? null :
    var.spec.identity.type == "SYSTEM_ASSIGNED" ? "SystemAssigned" :
    var.spec.identity.type == "USER_ASSIGNED" ? "UserAssigned" :
    var.spec.identity.type == "SYSTEM_AND_USER_ASSIGNED" ? "SystemAssigned, UserAssigned" : null
  )

  # Spot: presence of the spot message makes the VM a spot instance.
  priority        = var.spec.spot != null ? "Spot" : "Regular"
  eviction_policy = var.spec.spot == null ? null : var.spec.spot.eviction_policy == "DELETE" ? "Delete" : "Deallocate"
  max_bid_price   = var.spec.spot == null ? null : var.spec.spot.max_bid_price

  availability = var.spec.availability

  # Linux patch modes: ImageDefault / AutomaticByPlatform. The wire
  # format carries the full proto enum names (LINUX_-prefixed -- the
  # per-OS vocabularies share a spec so their values are disambiguated).
  linux_patch_mode = (
    local.linux == null ? null :
    local.linux.patch_mode == "LINUX_IMAGE_DEFAULT" ? "ImageDefault" :
    local.linux.patch_mode == "LINUX_AUTOMATIC_BY_PLATFORM" ? "AutomaticByPlatform" : null
  )

  # Windows patch modes: Manual / AutomaticByOS / AutomaticByPlatform.
  # Only the platform arm carries the WINDOWS_ disambiguation prefix.
  windows_patch_mode = (
    local.windows == null ? null :
    local.windows.patch_mode == "MANUAL" ? "Manual" :
    local.windows.patch_mode == "AUTOMATIC_BY_OS" ? "AutomaticByOS" :
    local.windows.patch_mode == "WINDOWS_AUTOMATIC_BY_PLATFORM" ? "AutomaticByPlatform" : null
  )

  patch_assessment_mode = (
    var.spec.patching == null ? null :
    var.spec.patching.assessment_mode == "ASSESSMENT_IMAGE_DEFAULT" ? "ImageDefault" :
    var.spec.patching.assessment_mode == "ASSESSMENT_AUTOMATIC_BY_PLATFORM" ? "AutomaticByPlatform" : null
  )

  reboot_setting = (
    var.spec.patching == null ? null :
    var.spec.patching.reboot_setting == "ALWAYS" ? "Always" :
    var.spec.patching.reboot_setting == "IF_REQUIRED" ? "IfRequired" :
    var.spec.patching.reboot_setting == "NEVER" ? "Never" : null
  )

  bypass_safety_checks = var.spec.patching != null ? var.spec.patching.bypass_platform_safety_checks_on_user_schedule_enabled : false

  # Windows license: ARM's literal values (None is an explicit choice).
  windows_license_type = (
    local.windows == null ? null :
    local.windows.license_type == "WINDOWS_LICENSE_NONE" ? "None" :
    local.windows.license_type == "WINDOWS_CLIENT" ? "Windows_Client" :
    local.windows.license_type == "WINDOWS_SERVER" ? "Windows_Server" : null
  )

  # Linux license: the enum names ARE ARM's literal values.
  linux_license_type = (
    local.linux == null || local.linux.license_type == null ? null : local.linux.license_type
  )

  disk_controller_type = (
    var.spec.disk_controller_type == "SCSI" ? "SCSI" :
    var.spec.disk_controller_type == "NVME" ? "NVMe" : null
  )

  unattend_setting_map = {
    "AUTO_LOGON"           = "AutoLogon"
    "FIRST_LOGON_COMMANDS" = "FirstLogonCommands"
  }
}
