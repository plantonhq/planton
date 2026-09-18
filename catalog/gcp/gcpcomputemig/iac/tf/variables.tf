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
  description = "GcpComputeMig specification"
  type = object({
    project_id         = optional(string, "")
    mig_name           = optional(string, "")
    zone               = optional(string, "")
    region             = optional(string, "")
    description        = optional(string, "")
    base_instance_name = optional(string, "")
    template = object({
      machine_type         = string
      description          = optional(string, "")
      instance_description = optional(string, "")
      disks = list(object({
        boot                   = optional(bool, false)
        source_image           = optional(string, "")
        source_snapshot        = optional(string, "")
        source                 = optional(string, "")
        size_gb                = optional(number, 0)
        disk_type              = optional(string, "")
        type                   = optional(string, "")
        auto_delete            = optional(bool)
        device_name            = optional(string, "")
        disk_name              = optional(string, "")
        mode                   = optional(string, "")
        interface              = optional(string, "")
        disk_labels            = optional(map(string), {})
        provisioned_iops       = optional(number)
        provisioned_throughput = optional(number)
        architecture           = optional(string, "")
        guest_os_features      = optional(list(string), [])
        resource_policies      = optional(list(string), [])
        resource_manager_tags  = optional(map(string), {})
        storage_pool           = optional(string, "")
        disk_encryption = optional(object({
          kms_key                 = string
          kms_key_service_account = optional(string, "")
        }))
        source_image_encryption = optional(object({
          kms_key                 = string
          kms_key_service_account = optional(string, "")
        }))
        source_snapshot_encryption = optional(object({
          kms_key                 = string
          kms_key_service_account = optional(string, "")
        }))
      }))
      network_interfaces = list(object({
        network            = optional(string, "")
        subnetwork         = optional(string, "")
        subnetwork_project = optional(string, "")
        network_ip         = optional(string, "")
        access_configs = optional(list(object({
          nat_ip       = optional(string, "")
          network_tier = optional(string, "")
        })), [])
        ipv6_access_configs = optional(list(object({
          network_tier = string
        })), [])
        stack_type  = optional(string, "")
        nic_type    = optional(string, "")
        queue_count = optional(number)
        alias_ip_ranges = optional(list(object({
          ip_cidr_range         = string
          subnetwork_range_name = optional(string, "")
        })), [])
        network_attachment          = optional(string, "")
        vlan                        = optional(number)
        igmp_query                  = optional(string, "")
        ipv6_address                = optional(string, "")
        internal_ipv6_prefix_length = optional(number)
      }))
      service_account = optional(object({
        email  = optional(string, "")
        scopes = list(string)
      }))
      scheduling = optional(object({
        provisioning_model          = optional(string, "")
        automatic_restart           = optional(bool)
        on_host_maintenance         = optional(string, "")
        instance_termination_action = optional(string, "")
        max_run_duration_seconds    = optional(number)
        termination_time            = optional(string, "")
        discard_local_ssds_on_stop  = optional(bool)
        availability_domain         = optional(number)
        min_node_cpus               = optional(number)
        node_affinities = optional(list(object({
          key      = string
          operator = string
          values   = list(string)
        })), [])
        local_ssd_recovery_timeout_seconds = optional(number)
        # Host-error detection timeout in seconds: 90..330 in steps of 30.
        host_error_timeout_seconds = optional(number)
      }))
      shielded_instance_config = optional(object({
        enable_secure_boot          = optional(bool)
        enable_vtpm                 = optional(bool)
        enable_integrity_monitoring = optional(bool)
      }))
      confidential_instance_config = optional(object({
        confidential_instance_type = string
      }))
      advanced_machine_features = optional(object({
        enable_nested_virtualization = optional(bool)
        threads_per_core             = optional(number)
        visible_core_count           = optional(number)
        enable_uefi_networking       = optional(bool)
        performance_monitoring_unit  = optional(string, "")
        turbo_mode                   = optional(string, "")
      }))
      guest_accelerators = optional(list(object({
        type  = string
        count = number
      })), [])
      reservation_affinity = optional(object({
        type = string
        specific_reservation = optional(object({
          key    = string
          values = list(string)
        }))
      }))
      total_egress_bandwidth_tier = optional(string, "")
      metadata                    = optional(map(string), {})
      startup_script              = optional(string, "")
      tags                        = optional(list(string), [])
      labels                      = optional(map(string), {})
      resource_manager_tags       = optional(map(string), {})
      min_cpu_platform            = optional(string, "")
      can_ip_forward              = optional(bool, false)
      key_revocation_action_type  = optional(string, "")
      resource_policies           = optional(list(string), [])
      # Managed workload identity (SPIFFE ID, optional X.509 certificates).
      workload_identity_config = optional(object({
        identity                     = string
        identity_certificate_enabled = optional(bool, false)
      }), null)
    })
    versions = optional(list(object({
      version_name        = optional(string, "")
      template_self_link  = optional(string, "")
      target_size_fixed   = optional(number)
      target_size_percent = optional(number)
    })), [])
    target_size = optional(number)
    named_ports = optional(list(object({
      name = string
      port = number
    })), [])
    update_policy = optional(object({
      minimal_action                 = string
      type                           = string
      most_disruptive_allowed_action = optional(string, "")
      replacement_method             = optional(string, "")
      max_surge_fixed                = optional(number)
      max_surge_percent              = optional(number)
      max_unavailable_fixed          = optional(number)
      max_unavailable_percent        = optional(number)
      instance_redistribution_type   = optional(string, "")
    }))
    auto_healing = optional(object({
      health_check      = string
      initial_delay_sec = number
    }))
    standby_policy = optional(object({
      initial_delay_sec = optional(number)
      mode              = optional(string, "")
    }))
    target_suspended_size = optional(number)
    target_stopped_size   = optional(number)
    stateful_disks = optional(list(object({
      device_name = string
      delete_rule = optional(string, "")
    })), [])
    stateful_external_ips = optional(list(object({
      interface_name = optional(string, "")
      delete_rule    = optional(string, "")
    })), [])
    stateful_internal_ips = optional(list(object({
      interface_name = optional(string, "")
      delete_rule    = optional(string, "")
    })), [])
    instance_lifecycle_policy = optional(object({
      default_action_on_failure     = optional(string, "")
      force_update_on_repair        = optional(string, "")
      on_failed_health_check        = optional(string, "")
      on_repair_allow_changing_zone = optional(string, "")
    }))
    all_instances_config = optional(object({
      labels   = optional(map(string), {})
      metadata = optional(map(string), {})
    }))
    list_managed_instances_results = optional(string, "")
    workload_policy                = optional(string, "")
    target_pools                   = optional(list(string), [])
    wait_for_instances             = optional(bool)
    wait_for_instances_status      = optional(string, "")
    distribution_policy = optional(object({
      zones        = optional(list(string), [])
      target_shape = optional(string, "")
    }))
    instance_flexibility_policy = optional(object({
      instance_selections = list(object({
        name          = string
        machine_types = list(string)
        rank          = optional(number)
      }))
    }))
    target_size_policy_mode = optional(string, "")
    autoscaler = optional(object({
      autoscaler_name       = optional(string, "")
      description           = optional(string, "")
      min_replicas          = optional(number, 0)
      max_replicas          = number
      cooldown_period       = optional(number)
      mode                  = optional(string, "")
      cpu_target            = optional(number)
      cpu_predictive_method = optional(string, "")
      load_balancing_target = optional(number)
      metrics = optional(list(object({
        name                       = string
        target                     = optional(number)
        type                       = optional(string, "")
        filter                     = optional(string, "")
        single_instance_assignment = optional(number)
      })), [])
      scale_in_control = optional(object({
        max_scaled_in_replicas_fixed   = optional(number)
        max_scaled_in_replicas_percent = optional(number)
        time_window_sec                = optional(number)
      }))
      schedules = optional(list(object({
        schedule_name         = string
        schedule              = string
        duration_sec          = number
        min_required_replicas = number
        disabled              = optional(bool)
        time_zone             = optional(string, "")
        description           = optional(string, "")
      })), [])
      stabilization_period = optional(number)
    }))
    per_instance_configs = optional(list(object({
      config_name = string
      preserved_state = optional(object({
        metadata = optional(map(string), {})
        disks = optional(list(object({
          device_name = string
          source      = string
          mode        = optional(string, "")
          delete_rule = optional(string, "")
        })), [])
        external_ips = optional(list(object({
          interface_name = string
          address        = optional(string, "")
          auto_delete    = optional(string, "")
        })), [])
        internal_ips = optional(list(object({
          interface_name = string
          address        = optional(string, "")
          auto_delete    = optional(string, "")
        })), [])
      }))
      minimal_action                   = optional(string, "")
      most_disruptive_allowed_action   = optional(string, "")
      remove_instance_on_destroy       = optional(bool)
      remove_instance_state_on_destroy = optional(bool)
    })), [])
    resize_requests = optional(list(object({
      request_name                   = string
      description                    = optional(string, "")
      resize_by                      = number
      requested_run_duration_seconds = optional(number)
    })), [])
    deletion_policy = optional(string, "")
  })

  validation {
    condition     = var.spec.template.scheduling == null || var.spec.template.scheduling.host_error_timeout_seconds == null || (var.spec.template.scheduling.host_error_timeout_seconds >= 90 && var.spec.template.scheduling.host_error_timeout_seconds <= 330 && var.spec.template.scheduling.host_error_timeout_seconds % 30 == 0)
    error_message = "template.scheduling.host_error_timeout_seconds must be a multiple of 30 between 90 and 330."
  }

  validation {
    condition     = var.spec.template.workload_identity_config == null || startswith(var.spec.template.workload_identity_config.identity, "spiffe://")
    error_message = "template.workload_identity_config.identity must be a SPIFFE ID starting with spiffe://."
  }
}
