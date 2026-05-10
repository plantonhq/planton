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
    # Kubernetes namespace
    namespace = string

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, true)

    # A list of Kafka topics to be created in the Kafka cluster.
    kafka_topics = optional(list(object({

      # The name of the Kafka topic.
      # Must be between 1 and 249 characters in length.
      # The name must start and end with an alphanumeric character, can contain alphanumeric characters, '.', '_', and '-'.
      # Must not contain '..' or non-ASCII characters.
      name = string

      # The number of partitions for the topic.
      # Recommended default is 1.
      partitions = number

      # The number of replicas for the topic.
      # Recommended default is 1.
      replicas = number

      # Additional configuration for the Kafka topic.
      # If not provided, default values will be set.
      # For example, the default `delete.policy` is `delete`, but it can be set to `compact`.
      config = optional(map(string))
    })), [])

    # The specifications for the Kafka broker containers.
    broker_container = object({

      # The number of Kafka brokers to deploy.
      # Defaults to 1 if the client sets the value to 0.
      # Recommended default value is 1.
      replicas = number

      # The CPU and memory resources allocated to the Kafka broker containers.
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

      # The size of the disk to be attached to each broker instance (e.g., "30Gi").
      # A default value is set if not provided by the client.
      disk_size = string
    })

    # The specifications for the Zookeeper containers.
    zookeeper_container = object({

      # The number of Zookeeper container replicas.
      # Zookeeper requires at least 3 replicas for high availability (HA) mode.
      # Zookeeper uses the Raft consensus algorithm; refer to https://raft.github.io/ for more information on how replica
      # count affects availability.
      replicas = number

      # The CPU and memory resources allocated to the Zookeeper containers.
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

      # The size of the disk to be attached to each Zookeeper instance (e.g., "30Gi").
      # A default value is set if not provided by the client.
      disk_size = string
    })

    # The specifications for the Schema Registry containers.
    schema_registry_container = optional(object({

      # A flag to control whether the Schema Registry is created for the Kafka deployment.
      # Defaults to `false`.
      is_enabled = bool

      # The number of Schema Registry replicas.
      # Recommended default value is "1".
      # This value has no effect if `is_enabled` is set to `false`.
      replicas = number

      # The CPU and memory resources allocated to the Schema Registry containers.
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
    }))

    # The ingress configuration for the Kafka deployment.
    ingress = optional(object({

      # A flag to enable or disable ingress.
      is_enabled = bool

      # The dns domain.
      dns_domain = string
    }))

    # A flag to toggle the deployment of the Kafka UI component.
    is_deploy_kafka_ui = optional(bool, false)
  })
}