variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "KubernetesRayCluster specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    ray_version = string
    image = optional(string, "")
    head = object({
      resources = object({
        limits = optional(object({
          cpu = optional(string, "")
          memory = optional(string, "")
        }))
        requests = optional(object({
          cpu = optional(string, "")
          memory = optional(string, "")
        }))
      })
      schedule_tasks_on_head = optional(bool)
      ray_start_params = optional(map(string), {})
      scheduling = optional(object({
        node_selector = optional(map(string), {})
        tolerations = optional(list(object({
          key = optional(string, "")
          operator = optional(string, "")
          value = optional(string, "")
          effect = optional(string, "")
          toleration_seconds = optional(number)
        })), [])
        priority_class_name = optional(string, "")
      }))
    })
    worker_groups = optional(list(object({
      name = string
      replicas = optional(number)
      min_replicas = optional(number)
      max_replicas = optional(number)
      resources = object({
        limits = optional(object({
          cpu = optional(string, "")
          memory = optional(string, "")
        }))
        requests = optional(object({
          cpu = optional(string, "")
          memory = optional(string, "")
        }))
      })
      ray_start_params = optional(map(string), {})
      scheduling = optional(object({
        node_selector = optional(map(string), {})
        tolerations = optional(list(object({
          key = optional(string, "")
          operator = optional(string, "")
          value = optional(string, "")
          effect = optional(string, "")
          toleration_seconds = optional(number)
        })), [])
        priority_class_name = optional(string, "")
      }))
      extra_resource_limits = optional(map(string), {})
    })), [])
    autoscaling = optional(object({
      enabled = optional(bool, false)
      idle_timeout_seconds = optional(number)
      upscaling_mode = optional(string, "")
      resources = optional(object({
        limits = optional(object({
          cpu = optional(string, "")
          memory = optional(string, "")
        }))
        requests = optional(object({
          cpu = optional(string, "")
          memory = optional(string, "")
        }))
      }))
    }))
    gcs_fault_tolerance = optional(object({
      enabled = optional(bool, false)
      redis_address = optional(string, "")
      redis_password_secret = optional(object({
        name = string
        key = string
      }))
      external_storage_namespace = optional(string, "")
    }))
    auth = optional(object({
      mode = optional(string)
      existing_token_secret_name = optional(string, "")
    }))
    suspend = optional(bool, false)
  })
}