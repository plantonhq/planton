variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "KubernetesOpenBao specification"
  type = object({
    # Kubernetes namespace to install OpenBao.
    namespace = string

    # Flag to indicate if the namespace should be created.
    create_namespace = bool

    # Helm chart version override (optional).
    helm_chart_version = optional(string)

    # Server container specifications.
    server_container = object({
      # Number of OpenBao server replicas.
      replicas = number

      # CPU and memory resources for the server container.
      resources = object({
        limits = object({
          cpu    = string
          memory = string
        })
        requests = object({
          cpu    = string
          memory = string
        })
      })

      # Size of the persistent volume for data storage.
      data_storage_size = string
    })

    # High Availability configuration.
    high_availability = optional(object({
      # Enable HA mode with Raft integrated storage.
      enabled = bool
      # Number of HA replicas.
      replicas = optional(number)
    }))

    # Ingress configuration.
    ingress = optional(object({
      # Enable ingress for external access.
      enabled = optional(bool)
      # Hostname for external access.
      hostname = optional(string)
    }))

    # Enable OpenBao UI.
    ui_enabled = optional(bool)

    # Agent Injector configuration.
    injector = optional(object({
      # Enable the Agent Injector.
      enabled = bool
      # Number of injector replicas.
      replicas = optional(number)
    }))

    # Enable TLS encryption.
    tls_enabled = optional(bool)

    # Auto-unseal configuration. Exactly one seal type should be specified.
    auto_unseal = optional(object({
      gcp_kms = optional(object({
        project                            = string
        region                             = string
        key_ring                           = string
        crypto_key                         = string
        workload_identity_service_account  = optional(string)
      }))
      aws_kms = optional(object({
        region                  = string
        kms_key_id              = string
        credentials_secret_name = optional(string)
      }))
      azure_key_vault = optional(object({
        vault_name              = string
        key_name                = string
        tenant_id               = string
        credentials_secret_name = optional(string)
      }))
      transit = optional(object({
        address           = string
        key_name          = string
        mount_path        = optional(string)
        token_secret_name = string
      }))
    }))
  })
}
