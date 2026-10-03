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
  description = "KubernetesLocust specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    image = optional(object({
      repository = optional(string)
      tag = optional(string)
    }))
    image_pull_secrets = optional(list(string), [])
    load_test = object({
      name = optional(string)
      inline = optional(object({
        locustfile_content = string
        lib_files = optional(map(string), {})
      }))
      existing_config_maps = optional(object({
        locustfile_config_map = string
        locustfile_name = optional(string)
        lib_config_map = optional(string, "")
      }))
      target_host = optional(string, "")
      pip_packages = optional(list(string), [])
      pip_requirements_config_map = optional(string, "")
      environment = optional(map(string), {})
      env_from_secrets = optional(list(string), [])
      env_from_secret_keys = optional(list(object({
        secret_name = string
        keys = list(string)
      })), [])
      tags = optional(list(string), [])
      exclude_tags = optional(list(string), [])
      headless = optional(bool, false)
    })
    master = optional(object({
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
      log_level = optional(string)
      pdb_enabled = optional(bool, false)
      scheduling = optional(object({
        node_selector = optional(map(string), {})
        tolerations = optional(list(object({
          key = optional(string, "")
          operator = optional(string, "")
          value = optional(string, "")
          effect = optional(string, "")
          toleration_seconds = optional(number)
        })), [])
      }))
    }))
    workers = optional(object({
      replicas = optional(number)
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
      log_level = optional(string)
      pdb_enabled = optional(bool, false)
      scheduling = optional(object({
        node_selector = optional(map(string), {})
        tolerations = optional(list(object({
          key = optional(string, "")
          operator = optional(string, "")
          value = optional(string, "")
          effect = optional(string, "")
          toleration_seconds = optional(number)
        })), [])
      }))
      hpa = optional(object({
        min_replicas = optional(number)
        max_replicas = optional(number, 0)
        target_cpu_utilization_percent = optional(number)
      }))
      keda = optional(object({
        min_replicas = optional(number)
        max_replicas = optional(number, 0)
        target_users_per_worker = optional(number)
        polling_interval_seconds = optional(number)
        cooldown_period_seconds = optional(number)
        custom_triggers = optional(string, "")
      }))
    }))
    web_ui_auth = optional(object({
      enabled = optional(bool)
      username = optional(string)
    }))
    service = optional(object({
      type = optional(string)
      annotations = optional(map(string), {})
    }))
    helm_values = optional(string, "")
  })
}