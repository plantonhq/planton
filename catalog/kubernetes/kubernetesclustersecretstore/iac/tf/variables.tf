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
  description = "KubernetesClusterSecretStore specification"
  type = object({
    secrets_namespace = string
    config = object({
      aws = optional(object({
        service                   = optional(string)
        region                    = string
        role                      = optional(string, "")
        prefix                    = optional(string, "")
        service_account_name      = optional(string, "")
        service_account_namespace = optional(string, "")
        access_key_id             = optional(string, "")
        secret_access_key         = optional(string, "")
      }))
      gcp_secret_manager = optional(object({
        project_id                = string
        location                  = optional(string, "")
        service_account_name      = optional(string, "")
        service_account_namespace = optional(string, "")
        service_account_key_json  = optional(string, "")
      }))
      azure_key_vault = optional(object({
        vault_url                 = string
        tenant_id                 = optional(string, "")
        auth_type                 = optional(string)
        identity_id               = optional(string, "")
        service_account_name      = optional(string, "")
        service_account_namespace = optional(string, "")
        client_id                 = optional(string, "")
        client_secret             = optional(string, "")
      }))
      vault = optional(object({
        server    = string
        path      = optional(string, "")
        version   = optional(string)
        namespace = optional(string, "")
        ca_bundle = optional(string, "")
        token = optional(object({
          token = string
        }))
        app_role = optional(object({
          path      = optional(string, "")
          role_id   = string
          secret_id = string
        }))
        kubernetes = optional(object({
          mount_path           = optional(string, "")
          role                 = string
          service_account_name = optional(string, "")
        }))
      }))
      kubernetes = optional(object({
        server_url           = optional(string, "")
        ca_bundle            = optional(string, "")
        remote_namespace     = optional(string, "")
        token                = optional(string, "")
        service_account_name = optional(string, "")
      }))
      fake = optional(object({
        data = list(object({
          key     = string
          value   = string
          version = optional(string, "")
        }))
      }))
      controller_class = optional(string, "")
      refresh_interval = optional(string, "")
      retry = optional(object({
        max_retries    = optional(number)
        retry_interval = optional(string, "")
      }))
    })
    conditions = optional(list(object({
      namespaces               = optional(list(string), [])
      namespace_label_selector = optional(map(string), {})
      namespace_regexes        = optional(list(string), [])
    })), [])
  })
}
