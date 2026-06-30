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
    # Kubernetes namespace to install the operator.
    namespace = string

    # Flag to indicate if the namespace should be created
    create_namespace = bool

    # The container specifications for the Neo4j deployment.
    container = object({

      # The CPU and memory resources allocated to the Neo4j container.
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

      # Flag to enable or disable data persistence
      persistence_enabled = optional(bool, false)

      # Size of the persistent volume if persistence is enabled
      disk_size = optional(string, "1Gi")
    })

    # Optional memory configuration for Neo4j
    memory_config = optional(object({
      # Maximum Java heap size
      heap_max = optional(string, "")

      # Page cache size for on-disk data
      page_cache = optional(string, "")
    }), {
      heap_max   = ""
      page_cache = ""
    })

    # The ingress configuration for the Neo4j deployment.
    ingress = object({

      # A flag to enable or disable ingress.
      enabled = bool

      # The full hostname for external access.
      hostname = string
    })
  })
}