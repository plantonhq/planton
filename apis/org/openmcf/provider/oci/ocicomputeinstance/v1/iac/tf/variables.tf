variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "OciComputeInstance specification"
  type = object({
    compartment_id = object({
      value = string
    })

    availability_domain = string
    shape               = string
    display_name        = optional(string, "")
    fault_domain        = optional(string, "")

    is_pv_encryption_in_transit_enabled = optional(bool, null)

    shape_config = optional(object({
      ocpus                      = optional(number, null)
      memory_in_gbs              = optional(number, null)
      baseline_ocpu_utilization  = optional(string, "")
      nvmes                      = optional(number, null)
    }), null)

    source_details = object({
      source_type              = string
      source_id                = string
      boot_volume_size_in_gbs  = optional(number, null)
      boot_volume_vpus_per_gb  = optional(number, null)
      kms_key_id = optional(object({
        value = string
      }), null)
    })

    create_vnic_details = object({
      subnet_id = object({
        value = string
      })
      nsg_ids = optional(list(object({
        value = string
      })), [])
      assign_public_ip          = optional(bool, null)
      display_name              = optional(string, "")
      hostname_label            = optional(string, "")
      private_ip                = optional(string, "")
      skip_source_dest_check    = optional(bool, null)
      assign_private_dns_record = optional(bool, null)
    })

    metadata = optional(map(string), {})

    agent_config = optional(object({
      are_all_plugins_disabled = optional(bool, null)
      is_management_disabled   = optional(bool, null)
      is_monitoring_disabled   = optional(bool, null)
      plugins_config = optional(list(object({
        name          = string
        desired_state = string
      })), [])
    }), null)

    availability_config = optional(object({
      is_live_migration_preferred = optional(bool, null)
      recovery_action             = optional(string, "")
    }), null)

    launch_options = optional(object({
      boot_volume_type                    = optional(string, "")
      network_type                        = optional(string, "")
      firmware                            = optional(string, "")
      is_pv_encryption_in_transit_enabled = optional(bool, null)
      is_consistent_volume_naming_enabled = optional(bool, null)
    }), null)

    instance_options = optional(object({
      are_legacy_imds_endpoints_disabled = optional(bool, null)
    }), null)

    preemptible_instance_config = optional(object({
      preserve_boot_volume = optional(bool, false)
    }), null)

    capacity_reservation_id = optional(object({
      value = string
    }), null)

    dedicated_vm_host_id = optional(object({
      value = string
    }), null)

    platform_config = optional(object({
      type                                             = string
      is_secure_boot_enabled                           = optional(bool, null)
      is_measured_boot_enabled                         = optional(bool, null)
      is_trusted_platform_module_enabled               = optional(bool, null)
      is_memory_encryption_enabled                     = optional(bool, null)
      is_symmetric_multi_threading_enabled             = optional(bool, null)
      are_virtual_instructions_enabled                 = optional(bool, null)
      is_access_control_service_enabled                = optional(bool, null)
      is_input_output_memory_management_unit_enabled   = optional(bool, null)
      numa_nodes_per_socket                            = optional(string, "")
      percentage_of_cores_enabled                      = optional(number, null)
    }), null)
  })
}
