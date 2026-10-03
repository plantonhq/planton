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
  description = "KubernetesFlinkDeployment specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    flink_version = string
    image = optional(string, "")
    job = optional(object({
      jar_uri = string
      entry_class = optional(string, "")
      args = optional(list(string), [])
      parallelism = optional(number)
      state = optional(string)
      upgrade_mode = optional(string)
      initial_savepoint_path = optional(string, "")
      allow_non_restored_state = optional(bool, false)
      savepoint_trigger_nonce = optional(number, 0)
    }))
    job_manager = optional(object({
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
      replicas = optional(number)
    }))
    task_manager = optional(object({
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
      replicas = optional(number)
    }))
    flink_configuration = optional(map(string), {})
    state = optional(object({
      checkpoints_dir = optional(string, "")
      savepoints_dir = optional(string, "")
      high_availability = optional(object({
        enabled = optional(bool, false)
        storage_dir = optional(string, "")
      }))
      s3 = optional(object({
        endpoint = string
        path_style_access = optional(bool)
        access_key_secret = object({
          name = string
          key = string
        })
        secret_key_secret = object({
          name = string
          key = string
        })
        # Exact jar under /opt/flink/opt — without this the S3 plugin
        # stays disabled and every s3:// path fails (verified live: TF
        # twin silently dropped the field when it was absent here).
        builtin_plugin_jar = optional(string, "")
      }))
    }))
    mode = optional(string)
    service_account = optional(string)
    log_configuration = optional(map(string), {})
    image_pull_secrets = optional(list(string), [])
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
    restart_nonce = optional(number, 0)
  })
}