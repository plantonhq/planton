resource "oci_core_instance" "this" {
  compartment_id      = var.spec.compartment_id.value
  availability_domain = var.spec.availability_domain
  shape               = var.spec.shape
  display_name        = local.display_name
  freeform_tags       = local.freeform_tags

  fault_domain                      = var.spec.fault_domain != "" ? var.spec.fault_domain : null
  is_pv_encryption_in_transit_enabled = var.spec.is_pv_encryption_in_transit_enabled
  capacity_reservation_id           = var.spec.capacity_reservation_id != null ? var.spec.capacity_reservation_id.value : null
  dedicated_vm_host_id              = var.spec.dedicated_vm_host_id != null ? var.spec.dedicated_vm_host_id.value : null

  metadata = var.spec.metadata

  source_details {
    source_type             = lookup(local.source_type_map, var.spec.source_details.source_type, "image")
    source_id               = var.spec.source_details.source_id
    boot_volume_size_in_gbs = var.spec.source_details.boot_volume_size_in_gbs
    boot_volume_vpus_per_gb = var.spec.source_details.boot_volume_vpus_per_gb
    kms_key_id              = var.spec.source_details.kms_key_id != null ? var.spec.source_details.kms_key_id.value : null
  }

  create_vnic_details {
    subnet_id              = var.spec.create_vnic_details.subnet_id.value
    nsg_ids                = local.nsg_ids
    assign_public_ip       = var.spec.create_vnic_details.assign_public_ip != null ? tostring(var.spec.create_vnic_details.assign_public_ip) : null
    display_name           = var.spec.create_vnic_details.display_name != "" ? var.spec.create_vnic_details.display_name : null
    hostname_label         = var.spec.create_vnic_details.hostname_label != "" ? var.spec.create_vnic_details.hostname_label : null
    private_ip             = var.spec.create_vnic_details.private_ip != "" ? var.spec.create_vnic_details.private_ip : null
    skip_source_dest_check = var.spec.create_vnic_details.skip_source_dest_check
    assign_private_dns_record = var.spec.create_vnic_details.assign_private_dns_record
  }

  dynamic "shape_config" {
    for_each = var.spec.shape_config != null ? [var.spec.shape_config] : []
    content {
      ocpus                     = shape_config.value.ocpus
      memory_in_gbs             = shape_config.value.memory_in_gbs
      baseline_ocpu_utilization = shape_config.value.baseline_ocpu_utilization != "" ? shape_config.value.baseline_ocpu_utilization : null
      nvmes                     = shape_config.value.nvmes
    }
  }

  dynamic "agent_config" {
    for_each = var.spec.agent_config != null ? [var.spec.agent_config] : []
    content {
      are_all_plugins_disabled = agent_config.value.are_all_plugins_disabled
      is_management_disabled   = agent_config.value.is_management_disabled
      is_monitoring_disabled   = agent_config.value.is_monitoring_disabled

      dynamic "plugins_config" {
        for_each = agent_config.value.plugins_config
        content {
          name          = plugins_config.value.name
          desired_state = upper(plugins_config.value.desired_state)
        }
      }
    }
  }

  dynamic "availability_config" {
    for_each = var.spec.availability_config != null ? [var.spec.availability_config] : []
    content {
      is_live_migration_preferred = availability_config.value.is_live_migration_preferred
      recovery_action             = availability_config.value.recovery_action != "" ? lookup(local.recovery_action_map, availability_config.value.recovery_action, null) : null
    }
  }

  dynamic "launch_options" {
    for_each = var.spec.launch_options != null ? [var.spec.launch_options] : []
    content {
      boot_volume_type                    = launch_options.value.boot_volume_type != "" ? launch_options.value.boot_volume_type : null
      network_type                        = launch_options.value.network_type != "" ? launch_options.value.network_type : null
      firmware                            = launch_options.value.firmware != "" ? lookup(local.firmware_map, launch_options.value.firmware, null) : null
      is_pv_encryption_in_transit_enabled = launch_options.value.is_pv_encryption_in_transit_enabled
      is_consistent_volume_naming_enabled = launch_options.value.is_consistent_volume_naming_enabled
    }
  }

  dynamic "instance_options" {
    for_each = var.spec.instance_options != null ? [var.spec.instance_options] : []
    content {
      are_legacy_imds_endpoints_disabled = instance_options.value.are_legacy_imds_endpoints_disabled
    }
  }

  dynamic "preemptible_instance_config" {
    for_each = var.spec.preemptible_instance_config != null ? [var.spec.preemptible_instance_config] : []
    content {
      preemption_action {
        type                 = "TERMINATE"
        preserve_boot_volume = preemptible_instance_config.value.preserve_boot_volume
      }
    }
  }

  dynamic "platform_config" {
    for_each = var.spec.platform_config != null ? [var.spec.platform_config] : []
    content {
      type                                             = lookup(local.platform_type_map, platform_config.value.type, platform_config.value.type)
      is_secure_boot_enabled                           = platform_config.value.is_secure_boot_enabled
      is_measured_boot_enabled                         = platform_config.value.is_measured_boot_enabled
      is_trusted_platform_module_enabled               = platform_config.value.is_trusted_platform_module_enabled
      is_memory_encryption_enabled                     = platform_config.value.is_memory_encryption_enabled
      is_symmetric_multi_threading_enabled             = platform_config.value.is_symmetric_multi_threading_enabled
      are_virtual_instructions_enabled                 = platform_config.value.are_virtual_instructions_enabled
      is_access_control_service_enabled                = platform_config.value.is_access_control_service_enabled
      is_input_output_memory_management_unit_enabled   = platform_config.value.is_input_output_memory_management_unit_enabled
      numa_nodes_per_socket                            = platform_config.value.numa_nodes_per_socket != "" ? platform_config.value.numa_nodes_per_socket : null
      percentage_of_cores_enabled                      = platform_config.value.percentage_of_cores_enabled
    }
  }
}
