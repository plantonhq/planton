# Enable the Cloud TPU API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one TPU must never
# disable the API for everything else in the project.
resource "google_project_service" "tpu_api" {
  project = local.project_id
  service = "tpu.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The TPU VM, on the beta provider under the recorded admission. Only the
# description, labels, metadata, tags, and data disks update in place;
# everything else replaces the TPU. With neither accelerator_type nor
# accelerator_config set, the provider asks Google for a v2-8.
resource "google_tpu_v2_vm" "this" {
  provider = google-beta

  project          = local.project_id
  zone             = var.spec.zone
  name             = local.node_id
  runtime_version  = var.spec.runtime_version
  accelerator_type = local.accelerator_type
  description      = local.description
  cidr_block       = local.cidr_block
  labels           = local.final_labels
  metadata         = local.metadata
  tags             = local.tags

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  deletion_policy = local.deletion_policy

  dynamic "accelerator_config" {
    for_each = var.spec.accelerator_config != null ? [var.spec.accelerator_config] : []
    content {
      type     = accelerator_config.value.type
      topology = accelerator_config.value.topology
    }
  }

  # One interface renders the singular block; several render the repeated
  # one -- the spec allows only one of the two, as Google does.
  dynamic "network_config" {
    for_each = var.spec.network_config != null ? local.network_configs : []
    content {
      network             = network_config.value.network
      subnetwork          = network_config.value.subnetwork
      enable_external_ips = network_config.value.enable_external_ips
      can_ip_forward      = network_config.value.can_ip_forward
      queue_count         = network_config.value.queue_count
    }
  }

  dynamic "network_configs" {
    for_each = var.spec.network_config == null ? local.network_configs : []
    content {
      network             = network_configs.value.network
      subnetwork          = network_configs.value.subnetwork
      enable_external_ips = network_configs.value.enable_external_ips
      can_ip_forward      = network_configs.value.can_ip_forward
      queue_count         = network_configs.value.queue_count
    }
  }

  dynamic "service_account" {
    for_each = var.spec.service_account != null ? [var.spec.service_account] : []
    content {
      email = service_account.value.email != "" ? service_account.value.email : null
      scope = length(service_account.value.scopes) > 0 ? service_account.value.scopes : null
    }
  }

  dynamic "scheduling_config" {
    for_each = var.spec.scheduling_config != null ? [var.spec.scheduling_config] : []
    content {
      preemptible = scheduling_config.value.preemptible
      spot        = scheduling_config.value.spot
      reserved    = scheduling_config.value.reserved
    }
  }

  dynamic "data_disks" {
    for_each = var.spec.data_disks
    content {
      source_disk = trimprefix(data_disks.value.source_disk, local.compute_prefix)
      mode        = data_disks.value.mode != "" ? data_disks.value.mode : null
    }
  }

  # The spec lifts Google's one-leaf shielded_instance_config wrapper; the
  # block is sent only when Secure Boot is on.
  dynamic "shielded_instance_config" {
    for_each = var.spec.enable_secure_boot ? [true] : []
    content {
      enable_secure_boot = true
    }
  }

  depends_on = [google_project_service.tpu_api]
}
