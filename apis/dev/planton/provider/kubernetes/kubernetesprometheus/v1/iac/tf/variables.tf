variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name = string,
    id = optional(string),
    org = optional(string),
    env = optional(string),
    labels = optional(map(string)),
    tags = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}


variable "spec" {
  description = "spec"
  type = object({
    # Kubernetes namespace to install Prometheus.
    namespace = string

    # Flag to indicate if the namespace should be created
    create_namespace = bool

    # The container specifications for the Prometheus deployment.
    container = object({

      # The number of Prometheus pods to deploy.
      replicas = number

      # The CPU and memory resources allocated to the Prometheus container.
      resources = object({

        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = object({

          # The amount of CPU allocated (e.g., "500m" for 0.5 CPU cores).
          cpu = string

          # The amount of memory allocated (e.g., "256Mi" for 256 mebibytes).
          memory = string
        })

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = object({

          # The amount of CPU allocated (e.g., "500m" for 0.5 CPU cores).
          cpu = string

          # The amount of memory allocated (e.g., "256Mi" for 256 mebibytes).
          memory = string
        })
      })

      # A flag to enable or disable data persistence for Prometheus.
      # When enabled, in-memory data is persisted to a storage volume, allowing data to survive pod restarts.
      # The backup data from the persistent volume is restored into Prometheus memory between pod restarts.
      # Defaults to `false`.
      persistence_enabled = bool

      # Description for disk_size
      disk_size = string
    })

    # The ingress configuration for the Prometheus deployment.
    ingress = object({

      # A flag to enable or disable ingress.
      is_enabled = bool

      # The dns domain.
      dns_domain = string
    })
  })
}