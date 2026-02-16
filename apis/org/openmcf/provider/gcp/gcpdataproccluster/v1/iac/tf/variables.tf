variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key_base64 = optional(string, "")
  })
  default = {}
}

variable "metadata" {
  description = "Resource metadata"
  type = object({
    name = string
    org  = optional(string, "")
    env  = optional(string, "")
    id   = optional(string, "")
  })
}

variable "spec" {
  description = "GcpDataprocCluster specification"
  type = object({
    project_id   = string
    region       = string
    cluster_name = string
    graceful_decommission_timeout = optional(string, "")
    cluster_config = optional(object({
      staging_bucket = optional(string, "")
      temp_bucket    = optional(string, "")
      gce_config = optional(object({
        network                = optional(string, "")
        subnetwork             = optional(string, "")
        service_account        = optional(string, "")
        service_account_scopes = optional(list(string), [])
        zone                   = optional(string, "")
        internal_ip_only       = optional(bool, false)
        tags                   = optional(list(string), [])
        metadata               = optional(map(string), {})
      }), null)
      master_config = optional(object({
        num_instances    = optional(number, 0)
        machine_type     = optional(string, "")
        min_cpu_platform = optional(string, "")
        image_uri        = optional(string, "")
        disk_config = optional(object({
          boot_disk_size_gb = optional(number, 0)
          boot_disk_type    = optional(string, "")
          num_local_ssds    = optional(number, 0)
        }), null)
        accelerators = optional(list(object({
          accelerator_type  = string
          accelerator_count = number
        })), [])
      }), null)
      worker_config = optional(object({
        num_instances     = optional(number, 0)
        machine_type      = optional(string, "")
        min_cpu_platform  = optional(string, "")
        image_uri         = optional(string, "")
        min_num_instances = optional(number, 0)
        disk_config = optional(object({
          boot_disk_size_gb = optional(number, 0)
          boot_disk_type    = optional(string, "")
          num_local_ssds    = optional(number, 0)
        }), null)
        accelerators = optional(list(object({
          accelerator_type  = string
          accelerator_count = number
        })), [])
      }), null)
      secondary_worker_config = optional(object({
        num_instances  = optional(number, 0)
        preemptibility = optional(string, "")
        disk_config = optional(object({
          boot_disk_size_gb = optional(number, 0)
          boot_disk_type    = optional(string, "")
          num_local_ssds    = optional(number, 0)
        }), null)
      }), null)
      software_config = optional(object({
        image_version       = optional(string, "")
        optional_components = optional(list(string), [])
        properties          = optional(map(string), {})
      }), null)
      initialization_actions = optional(list(object({
        script      = string
        timeout_sec = optional(number, 0)
      })), [])
      autoscaling_policy_uri    = optional(string, "")
      encryption_kms_key_name   = optional(string, "")
      endpoint_config = optional(object({
        enable_http_port_access = optional(bool, false)
      }), null)
      lifecycle_config = optional(object({
        idle_delete_ttl  = optional(string, "")
        auto_delete_time = optional(string, "")
      }), null)
    }), null)
  })
}
