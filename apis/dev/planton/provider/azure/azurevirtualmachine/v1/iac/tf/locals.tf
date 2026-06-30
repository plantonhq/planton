locals {
  name           = var.metadata.name
  resource_group = var.spec.resource_group
  region         = var.spec.region
  vm_size        = var.spec.vm_size
  admin_username = var.spec.admin_username

  # Storage type mapping
  storage_type_map = {
    "standard_lrs"     = "Standard_LRS"
    "standard_ssd_lrs" = "StandardSSD_LRS"
    "premium_lrs"      = "Premium_LRS"
    "premium_zrs"      = "Premium_ZRS"
  }

  # Disk caching mapping
  disk_caching_map = {
    "none"       = "None"
    "read_only"  = "ReadOnly"
    "read_write" = "ReadWrite"
  }

  # OS disk configuration
  os_disk_storage_type = lookup(local.storage_type_map, try(var.spec.os_disk.storage_type, "premium_lrs"), "Premium_LRS")
  os_disk_caching      = lookup(local.disk_caching_map, try(var.spec.os_disk.caching, "read_write"), "ReadWrite")

  # Network configuration
  enable_public_ip              = try(var.spec.network.enable_public_ip, false)
  enable_accelerated_networking = try(var.spec.network.enable_accelerated_networking, true)
  private_ip_allocation         = try(var.spec.network.private_ip_allocation, "private_dynamic") == "private_static" ? "Static" : "Dynamic"
  private_ip_address            = try(var.spec.network.private_ip_address, null)
  network_security_group_id     = try(var.spec.network.network_security_group_id, null)

  # Authentication
  use_ssh_key = var.spec.ssh_public_key != null && var.spec.ssh_public_key != ""

  # Identity
  enable_identity = var.spec.enable_system_assigned_identity || length(var.spec.user_assigned_identity_ids) > 0

  identity_type = (
    var.spec.enable_system_assigned_identity && length(var.spec.user_assigned_identity_ids) > 0 ? "SystemAssigned, UserAssigned" :
    var.spec.enable_system_assigned_identity ? "SystemAssigned" :
    length(var.spec.user_assigned_identity_ids) > 0 ? "UserAssigned" :
    null
  )

  # Tags
  tags = merge(var.spec.tags, {
    managed_by = "planton"
  })
}
