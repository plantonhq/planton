variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "KubernetesDaemonSet specification"
  type = object({
    # Kubernetes namespace to deploy the DaemonSet
    namespace = string

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, false)

    # The container specifications for the DaemonSet
    container = object({
      # The main application container specifications
      app = object({
        # The container image configuration
        image = object({
          # The repository of the image (e.g., "fluent/fluentd")
          repo = string
          # The tag of the image (e.g., "v1.16")
          tag = string
          # The name of the image pull secret for private registries
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
          # A map of secret environment variable names to their values
          # Each secret can be provided either as a literal string value (value)
          # or as a reference to an existing Kubernetes Secret (secret_ref)
          secrets = optional(map(object({
            # A literal string value for the secret (for development/testing)
            value = optional(string)
            # A reference to a key within a Kubernetes Secret (recommended for production)
            secret_ref = optional(object({
              # The namespace of the Kubernetes Secret (optional, defaults to DaemonSet namespace)
              namespace = optional(string)
              # The name of the Kubernetes Secret
              name = string
              # The key within the Secret that contains the value
              key = string
            }))
          })))
        }))

        # A list of ports to be configured for the container
        ports = optional(list(object({
          # The name of the port (e.g., "metrics", "health")
          name = string
          # The port number on the container
          container_port = number
          # The network protocol (TCP, UDP, SCTP)
          network_protocol = string
          # Host port to expose (use with caution)
          host_port = optional(number)
        })), [])

        # Volume mounts for the container
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

          # HostPath volume source (common for DaemonSets)
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

        # Command to run in the container (overrides ENTRYPOINT)
        command = optional(list(string), [])

        # Arguments to pass to the command (overrides CMD)
        args = optional(list(string), [])

        # Security context for the container
        security_context = optional(object({
          privileged             = optional(bool, false)
          run_as_user            = optional(number)
          run_as_group           = optional(number)
          run_as_non_root        = optional(bool)
          read_only_root_filesystem = optional(bool, false)
          capabilities = optional(object({
            add  = optional(list(string), [])
            drop = optional(list(string), [])
          }))
        }))
      })

      # Sidecar containers (optional)
      sidecars = optional(list(object({
        name  = string
        image = string
        ports = optional(list(object({
          name           = string
          container_port = number
          protocol       = string
        })), [])
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
        env = optional(list(object({
          name  = string
          value = string
        })), [])
      })), [])
    })

    # Node selector for constraining pods to specific nodes
    node_selector = optional(map(string), {})

    # Tolerations for scheduling on tainted nodes
    tolerations = optional(list(object({
      key                = optional(string)
      operator           = optional(string)
      value              = optional(string)
      effect             = optional(string)
      toleration_seconds = optional(number)
    })), [])

    # Update strategy for the DaemonSet
    update_strategy = optional(object({
      type = string
      rolling_update = optional(object({
        max_unavailable = optional(string)
        max_surge       = optional(string)
      }))
    }))

    # Minimum ready seconds before pod is considered available
    min_ready_seconds = optional(number, 0)

    # Whether to create a ServiceAccount
    create_service_account = optional(bool, false)

    # Name of the ServiceAccount
    service_account_name = optional(string)

    # ConfigMaps to create alongside the DaemonSet
    config_maps = optional(map(string), {})

    # RBAC configuration
    rbac = optional(object({
      cluster_rules = optional(list(object({
        api_groups     = list(string)
        resources      = list(string)
        verbs          = list(string)
        resource_names = optional(list(string), [])
      })), [])
      namespace_rules = optional(list(object({
        api_groups     = list(string)
        resources      = list(string)
        verbs          = list(string)
        resource_names = optional(list(string), [])
      })), [])
    }))
  })
}

