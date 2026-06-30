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
  description = "Specification for KubernetesElasticOperator"
  type = object({

    # The Kubernetes cluster to install this operator on.
    target_cluster = optional(object({
      cluster_name = string
      cluster_kind = optional(number)
    }))

    # Kubernetes namespace to install the operator.
    namespace = string

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, false)

    # The container specifications for the ECK operator.
    container = object({

      # The CPU and memory resources allocated to the ECK operator container.
      resources = object({

        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = object({

          # The amount of CPU allocated (e.g., "1000m" for 1 CPU core).
          cpu = string

          # The amount of memory allocated (e.g., "1Gi" for 1 gibibyte).
          memory = string
        })

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = object({

          # The amount of CPU allocated (e.g., "50m" for 0.05 CPU cores).
          cpu = string

          # The amount of memory allocated (e.g., "100Mi" for 100 mebibytes).
          memory = string
        })
      })
    })
  })
}