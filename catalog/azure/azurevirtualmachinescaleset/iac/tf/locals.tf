locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurevirtualmachinescaleset"
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

  # Metadata-derived tags first, then the user's spec tags merged over
  # them: user tags deliberately win so an org's governance conventions
  # (cost center, owner) can override the derived values where they
  # collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # The dispatch axes: ONE proto surface realizes onto azurerm's three
  # scale-set resources -- linux/windows (Uniform) and orchestrated
  # (Flexible). Unset orchestration_mode applies FLEXIBLE (Azure's
  # recommendation for new workloads).
  is_uniform = var.spec.orchestration_mode == "UNIFORM"
  is_linux   = var.spec.os_profile.linux != null

  linux_profile   = var.spec.os_profile.linux
  windows_profile = var.spec.os_profile.windows

  # Enum values arrive as the FULL proto value name -- every map below
  # matches them verbatim (the DATA_ / PROBE_ style prefixes exist only
  # to keep proto enum value names unique within the package).
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
  data_disk_storage_map = {
    "DATA_STANDARD_LRS"     = "Standard_LRS"
    "DATA_STANDARD_SSD_LRS" = "StandardSSD_LRS"
    "DATA_PREMIUM_LRS"      = "Premium_LRS"
    "DATA_PREMIUM_ZRS"      = "Premium_ZRS"
    "ULTRA_SSD_LRS"         = "UltraSSD_LRS"
    "PREMIUM_V2_LRS"        = "PremiumV2_LRS"
    "DATA_STANDARD_SSD_ZRS" = "StandardSSD_ZRS"
  }
  create_option_map = {
    "EMPTY"      = "Empty"
    "FROM_IMAGE" = "FromImage"
  }
  diff_disk_placement_map = {
    "CACHE_DISK"    = "CacheDisk"
    "RESOURCE_DISK" = "ResourceDisk"
    "NVME_DISK"     = "NvmeDisk"
  }
  security_encryption_map = {
    "VM_GUEST_STATE_ONLY"      = "VMGuestStateOnly"
    "DISK_WITH_VM_GUEST_STATE" = "DiskWithVMGuestState"
  }
  ip_version_map = {
    "IPV4" = "IPv4"
    "IPV6" = "IPv6"
  }
  upgrade_mode_map = {
    "MANUAL"    = "Manual"
    "AUTOMATIC" = "Automatic"
    "ROLLING"   = "Rolling"
  }
  eviction_policy_map = {
    "DEALLOCATE" = "Deallocate"
    "DELETE"     = "Delete"
  }
  repair_action_map = {
    "REPLACE" = "Replace"
    "RESTART" = "Restart"
    "REIMAGE" = "Reimage"
  }
  scale_in_rule_map = {
    "DEFAULT"   = "Default"
    "NEWEST_VM" = "NewestVM"
    "OLDEST_VM" = "OldestVM"
  }
  identity_type_map = {
    "SYSTEM_ASSIGNED"          = "SystemAssigned"
    "USER_ASSIGNED"            = "UserAssigned"
    "SYSTEM_AND_USER_ASSIGNED" = "SystemAssigned, UserAssigned"
  }
  linux_patch_mode_map = {
    "LINUX_IMAGE_DEFAULT"         = "ImageDefault"
    "LINUX_AUTOMATIC_BY_PLATFORM" = "AutomaticByPlatform"
  }
  windows_patch_mode_map = {
    "WINDOWS_MANUAL"                = "Manual"
    "AUTOMATIC_BY_OS"               = "AutomaticByOS"
    "WINDOWS_AUTOMATIC_BY_PLATFORM" = "AutomaticByPlatform"
  }
  assessment_mode_map = {
    "ASSESSMENT_IMAGE_DEFAULT"         = "ImageDefault"
    "ASSESSMENT_AUTOMATIC_BY_PLATFORM" = "AutomaticByPlatform"
  }
  windows_license_map = {
    "WINDOWS_LICENSE_NONE" = "None"
    "WINDOWS_CLIENT"       = "Windows_Client"
    "WINDOWS_SERVER"       = "Windows_Server"
  }
  winrm_protocol_map = {
    "HTTP"  = "Http"
    "HTTPS" = "Https"
  }
  unattend_setting_map = {
    "AUTO_LOGON"           = "AutoLogon"
    "FIRST_LOGON_COMMANDS" = "FirstLogonCommands"
  }
  allocation_strategy_map = {
    "LOWEST_PRICE"       = "LowestPrice"
    "CAPACITY_OPTIMIZED" = "CapacityOptimized"
    "PRIORITIZED"        = "Prioritized"
  }
  auxiliary_mode_map = {
    "ACCELERATED_CONNECTIONS" = "AcceleratedConnections"
    "FLOATING"                = "Floating"
  }
  auxiliary_sku_map = {
    "A1" = "A1"
    "A2" = "A2"
    "A4" = "A4"
    "A8" = "A8"
  }

  # Spot presence is the priority switch: present means "Spot" (with an
  # explicit eviction policy), absent means a regular on-demand fleet.
  priority        = var.spec.spot != null ? "Spot" : "Regular"
  eviction_policy = var.spec.spot != null ? local.eviction_policy_map[var.spec.spot.eviction_policy] : null
  max_bid_price   = var.spec.spot != null && try(var.spec.spot.max_bid_price, null) != null ? var.spec.spot.max_bid_price : null

  upgrade_mode = var.spec.upgrade_policy != null && try(var.spec.upgrade_policy.mode, null) != null ? local.upgrade_mode_map[var.spec.upgrade_policy.mode] : "Manual"

  # Instance computer names derive from this prefix (Azure appends a
  # unique suffix); unset defaults to the scale-set name.
  computer_name_prefix = (
    var.spec.os_profile.computer_name_prefix != null && var.spec.os_profile.computer_name_prefix != ""
    ? var.spec.os_profile.computer_name_prefix
    : null
  )
}
