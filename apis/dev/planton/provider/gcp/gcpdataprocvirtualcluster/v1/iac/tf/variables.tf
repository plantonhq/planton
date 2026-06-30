variable "spec" {
  description = "GcpDataprocVirtualClusterSpec"
  type = object({
    project_id = object({
      value = string
    })
    region       = string
    cluster_name = optional(string, "")
    gke_cluster_target = object({
      value = string
    })
    kubernetes_namespace = optional(object({
      value = string
    }), null)
    staging_bucket = optional(object({
      value = string
    }), null)
    software_config = object({
      component_version = map(string)
      properties        = optional(map(string), {})
    })
    node_pool_targets = list(object({
      node_pool = object({
        value = string
      })
      roles = list(string)
      node_pool_config = optional(object({
        locations        = optional(list(string), [])
        machine_type     = optional(string, "")
        local_ssd_count  = optional(number, 0)
        min_cpu_platform = optional(string, "")
        preemptible      = optional(bool, false)
        spot             = optional(bool, false)
        autoscaling = optional(object({
          min_node_count = number
          max_node_count = number
        }), null)
      }), null)
    }))
    auxiliary_services_config = optional(object({
      metastore_service              = optional(string, "")
      spark_history_server_cluster   = optional(string, "")
    }), null)
  })
}

variable "metadata" {
  description = "Resource metadata"
  type = object({
    name = string
    org  = optional(string, "")
    env = optional(object({
      id = string
    }), null)
    id = optional(string, "")
  })
  default = {
    name = ""
  }
}
