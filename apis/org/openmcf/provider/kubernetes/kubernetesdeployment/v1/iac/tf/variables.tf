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
    # Kubernetes namespace to install the microservice.
    namespace = string

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, true)

    # The version of the microservice being deployed.
    # This is usually either "main" (the default git branch name) or "review-<id>" where <id> is the merge request number.
    # It must be between 1 and 20 characters and can only contain lowercase letters, numbers, and hyphens.
    version = string

    # The container specifications for the microservice deployment.
    # This includes configurations for the main application container and any sidecar containers.
    container = object({

      # The main application container specifications.
      app = object({

        # The container image to be used for the application.
        # This value is computed during creation but can be updated.
        # It is derived by combining the Docker repository of the artifact store configured for the environment and the code project path.
        # The `pull_secret_name` is the name of the image pull secret to be configured in the Kubernetes Deployment resource.
        # It is determined by looking up the `container_image_artifact_store_id` from the environment where the microservice is deployed.
        image = object({

          # The repository of the image (e.g., "gcr.io/project/image").
          repo = string

          # The tag of the image (e.g., "latest" or "1.0.0").
          tag = string

          # The name of the image pull secret for private image repositories.
          pull_secret_name = optional(string)
        })

        # The CPU and memory resources allocated to the application container.
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

        # The environment variables and secrets for the application container.
        env = optional(object({
          # Environment variables (non-sensitive). Each entry supports multiple value sources.
          variables = optional(list(object({
            name  = string
            value = optional(string)
            value_from = optional(object({
              kind       = optional(string)
              env        = optional(string)
              name       = string
              field_path = optional(string)
            }))
            config_map_key_ref = optional(object({
              name     = string
              key      = string
              optional = optional(bool, false)
            }))
            field_ref = optional(object({
              api_version = optional(string)
              field_path  = string
            }))
            resource_field_ref = optional(object({
              container_name = optional(string)
              resource       = string
              divisor        = optional(string)
            }))
          })), [])
          # Secret environment variables (sensitive).
          secrets = optional(list(object({
            name  = string
            value = optional(string)
            secret_ref = optional(object({
              namespace = optional(string)
              name      = string
              key       = string
              optional  = optional(bool, false)
            }))
            value_from = optional(object({
              kind       = optional(string)
              env        = optional(string)
              name       = string
              field_path = optional(string)
            }))
          })), [])
          # Bulk import of environment variables from ConfigMaps or Secrets.
          env_from = optional(list(object({
            prefix = optional(string)
            config_map_ref = optional(object({
              name     = string
              optional = optional(bool, false)
            }))
            secret_ref = optional(object({
              name     = string
              optional = optional(bool, false)
            }))
          })), [])
        }))

        # A list of ports to be configured for the application container.
        ports = list(object({

          # The name of the port (e.g., "http", "grpc").
          # The name must only contain lowercase alphanumeric characters and hyphens.
          # Port names must also start and end with an alphanumeric character.
          # For example, "123-abc" and "web" are valid, but "123_abc" and "-web" are not.
          name = string

          # The port number on the container.
          container_port = number

          # The network protocol used by the port (e.g., "TCP", "UDP", "SCTP").
          # Must be one of "TCP", "UDP", or "SCTP".
          network_protocol = string

          # The application protocol for the microservice (e.g., "http").
          # This field is used for setting up the name of the service port in Kubernetes.
          # It is used during microservice deployment and is relevant for deployment and stateful set pod managers.
          # Refer to: https://kubernetes.io/docs/concepts/services-networking/service/#application-protocol
          app_protocol = string

          # The port number on the Kubernetes service.
          service_port = number

          # A flag indicating whether this port should be exposed via ingress.
          is_ingress_port = optional(bool, false)
        }))

        # Volume mounts for the application container.
        # Supports mounting ConfigMaps, Secrets, HostPaths, EmptyDirs, and PVCs.
        volume_mounts = optional(list(object({
          # Name of the volume mount. Must be unique within the container.
          name = string

          # Path within the container at which the volume should be mounted.
          mount_path = string

          # Whether the volume should be mounted read-only.
          read_only = optional(bool, false)

          # Path within the volume from which the container's volume should be mounted.
          sub_path = optional(string)

          # ConfigMap volume source.
          config_map = optional(object({
            name         = string
            key          = optional(string)
            path         = optional(string)
            default_mode = optional(number)
          }))

          # Secret volume source.
          secret = optional(object({
            name         = string
            key          = optional(string)
            path         = optional(string)
            default_mode = optional(number)
          }))

          # HostPath volume source.
          host_path = optional(object({
            path = string
            type = optional(string)
          }))

          # EmptyDir volume source.
          empty_dir = optional(object({
            medium     = optional(string)
            size_limit = optional(string)
          }))

          # PVC volume source.
          pvc = optional(object({
            claim_name = string
            read_only  = optional(bool, false)
          }))
        })), [])

        # Command to run in the container (overrides the container image's ENTRYPOINT).
        command = optional(list(string), [])

        # Arguments to pass to the command (overrides the container image's CMD).
        args = optional(list(string), [])
      })

      # A list of sidecar containers to be deployed alongside the main application container.
      sidecars = optional(list(object({

        # The name of the container.
        name = string

        # The container image to be used.
        image = string

        # A list of ports exposed by the container.
        ports = list(object({

          # The name of the port.
          name = string

          # The port number on the container.
          # **Note:** The attribute names must use camel case to marshal into the Kubernetes Container spec.
          container_port = number

          # The protocol used by the port (e.g., "TCP" or "UDP").
          protocol = string
        }))

        # Resource specifications for the container, including CPU and memory limits and requests.
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

        # A list of environment variables to be set in the container.
        env = list(object({

          # The name of the environment variable.
          name = string

          # The value of the environment variable.
          value = string
        }))
      })))
    })

    # The ingress configuration for the microservice.
    # This defines how the microservice can be accessed externally.
    ingress = optional(object({

      # A flag to enable or disable ingress.
      is_enabled = bool

      # The dns domain.
      dns_domain = string
    }))

    # ConfigMaps to create alongside the deployment.
    # Key is the ConfigMap name, value is the content.
    # These ConfigMaps can be referenced in volume mounts.
    config_maps = optional(map(string), {})

    # The availability configuration for the microservice.
    # This includes settings for minimum replicas and autoscaling options.
    availability = optional(object({

      # The minimum number of pod replicas to maintain.
      min_replicas = number

      # The configuration for horizontal pod autoscaling.
      horizontal_pod_autoscaling = optional(object({

        # A flag to enable or disable horizontal pod autoscaling.
        is_enabled = bool

        # The target CPU utilization percentage to trigger autoscaling (e.g., 60.0).
        target_cpu_utilization_percent = optional(number)

        # The target memory utilization to trigger autoscaling (e.g., "1Gi").
        target_memory_utilization = optional(string)
      }))
    }))
  })
}
