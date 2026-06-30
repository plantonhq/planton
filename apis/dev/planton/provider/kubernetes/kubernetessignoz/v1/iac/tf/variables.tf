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
  description = "Specification for SigNoz Kubernetes deployment"
  type = object({
    # Kubernetes namespace to install the component
    namespace = string

    # Flag to indicate if the namespace should be created
    create_namespace = bool

    # The container specifications for the main SigNoz binary (UI, API server, Ruler, Alertmanager).
    signoz_container = object({

      # The number of SigNoz pods to deploy.
      replicas = number

      # The CPU and memory resources allocated to the SigNoz container.
      resources = object({

        # The resource limits for the container.
        limits = object({
          cpu    = string
          memory = string
        })

        # The resource requests for the container.
        requests = object({
          cpu    = string
          memory = string
        })
      })

      # Optional container image configuration
      image = optional(object({
        repo = string
        tag  = string
      }))
    })

    # The container specifications for the OpenTelemetry Collector (data ingestion gateway).
    otel_collector_container = object({

      # The number of OTel Collector pods to deploy.
      replicas = number

      # The CPU and memory resources allocated to the OTel Collector container.
      resources = object({

        # The resource limits for the container.
        limits = object({
          cpu    = string
          memory = string
        })

        # The resource requests for the container.
        requests = object({
          cpu    = string
          memory = string
        })
      })

      # Optional container image configuration
      image = optional(object({
        repo = string
        tag  = string
      }))
    })

    # The database configuration for SigNoz, supporting both self-managed and external ClickHouse.
    database = object({

      # Flag to enable using an external ClickHouse database.
      # When false (default), SigNoz will deploy and manage its own ClickHouse instance.
      # When true, the external_database field must be configured.
      is_external = bool

      # External ClickHouse database connection details.
      # This field is required when is_external is true and ignored when false.
      external_database = optional(object({

        # The hostname or endpoint of the external ClickHouse instance.
        host = string

        # The HTTP port for ClickHouse (default is 8123).
        http_port = number

        # The TCP port for ClickHouse native protocol (default is 9000).
        tcp_port = number

        # The name of the distributed cluster in ClickHouse configuration.
        cluster_name = string

        # Whether to use secure (TLS) connection to ClickHouse.
        is_secure = bool

        # The username for authenticating to ClickHouse.
        username = string

        # The password for authenticating to ClickHouse.
        # Can be provided either as a plain string value or as a reference to an existing Kubernetes Secret.
        # Using a secret reference is recommended for production deployments.
        # Example with string value: { string_value = "my-password" }
        # Example with secret ref: { secret_ref = { name = "clickhouse-secret", key = "password" } }
        password = object({
          # Plain text password value (not recommended for production)
          string_value = optional(string)
          # Reference to an existing Kubernetes Secret
          secret_ref = optional(object({
            # The namespace of the Kubernetes Secret (optional - defaults to deployment namespace)
            namespace = optional(string)
            # The name of the Kubernetes Secret
            name = string
            # The key within the Kubernetes Secret that contains the password
            key = string
          }))
        })
      }))

      # Self-managed ClickHouse configuration.
      # This field is used when is_external is false and configures the in-cluster ClickHouse deployment.
      managed_database = optional(object({

        # The container specifications for ClickHouse.
        container = object({

          # The number of ClickHouse pods to deploy.
          replicas = number

          # The CPU and memory resources allocated to the ClickHouse container.
          resources = object({

            # The resource limits for the container.
            limits = object({
              cpu    = string
              memory = string
            })

            # The resource requests for the container.
            requests = object({
              cpu    = string
              memory = string
            })
          })

          # Optional container image configuration
          image = optional(object({
            repo = string
            tag  = string
          }))

          # Flag to enable or disable data persistence for ClickHouse.
          # When enabled, data is persisted to a storage volume, allowing data to survive pod restarts.
          persistence_enabled = bool

          # The size of the persistent volume attached to each ClickHouse pod (e.g., "20Gi").
          # This attribute is ignored when persistence is not enabled.
          disk_size = string
        })

        # The cluster configuration for ClickHouse (sharding and replication).
        cluster = optional(object({

          # Flag to enable or disable clustering mode for ClickHouse.
          # When enabled, ClickHouse will be deployed in a distributed cluster configuration.
          is_enabled = bool

          # The number of shards in the ClickHouse cluster.
          # Sharding distributes data across multiple nodes for horizontal scaling.
          shard_count = number

          # The number of replicas for each shard.
          # Replication provides data redundancy and high availability.
          replica_count = number
        }))

        # The Zookeeper configuration (required for distributed ClickHouse clusters).
        zookeeper = optional(object({

          # Flag to enable or disable Zookeeper deployment.
          # This must be true if ClickHouse clustering is enabled.
          is_enabled = bool

          # The container specifications for Zookeeper.
          container = optional(object({

            # The number of Zookeeper pods to deploy.
            # For production, this should be an odd number (3 or 5) to maintain quorum.
            replicas = number

            # The CPU and memory resources allocated to the Zookeeper container.
            resources = object({

              # The resource limits for the container.
              limits = object({
                cpu    = string
                memory = string
              })

              # The resource requests for the container.
              requests = object({
                cpu    = string
                memory = string
              })
            })

            # Optional container image configuration
            image = optional(object({
              repo = string
              tag  = string
            }))

            # The size of the persistent volume attached to each Zookeeper pod (e.g., "8Gi").
            disk_size = string
          }))
        }))
      }))
    })

    # The ingress configuration for SigNoz UI and OpenTelemetry Collector endpoints.
    ingress = optional(object({

      # Ingress configuration for SigNoz UI and API.
      ui = optional(object({

        # Flag to enable or disable ingress for the UI.
        enabled = bool

        # The full hostname for external access to the UI (e.g., "signoz.example.com").
        hostname = string
      }))

      # Ingress configuration for OpenTelemetry Collector data ingestion endpoint.
      otel_collector = optional(object({

        # Flag to enable or disable ingress for the OpenTelemetry Collector.
        enabled = bool

        # The full hostname for external access to the collector (e.g., "signoz-ingest.example.com").
        hostname = string
      }))
    }))

    # A map of key-value pairs that provide additional customization options for the SigNoz Helm chart.
    # These values allow for further refinement of the deployment, such as setting environment variables,
    # configuring alerting integrations, or customizing retention policies.
    # For detailed information on available options, refer to the Helm chart documentation at:
    # https://github.com/SigNoz/charts
    helm_values = optional(map(string))
  })
}

