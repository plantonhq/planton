###############################################################################
# Compute Engine Instance
###############################################################################
resource "google_compute_instance" "instance" {
  name         = var.metadata.name
  project      = local.project_id
  zone         = var.spec.zone
  machine_type = var.spec.machine_type

  # Boot disk configuration
  boot_disk {
    auto_delete = var.spec.boot_disk.auto_delete

    initialize_params {
      image = var.spec.boot_disk.image
      size  = var.spec.boot_disk.size_gb
      type  = var.spec.boot_disk.type
    }
  }

  # Network interfaces
  dynamic "network_interface" {
    for_each = var.spec.network_interfaces
    content {
      network    = network_interface.value.network != null ? network_interface.value.network.value : null
      subnetwork = network_interface.value.subnetwork != null ? network_interface.value.subnetwork.value : null

      # Access configurations for external IP
      dynamic "access_config" {
        for_each = network_interface.value.access_configs
        content {
          nat_ip       = access_config.value.nat_ip
          network_tier = access_config.value.network_tier
        }
      }

      # Alias IP ranges
      dynamic "alias_ip_range" {
        for_each = network_interface.value.alias_ip_ranges
        content {
          ip_cidr_range         = alias_ip_range.value.ip_cidr_range
          subnetwork_range_name = alias_ip_range.value.subnetwork_range_name
        }
      }
    }
  }

  # Attached disks
  dynamic "attached_disk" {
    for_each = var.spec.attached_disks
    content {
      source      = attached_disk.value.source
      device_name = attached_disk.value.device_name
      mode        = attached_disk.value.mode
    }
  }

  # Service account
  dynamic "service_account" {
    for_each = var.spec.service_account != null ? [var.spec.service_account] : []
    content {
      email  = service_account.value.email != null ? service_account.value.email.value : null
      scopes = service_account.value.scopes
    }
  }

  # Scheduling
  scheduling {
    preemptible                 = local.is_preemptible
    automatic_restart           = local.automatic_restart
    on_host_maintenance         = local.on_host_maintenance
    provisioning_model          = local.provisioning_model
    instance_termination_action = var.spec.scheduling != null ? var.spec.scheduling.instance_termination_action : null
  }

  # Labels
  labels = local.final_gcp_labels

  # Tags
  tags = var.spec.tags

  # Metadata
  metadata = local.final_metadata

  # Startup script
  metadata_startup_script = var.spec.startup_script

  # Deletion protection
  deletion_protection = var.spec.deletion_protection

  # Allow stopping for update
  allow_stopping_for_update = var.spec.allow_stopping_for_update
}
