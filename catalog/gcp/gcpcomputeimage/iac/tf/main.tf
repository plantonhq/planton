# Enable the Compute Engine API -- the control plane that owns images.
# disable_on_destroy is false: tearing down one image must never disable
# the API for everything else in the project.
resource "google_project_service" "compute_api" {
  project = local.project_id
  service = "compute.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The custom image. Everything except labels is ForceNew: a changed source,
# family, key, or guest OS feature replaces the image, which is why images
# are versioned by name and consumed through their family. Exactly one
# source is set (the spec enforces it).
resource "google_compute_image" "this" {
  project         = local.project_id
  name            = local.image_name
  description     = local.description
  family          = local.family
  disk_size_gb    = var.spec.disk_size_gb
  licenses        = local.licenses
  labels          = local.final_labels
  deletion_policy = local.deletion_policy

  source_disk     = local.source_disk
  source_image    = local.source_image
  source_snapshot = local.source_snapshot

  storage_locations = local.storage_locations

  dynamic "raw_disk" {
    for_each = var.spec.raw_disk != null ? [var.spec.raw_disk] : []
    content {
      source         = raw_disk.value.source
      sha1           = raw_disk.value.sha1 != "" ? raw_disk.value.sha1 : null
      container_type = raw_disk.value.container_type != "" ? raw_disk.value.container_type : null
    }
  }

  dynamic "guest_os_features" {
    for_each = local.guest_os_features
    content {
      type = guest_os_features.value
    }
  }

  dynamic "image_encryption_key" {
    for_each = local.kms_key != null ? [local.kms_key] : []
    content {
      kms_key_self_link       = image_encryption_key.value
      kms_key_service_account = local.kms_key_service_account
    }
  }

  dynamic "source_disk_encryption_key" {
    for_each = local.source_encryptions.source_disk_encryption_key != null ? [local.source_encryptions.source_disk_encryption_key] : []
    content {
      kms_key_self_link       = source_disk_encryption_key.value.kms_key
      kms_key_service_account = source_disk_encryption_key.value.kms_key_service_account != "" ? source_disk_encryption_key.value.kms_key_service_account : null
    }
  }

  dynamic "source_image_encryption_key" {
    for_each = local.source_encryptions.source_image_encryption_key != null ? [local.source_encryptions.source_image_encryption_key] : []
    content {
      kms_key_self_link       = source_image_encryption_key.value.kms_key
      kms_key_service_account = source_image_encryption_key.value.kms_key_service_account != "" ? source_image_encryption_key.value.kms_key_service_account : null
    }
  }

  dynamic "source_snapshot_encryption_key" {
    for_each = local.source_encryptions.source_snapshot_encryption_key != null ? [local.source_encryptions.source_snapshot_encryption_key] : []
    content {
      kms_key_self_link       = source_snapshot_encryption_key.value.kms_key
      kms_key_service_account = source_snapshot_encryption_key.value.kms_key_service_account != "" ? source_snapshot_encryption_key.value.kms_key_service_account : null
    }
  }

  # Optional+Computed: sent only when declared, so Google's default Secure
  # Boot certificates never show as a diff.
  dynamic "shielded_instance_initial_state" {
    for_each = var.spec.shielded_instance_initial_state != null ? [var.spec.shielded_instance_initial_state] : []
    content {
      dynamic "pk" {
        for_each = shielded_instance_initial_state.value.pk != null ? [shielded_instance_initial_state.value.pk] : []
        content {
          content   = pk.value.content
          file_type = pk.value.file_type != "" ? pk.value.file_type : null
        }
      }
      dynamic "keks" {
        for_each = shielded_instance_initial_state.value.keks
        content {
          content   = keks.value.content
          file_type = keks.value.file_type != "" ? keks.value.file_type : null
        }
      }
      dynamic "dbs" {
        for_each = shielded_instance_initial_state.value.dbs
        content {
          content   = dbs.value.content
          file_type = dbs.value.file_type != "" ? dbs.value.file_type : null
        }
      }
      dynamic "dbxs" {
        for_each = shielded_instance_initial_state.value.dbxs
        content {
          content   = dbxs.value.content
          file_type = dbxs.value.file_type != "" ? dbxs.value.file_type : null
        }
      }
    }
  }

  dynamic "params" {
    for_each = length(var.spec.resource_manager_tags) > 0 ? [var.spec.resource_manager_tags] : []
    content {
      resource_manager_tags = params.value
    }
  }

  depends_on = [google_project_service.compute_api]
}
