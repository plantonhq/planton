# Public IP (optional)
resource "azurerm_public_ip" "this" {
  count               = local.enable_public_ip ? 1 : 0
  name                = "${local.name}-pip"
  resource_group_name = local.resource_group
  location            = local.region
  allocation_method   = try(var.spec.network.public_ip_allocation, "static") == "dynamic" ? "Dynamic" : "Static"
  sku                 = try(var.spec.network.public_ip_sku, "standard") == "basic" ? "Basic" : "Standard"
  zones               = var.spec.availability_zone != null ? [var.spec.availability_zone] : null
  tags                = local.tags
}

# Network Interface
resource "azurerm_network_interface" "this" {
  name                          = "${local.name}-nic"
  resource_group_name           = local.resource_group
  location                      = local.region
  enable_accelerated_networking = local.enable_accelerated_networking
  tags                          = local.tags

  ip_configuration {
    name                          = "primary"
    subnet_id                     = var.spec.subnet_id
    private_ip_address_allocation = local.private_ip_allocation
    private_ip_address            = local.private_ip_allocation == "Static" ? local.private_ip_address : null
    public_ip_address_id          = local.enable_public_ip ? azurerm_public_ip.this[0].id : null
    primary                       = true
  }
}

# Network Security Group Association (optional)
resource "azurerm_network_interface_security_group_association" "this" {
  count                     = local.network_security_group_id != null ? 1 : 0
  network_interface_id      = azurerm_network_interface.this.id
  network_security_group_id = local.network_security_group_id
}

# Linux Virtual Machine
resource "azurerm_linux_virtual_machine" "this" {
  count                 = local.use_ssh_key ? 1 : 0
  name                  = local.name
  resource_group_name   = local.resource_group
  location              = local.region
  size                  = local.vm_size
  admin_username        = local.admin_username
  zone                  = var.spec.availability_zone
  priority              = var.spec.is_spot_instance ? "Spot" : "Regular"
  eviction_policy       = var.spec.is_spot_instance ? "Deallocate" : null
  max_bid_price         = var.spec.is_spot_instance && var.spec.spot_max_price > 0 ? var.spec.spot_max_price : null
  custom_data           = var.spec.custom_data != null ? base64encode(var.spec.custom_data) : null
  tags                  = local.tags

  network_interface_ids = [azurerm_network_interface.this.id]

  admin_ssh_key {
    username   = local.admin_username
    public_key = var.spec.ssh_public_key
  }

  os_disk {
    name                 = "${local.name}-osdisk"
    caching              = local.os_disk_caching
    storage_account_type = local.os_disk_storage_type
    disk_size_gb         = try(var.spec.os_disk.size_gb, null)
  }

  dynamic "source_image_reference" {
    for_each = var.spec.image.custom_image_id == null ? [1] : []
    content {
      publisher = var.spec.image.publisher
      offer     = var.spec.image.offer
      sku       = var.spec.image.sku
      version   = var.spec.image.version
    }
  }

  source_image_id = var.spec.image.custom_image_id

  dynamic "identity" {
    for_each = local.enable_identity ? [1] : []
    content {
      type         = local.identity_type
      identity_ids = length(var.spec.user_assigned_identity_ids) > 0 ? var.spec.user_assigned_identity_ids : null
    }
  }

  boot_diagnostics {
    storage_account_uri = var.spec.enable_boot_diagnostics ? null : null # Uses managed storage when null
  }
}

# Windows Virtual Machine (when password is used and no SSH key)
resource "azurerm_windows_virtual_machine" "this" {
  count                 = local.use_ssh_key ? 0 : 1
  name                  = local.name
  resource_group_name   = local.resource_group
  location              = local.region
  size                  = local.vm_size
  admin_username        = local.admin_username
  admin_password        = var.spec.admin_password
  zone                  = var.spec.availability_zone
  priority              = var.spec.is_spot_instance ? "Spot" : "Regular"
  eviction_policy       = var.spec.is_spot_instance ? "Deallocate" : null
  max_bid_price         = var.spec.is_spot_instance && var.spec.spot_max_price > 0 ? var.spec.spot_max_price : null
  custom_data           = var.spec.custom_data != null ? base64encode(var.spec.custom_data) : null
  tags                  = local.tags

  network_interface_ids = [azurerm_network_interface.this.id]

  os_disk {
    name                 = "${local.name}-osdisk"
    caching              = local.os_disk_caching
    storage_account_type = local.os_disk_storage_type
    disk_size_gb         = try(var.spec.os_disk.size_gb, null)
  }

  dynamic "source_image_reference" {
    for_each = var.spec.image.custom_image_id == null ? [1] : []
    content {
      publisher = var.spec.image.publisher
      offer     = var.spec.image.offer
      sku       = var.spec.image.sku
      version   = var.spec.image.version
    }
  }

  source_image_id = var.spec.image.custom_image_id

  dynamic "identity" {
    for_each = local.enable_identity ? [1] : []
    content {
      type         = local.identity_type
      identity_ids = length(var.spec.user_assigned_identity_ids) > 0 ? var.spec.user_assigned_identity_ids : null
    }
  }

  boot_diagnostics {
    storage_account_uri = var.spec.enable_boot_diagnostics ? null : null
  }
}

# Data Disks
resource "azurerm_managed_disk" "data_disks" {
  for_each             = { for idx, disk in var.spec.data_disks : disk.name => disk }
  name                 = each.value.name
  resource_group_name  = local.resource_group
  location             = local.region
  storage_account_type = lookup(local.storage_type_map, each.value.storage_type, "Premium_LRS")
  disk_size_gb         = each.value.size_gb
  create_option        = "Empty"
  zone                 = var.spec.availability_zone
  tags                 = local.tags
}

resource "azurerm_virtual_machine_data_disk_attachment" "data_disks" {
  for_each           = { for idx, disk in var.spec.data_disks : disk.name => disk }
  managed_disk_id    = azurerm_managed_disk.data_disks[each.key].id
  virtual_machine_id = local.use_ssh_key ? azurerm_linux_virtual_machine.this[0].id : azurerm_windows_virtual_machine.this[0].id
  lun                = each.value.lun
  caching            = lookup(local.disk_caching_map, each.value.caching, "ReadOnly")
}
