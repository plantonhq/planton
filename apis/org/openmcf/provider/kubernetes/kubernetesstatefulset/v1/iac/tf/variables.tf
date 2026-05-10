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
  description = "Specification for the KubernetesStatefulSet"
  type = object({
    # Kubernetes namespace to install the statefulset
    namespace = string

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, true)

    # The container specifications for the statefulset
    container = object({
      # The main application container specifications
      app = object({
        # The container image to be used for the application
        image = object({
          repo             = string
          tag              = string
          pull_secret_name = optional(string)
        })

        # The CPU and memory resources allocated to the container
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

        # The environment variables and secrets for the container
        env = optional(object({
          # A map of environment variable names to their values.
          # Each variable can be provided either as a direct string value (value)
          # or as a reference to another OpenMCF resource's field (value_from).
          # The orchestrator resolves value_from references and populates .value before invoking Terraform.
          variables = optional(map(string))
          secrets = optional(map(object({
            value = optional(string)
            secret_ref = optional(object({
              namespace = optional(string)
              name      = string
              key       = string
            }))
          })))
        }))

        # A list of ports to be configured for the container
        ports = list(object({
          name             = string
          container_port   = number
          network_protocol = string
          app_protocol     = string
          service_port     = number
          is_ingress_port  = optional(bool, false)
        }))

        # Volume mounts for the container
        # Supports mounting ConfigMaps, Secrets, HostPaths, EmptyDirs, and PVCs
        volume_mounts = optional(list(object({
          name       = string
          mount_path = string
          read_only  = optional(bool, false)
          sub_path   = optional(string)

          # ConfigMap volume source
          config_map = optional(object({
            name         = string
            key          = optional(string)
            path         = optional(string)
            default_mode = optional(number)
          }))

          # Secret volume source
          secret = optional(object({
            name         = string
            key          = optional(string)
            path         = optional(string)
            default_mode = optional(number)
          }))

          # HostPath volume source
          host_path = optional(object({
            path = string
            type = optional(string)
          }))

          # EmptyDir volume source
          empty_dir = optional(object({
            medium     = optional(string)
            size_limit = optional(string)
          }))

          # PVC volume source
          pvc = optional(object({
            claim_name = string
            read_only  = optional(bool, false)
          }))
        })), [])

        # Optional command to run instead of the image's default entrypoint
        command = optional(list(string))

        # Optional arguments to pass to the command
        args = optional(list(string))
      })

      # A list of sidecar containers
      sidecars = optional(list(object({
        name  = string
        image = string
        ports = list(object({
          name           = string
          container_port = number
          protocol       = string
        }))
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
        env = list(object({
          name  = string
          value = string
        }))
      })))
    })

    # The ingress configuration for the statefulset
    ingress = optional(object({
      enabled  = bool
      hostname = optional(string)
    }))

    # The availability configuration for the statefulset
    availability = optional(object({
      replicas = number

      pod_disruption_budget = optional(object({
        enabled         = bool
        min_available   = optional(string)
        max_unavailable = optional(string)
      }))
    }))

    # Persistent volume claims for the statefulset
    volume_claim_templates = optional(list(object({
      name          = string
      storage_class = optional(string)
      size          = string
      access_modes  = optional(list(string))
    })))

    # Pod management policy for the statefulset
    pod_management_policy = optional(string)

    # ConfigMaps to create alongside the StatefulSet
    # Key is the ConfigMap name, value is the content
    config_maps = optional(map(string), {})
  })
}
