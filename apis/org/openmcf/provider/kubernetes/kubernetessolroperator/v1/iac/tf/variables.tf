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
  description = "Specification for Kubernetes Solr Operator deployment"
  type = object({
    # Target Kubernetes cluster (orchestrator-only, not included in generated tfvars)
    target_cluster_name = optional(string)

    # Kubernetes namespace where operator will be deployed
    namespace = optional(string, "solr-operator-system")

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, false)

    # The version of the Apache Solr Operator to deploy
    # https://github.com/apache/solr-operator/releases
    operator_version = optional(string, "v0.9.1")

    # The container specifications for the Apache Solr Operator deployment.
    container = object({

      # The CPU and memory resources allocated to the operator container.
      resources = optional(object({

        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = optional(object({

          # The amount of CPU allocated (e.g., "1000m" for 1 CPU core).
          cpu = optional(string, "1000m")

          # The amount of memory allocated (e.g., "1Gi" for 1 gibibyte).
          memory = optional(string, "1Gi")
        }))

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = optional(object({

          # The amount of CPU allocated (e.g., "50m" for 0.05 CPU cores).
          cpu = optional(string, "50m")

          # The amount of memory allocated (e.g., "100Mi" for 100 mebibytes).
          memory = optional(string, "100Mi")
        }))
      }))
    })
  })
}