resource "google_dataproc_cluster" "cluster" {
  name    = var.spec.cluster_name
  region  = var.spec.region
  project = var.spec.project_id
  labels  = local.labels

  graceful_decommission_timeout = var.spec.graceful_decommission_timeout != "" ? var.spec.graceful_decommission_timeout : null

  dynamic "cluster_config" {
    for_each = var.spec.cluster_config != null ? [var.spec.cluster_config] : []
    content {
      staging_bucket = cluster_config.value.staging_bucket != "" ? cluster_config.value.staging_bucket : null
      temp_bucket    = cluster_config.value.temp_bucket != "" ? cluster_config.value.temp_bucket : null

      # GCE cluster config (networking, service account, zone, tags, metadata)
      dynamic "gce_cluster_config" {
        for_each = cluster_config.value.gce_config != null ? [cluster_config.value.gce_config] : []
        content {
          network            = gce_cluster_config.value.network != "" ? gce_cluster_config.value.network : null
          subnetwork         = gce_cluster_config.value.subnetwork != "" ? gce_cluster_config.value.subnetwork : null
          service_account    = gce_cluster_config.value.service_account != "" ? gce_cluster_config.value.service_account : null
          service_account_scopes = length(gce_cluster_config.value.service_account_scopes) > 0 ? gce_cluster_config.value.service_account_scopes : null
          zone               = gce_cluster_config.value.zone != "" ? gce_cluster_config.value.zone : null
          internal_ip_only   = gce_cluster_config.value.internal_ip_only
          tags               = length(gce_cluster_config.value.tags) > 0 ? gce_cluster_config.value.tags : null
          metadata           = length(gce_cluster_config.value.metadata) > 0 ? gce_cluster_config.value.metadata : null
        }
      }

      # Master config
      dynamic "master_config" {
        for_each = cluster_config.value.master_config != null ? [cluster_config.value.master_config] : []
        content {
          num_instances    = master_config.value.num_instances > 0 ? master_config.value.num_instances : null
          machine_type     = master_config.value.machine_type != "" ? master_config.value.machine_type : null
          min_cpu_platform = master_config.value.min_cpu_platform != "" ? master_config.value.min_cpu_platform : null
          image_uri        = master_config.value.image_uri != "" ? master_config.value.image_uri : null

          dynamic "disk_config" {
            for_each = master_config.value.disk_config != null ? [master_config.value.disk_config] : []
            content {
              boot_disk_size_gb = disk_config.value.boot_disk_size_gb > 0 ? disk_config.value.boot_disk_size_gb : null
              boot_disk_type    = disk_config.value.boot_disk_type != "" ? disk_config.value.boot_disk_type : null
              num_local_ssds    = disk_config.value.num_local_ssds > 0 ? disk_config.value.num_local_ssds : null
            }
          }

          dynamic "accelerators" {
            for_each = master_config.value.accelerators
            content {
              accelerator_type  = accelerators.value.accelerator_type
              accelerator_count = accelerators.value.accelerator_count
            }
          }
        }
      }

      # Worker config
      dynamic "worker_config" {
        for_each = cluster_config.value.worker_config != null ? [cluster_config.value.worker_config] : []
        content {
          num_instances     = worker_config.value.num_instances > 0 ? worker_config.value.num_instances : null
          machine_type      = worker_config.value.machine_type != "" ? worker_config.value.machine_type : null
          min_cpu_platform  = worker_config.value.min_cpu_platform != "" ? worker_config.value.min_cpu_platform : null
          image_uri         = worker_config.value.image_uri != "" ? worker_config.value.image_uri : null
          min_num_instances = worker_config.value.min_num_instances > 0 ? worker_config.value.min_num_instances : null

          dynamic "disk_config" {
            for_each = worker_config.value.disk_config != null ? [worker_config.value.disk_config] : []
            content {
              boot_disk_size_gb = disk_config.value.boot_disk_size_gb > 0 ? disk_config.value.boot_disk_size_gb : null
              boot_disk_type    = disk_config.value.boot_disk_type != "" ? disk_config.value.boot_disk_type : null
              num_local_ssds    = disk_config.value.num_local_ssds > 0 ? disk_config.value.num_local_ssds : null
            }
          }

          dynamic "accelerators" {
            for_each = worker_config.value.accelerators
            content {
              accelerator_type  = accelerators.value.accelerator_type
              accelerator_count = accelerators.value.accelerator_count
            }
          }
        }
      }

      # Secondary worker config (preemptible/spot)
      dynamic "preemptible_worker_config" {
        for_each = cluster_config.value.secondary_worker_config != null ? [cluster_config.value.secondary_worker_config] : []
        content {
          num_instances  = preemptible_worker_config.value.num_instances > 0 ? preemptible_worker_config.value.num_instances : null
          preemptibility = preemptible_worker_config.value.preemptibility != "" ? preemptible_worker_config.value.preemptibility : null

          dynamic "disk_config" {
            for_each = preemptible_worker_config.value.disk_config != null ? [preemptible_worker_config.value.disk_config] : []
            content {
              boot_disk_size_gb = disk_config.value.boot_disk_size_gb > 0 ? disk_config.value.boot_disk_size_gb : null
              boot_disk_type    = disk_config.value.boot_disk_type != "" ? disk_config.value.boot_disk_type : null
              num_local_ssds    = disk_config.value.num_local_ssds > 0 ? disk_config.value.num_local_ssds : null
            }
          }
        }
      }

      # Software config
      dynamic "software_config" {
        for_each = cluster_config.value.software_config != null ? [cluster_config.value.software_config] : []
        content {
          image_version       = software_config.value.image_version != "" ? software_config.value.image_version : null
          optional_components = length(software_config.value.optional_components) > 0 ? software_config.value.optional_components : null
          override_properties = length(software_config.value.properties) > 0 ? software_config.value.properties : null
        }
      }

      # Initialization actions
      dynamic "initialization_action" {
        for_each = cluster_config.value.initialization_actions
        content {
          script      = initialization_action.value.script
          timeout_sec = initialization_action.value.timeout_sec > 0 ? initialization_action.value.timeout_sec : null
        }
      }

      # Autoscaling policy
      dynamic "autoscaling_config" {
        for_each = cluster_config.value.autoscaling_policy_uri != "" ? [1] : []
        content {
          policy_uri = cluster_config.value.autoscaling_policy_uri
        }
      }

      # CMEK encryption
      dynamic "encryption_config" {
        for_each = cluster_config.value.encryption_kms_key_name != "" ? [1] : []
        content {
          kms_key_name = cluster_config.value.encryption_kms_key_name
        }
      }

      # Component Gateway
      dynamic "endpoint_config" {
        for_each = cluster_config.value.endpoint_config != null ? [cluster_config.value.endpoint_config] : []
        content {
          enable_http_port_access = endpoint_config.value.enable_http_port_access
        }
      }

      # Lifecycle config (auto-delete, idle shutdown)
      dynamic "lifecycle_config" {
        for_each = cluster_config.value.lifecycle_config != null ? [cluster_config.value.lifecycle_config] : []
        content {
          idle_delete_ttl  = lifecycle_config.value.idle_delete_ttl != "" ? lifecycle_config.value.idle_delete_ttl : null
          auto_delete_time = lifecycle_config.value.auto_delete_time != "" ? lifecycle_config.value.auto_delete_time : null
        }
      }
    }
  }
}
