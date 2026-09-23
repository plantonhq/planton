# Enable the Vertex AI API (Colab Enterprise's API) first so a fresh project
# works on the first deploy. disable_on_destroy is false: tearing down one
# template must never disable the API for everything else in the project.
resource "google_project_service" "aiplatform_api" {
  project = local.project_id
  service = "aiplatform.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The runtime template. The display name, the key, and the software update
# in place; everything else (labels included) is fixed at creation.
# Optional+Computed settings are sent only when set so Google's defaults
# stay in charge otherwise.
resource "google_colab_runtime_template" "this" {
  project      = local.project_id
  location     = var.spec.location
  name         = local.runtime_template_id
  display_name = local.display_name
  description  = local.description
  labels       = local.final_labels
  network_tags = local.network_tags

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  deletion_policy = local.deletion_policy

  dynamic "machine_spec" {
    for_each = var.spec.machine_spec != null ? [var.spec.machine_spec] : []
    content {
      machine_type      = machine_spec.value.machine_type != "" ? machine_spec.value.machine_type : null
      accelerator_type  = machine_spec.value.accelerator_type != "" ? machine_spec.value.accelerator_type : null
      accelerator_count = machine_spec.value.accelerator_count != 0 ? machine_spec.value.accelerator_count : null
    }
  }

  # Google types the disk size as a decimal string.
  dynamic "data_persistent_disk_spec" {
    for_each = var.spec.data_persistent_disk_spec != null ? [var.spec.data_persistent_disk_spec] : []
    content {
      disk_type    = data_persistent_disk_spec.value.disk_type != "" ? data_persistent_disk_spec.value.disk_type : null
      disk_size_gb = data_persistent_disk_spec.value.disk_size_gb != 0 ? tostring(data_persistent_disk_spec.value.disk_size_gb) : null
    }
  }

  dynamic "network_spec" {
    for_each = var.spec.network_spec != null ? [var.spec.network_spec] : []
    content {
      enable_internet_access = network_spec.value.enable_internet_access
      network                = network_spec.value.network != "" ? network_spec.value.network : null
      subnetwork             = local.subnetwork
    }
  }

  # The spec lifts Google's one-leaf wrappers; each block is sent only when
  # its field is set.
  dynamic "idle_shutdown_config" {
    for_each = var.spec.idle_timeout != "" ? [var.spec.idle_timeout] : []
    content {
      idle_timeout = idle_shutdown_config.value
    }
  }

  dynamic "euc_config" {
    for_each = var.spec.euc_disabled ? [true] : []
    content {
      euc_disabled = true
    }
  }

  dynamic "shielded_vm_config" {
    for_each = var.spec.enable_secure_boot ? [true] : []
    content {
      enable_secure_boot = true
    }
  }

  dynamic "encryption_spec" {
    for_each = local.kms_key_name != null ? [local.kms_key_name] : []
    content {
      kms_key_name = encryption_spec.value
    }
  }

  dynamic "software_config" {
    for_each = var.spec.software_config != null ? [var.spec.software_config] : []
    content {
      dynamic "env" {
        for_each = software_config.value.env
        content {
          name  = env.value.name
          value = env.value.value != "" ? env.value.value : null
        }
      }

      dynamic "post_startup_script_config" {
        for_each = software_config.value.post_startup_script_config != null ? [software_config.value.post_startup_script_config] : []
        content {
          post_startup_script          = post_startup_script_config.value.post_startup_script != "" ? post_startup_script_config.value.post_startup_script : null
          post_startup_script_url      = post_startup_script_config.value.post_startup_script_url != "" ? post_startup_script_config.value.post_startup_script_url : null
          post_startup_script_behavior = post_startup_script_config.value.post_startup_script_behavior != "" ? post_startup_script_config.value.post_startup_script_behavior : null
        }
      }

      dynamic "colab_image" {
        for_each = software_config.value.colab_image != null ? [software_config.value.colab_image] : []
        content {
          release_name = colab_image.value.release_name != "" ? colab_image.value.release_name : null
        }
      }
    }
  }

  depends_on = [google_project_service.aiplatform_api]
}
