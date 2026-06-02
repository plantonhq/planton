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
    # Kubernetes namespace to install the component.
    namespace = string

    # Flag to indicate if the namespace should be created
    create_namespace = bool

    # Istio version to deploy (full patch, e.g. "1.26.8"). Drives the Helm chart
    # version. If null, falls back to the module default. Istio supports only
    # sequential single-minor upgrades; pin an existing mesh's current version here
    # before redeploying to avoid an unsupported multi-minor jump.
    version = optional(string)

    # The container specifications for the Istio control plane deployment.
    container = object({

      # The CPU and memory resources allocated to the Istio control plane container.
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
    })

  })
}