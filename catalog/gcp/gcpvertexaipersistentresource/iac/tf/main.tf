# Enable the Vertex AI API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one persistent resource
# must never disable the API for everything else in the project.
resource "google_project_service" "aiplatform_api" {
  project = local.project_id
  service = "aiplatform.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The project NUMBER the peered network's path needs, read only when the
# path carries a project ID; a numeric path (or no network) performs no
# read.
data "google_project" "network" {
  count      = local.network_needs_number ? 1 : 0
  project_id = local.network_project
}

# The persistent resource: pools of machines Vertex AI keeps provisioned
# for training jobs. Everything but the display name, labels, and each
# pool's replica_count is immutable.
resource "google_vertex_ai_persistent_resource" "this" {
  project      = local.project_id
  location     = var.spec.location
  name         = local.persistent_resource_id
  display_name = local.display_name
  labels       = local.final_labels

  network            = local.network
  reserved_ip_ranges = length(var.spec.reserved_ip_ranges) > 0 ? var.spec.reserved_ip_ranges : null

  # Client-side destroy behavior: DELETE (default; releases the machines),
  # PREVENT, or ABANDON. Sent only when set so the provider default stays
  # in charge otherwise.
  deletion_policy = local.deletion_policy

  dynamic "resource_pools" {
    for_each = var.spec.resource_pools
    content {
      # Optional+Computed: Google generates the pool id when unset.
      id = resource_pools.value.id != "" ? resource_pools.value.id : null
      # Counts are int64 in the spec and decimal strings on the wire.
      replica_count = resource_pools.value.replica_count != null ? tostring(resource_pools.value.replica_count) : null

      machine_spec {
        machine_type      = resource_pools.value.machine_spec.machine_type != "" ? resource_pools.value.machine_spec.machine_type : null
        accelerator_type  = resource_pools.value.machine_spec.accelerator_type != "" ? resource_pools.value.machine_spec.accelerator_type : null
        accelerator_count = resource_pools.value.machine_spec.accelerator_count > 0 ? resource_pools.value.machine_spec.accelerator_count : null
      }

      dynamic "autoscaling_spec" {
        for_each = resource_pools.value.autoscaling_spec != null ? [resource_pools.value.autoscaling_spec] : []
        content {
          min_replica_count = autoscaling_spec.value.min_replica_count != null ? tostring(autoscaling_spec.value.min_replica_count) : null
          max_replica_count = autoscaling_spec.value.max_replica_count != null ? tostring(autoscaling_spec.value.max_replica_count) : null
        }
      }

      # Optional+Computed: Google's boot disk defaults stay in charge unless
      # the spec shapes the disk.
      dynamic "disk_spec" {
        for_each = resource_pools.value.disk_spec != null ? [resource_pools.value.disk_spec] : []
        content {
          boot_disk_size_gb = disk_spec.value.boot_disk_size_gb
          boot_disk_type    = disk_spec.value.boot_disk_type != "" ? disk_spec.value.boot_disk_type : null
        }
      }
    }
  }

  dynamic "psc_interface_config" {
    for_each = var.spec.psc_interface_config != null ? [var.spec.psc_interface_config] : []
    content {
      network_attachment = psc_interface_config.value.network_attachment != "" ? psc_interface_config.value.network_attachment : null

      dynamic "dns_peering_configs" {
        for_each = psc_interface_config.value.dns_peering_configs
        content {
          domain         = dns_peering_configs.value.domain
          target_project = dns_peering_configs.value.target_project
          target_network = dns_peering_configs.value.target_network
        }
      }
    }
  }

  # The spec lifts the two one-leaf wrappers into one bool; false and an
  # absent block are the same fact to Google, so the block is sent only
  # when true.
  dynamic "resource_runtime_spec" {
    for_each = var.spec.enable_custom_service_account ? [true] : []
    content {
      service_account_spec {
        enable_custom_service_account = true
      }
    }
  }

  dynamic "encryption_spec" {
    for_each = local.kms_key_name != null ? [local.kms_key_name] : []
    content {
      kms_key_name = encryption_spec.value
    }
  }

  depends_on = [google_project_service.aiplatform_api]
}
