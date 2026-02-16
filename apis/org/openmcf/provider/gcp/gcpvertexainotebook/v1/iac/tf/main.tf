resource "google_workbench_instance" "this" {
  name     = local.instance_name
  location = local.location
  project  = local.project_id
  labels   = local.gcp_labels

  instance_id          = local.instance_name
  disable_proxy_access = var.spec.disable_proxy_access ? true : null
  desired_state        = var.spec.desired_state != "" ? var.spec.desired_state : null
  instance_owners      = length(var.spec.instance_owners) > 0 ? var.spec.instance_owners : null

  gce_setup {
    machine_type = var.spec.machine_type

    # Boot disk.
    dynamic "boot_disk" {
      for_each = var.spec.boot_disk != null ? [var.spec.boot_disk] : []
      content {
        disk_type       = boot_disk.value.disk_type != "" ? boot_disk.value.disk_type : null
        disk_size_gb    = boot_disk.value.disk_size_gb != 0 ? tostring(boot_disk.value.disk_size_gb) : null
        disk_encryption = boot_disk.value.kms_key != null ? "CMEK" : null
        kms_key         = boot_disk.value.kms_key != null ? boot_disk.value.kms_key.value : null
      }
    }

    # Data disk.
    dynamic "data_disks" {
      for_each = var.spec.data_disk != null ? [var.spec.data_disk] : []
      content {
        disk_type       = data_disks.value.disk_type != "" ? data_disks.value.disk_type : null
        disk_size_gb    = data_disks.value.disk_size_gb != 0 ? tostring(data_disks.value.disk_size_gb) : null
        disk_encryption = data_disks.value.kms_key != null ? "CMEK" : null
        kms_key         = data_disks.value.kms_key != null ? data_disks.value.kms_key.value : null
      }
    }

    # Accelerator config.
    dynamic "accelerator_configs" {
      for_each = var.spec.accelerator_config != null && var.spec.accelerator_config.type != "" ? [var.spec.accelerator_config] : []
      content {
        type       = accelerator_configs.value.type
        core_count = accelerator_configs.value.core_count != 0 ? tostring(accelerator_configs.value.core_count) : null
      }
    }

    # Network interface.
    dynamic "network_interfaces" {
      for_each = var.spec.network_interface != null ? [var.spec.network_interface] : []
      content {
        network  = network_interfaces.value.network != null ? network_interfaces.value.network.value : null
        subnet   = network_interfaces.value.subnet != null ? network_interfaces.value.subnet.value : null
        nic_type = network_interfaces.value.nic_type != "" ? network_interfaces.value.nic_type : null
      }
    }

    disable_public_ip  = var.spec.disable_public_ip ? true : null
    enable_ip_forwarding = var.spec.enable_ip_forwarding ? true : null

    # Service account.
    dynamic "service_accounts" {
      for_each = var.spec.service_account != null ? [var.spec.service_account] : []
      content {
        email = service_accounts.value.value
      }
    }

    tags     = length(var.spec.tags) > 0 ? var.spec.tags : null
    metadata = length(var.spec.metadata) > 0 ? var.spec.metadata : null

    # VM image (mutually exclusive with container_image).
    dynamic "vm_image" {
      for_each = var.spec.vm_image != null ? [var.spec.vm_image] : []
      content {
        project = vm_image.value.project != "" ? vm_image.value.project : null
        family  = vm_image.value.family != "" ? vm_image.value.family : null
        name    = vm_image.value.name != "" ? vm_image.value.name : null
      }
    }

    # Container image (mutually exclusive with vm_image).
    dynamic "container_image" {
      for_each = var.spec.container_image != null ? [var.spec.container_image] : []
      content {
        repository = container_image.value.repository
        tag        = container_image.value.tag != "" ? container_image.value.tag : null
      }
    }

    # Shielded instance config.
    dynamic "shielded_instance_config" {
      for_each = var.spec.shielded_instance_config != null ? [var.spec.shielded_instance_config] : []
      content {
        enable_secure_boot          = shielded_instance_config.value.enable_secure_boot
        enable_vtpm                 = shielded_instance_config.value.enable_vtpm
        enable_integrity_monitoring = shielded_instance_config.value.enable_integrity_monitoring
      }
    }
  }
}
