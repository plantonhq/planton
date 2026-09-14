# Create the virtual machine -- the compute shell wired to first-class
# referenced resources.
#
# The VM is deliberately just the machine (matching Azure's own model):
# - Network presence comes from referenced network interfaces
#   (network_interface_ids); public IPs, NSG filtering, and subnet
#   placement live NIC-side.
# - Data volumes are referenced managed disks realized as attachment
#   resources below -- the data outlives the machine.
# - Only the OS disk is inline: it is born and dies with the VM, unless
#   the VM boots from an existing referenced OS disk (os_managed_disk_id).
#
# ARM models Linux and Windows VMs as separate management surfaces
# (different auth contracts, patch vocabularies, and OS settings), so the
# module deploys exactly one of the two resources below from the spec's
# explicit OS discriminator (os_profile.linux XOR os_profile.windows).
#
# Lifecycle notes worth knowing before operating this resource:
# - Name, region, zone, image source, admin credentials, custom_data,
#   and the security/confidential posture are the VM's identity --
#   changing any of them replaces the VM (the OS disk with it; data
#   disks and NICs survive, which is exactly why they are referenced).
# - Resizing (size) reboots in place. Spot settings are fixed at
#   creation.
resource "azurerm_linux_virtual_machine" "main" {
  count = local.is_linux ? 1 : 0

  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  size                = var.spec.size

  network_interface_ids = var.spec.network_interface_ids

  computer_name = var.spec.os_profile.computer_name

  # Authentication: SSH-first. Absent entirely when booting from an
  # existing OS disk (the disk already contains its users; spec-level
  # validation enforces the pairing).
  admin_username                  = local.linux.admin_username != "" ? local.linux.admin_username : null
  admin_password                  = local.linux.admin_password
  disable_password_authentication = local.linux.disable_password_authentication

  dynamic "admin_ssh_key" {
    for_each = local.linux.ssh_public_keys
    content {
      public_key = admin_ssh_key.value.public_key
      # An unset key username defaults to the admin account -- the
      # common case.
      username = coalesce(admin_ssh_key.value.username, local.linux.admin_username)
    }
  }

  os_disk {
    caching              = local.caching_map[var.spec.os_disk.caching]
    storage_account_type = local.os_disk_storage_map[var.spec.os_disk.storage_account_type]
    disk_size_gb         = var.spec.os_disk.disk_size_gb
    name                 = var.spec.os_disk.name

    # Ephemeral OS disk: lives on local VM storage, wiped on every
    # stop/deallocate -- stateless fleets only.
    dynamic "diff_disk_settings" {
      for_each = var.spec.os_disk.diff_disk_settings != null ? [1] : []
      content {
        option    = "Local"
        placement = local.diff_disk_placement
      }
    }

    disk_encryption_set_id           = var.spec.os_disk.disk_encryption_set_id
    secure_vm_disk_encryption_set_id = var.spec.os_disk.secure_vm_disk_encryption_set_id
    security_encryption_type         = local.security_encryption_type
    write_accelerator_enabled        = var.spec.os_disk.write_accelerator_enabled
  }

  # Exactly one image source (spec-level validation): marketplace
  # coordinates, a custom/gallery image id, or an existing OS disk.
  dynamic "source_image_reference" {
    for_each = var.spec.source_image_reference != null ? [var.spec.source_image_reference] : []
    content {
      publisher = source_image_reference.value.publisher
      offer     = source_image_reference.value.offer
      sku       = source_image_reference.value.sku
      version   = source_image_reference.value.version
    }
  }
  source_image_id    = var.spec.source_image_id
  os_managed_disk_id = var.spec.os_managed_disk_id

  dynamic "identity" {
    for_each = var.spec.identity != null ? [var.spec.identity] : []
    content {
      type         = local.identity_type
      identity_ids = identity.value.identity_ids
    }
  }

  # Spot: presence of the spec's spot message makes the VM evictable,
  # deeply discounted capacity.
  priority        = local.priority
  eviction_policy = local.eviction_policy
  max_bid_price   = local.max_bid_price

  # Placement: zone XOR availability set (spec-level validation), plus
  # the specialized placement seams.
  zone                          = local.availability != null ? local.availability.zone : null
  availability_set_id           = local.availability != null ? local.availability.availability_set_id : null
  proximity_placement_group_id  = local.availability != null ? local.availability.proximity_placement_group_id : null
  capacity_reservation_group_id = local.availability != null ? local.availability.capacity_reservation_group_id : null
  dedicated_host_id             = local.availability != null ? local.availability.dedicated_host_id : null
  dedicated_host_group_id       = local.availability != null ? local.availability.dedicated_host_group_id : null
  virtual_machine_scale_set_id  = local.availability != null ? local.availability.virtual_machine_scale_set_id : null
  platform_fault_domain         = local.availability != null ? local.availability.platform_fault_domain : null

  # Trusted launch / encryption posture (fixed at creation).
  secure_boot_enabled        = var.spec.security != null ? var.spec.security.secure_boot_enabled : null
  vtpm_enabled               = var.spec.security != null ? var.spec.security.vtpm_enabled : null
  encryption_at_host_enabled = var.spec.security != null ? var.spec.security.encryption_at_host_enabled : null

  # Patch orchestration: the MODE vocabulary is Linux-specific; the
  # shared dials come from spec.patching.
  patch_mode                                             = local.linux_patch_mode
  patch_assessment_mode                                  = local.patch_assessment_mode
  reboot_setting                                         = local.reboot_setting
  bypass_platform_safety_checks_on_user_schedule_enabled = local.bypass_safety_checks

  license_type = local.linux_license_type

  # Presence enables boot diagnostics; an empty URI uses Azure's managed
  # storage (the right default).
  dynamic "boot_diagnostics" {
    for_each = var.spec.boot_diagnostics != null && coalesce(var.spec.boot_diagnostics.enabled, true) ? [var.spec.boot_diagnostics] : []
    content {
      storage_account_uri = boot_diagnostics.value.storage_account_uri
    }
  }

  dynamic "gallery_application" {
    for_each = var.spec.gallery_applications
    content {
      version_id                                  = gallery_application.value.version_id
      order                                       = gallery_application.value.order
      tag                                         = gallery_application.value.tag
      configuration_blob_uri                      = gallery_application.value.configuration_blob_uri
      automatic_upgrade_enabled                   = gallery_application.value.automatic_upgrade_enabled
      treat_failure_as_deployment_failure_enabled = gallery_application.value.treat_failure_as_deployment_failure_enabled
    }
  }

  dynamic "termination_notification" {
    for_each = var.spec.termination_notification != null ? [var.spec.termination_notification] : []
    content {
      enabled = true
      timeout = termination_notification.value.timeout
    }
  }

  dynamic "os_image_notification" {
    for_each = var.spec.os_image_notification != null ? [var.spec.os_image_notification] : []
    content {
      timeout = os_image_notification.value.timeout
    }
  }

  dynamic "plan" {
    for_each = var.spec.plan != null ? [var.spec.plan] : []
    content {
      name      = plan.value.name
      product   = plan.value.product
      publisher = plan.value.publisher
    }
  }

  # custom_data is delivered once at first boot (may embed bootstrap
  # secrets); user_data is IMDS-readable and updatable -- never secret.
  custom_data = var.spec.custom_data
  user_data   = var.spec.user_data

  extensions_time_budget     = var.spec.extensions_time_budget
  provision_vm_agent         = var.spec.provision_vm_agent
  allow_extension_operations = var.spec.allow_extension_operations

  disk_controller_type = local.disk_controller_type

  dynamic "additional_capabilities" {
    for_each = var.spec.additional_capabilities != null ? [var.spec.additional_capabilities] : []
    content {
      ultra_ssd_enabled   = additional_capabilities.value.ultra_ssd_enabled
      hibernation_enabled = additional_capabilities.value.hibernation_enabled
    }
  }

  dynamic "secret" {
    for_each = var.spec.secrets
    content {
      key_vault_id = secret.value.key_vault_id
      dynamic "certificate" {
        for_each = secret.value.certificates
        content {
          url = certificate.value.url
        }
      }
    }
  }

  edge_zone = var.spec.edge_zone

  tags = local.final_tags
}

resource "azurerm_windows_virtual_machine" "main" {
  count = local.is_linux ? 0 : 1

  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  size                = var.spec.size

  network_interface_ids = var.spec.network_interface_ids

  computer_name = var.spec.os_profile.computer_name

  # Authentication: username + password (Windows has no SSH-key
  # concept). Absent entirely when booting from an existing OS disk.
  admin_username = local.windows.admin_username != "" ? local.windows.admin_username : null
  admin_password = local.windows.admin_password

  os_disk {
    caching              = local.caching_map[var.spec.os_disk.caching]
    storage_account_type = local.os_disk_storage_map[var.spec.os_disk.storage_account_type]
    disk_size_gb         = var.spec.os_disk.disk_size_gb
    name                 = var.spec.os_disk.name

    dynamic "diff_disk_settings" {
      for_each = var.spec.os_disk.diff_disk_settings != null ? [1] : []
      content {
        option    = "Local"
        placement = local.diff_disk_placement
      }
    }

    disk_encryption_set_id           = var.spec.os_disk.disk_encryption_set_id
    secure_vm_disk_encryption_set_id = var.spec.os_disk.secure_vm_disk_encryption_set_id
    security_encryption_type         = local.security_encryption_type
    write_accelerator_enabled        = var.spec.os_disk.write_accelerator_enabled
  }

  dynamic "source_image_reference" {
    for_each = var.spec.source_image_reference != null ? [var.spec.source_image_reference] : []
    content {
      publisher = source_image_reference.value.publisher
      offer     = source_image_reference.value.offer
      sku       = source_image_reference.value.sku
      version   = source_image_reference.value.version
    }
  }
  source_image_id    = var.spec.source_image_id
  os_managed_disk_id = var.spec.os_managed_disk_id

  dynamic "identity" {
    for_each = var.spec.identity != null ? [var.spec.identity] : []
    content {
      type         = local.identity_type
      identity_ids = identity.value.identity_ids
    }
  }

  priority        = local.priority
  eviction_policy = local.eviction_policy
  max_bid_price   = local.max_bid_price

  zone                          = local.availability != null ? local.availability.zone : null
  availability_set_id           = local.availability != null ? local.availability.availability_set_id : null
  proximity_placement_group_id  = local.availability != null ? local.availability.proximity_placement_group_id : null
  capacity_reservation_group_id = local.availability != null ? local.availability.capacity_reservation_group_id : null
  dedicated_host_id             = local.availability != null ? local.availability.dedicated_host_id : null
  dedicated_host_group_id       = local.availability != null ? local.availability.dedicated_host_group_id : null
  virtual_machine_scale_set_id  = local.availability != null ? local.availability.virtual_machine_scale_set_id : null
  platform_fault_domain         = local.availability != null ? local.availability.platform_fault_domain : null

  secure_boot_enabled        = var.spec.security != null ? var.spec.security.secure_boot_enabled : null
  vtpm_enabled               = var.spec.security != null ? var.spec.security.vtpm_enabled : null
  encryption_at_host_enabled = var.spec.security != null ? var.spec.security.encryption_at_host_enabled : null

  # Patch orchestration: the Windows MODE vocabulary, plus the
  # Windows-only knobs (automatic updates, hotpatching, timezone).
  patch_mode                                             = local.windows_patch_mode
  patch_assessment_mode                                  = local.patch_assessment_mode
  reboot_setting                                         = local.reboot_setting
  bypass_platform_safety_checks_on_user_schedule_enabled = local.bypass_safety_checks

  automatic_updates_enabled = local.windows.automatic_updates_enabled
  hotpatching_enabled       = local.windows.hotpatching_enabled
  timezone                  = local.windows.timezone

  dynamic "winrm_listener" {
    for_each = local.windows.winrm_listeners
    content {
      protocol        = winrm_listener.value.protocol == "HTTPS" ? "Https" : "Http"
      certificate_url = winrm_listener.value.certificate_url
    }
  }

  dynamic "additional_unattend_content" {
    for_each = local.windows.additional_unattend_contents
    content {
      setting = local.unattend_setting_map[additional_unattend_content.value.setting]
      content = additional_unattend_content.value.content
    }
  }

  license_type = local.windows_license_type

  dynamic "boot_diagnostics" {
    for_each = var.spec.boot_diagnostics != null && coalesce(var.spec.boot_diagnostics.enabled, true) ? [var.spec.boot_diagnostics] : []
    content {
      storage_account_uri = boot_diagnostics.value.storage_account_uri
    }
  }

  dynamic "gallery_application" {
    for_each = var.spec.gallery_applications
    content {
      version_id                                  = gallery_application.value.version_id
      order                                       = gallery_application.value.order
      tag                                         = gallery_application.value.tag
      configuration_blob_uri                      = gallery_application.value.configuration_blob_uri
      automatic_upgrade_enabled                   = gallery_application.value.automatic_upgrade_enabled
      treat_failure_as_deployment_failure_enabled = gallery_application.value.treat_failure_as_deployment_failure_enabled
    }
  }

  dynamic "termination_notification" {
    for_each = var.spec.termination_notification != null ? [var.spec.termination_notification] : []
    content {
      enabled = true
      timeout = termination_notification.value.timeout
    }
  }

  dynamic "os_image_notification" {
    for_each = var.spec.os_image_notification != null ? [var.spec.os_image_notification] : []
    content {
      timeout = os_image_notification.value.timeout
    }
  }

  dynamic "plan" {
    for_each = var.spec.plan != null ? [var.spec.plan] : []
    content {
      name      = plan.value.name
      product   = plan.value.product
      publisher = plan.value.publisher
    }
  }

  custom_data = var.spec.custom_data
  user_data   = var.spec.user_data

  extensions_time_budget     = var.spec.extensions_time_budget
  provision_vm_agent         = var.spec.provision_vm_agent
  allow_extension_operations = var.spec.allow_extension_operations

  disk_controller_type = local.disk_controller_type

  dynamic "additional_capabilities" {
    for_each = var.spec.additional_capabilities != null ? [var.spec.additional_capabilities] : []
    content {
      ultra_ssd_enabled   = additional_capabilities.value.ultra_ssd_enabled
      hibernation_enabled = additional_capabilities.value.hibernation_enabled
    }
  }

  dynamic "secret" {
    for_each = var.spec.secrets
    content {
      key_vault_id = secret.value.key_vault_id
      dynamic "certificate" {
        for_each = secret.value.certificates
        content {
          url = certificate.value.url
          # Windows installs into a named certificate store.
          store = certificate.value.store
        }
      }
    }
  }

  edge_zone = var.spec.edge_zone

  tags = local.final_tags
}

# Attach the referenced first-class data disks. Each attachment is its
# own ARM operation (Azure's model): the disk -- and its data -- outlives
# the VM, and detaching is just removing the spec entry.
resource "azurerm_virtual_machine_data_disk_attachment" "main" {
  for_each = { for attachment in var.spec.data_disk_attachments : attachment.lun => attachment }

  managed_disk_id           = each.value.managed_disk_id
  virtual_machine_id        = local.is_linux ? azurerm_linux_virtual_machine.main[0].id : azurerm_windows_virtual_machine.main[0].id
  lun                       = each.value.lun
  caching                   = local.caching_map[each.value.caching]
  write_accelerator_enabled = each.value.write_accelerator_enabled
}
