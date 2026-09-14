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
  description = "Azure Virtual Machine specification"
  type = object({
    # The Azure region the VM runs in (must match every referenced NIC
    # and disk).
    region = string

    # The resource group the VM lives in. References are resolved to a
    # literal name by the platform before the module runs.
    resource_group = string

    # The VM's name, unique within the resource group.
    name = string

    # The VM size (SKU), e.g. "Standard_D2s_v3".
    size = string

    # The attached network interfaces, as resolved ARM IDs (at least
    # one; the first is primary).
    network_interface_ids = list(string)

    # The OS profile: exactly one of linux/windows (spec-level
    # validation enforces it), plus the optional hostname override.
    os_profile = object({
      computer_name = optional(string)

      linux = optional(object({
        admin_username = optional(string)
        ssh_public_keys = optional(list(object({
          public_key = string
          username   = optional(string)
        })), [])
        admin_password                  = optional(string)
        disable_password_authentication = optional(bool, true)
        # Patch mode / license type as the spec enums' name strings.
        patch_mode   = optional(string)
        license_type = optional(string)
      }))

      windows = optional(object({
        admin_username            = optional(string)
        admin_password            = optional(string)
        patch_mode                = optional(string)
        automatic_updates_enabled = optional(bool, true)
        hotpatching_enabled       = optional(bool, false)
        timezone                  = optional(string)
        winrm_listeners = optional(list(object({
          protocol        = string
          certificate_url = optional(string)
        })), [])
        additional_unattend_contents = optional(list(object({
          setting = string
          content = string
        })), [])
        license_type = optional(string)
      }))
    })

    # The OS disk (always required; describes caching/storage even when
    # booting from an existing disk).
    os_disk = object({
      caching              = string
      storage_account_type = string
      disk_size_gb         = optional(number)
      name                 = optional(string)
      diff_disk_settings = optional(object({
        placement = optional(string)
      }))
      disk_encryption_set_id           = optional(string)
      secure_vm_disk_encryption_set_id = optional(string)
      security_encryption_type         = optional(string)
      write_accelerator_enabled        = optional(bool, false)
    })

    # Exactly one image source (spec-level validation enforces it).
    source_image_reference = optional(object({
      publisher = string
      offer     = string
      sku       = string
      version   = string
    }))
    source_image_id    = optional(string)
    os_managed_disk_id = optional(string)

    # Referenced first-class disks mounted at LUNs; each is realized as
    # an attachment resource.
    data_disk_attachments = optional(list(object({
      managed_disk_id           = string
      lun                       = number
      caching                   = string
      write_accelerator_enabled = optional(bool, false)
    })), [])

    # Managed identity: type as the spec enum's name string;
    # identity_ids as resolved ARM IDs.
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))

    # Spot capacity: presence makes the VM a spot instance.
    spot = optional(object({
      eviction_policy = string
      max_bid_price   = optional(number, -1)
    }))

    # Placement relative to Azure's fault machinery.
    availability = optional(object({
      zone                          = optional(string)
      availability_set_id           = optional(string)
      proximity_placement_group_id  = optional(string)
      capacity_reservation_group_id = optional(string)
      dedicated_host_id             = optional(string)
      dedicated_host_group_id       = optional(string)
      virtual_machine_scale_set_id  = optional(string)
      platform_fault_domain         = optional(number)
    }))

    # Trusted-launch / encryption posture.
    security = optional(object({
      secure_boot_enabled        = optional(bool, false)
      vtpm_enabled               = optional(bool, false)
      encryption_at_host_enabled = optional(bool, false)
    }))

    # Patch orchestration shared across OSes (the per-OS MODE lives in
    # os_profile).
    patching = optional(object({
      assessment_mode                                        = optional(string)
      reboot_setting                                         = optional(string)
      bypass_platform_safety_checks_on_user_schedule_enabled = optional(bool, false)
    }))

    # Boot diagnostics: presence enables; empty URI = managed storage.
    boot_diagnostics = optional(object({
      storage_account_uri = optional(string)
      enabled             = optional(bool)
    }))

    # VM Applications installed at deployment.
    gallery_applications = optional(list(object({
      version_id                                  = string
      order                                       = optional(number)
      tag                                         = optional(string)
      configuration_blob_uri                      = optional(string)
      automatic_upgrade_enabled                   = optional(bool, false)
      treat_failure_as_deployment_failure_enabled = optional(bool, false)
    })), [])

    # Scheduled events: presence enables.
    termination_notification = optional(object({
      timeout = optional(string)
    }))
    os_image_notification = optional(object({
      timeout = optional(string)
    }))

    # Marketplace purchase plan (third-party images only).
    plan = optional(object({
      name      = string
      product   = string
      publisher = string
    }))

    # Cloud-init (fixed at creation, may embed secrets) and IMDS user
    # data (updatable, never secret) -- both base64.
    custom_data = optional(string)
    user_data   = optional(string)

    # Collective extension provisioning budget (ISO 8601; Azure defaults
    # to PT1H30M).
    extensions_time_budget = optional(string)

    # VM agent / extension gates (Azure defaults to true for both).
    provision_vm_agent         = optional(bool, true)
    allow_extension_operations = optional(bool, true)

    # Disk controller as the spec enum's name string (SCSI/NVME); unset
    # applies Azure's default.
    disk_controller_type = optional(string)

    # Niche capability toggles.
    additional_capabilities = optional(object({
      ultra_ssd_enabled   = optional(bool, false)
      hibernation_enabled = optional(bool, false)
    }))

    # Key Vault certificates installed at provisioning.
    secrets = optional(list(object({
      key_vault_id = string
      certificates = list(object({
        url   = string
        store = optional(string)
      }))
    })), [])

    # Edge Zone pinning (fixed at creation).
    edge_zone = optional(string)

    # Free-form user tags, merged over the metadata-derived tags (user
    # tags win on key collision).
    tags = optional(map(string), {})
  })
}
