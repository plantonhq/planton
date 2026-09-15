# ONE spec surface realizes onto azurerm's three scale-set resources:
# linux/windows (UNIFORM orchestration) and orchestrated (FLEXIBLE).
# ARM has a single scale-set resource type with an orchestration-mode
# property; the three-resource split is the provider's ergonomics, so
# the dispatch lives here rather than in the user's model.
#
# Lifecycle notes worth knowing before operating this resource:
# - Orchestration mode, zones (removal), fault domains, placement, and
#   the plan are fixed at creation -- changing them replaces the fleet.
# - How template changes reach EXISTING instances is governed by
#   upgrade_mode: Manual leaves them on the old model until upgraded,
#   Automatic applies immediately, Rolling batches with health checks.
# - single_placement_group can go true->false but never back.

# ---------------------------------------------------------------------
# UNIFORM + Linux
# ---------------------------------------------------------------------
resource "azurerm_linux_virtual_machine_scale_set" "main" {
  count = local.is_uniform && local.is_linux ? 1 : 0

  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  sku                 = var.spec.sku_name
  instances           = var.spec.instances
  tags                = local.final_tags

  admin_username                  = local.linux_profile.admin_username
  admin_password                  = local.linux_profile.admin_password
  disable_password_authentication = local.linux_profile.disable_password_authentication
  computer_name_prefix            = local.computer_name_prefix

  dynamic "admin_ssh_key" {
    for_each = local.linux_profile.ssh_public_keys
    content {
      public_key = admin_ssh_key.value.public_key
      # Unset key usernames default to the admin account -- the common
      # case.
      username = coalesce(admin_ssh_key.value.username, local.linux_profile.admin_username)
    }
  }

  custom_data = var.spec.custom_data != null && var.spec.custom_data != "" ? var.spec.custom_data : null
  user_data   = var.spec.user_data != null && var.spec.user_data != "" ? var.spec.user_data : null

  source_image_id = var.spec.source_image_id
  dynamic "source_image_reference" {
    for_each = var.spec.source_image_reference != null ? [var.spec.source_image_reference] : []
    content {
      publisher = source_image_reference.value.publisher
      offer     = source_image_reference.value.offer
      sku       = source_image_reference.value.sku
      version   = source_image_reference.value.version
    }
  }

  os_disk {
    caching                          = local.caching_map[var.spec.os_disk.caching]
    storage_account_type             = local.os_disk_storage_map[var.spec.os_disk.storage_account_type]
    disk_size_gb                     = var.spec.os_disk.disk_size_gb
    write_accelerator_enabled        = var.spec.os_disk.write_accelerator_enabled
    disk_encryption_set_id           = var.spec.os_disk.disk_encryption_set_id
    secure_vm_disk_encryption_set_id = var.spec.os_disk.secure_vm_disk_encryption_set_id
    security_encryption_type         = var.spec.os_disk.security_encryption_type != null ? local.security_encryption_map[var.spec.os_disk.security_encryption_type] : null

    dynamic "diff_disk_settings" {
      for_each = var.spec.os_disk.diff_disk_settings != null ? [var.spec.os_disk.diff_disk_settings] : []
      content {
        option    = "Local"
        placement = diff_disk_settings.value.placement != null ? local.diff_disk_placement_map[diff_disk_settings.value.placement] : null
      }
    }
  }

  dynamic "data_disk" {
    for_each = var.spec.data_disks
    content {
      lun                       = data_disk.value.lun
      caching                   = local.caching_map[data_disk.value.caching]
      disk_size_gb              = data_disk.value.disk_size_gb
      storage_account_type      = local.data_disk_storage_map[data_disk.value.storage_account_type]
      create_option             = data_disk.value.create_option != null ? local.create_option_map[data_disk.value.create_option] : "Empty"
      write_accelerator_enabled = data_disk.value.write_accelerator_enabled
      disk_encryption_set_id    = data_disk.value.disk_encryption_set_id
      disk_iops_read_write      = data_disk.value.ultra_ssd_disk_iops_read_write
      disk_mbps_read_write      = data_disk.value.ultra_ssd_disk_mbps_read_write
    }
  }

  dynamic "network_interface" {
    for_each = var.spec.network_interfaces
    content {
      name                           = network_interface.value.name
      primary                        = network_interface.value.primary
      dns_servers                    = network_interface.value.dns_servers
      accelerated_networking_enabled = network_interface.value.accelerated_networking_enabled
      ip_forwarding_enabled          = network_interface.value.ip_forwarding_enabled
      network_security_group_id      = network_interface.value.network_security_group_id
      auxiliary_mode                 = network_interface.value.auxiliary_mode != null ? local.auxiliary_mode_map[network_interface.value.auxiliary_mode] : null
      auxiliary_sku                  = network_interface.value.auxiliary_sku != null ? local.auxiliary_sku_map[network_interface.value.auxiliary_sku] : null

      dynamic "ip_configuration" {
        for_each = network_interface.value.ip_configurations
        content {
          name      = ip_configuration.value.name
          primary   = ip_configuration.value.primary
          subnet_id = ip_configuration.value.subnet_id
          version   = ip_configuration.value.version != null ? local.ip_version_map[ip_configuration.value.version] : "IPv4"

          load_balancer_backend_address_pool_ids       = ip_configuration.value.load_balancer_backend_address_pool_ids
          load_balancer_inbound_nat_rules_ids          = ip_configuration.value.load_balancer_inbound_nat_rule_ids
          application_gateway_backend_address_pool_ids = ip_configuration.value.application_gateway_backend_address_pool_ids
          application_security_group_ids               = ip_configuration.value.application_security_group_ids

          dynamic "public_ip_address" {
            for_each = ip_configuration.value.public_ip_address != null ? [ip_configuration.value.public_ip_address] : []
            content {
              name                    = public_ip_address.value.name
              domain_name_label       = public_ip_address.value.domain_name_label
              idle_timeout_in_minutes = public_ip_address.value.idle_timeout_in_minutes
              version                 = public_ip_address.value.version != null ? local.ip_version_map[public_ip_address.value.version] : null
              public_ip_prefix_id     = public_ip_address.value.public_ip_prefix_id

              dynamic "ip_tag" {
                for_each = public_ip_address.value.ip_tags
                content {
                  type = ip_tag.value.type
                  tag  = ip_tag.value.tag
                }
              }
            }
          }
        }
      }
    }
  }

  upgrade_mode    = local.upgrade_mode
  health_probe_id = var.spec.upgrade_policy != null ? var.spec.upgrade_policy.health_probe_id : null

  dynamic "rolling_upgrade_policy" {
    for_each = var.spec.upgrade_policy != null && try(var.spec.upgrade_policy.rolling, null) != null ? [var.spec.upgrade_policy.rolling] : []
    content {
      max_batch_instance_percent              = rolling_upgrade_policy.value.max_batch_instance_percent
      max_unhealthy_instance_percent          = rolling_upgrade_policy.value.max_unhealthy_instance_percent
      max_unhealthy_upgraded_instance_percent = rolling_upgrade_policy.value.max_unhealthy_upgraded_instance_percent
      pause_time_between_batches              = rolling_upgrade_policy.value.pause_time_between_batches
      cross_zone_upgrades_enabled             = rolling_upgrade_policy.value.cross_zone_upgrades_enabled
      prioritize_unhealthy_instances_enabled  = rolling_upgrade_policy.value.prioritize_unhealthy_instances_enabled
      maximum_surge_instances_enabled         = rolling_upgrade_policy.value.maximum_surge_instances_enabled
    }
  }

  dynamic "automatic_os_upgrade_policy" {
    for_each = var.spec.upgrade_policy != null && try(var.spec.upgrade_policy.automatic_os_upgrade, null) != null ? [var.spec.upgrade_policy.automatic_os_upgrade] : []
    content {
      # The provider expresses rollback as a positive flag; the spec keeps
      # the ARM-side disable_automatic_rollback name, so the value inverts
      # here (and only here).
      automatic_os_upgrade_enabled = automatic_os_upgrade_policy.value.enabled
      automatic_rollback_enabled   = !automatic_os_upgrade_policy.value.disable_automatic_rollback
    }
  }

  # Spot presence is the priority switch; the eviction policy is the
  # explicit fleet-level choice.
  priority        = local.priority
  eviction_policy = local.eviction_policy
  max_bid_price   = local.max_bid_price

  dynamic "spot_restore" {
    for_each = var.spec.spot != null && try(var.spec.spot.restore, null) != null ? [var.spec.spot.restore] : []
    content {
      enabled = true
      timeout = spot_restore.value.timeout != null && spot_restore.value.timeout != "" ? spot_restore.value.timeout : null
    }
  }

  dynamic "automatic_instance_repair" {
    for_each = var.spec.automatic_instance_repair != null ? [var.spec.automatic_instance_repair] : []
    content {
      enabled      = automatic_instance_repair.value.enabled
      grace_period = automatic_instance_repair.value.grace_period != null && automatic_instance_repair.value.grace_period != "" ? automatic_instance_repair.value.grace_period : null
      action       = automatic_instance_repair.value.action != null ? local.repair_action_map[automatic_instance_repair.value.action] : null
    }
  }

  dynamic "termination_notification" {
    for_each = var.spec.termination_notification != null ? [var.spec.termination_notification] : []
    content {
      enabled = true
      timeout = termination_notification.value.timeout != null && termination_notification.value.timeout != "" ? termination_notification.value.timeout : null
    }
  }

  dynamic "scale_in" {
    for_each = var.spec.scale_in != null ? [var.spec.scale_in] : []
    content {
      rule                   = scale_in.value.rule != null ? local.scale_in_rule_map[scale_in.value.rule] : "Default"
      force_deletion_enabled = scale_in.value.force_deletion_enabled
    }
  }

  dynamic "identity" {
    for_each = var.spec.identity != null ? [var.spec.identity] : []
    content {
      type         = local.identity_type_map[identity.value.type]
      identity_ids = identity.value.identity_ids
    }
  }

  dynamic "boot_diagnostics" {
    for_each = var.spec.boot_diagnostics != null && coalesce(var.spec.boot_diagnostics.enabled, true) ? [var.spec.boot_diagnostics] : []
    content {
      # Empty URI selects Azure's managed storage -- the right default.
      storage_account_uri = boot_diagnostics.value.storage_account_uri != null && boot_diagnostics.value.storage_account_uri != "" ? boot_diagnostics.value.storage_account_uri : null
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

  dynamic "extension" {
    for_each = var.spec.extensions
    content {
      name                       = extension.value.name
      publisher                  = extension.value.publisher
      type                       = extension.value.type
      type_handler_version       = extension.value.type_handler_version
      auto_upgrade_minor_version = extension.value.auto_upgrade_minor_version_enabled
      automatic_upgrade_enabled  = extension.value.automatic_upgrade_enabled
      settings                   = extension.value.settings
      protected_settings         = extension.value.protected_settings
      provision_after_extensions = extension.value.provision_after_extensions
      force_update_tag           = extension.value.force_update_tag

      dynamic "protected_settings_from_key_vault" {
        for_each = extension.value.protected_settings_from_key_vault != null ? [extension.value.protected_settings_from_key_vault] : []
        content {
          secret_url      = protected_settings_from_key_vault.value.secret_url
          source_vault_id = protected_settings_from_key_vault.value.source_vault_id
        }
      }
    }
  }
  extensions_time_budget = var.spec.extensions_time_budget != null && var.spec.extensions_time_budget != "" ? var.spec.extensions_time_budget : null

  do_not_run_extensions_on_overprovisioned_machines = var.spec.do_not_run_extensions_on_overprovisioned_machines != null ? var.spec.do_not_run_extensions_on_overprovisioned_machines : false

  overprovision          = var.spec.overprovision != null ? var.spec.overprovision : true
  single_placement_group = var.spec.placement != null ? var.spec.placement.single_placement_group : null

  zones                       = length(var.spec.zones) > 0 ? var.spec.zones : null
  zone_balance                = var.spec.zone_balance
  platform_fault_domain_count = var.spec.platform_fault_domain_count

  proximity_placement_group_id  = var.spec.placement != null ? var.spec.placement.proximity_placement_group_id : null
  capacity_reservation_group_id = var.spec.placement != null ? var.spec.placement.capacity_reservation_group_id : null
  host_group_id                 = var.spec.placement != null ? var.spec.placement.host_group_id : null

  encryption_at_host_enabled = var.spec.security != null ? var.spec.security.encryption_at_host_enabled : false
  secure_boot_enabled        = var.spec.security != null ? var.spec.security.secure_boot_enabled : false
  vtpm_enabled               = var.spec.security != null ? var.spec.security.vtpm_enabled : false

  dynamic "plan" {
    for_each = var.spec.plan != null ? [var.spec.plan] : []
    content {
      name      = plan.value.name
      product   = plan.value.product
      publisher = plan.value.publisher
    }
  }

  dynamic "gallery_application" {
    for_each = var.spec.gallery_applications
    content {
      version_id             = gallery_application.value.version_id
      order                  = gallery_application.value.order
      tag                    = gallery_application.value.tag
      configuration_blob_uri = gallery_application.value.configuration_blob_uri
    }
  }

  dynamic "additional_capabilities" {
    for_each = var.spec.additional_capabilities != null ? [var.spec.additional_capabilities] : []
    content {
      ultra_ssd_enabled = additional_capabilities.value.ultra_ssd_enabled
    }
  }

  provision_vm_agent           = var.spec.provision_vm_agent
  extension_operations_enabled = var.spec.extension_operations_enabled
  edge_zone                    = var.spec.edge_zone
}

# ---------------------------------------------------------------------
# UNIFORM + Windows
# ---------------------------------------------------------------------
resource "azurerm_windows_virtual_machine_scale_set" "main" {
  count = local.is_uniform && !local.is_linux ? 1 : 0

  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  sku                 = var.spec.sku_name
  instances           = var.spec.instances != null ? var.spec.instances : 0
  tags                = local.final_tags

  admin_username       = local.windows_profile.admin_username
  admin_password       = local.windows_profile.admin_password
  computer_name_prefix = local.computer_name_prefix

  automatic_updates_enabled = local.windows_profile.automatic_updates_enabled
  timezone                  = local.windows_profile.timezone
  license_type              = local.windows_profile.license_type != null ? local.windows_license_map[local.windows_profile.license_type] : null

  dynamic "winrm_listener" {
    for_each = local.windows_profile.winrm_listeners
    content {
      protocol        = local.winrm_protocol_map[winrm_listener.value.protocol]
      certificate_url = winrm_listener.value.certificate_url
    }
  }

  dynamic "additional_unattend_content" {
    for_each = local.windows_profile.additional_unattend_contents
    content {
      setting = local.unattend_setting_map[additional_unattend_content.value.setting]
      content = additional_unattend_content.value.content
    }
  }

  custom_data = var.spec.custom_data != null && var.spec.custom_data != "" ? var.spec.custom_data : null
  user_data   = var.spec.user_data != null && var.spec.user_data != "" ? var.spec.user_data : null

  source_image_id = var.spec.source_image_id
  dynamic "source_image_reference" {
    for_each = var.spec.source_image_reference != null ? [var.spec.source_image_reference] : []
    content {
      publisher = source_image_reference.value.publisher
      offer     = source_image_reference.value.offer
      sku       = source_image_reference.value.sku
      version   = source_image_reference.value.version
    }
  }

  os_disk {
    caching                          = local.caching_map[var.spec.os_disk.caching]
    storage_account_type             = local.os_disk_storage_map[var.spec.os_disk.storage_account_type]
    disk_size_gb                     = var.spec.os_disk.disk_size_gb
    write_accelerator_enabled        = var.spec.os_disk.write_accelerator_enabled
    disk_encryption_set_id           = var.spec.os_disk.disk_encryption_set_id
    secure_vm_disk_encryption_set_id = var.spec.os_disk.secure_vm_disk_encryption_set_id
    security_encryption_type         = var.spec.os_disk.security_encryption_type != null ? local.security_encryption_map[var.spec.os_disk.security_encryption_type] : null

    dynamic "diff_disk_settings" {
      for_each = var.spec.os_disk.diff_disk_settings != null ? [var.spec.os_disk.diff_disk_settings] : []
      content {
        option    = "Local"
        placement = diff_disk_settings.value.placement != null ? local.diff_disk_placement_map[diff_disk_settings.value.placement] : null
      }
    }
  }

  dynamic "data_disk" {
    for_each = var.spec.data_disks
    content {
      lun                       = data_disk.value.lun
      caching                   = local.caching_map[data_disk.value.caching]
      disk_size_gb              = data_disk.value.disk_size_gb
      storage_account_type      = local.data_disk_storage_map[data_disk.value.storage_account_type]
      create_option             = data_disk.value.create_option != null ? local.create_option_map[data_disk.value.create_option] : "Empty"
      write_accelerator_enabled = data_disk.value.write_accelerator_enabled
      disk_encryption_set_id    = data_disk.value.disk_encryption_set_id
      disk_iops_read_write      = data_disk.value.ultra_ssd_disk_iops_read_write
      disk_mbps_read_write      = data_disk.value.ultra_ssd_disk_mbps_read_write
    }
  }

  dynamic "network_interface" {
    for_each = var.spec.network_interfaces
    content {
      name                           = network_interface.value.name
      primary                        = network_interface.value.primary
      dns_servers                    = network_interface.value.dns_servers
      accelerated_networking_enabled = network_interface.value.accelerated_networking_enabled
      ip_forwarding_enabled          = network_interface.value.ip_forwarding_enabled
      network_security_group_id      = network_interface.value.network_security_group_id
      auxiliary_mode                 = network_interface.value.auxiliary_mode != null ? local.auxiliary_mode_map[network_interface.value.auxiliary_mode] : null
      auxiliary_sku                  = network_interface.value.auxiliary_sku != null ? local.auxiliary_sku_map[network_interface.value.auxiliary_sku] : null

      dynamic "ip_configuration" {
        for_each = network_interface.value.ip_configurations
        content {
          name      = ip_configuration.value.name
          primary   = ip_configuration.value.primary
          subnet_id = ip_configuration.value.subnet_id
          version   = ip_configuration.value.version != null ? local.ip_version_map[ip_configuration.value.version] : "IPv4"

          load_balancer_backend_address_pool_ids       = ip_configuration.value.load_balancer_backend_address_pool_ids
          load_balancer_inbound_nat_rules_ids          = ip_configuration.value.load_balancer_inbound_nat_rule_ids
          application_gateway_backend_address_pool_ids = ip_configuration.value.application_gateway_backend_address_pool_ids
          application_security_group_ids               = ip_configuration.value.application_security_group_ids

          dynamic "public_ip_address" {
            for_each = ip_configuration.value.public_ip_address != null ? [ip_configuration.value.public_ip_address] : []
            content {
              name                    = public_ip_address.value.name
              domain_name_label       = public_ip_address.value.domain_name_label
              idle_timeout_in_minutes = public_ip_address.value.idle_timeout_in_minutes
              version                 = public_ip_address.value.version != null ? local.ip_version_map[public_ip_address.value.version] : null
              public_ip_prefix_id     = public_ip_address.value.public_ip_prefix_id

              dynamic "ip_tag" {
                for_each = public_ip_address.value.ip_tags
                content {
                  type = ip_tag.value.type
                  tag  = ip_tag.value.tag
                }
              }
            }
          }
        }
      }
    }
  }

  upgrade_mode    = local.upgrade_mode
  health_probe_id = var.spec.upgrade_policy != null ? var.spec.upgrade_policy.health_probe_id : null

  dynamic "rolling_upgrade_policy" {
    for_each = var.spec.upgrade_policy != null && try(var.spec.upgrade_policy.rolling, null) != null ? [var.spec.upgrade_policy.rolling] : []
    content {
      max_batch_instance_percent              = rolling_upgrade_policy.value.max_batch_instance_percent
      max_unhealthy_instance_percent          = rolling_upgrade_policy.value.max_unhealthy_instance_percent
      max_unhealthy_upgraded_instance_percent = rolling_upgrade_policy.value.max_unhealthy_upgraded_instance_percent
      pause_time_between_batches              = rolling_upgrade_policy.value.pause_time_between_batches
      cross_zone_upgrades_enabled             = rolling_upgrade_policy.value.cross_zone_upgrades_enabled
      prioritize_unhealthy_instances_enabled  = rolling_upgrade_policy.value.prioritize_unhealthy_instances_enabled
      maximum_surge_instances_enabled         = rolling_upgrade_policy.value.maximum_surge_instances_enabled
    }
  }

  dynamic "automatic_os_upgrade_policy" {
    for_each = var.spec.upgrade_policy != null && try(var.spec.upgrade_policy.automatic_os_upgrade, null) != null ? [var.spec.upgrade_policy.automatic_os_upgrade] : []
    content {
      # The provider expresses rollback as a positive flag; the spec keeps
      # the ARM-side disable_automatic_rollback name, so the value inverts
      # here (and only here).
      automatic_os_upgrade_enabled = automatic_os_upgrade_policy.value.enabled
      automatic_rollback_enabled   = !automatic_os_upgrade_policy.value.disable_automatic_rollback
    }
  }

  priority        = local.priority
  eviction_policy = local.eviction_policy
  max_bid_price   = local.max_bid_price

  dynamic "spot_restore" {
    for_each = var.spec.spot != null && try(var.spec.spot.restore, null) != null ? [var.spec.spot.restore] : []
    content {
      enabled = true
      timeout = spot_restore.value.timeout != null && spot_restore.value.timeout != "" ? spot_restore.value.timeout : null
    }
  }

  dynamic "automatic_instance_repair" {
    for_each = var.spec.automatic_instance_repair != null ? [var.spec.automatic_instance_repair] : []
    content {
      enabled      = automatic_instance_repair.value.enabled
      grace_period = automatic_instance_repair.value.grace_period != null && automatic_instance_repair.value.grace_period != "" ? automatic_instance_repair.value.grace_period : null
      action       = automatic_instance_repair.value.action != null ? local.repair_action_map[automatic_instance_repair.value.action] : null
    }
  }

  dynamic "termination_notification" {
    for_each = var.spec.termination_notification != null ? [var.spec.termination_notification] : []
    content {
      enabled = true
      timeout = termination_notification.value.timeout != null && termination_notification.value.timeout != "" ? termination_notification.value.timeout : null
    }
  }

  dynamic "scale_in" {
    for_each = var.spec.scale_in != null ? [var.spec.scale_in] : []
    content {
      rule                   = scale_in.value.rule != null ? local.scale_in_rule_map[scale_in.value.rule] : "Default"
      force_deletion_enabled = scale_in.value.force_deletion_enabled
    }
  }

  dynamic "identity" {
    for_each = var.spec.identity != null ? [var.spec.identity] : []
    content {
      type         = local.identity_type_map[identity.value.type]
      identity_ids = identity.value.identity_ids
    }
  }

  dynamic "boot_diagnostics" {
    for_each = var.spec.boot_diagnostics != null && coalesce(var.spec.boot_diagnostics.enabled, true) ? [var.spec.boot_diagnostics] : []
    content {
      storage_account_uri = boot_diagnostics.value.storage_account_uri != null && boot_diagnostics.value.storage_account_uri != "" ? boot_diagnostics.value.storage_account_uri : null
    }
  }

  dynamic "secret" {
    for_each = var.spec.secrets
    content {
      key_vault_id = secret.value.key_vault_id
      dynamic "certificate" {
        for_each = secret.value.certificates
        content {
          url   = certificate.value.url
          store = certificate.value.store
        }
      }
    }
  }

  dynamic "extension" {
    for_each = var.spec.extensions
    content {
      name                       = extension.value.name
      publisher                  = extension.value.publisher
      type                       = extension.value.type
      type_handler_version       = extension.value.type_handler_version
      auto_upgrade_minor_version = extension.value.auto_upgrade_minor_version_enabled
      automatic_upgrade_enabled  = extension.value.automatic_upgrade_enabled
      settings                   = extension.value.settings
      protected_settings         = extension.value.protected_settings
      provision_after_extensions = extension.value.provision_after_extensions
      force_update_tag           = extension.value.force_update_tag

      dynamic "protected_settings_from_key_vault" {
        for_each = extension.value.protected_settings_from_key_vault != null ? [extension.value.protected_settings_from_key_vault] : []
        content {
          secret_url      = protected_settings_from_key_vault.value.secret_url
          source_vault_id = protected_settings_from_key_vault.value.source_vault_id
        }
      }
    }
  }
  extensions_time_budget = var.spec.extensions_time_budget != null && var.spec.extensions_time_budget != "" ? var.spec.extensions_time_budget : null

  do_not_run_extensions_on_overprovisioned_machines = var.spec.do_not_run_extensions_on_overprovisioned_machines != null ? var.spec.do_not_run_extensions_on_overprovisioned_machines : false

  overprovision          = var.spec.overprovision != null ? var.spec.overprovision : true
  single_placement_group = var.spec.placement != null ? var.spec.placement.single_placement_group : null

  zones                       = length(var.spec.zones) > 0 ? var.spec.zones : null
  zone_balance                = var.spec.zone_balance
  platform_fault_domain_count = var.spec.platform_fault_domain_count

  proximity_placement_group_id  = var.spec.placement != null ? var.spec.placement.proximity_placement_group_id : null
  capacity_reservation_group_id = var.spec.placement != null ? var.spec.placement.capacity_reservation_group_id : null
  host_group_id                 = var.spec.placement != null ? var.spec.placement.host_group_id : null

  encryption_at_host_enabled = var.spec.security != null ? var.spec.security.encryption_at_host_enabled : false
  secure_boot_enabled        = var.spec.security != null ? var.spec.security.secure_boot_enabled : false
  vtpm_enabled               = var.spec.security != null ? var.spec.security.vtpm_enabled : false

  dynamic "plan" {
    for_each = var.spec.plan != null ? [var.spec.plan] : []
    content {
      name      = plan.value.name
      product   = plan.value.product
      publisher = plan.value.publisher
    }
  }

  dynamic "gallery_application" {
    for_each = var.spec.gallery_applications
    content {
      version_id             = gallery_application.value.version_id
      order                  = gallery_application.value.order
      tag                    = gallery_application.value.tag
      configuration_blob_uri = gallery_application.value.configuration_blob_uri
    }
  }

  dynamic "additional_capabilities" {
    for_each = var.spec.additional_capabilities != null ? [var.spec.additional_capabilities] : []
    content {
      ultra_ssd_enabled = additional_capabilities.value.ultra_ssd_enabled
    }
  }

  provision_vm_agent           = var.spec.provision_vm_agent
  extension_operations_enabled = var.spec.extension_operations_enabled
  edge_zone                    = var.spec.edge_zone
}

# ---------------------------------------------------------------------
# FLEXIBLE (orchestrated) -- either OS
# ---------------------------------------------------------------------
resource "azurerm_orchestrated_virtual_machine_scale_set" "main" {
  count = local.is_uniform ? 0 : 1

  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  sku_name            = var.spec.sku_name
  instances           = var.spec.instances
  tags                = local.final_tags

  # Fault domains are the FLEXIBLE resilience contract: 1 with zones
  # (zones are the resilience unit) or the region's max for regional
  # spreading. Required by ARM.
  platform_fault_domain_count = var.spec.platform_fault_domain_count

  # Unset applies Azure's default ("2020-11-01"); "2022-11-01" unlocks
  # NIC auxiliary acceleration.
  network_api_version = var.spec.network_api_version != null && var.spec.network_api_version != "" ? var.spec.network_api_version : null

  # PARITY-EXCEPTION: this module realizes ranked virtual_machine_size
  # blocks (azurerm v5 supports per-size ranks), while the Pulumi module's
  # pinned pulumi-azure v6 SDK bridges the legacy sku_profile shape (plain
  # vm_sizes, no ranks) and fails loudly on any ranked profile. Sizes
  # deploy identically on both engines; output-neutral (sku_profile never
  # feeds stack outputs); revisit when the SDK catches up.
  dynamic "sku_profile" {
    for_each = var.spec.sku_profile != null ? [var.spec.sku_profile] : []
    content {
      allocation_strategy = local.allocation_strategy_map[sku_profile.value.allocation_strategy]

      dynamic "virtual_machine_size" {
        for_each = sku_profile.value.vm_sizes
        content {
          name = virtual_machine_size.value.name
          rank = virtual_machine_size.value.rank
        }
      }
    }
  }

  os_profile {
    custom_data = var.spec.custom_data != null && var.spec.custom_data != "" ? var.spec.custom_data : null

    dynamic "linux_configuration" {
      for_each = local.is_linux ? [local.linux_profile] : []
      content {
        admin_username                  = linux_configuration.value.admin_username
        admin_password                  = linux_configuration.value.admin_password
        disable_password_authentication = linux_configuration.value.disable_password_authentication
        computer_name_prefix            = local.computer_name_prefix
        provision_vm_agent              = var.spec.provision_vm_agent
        patch_mode                      = linux_configuration.value.patch_mode != null ? local.linux_patch_mode_map[linux_configuration.value.patch_mode] : null
        patch_assessment_mode           = linux_configuration.value.patch_assessment_mode != null ? local.assessment_mode_map[linux_configuration.value.patch_assessment_mode] : null

        dynamic "admin_ssh_key" {
          for_each = linux_configuration.value.ssh_public_keys
          content {
            public_key = admin_ssh_key.value.public_key
            username   = coalesce(admin_ssh_key.value.username, linux_configuration.value.admin_username)
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
      }
    }

    dynamic "windows_configuration" {
      for_each = local.is_linux ? [] : [local.windows_profile]
      content {
        admin_username            = windows_configuration.value.admin_username
        admin_password            = windows_configuration.value.admin_password
        computer_name_prefix      = local.computer_name_prefix
        automatic_updates_enabled = windows_configuration.value.automatic_updates_enabled
        hotpatching_enabled       = windows_configuration.value.hotpatching_enabled
        provision_vm_agent        = var.spec.provision_vm_agent
        patch_mode                = windows_configuration.value.patch_mode != null ? local.windows_patch_mode_map[windows_configuration.value.patch_mode] : null
        patch_assessment_mode     = windows_configuration.value.patch_assessment_mode != null ? local.assessment_mode_map[windows_configuration.value.patch_assessment_mode] : null
        timezone                  = windows_configuration.value.timezone

        dynamic "winrm_listener" {
          for_each = windows_configuration.value.winrm_listeners
          content {
            protocol        = local.winrm_protocol_map[winrm_listener.value.protocol]
            certificate_url = winrm_listener.value.certificate_url
          }
        }

        dynamic "additional_unattend_content" {
          for_each = windows_configuration.value.additional_unattend_contents
          content {
            setting = local.unattend_setting_map[additional_unattend_content.value.setting]
            content = additional_unattend_content.value.content
          }
        }

        dynamic "secret" {
          for_each = var.spec.secrets
          content {
            key_vault_id = secret.value.key_vault_id
            dynamic "certificate" {
              for_each = secret.value.certificates
              content {
                url   = certificate.value.url
                store = certificate.value.store
              }
            }
          }
        }
      }
    }
  }

  # Azure Hybrid Benefit lives top-level on the FLEXIBLE resource
  # (Windows fleets only).
  license_type = !local.is_linux && local.windows_profile.license_type != null ? local.windows_license_map[local.windows_profile.license_type] : null

  # The FLEXIBLE resource takes user data pre-encoded (the spec carries
  # base64 already).
  user_data_base64 = var.spec.user_data != null && var.spec.user_data != "" ? var.spec.user_data : null

  source_image_id = var.spec.source_image_id
  dynamic "source_image_reference" {
    for_each = var.spec.source_image_reference != null ? [var.spec.source_image_reference] : []
    content {
      publisher = source_image_reference.value.publisher
      offer     = source_image_reference.value.offer
      sku       = source_image_reference.value.sku
      version   = source_image_reference.value.version
    }
  }

  os_disk {
    caching                   = local.caching_map[var.spec.os_disk.caching]
    storage_account_type      = local.os_disk_storage_map[var.spec.os_disk.storage_account_type]
    disk_size_gb              = var.spec.os_disk.disk_size_gb
    write_accelerator_enabled = var.spec.os_disk.write_accelerator_enabled
    disk_encryption_set_id    = var.spec.os_disk.disk_encryption_set_id

    dynamic "diff_disk_settings" {
      for_each = var.spec.os_disk.diff_disk_settings != null ? [var.spec.os_disk.diff_disk_settings] : []
      content {
        option    = "Local"
        placement = diff_disk_settings.value.placement != null ? local.diff_disk_placement_map[diff_disk_settings.value.placement] : null
      }
    }
  }

  dynamic "data_disk" {
    for_each = var.spec.data_disks
    content {
      lun                       = data_disk.value.lun
      caching                   = local.caching_map[data_disk.value.caching]
      disk_size_gb              = data_disk.value.disk_size_gb
      storage_account_type      = local.data_disk_storage_map[data_disk.value.storage_account_type]
      create_option             = data_disk.value.create_option != null ? local.create_option_map[data_disk.value.create_option] : "Empty"
      write_accelerator_enabled = data_disk.value.write_accelerator_enabled
      disk_encryption_set_id    = data_disk.value.disk_encryption_set_id
      disk_iops_read_write      = data_disk.value.ultra_ssd_disk_iops_read_write
      disk_mbps_read_write      = data_disk.value.ultra_ssd_disk_mbps_read_write
    }
  }

  dynamic "network_interface" {
    for_each = var.spec.network_interfaces
    content {
      name                           = network_interface.value.name
      primary                        = network_interface.value.primary
      dns_servers                    = network_interface.value.dns_servers
      accelerated_networking_enabled = network_interface.value.accelerated_networking_enabled
      ip_forwarding_enabled          = network_interface.value.ip_forwarding_enabled
      network_security_group_id      = network_interface.value.network_security_group_id
      auxiliary_mode                 = network_interface.value.auxiliary_mode != null ? local.auxiliary_mode_map[network_interface.value.auxiliary_mode] : null
      auxiliary_sku                  = network_interface.value.auxiliary_sku != null ? local.auxiliary_sku_map[network_interface.value.auxiliary_sku] : null

      dynamic "ip_configuration" {
        for_each = network_interface.value.ip_configurations
        content {
          name      = ip_configuration.value.name
          primary   = ip_configuration.value.primary
          subnet_id = ip_configuration.value.subnet_id
          version   = ip_configuration.value.version != null ? local.ip_version_map[ip_configuration.value.version] : "IPv4"

          load_balancer_backend_address_pool_ids       = ip_configuration.value.load_balancer_backend_address_pool_ids
          application_gateway_backend_address_pool_ids = ip_configuration.value.application_gateway_backend_address_pool_ids
          application_security_group_ids               = ip_configuration.value.application_security_group_ids

          dynamic "public_ip_address" {
            for_each = ip_configuration.value.public_ip_address != null ? [ip_configuration.value.public_ip_address] : []
            content {
              name                    = public_ip_address.value.name
              domain_name_label       = public_ip_address.value.domain_name_label
              idle_timeout_in_minutes = public_ip_address.value.idle_timeout_in_minutes
              version                 = public_ip_address.value.version != null ? local.ip_version_map[public_ip_address.value.version] : null
              public_ip_prefix_id     = public_ip_address.value.public_ip_prefix_id

              dynamic "ip_tag" {
                for_each = public_ip_address.value.ip_tags
                content {
                  type = ip_tag.value.type
                  tag  = ip_tag.value.tag
                }
              }
            }
          }
        }
      }
    }
  }

  upgrade_mode = local.upgrade_mode

  dynamic "rolling_upgrade_policy" {
    for_each = var.spec.upgrade_policy != null && try(var.spec.upgrade_policy.rolling, null) != null ? [var.spec.upgrade_policy.rolling] : []
    content {
      max_batch_instance_percent              = rolling_upgrade_policy.value.max_batch_instance_percent
      max_unhealthy_instance_percent          = rolling_upgrade_policy.value.max_unhealthy_instance_percent
      max_unhealthy_upgraded_instance_percent = rolling_upgrade_policy.value.max_unhealthy_upgraded_instance_percent
      pause_time_between_batches              = rolling_upgrade_policy.value.pause_time_between_batches
      cross_zone_upgrades_enabled             = rolling_upgrade_policy.value.cross_zone_upgrades_enabled
      prioritize_unhealthy_instances_enabled  = rolling_upgrade_policy.value.prioritize_unhealthy_instances_enabled
      maximum_surge_instances_enabled         = rolling_upgrade_policy.value.maximum_surge_instances_enabled
    }
  }

  priority        = local.priority
  eviction_policy = local.eviction_policy
  max_bid_price   = local.max_bid_price

  dynamic "priority_mix" {
    for_each = var.spec.spot != null && try(var.spec.spot.priority_mix, null) != null ? [var.spec.spot.priority_mix] : []
    content {
      base_regular_count            = priority_mix.value.base_regular_count
      regular_percentage_above_base = priority_mix.value.regular_percentage_above_base
    }
  }

  dynamic "automatic_instance_repair" {
    for_each = var.spec.automatic_instance_repair != null ? [var.spec.automatic_instance_repair] : []
    content {
      enabled      = automatic_instance_repair.value.enabled
      grace_period = automatic_instance_repair.value.grace_period != null && automatic_instance_repair.value.grace_period != "" ? automatic_instance_repair.value.grace_period : null
      action       = automatic_instance_repair.value.action != null ? local.repair_action_map[automatic_instance_repair.value.action] : null
    }
  }

  dynamic "termination_notification" {
    for_each = var.spec.termination_notification != null ? [var.spec.termination_notification] : []
    content {
      enabled = true
      timeout = termination_notification.value.timeout != null && termination_notification.value.timeout != "" ? termination_notification.value.timeout : null
    }
  }

  dynamic "identity" {
    for_each = var.spec.identity != null ? [var.spec.identity] : []
    content {
      type         = local.identity_type_map[identity.value.type]
      identity_ids = identity.value.identity_ids
    }
  }

  dynamic "boot_diagnostics" {
    for_each = var.spec.boot_diagnostics != null && coalesce(var.spec.boot_diagnostics.enabled, true) ? [var.spec.boot_diagnostics] : []
    content {
      storage_account_uri = boot_diagnostics.value.storage_account_uri != null && boot_diagnostics.value.storage_account_uri != "" ? boot_diagnostics.value.storage_account_uri : null
    }
  }

  dynamic "extension" {
    for_each = var.spec.extensions
    content {
      name                               = extension.value.name
      publisher                          = extension.value.publisher
      type                               = extension.value.type
      type_handler_version               = extension.value.type_handler_version
      auto_upgrade_minor_version_enabled = extension.value.auto_upgrade_minor_version_enabled
      settings                           = extension.value.settings
      protected_settings                 = extension.value.protected_settings
      failure_suppression_enabled        = extension.value.failure_suppression_enabled

      extensions_to_provision_after_vm_creation = extension.value.provision_after_extensions
      force_extension_execution_on_change       = extension.value.force_update_tag

      dynamic "protected_settings_from_key_vault" {
        for_each = extension.value.protected_settings_from_key_vault != null ? [extension.value.protected_settings_from_key_vault] : []
        content {
          secret_url      = protected_settings_from_key_vault.value.secret_url
          source_vault_id = protected_settings_from_key_vault.value.source_vault_id
        }
      }
    }
  }
  extensions_time_budget = var.spec.extensions_time_budget != null && var.spec.extensions_time_budget != "" ? var.spec.extensions_time_budget : null

  extension_operations_enabled = var.spec.extension_operations_enabled

  zones        = length(var.spec.zones) > 0 ? var.spec.zones : null
  zone_balance = var.spec.zone_balance

  single_placement_group        = var.spec.placement != null ? var.spec.placement.single_placement_group : null
  proximity_placement_group_id  = var.spec.placement != null ? var.spec.placement.proximity_placement_group_id : null
  capacity_reservation_group_id = var.spec.placement != null ? var.spec.placement.capacity_reservation_group_id : null

  encryption_at_host_enabled = var.spec.security != null ? var.spec.security.encryption_at_host_enabled : false

  dynamic "plan" {
    for_each = var.spec.plan != null ? [var.spec.plan] : []
    content {
      name      = plan.value.name
      product   = plan.value.product
      publisher = plan.value.publisher
    }
  }

  dynamic "additional_capabilities" {
    for_each = var.spec.additional_capabilities != null ? [var.spec.additional_capabilities] : []
    content {
      ultra_ssd_enabled = additional_capabilities.value.ultra_ssd_enabled
    }
  }
}
