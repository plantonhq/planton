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
  description = "KubernetesArgocd specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    chart_version = optional(string)
    admin_enabled = optional(bool)
    domain = optional(string, "")
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
    }))
    server = optional(object({
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
      autoscaling = optional(object({
        enabled = optional(bool, false)
        min_replicas = optional(number)
        max_replicas = optional(number)
      }))
      insecure = optional(bool, false)
    }))
    repo_server = optional(object({
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
      autoscaling = optional(object({
        enabled = optional(bool, false)
        min_replicas = optional(number)
        max_replicas = optional(number)
      }))
    }))
    application_set = optional(object({
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
    }))
    notifications = optional(object({
      enabled = optional(bool)
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
    dex = optional(object({
      enabled = optional(bool)
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
    commit_server = optional(object({
      enabled = optional(bool)
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
    redis = optional(object({
      bundled = optional(object({
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
      ha = optional(object({
        replicas = optional(number)
      }))
      external = optional(object({
        host = string
        port = optional(number)
        credentials_secret_name = optional(string, "")
      }))
    }))
    sso = optional(object({
      oidc = optional(object({
        name = string
        issuer = string
        client_id = string
        client_secret_secret = optional(object({
          name = string
          key = string
        }))
      }))
      dex_config = optional(string, "")
    }))
    rbac = optional(object({
      policy_default = optional(string, "")
      policy_csv = optional(string, "")
      scopes = optional(string)
    }))
    exec_enabled = optional(bool, false)
    reconciliation_timeout = optional(string, "")
    repositories = optional(list(object({
      name = string
      url = string
      type = optional(string)
    })), [])
    crds = optional(object({
      install = optional(bool)
      keep = optional(bool)
    }))
    service_monitors_enabled = optional(bool, false)
    image = optional(object({
      repository = optional(string, "")
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