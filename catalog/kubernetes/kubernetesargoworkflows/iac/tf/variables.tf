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
  description = "KubernetesArgoWorkflows specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    chart_version = optional(string)
    controller = optional(object({
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
      workflow_namespaces = optional(list(string), [])
      instance_id = optional(string, "")
      parallelism = optional(number)
      namespace_parallelism = optional(number)
    }))
    server = optional(object({
      enabled = optional(bool)
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
      auth_modes = optional(list(string), [])
      secure = optional(bool, false)
      base_href = optional(string, "")
    }))
    workflow_service_account = optional(string)
    artifact_repository = optional(object({
      archive_logs = optional(bool, false)
      s3 = optional(object({
        bucket = string
        endpoint = optional(string, "")
        region = optional(string, "")
        insecure = optional(bool, false)
        use_ambient_credentials = optional(bool, false)
        credentials_secret = optional(object({
          secret_name = string
          access_key_id_key = optional(string)
          secret_access_key_key = optional(string)
        }))
      }))
      gcs = optional(object({
        bucket = string
        credentials_secret_name = optional(string, "")
      }))
      azure = optional(object({
        endpoint = string
        container = string
        credentials_secret_name = optional(string, "")
      }))
    }))
    archive = optional(object({
      engine = string
      host = string
      port = optional(number)
      database = string
      credentials_secret = object({
        name = string
        username_key = optional(string)
        password_key = optional(string)
      })
      ssl_mode = optional(string, "")
    }))
    retention_policy = optional(object({
      completed = optional(number)
      failed = optional(number)
      errored = optional(number)
    }))
    crds = optional(object({
      install = optional(bool)
      keep = optional(bool)
      full_schema = optional(bool)
      base_url = optional(string, "")
    }))
    service_monitor_enabled = optional(bool, false)
    image = optional(object({
      registry = optional(string, "")
      tag = optional(string, "")
      pull_secret_name = optional(string, "")
    }))
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
    helm_values = optional(string, "")
  })
}