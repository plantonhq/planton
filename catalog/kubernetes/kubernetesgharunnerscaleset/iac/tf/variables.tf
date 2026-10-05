# Typed mirror of KubernetesGhaRunnerScaleSetSpec (spec.proto). The spec
# arrives from the proto->tfvars converter in snake_case;
# StringValueOrRef fields arrive resolved to their literal string values.
# The auth oneof arrives with exactly one arm populated (spec CEL).

variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "KubernetesGhaRunnerScaleSet specification"
  type = object({
    namespace         = string
    create_namespace  = optional(bool, false)
    chart_version     = optional(string, "0.14.2")
    github_config_url = string
    auth = object({
      existing_secret_name = optional(string, "")
      pat = optional(object({
        token = string
      }))
      github_app = optional(object({
        app_id          = string
        installation_id = string
        private_key     = string
      }))
    })
    runner_scale_set_name = optional(string, "")
    runner_group          = optional(string, "")
    min_runners           = optional(number)
    max_runners           = optional(number)
    container_mode = optional(object({
      mode = string
      kubernetes_work_volume = optional(object({
        storage_class = string
        size          = string
      }))
    }))
    runner = optional(object({
      image = optional(string, "")
      resources = optional(object({
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
    }))
    proxy = optional(object({
      http = optional(object({
        url                    = string
        credential_secret_name = optional(string, "")
      }))
      https = optional(object({
        url                    = string
        credential_secret_name = optional(string, "")
      }))
      no_proxy = optional(list(string), [])
    }))
    github_server_tls = optional(object({
      config_map_name   = string
      key               = optional(string, "ca.crt")
      runner_mount_path = optional(string, "")
    }))
    controller_service_account = optional(object({
      namespace = string
      name      = string
    }))
    helm_values = optional(string, "")
  })
}
