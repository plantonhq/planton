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
  description = "spec"
  type = object({
    # Kubernetes namespace to install the component
    namespace = string

    # flag to indicate if the namespace should be created
    create_namespace = optional(bool, false)

    # The specifications for the MongoDB container deployment.
    container = object({

      # The number of MongoDB pods to deploy.
      replicas = number

      # The CPU and memory resources allocated to the MongoDB container.
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

      # A flag to enable or disable data persistence for MongoDB.
      # When enabled, in-memory data is persisted to a storage volume, allowing data to survive pod restarts.
      persistence_enabled = bool

      # Description for disk_size
      disk_size = string
    })

    # The ingress configuration for the MongoDB deployment.
    ingress = optional(object({

      # A flag to enable or disable ingress.
      enabled = bool

      # The full hostname for MongoDB access.
      hostname = string
    }))

    # A map of key-value pairs that provide additional customization options for the Helm chart used
    # to deploy MongoDB on Kubernetes. These values allow for further refinement of the deployment,
    # such as customizing resource limits, setting environment variables, or specifying version tags.
    # For detailed information on the available options, refer to the Helm chart documentation at:
    # https://artifacthub.io/packages/helm/bitnami/mongodb
    helm_values = optional(map(string))
  })
}
