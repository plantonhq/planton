variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Azure Virtual Machine Scale Set specification"
  type = object({
    # The Azure region the scale set runs in.
    region = string

    # The resource group name. References are resolved to a literal name
    # by the platform before the module runs.
    resource_group = string

    # The scale-set name, unique within the resource group.
    name = string

    # Orchestration mode, as the spec enum's name string
    # (FLEXIBLE/UNIFORM). Unset applies FLEXIBLE -- Azure's
    # recommendation for new workloads. Fixed at creation.
    orchestration_mode = optional(string)

    # The VM size, e.g. "Standard_D2s_v3"; "Mix" (FLEXIBLE only)
    # activates sku_profile.
    sku_name = string

    # The instance count (0-1000). Unset lets the platform manage it.
    instances = optional(number)

    # FLEXIBLE + "Mix" only: mixed sizes with an allocation strategy.
    sku_profile = optional(object({
      # LOWEST_PRICE / CAPACITY_OPTIMIZED / PRIORITIZED.
      allocation_strategy = string
      vm_sizes = list(object({
        name = string
        rank = optional(number)
      }))
    }))

    # The OS profile: exactly one of linux/windows (spec-level
    # validation enforces it).
    os_profile = object({
      # The computer-name prefix instances derive their hostnames from.
      computer_name_prefix = optional(string)

      linux = optional(object({
        admin_username = string
        ssh_public_keys = optional(list(object({
          public_key = string
          username   = optional(string)
        })), [])
        admin_password                  = optional(string)
        disable_password_authentication = optional(bool, true)
        # FLEXIBLE only, as the spec enum's name string
        # (LINUX_IMAGE_DEFAULT/LINUX_AUTOMATIC_BY_PLATFORM).
        patch_mode = optional(string)
        # FLEXIBLE only (ASSESSMENT_IMAGE_DEFAULT/
        # ASSESSMENT_AUTOMATIC_BY_PLATFORM).
        patch_assessment_mode = optional(string)
      }))

      windows = optional(object({
        admin_username = string
        admin_password = optional(string)
        # FLEXIBLE only (WINDOWS_MANUAL/AUTOMATIC_BY_OS/
        # WINDOWS_AUTOMATIC_BY_PLATFORM).
        patch_mode            = optional(string)
        patch_assessment_mode = optional(string)
        # Azure's default is true.
        automatic_updates_enabled = optional(bool, true)
        # FLEXIBLE only; requires platform patching + health extension.
        hotpatching_enabled = optional(bool, false)
        timezone            = optional(string)
        winrm_listeners = optional(list(object({
          # HTTP/HTTPS, as the spec enum's name string.
          protocol        = string
          certificate_url = optional(string)
        })), [])
        additional_unattend_contents = optional(list(object({
          # AUTO_LOGON/FIRST_LOGON_COMMANDS.
          setting = string
          content = string
        })), [])
        # WINDOWS_LICENSE_NONE/WINDOWS_CLIENT/WINDOWS_SERVER.
        license_type = optional(string)
      }))
    })

    # The OS disk template.
    os_disk = object({
      # NONE/READ_ONLY/READ_WRITE, as the spec enum's name string.
      caching = string
      # STANDARD_LRS/STANDARD_SSD_LRS/PREMIUM_LRS/STANDARD_SSD_ZRS/
      # PREMIUM_ZRS.
      storage_account_type = string
      disk_size_gb         = optional(number)
      # Presence makes the OS disk ephemeral (requires READ_ONLY
      # caching).
      diff_disk_settings = optional(object({
        # CACHE_DISK/RESOURCE_DISK/NVME_DISK; unset applies Azure's
        # default.
        placement = optional(string)
      }))
      disk_encryption_set_id = optional(string)
      # UNIFORM only.
      secure_vm_disk_encryption_set_id = optional(string)
      # UNIFORM only: VM_GUEST_STATE_ONLY/DISK_WITH_VM_GUEST_STATE.
      security_encryption_type  = optional(string)
      write_accelerator_enabled = optional(bool, false)
    })

    # Data-disk templates every instance stamps.
    data_disks = optional(list(object({
      lun          = number
      caching      = string
      disk_size_gb = number
      # DATA_STANDARD_LRS/DATA_STANDARD_SSD_LRS/DATA_PREMIUM_LRS/
      # DATA_PREMIUM_ZRS/ULTRA_SSD_LRS/PREMIUM_V2_LRS/
      # DATA_STANDARD_SSD_ZRS.
      storage_account_type = string
      # EMPTY/FROM_IMAGE; unset applies Azure's default (Empty).
      create_option                  = optional(string)
      name                           = optional(string)
      write_accelerator_enabled      = optional(bool, false)
      disk_encryption_set_id         = optional(string)
      ultra_ssd_disk_iops_read_write = optional(number)
      ultra_ssd_disk_mbps_read_write = optional(number)
    })), [])

    # Marketplace/platform image coordinates (exactly one image source;
    # spec-level validation enforces it).
    source_image_reference = optional(object({
      publisher = string
      offer     = string
      sku       = string
      version   = string
    }))

    # Custom/gallery image ARM ID.
    source_image_id = optional(string)

    # NIC templates (at least one; the first is primary when several).
    network_interfaces = list(object({
      name    = string
      primary = optional(bool, false)
      ip_configurations = list(object({
        name    = string
        primary = optional(bool, false)
        # The subnet, as a resolved ARM ID.
        subnet_id = optional(string)
        # IPV4/IPV6; unset applies Azure's default (IPv4).
        version = optional(string)
        # Load-balancer pools every instance joins, as resolved ARM IDs.
        load_balancer_backend_address_pool_ids = optional(list(string), [])
        # UNIFORM only: pool-style NAT rules, as resolved ARM IDs.
        load_balancer_inbound_nat_rule_ids = optional(list(string), [])
        # App Gateway pools, as ARM IDs.
        application_gateway_backend_address_pool_ids = optional(list(string), [])
        # ASG memberships, as ARM IDs (up to 20).
        application_security_group_ids = optional(list(string), [])
        # Per-instance public IP template.
        public_ip_address = optional(object({
          name                    = string
          domain_name_label       = optional(string)
          idle_timeout_in_minutes = optional(number)
          version                 = optional(string)
          public_ip_prefix_id     = optional(string)
          ip_tags = optional(list(object({
            type = string
            tag  = string
          })), [])
        }))
      }))
      dns_servers                    = optional(list(string), [])
      accelerated_networking_enabled = optional(bool, false)
      ip_forwarding_enabled          = optional(bool, false)
      network_security_group_id      = optional(string)
      # Paired NVA acceleration (preview), as the spec enums' name
      # strings.
      auxiliary_mode = optional(string)
      auxiliary_sku  = optional(string)
    }))

    # Upgrade orchestration.
    upgrade_policy = optional(object({
      # MANUAL/AUTOMATIC/ROLLING; unset applies Azure's default (Manual).
      mode = optional(string)
      rolling = optional(object({
        max_batch_instance_percent              = number
        max_unhealthy_instance_percent          = number
        max_unhealthy_upgraded_instance_percent = number
        pause_time_between_batches              = string
        cross_zone_upgrades_enabled             = optional(bool, false)
        prioritize_unhealthy_instances_enabled  = optional(bool, false)
        maximum_surge_instances_enabled         = optional(bool, false)
      }))
      # UNIFORM only.
      automatic_os_upgrade = optional(object({
        enabled                    = optional(bool, false)
        disable_automatic_rollback = optional(bool, false)
      }))
      # UNIFORM only: the LB health probe, as a resolved ARM ID.
      health_probe_id = optional(string)
    }))

    # Spot economics. Presence makes the fleet spot.
    spot = optional(object({
      # DEALLOCATE/DELETE.
      eviction_policy = string
      max_bid_price   = optional(number)
      # UNIFORM only. Presence enables restore.
      restore = optional(object({
        timeout = optional(string)
      }))
      # FLEXIBLE only.
      priority_mix = optional(object({
        base_regular_count            = optional(number)
        regular_percentage_above_base = optional(number)
      }))
    }))

    # Managed identity (FLEXIBLE sets: USER_ASSIGNED only).
    identity = optional(object({
      # SYSTEM_ASSIGNED/USER_ASSIGNED/SYSTEM_AND_USER_ASSIGNED.
      type         = string
      identity_ids = optional(list(string), [])
    }))

    # Trusted-launch / encryption posture (secure boot + vTPM are
    # UNIFORM only).
    security = optional(object({
      secure_boot_enabled        = optional(bool, false)
      vtpm_enabled               = optional(bool, false)
      encryption_at_host_enabled = optional(bool, false)
    }))

    # Automatic replacement of unhealthy instances.
    automatic_instance_repair = optional(object({
      enabled      = bool
      grace_period = optional(string)
      # REPLACE/RESTART/REIMAGE; unset applies Azure's default.
      action = optional(string)
    }))

    # Pre-termination scheduled event. Presence enables it.
    termination_notification = optional(object({
      timeout = optional(string)
    }))

    # Extensions installed onto every instance.
    extensions = optional(list(object({
      name                               = string
      publisher                          = string
      type                               = string
      type_handler_version               = string
      auto_upgrade_minor_version_enabled = optional(bool, true)
      automatic_upgrade_enabled          = optional(bool, false)
      settings                           = optional(string)
      protected_settings                 = optional(string)
      protected_settings_from_key_vault = optional(object({
        secret_url      = string
        source_vault_id = string
      }))
      provision_after_extensions = optional(list(string), [])
      force_update_tag           = optional(string)
      # FLEXIBLE only.
      failure_suppression_enabled = optional(bool, false)
    })), [])

    # ISO 8601 budget for all extensions (unset applies PT1H30M).
    extensions_time_budget = optional(string)

    # Boot diagnostics. Presence enables it; empty URI uses managed
    # storage.
    boot_diagnostics = optional(object({
      storage_account_uri = optional(string)
      enabled             = optional(bool)
    }))

    # Availability zones + strict balancing.
    zones        = optional(list(string), [])
    zone_balance = optional(bool, false)

    # Fault domains (REQUIRED on FLEXIBLE sets; spec-level validation
    # enforces it).
    platform_fault_domain_count = optional(number)

    # Placement constraints.
    placement = optional(object({
      proximity_placement_group_id  = optional(string)
      capacity_reservation_group_id = optional(string)
      # UNIFORM only.
      host_group_id          = optional(string)
      single_placement_group = optional(bool)
    }))

    # UNIFORM only knobs.
    overprovision = optional(bool)
    scale_in = optional(object({
      # DEFAULT/NEWEST_VM/OLDEST_VM.
      rule                   = optional(string)
      force_deletion_enabled = optional(bool, false)
    }))
    do_not_run_extensions_on_overprovisioned_machines = optional(bool)

    # Provisioning data (base64). custom_data may embed secrets.
    custom_data = optional(string)
    user_data   = optional(string)

    # Agent + extension gates (Azure defaults: true).
    provision_vm_agent           = optional(bool, true)
    extension_operations_enabled = optional(bool, true)

    # Key Vault certificates installed at provisioning time.
    secrets = optional(list(object({
      key_vault_id = string
      certificates = list(object({
        url   = string
        store = optional(string)
      }))
    })), [])

    # FLEXIBLE only: the Microsoft.Network API version for instance
    # networking ("2020-11-01"/"2022-11-01").
    network_api_version = optional(string)

    # Marketplace purchase plan (third-party images).
    plan = optional(object({
      name      = string
      product   = string
      publisher = string
    }))

    # UNIFORM only: gallery applications.
    gallery_applications = optional(list(object({
      version_id             = string
      order                  = optional(number)
      tag                    = optional(string)
      configuration_blob_uri = optional(string)
    })), [])

    # Niche capabilities.
    additional_capabilities = optional(object({
      ultra_ssd_enabled = optional(bool, false)
    }))

    # Edge Zone pinning (fixed at creation).
    edge_zone = optional(string)

    # Free-form user tags, merged over the metadata-derived tags (user
    # tags win on key collision).
    tags = optional(map(string), {})
  })
}
