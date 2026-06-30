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
  description = "Temporal Kubernetes deployment specification"
  type = object({
    # Kubernetes namespace to install Temporal
    namespace = string

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, false)

    # Database configuration
    database = object({
      # Selected database backend: cassandra, postgresql, or mysql
      backend = string

      # External database configuration (optional)
      external_database = optional(object({
        # Hostname for external database
        host = string

        # Port for external database
        port = number

        # Username for database
        username = string

        # Password for database
        # Can be provided either as a plain string value or as a reference to an existing Kubernetes Secret.
        # Using a secret reference is recommended for production deployments.
        # Example with string value: { string_value = "my-password" }
        # Example with secret ref: { secret_ref = { name = "db-credentials", key = "password" } }
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

      # Primary database or keyspace name (default: "temporal")
      database_name = optional(string, "temporal")

      # Visibility database or keyspace name (default: "temporal_visibility")
      visibility_name = optional(string, "temporal_visibility")

      # Disables automatic schema creation
      disable_auto_schema_setup = optional(bool, false)
    })

    # Disables Temporal web UI
    disable_web_ui = optional(bool, false)

    # Enables embedded Elasticsearch for Temporal
    # This is ignored if external Elasticsearch is set
    enable_embedded_elasticsearch = optional(bool, false)

    # Enables monitoring stack for Temporal
    # Enabling this will deploy Prometheus and Grafana
    enable_monitoring_stack = optional(bool, false)

    # Number of Cassandra nodes to be deployed
    # This is only honored when the backend is Cassandra and no external database is provided
    cassandra_replicas = optional(number, 1)

    # Ingress configuration for the Temporal deployment
    ingress = optional(object({
      # Frontend (gRPC + HTTP) ingress configuration
      frontend = optional(object({
        # Flag to enable or disable frontend ingress
        enabled = optional(bool, false)

        # The full hostname for gRPC access via LoadBalancer (e.g., "temporal-frontend-grpc.example.com")
        grpc_hostname = optional(string, "")

        # The full hostname for HTTP access via Gateway API (e.g., "temporal-frontend-http.example.com")
        http_hostname = optional(string, "")
      }))

      # Web UI ingress configuration
      web_ui = optional(object({
        # Flag to enable or disable web UI ingress
        enabled = optional(bool, false)

        # The full hostname for HTTP access via Gateway API (e.g., "temporal-ui.example.com")
        hostname = optional(string, "")
      }))
    }))

    # External Elasticsearch configuration
    external_elasticsearch = optional(object({
      # The host address of the existing Elasticsearch cluster
      host = string

      # The port for the existing Elasticsearch cluster
      port = number

      # Optional username, if the external cluster requires auth
      user = optional(string, "")

      # Optional password, if the external cluster requires auth
      # Can be provided either as a plain string value or as a reference to an existing Kubernetes Secret.
      # Example with string value: { string_value = "my-password" }
      # Example with secret ref: { secret_ref = { name = "es-credentials", key = "password" } }
      password = optional(object({
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
      }))
    }))

    # Version of the Temporal Helm chart to deploy (e.g., "0.62.0")
    # If not specified, the default version will be used
    version = optional(string, "0.62.0")

    # Dynamic configuration values for Temporal server runtime behavior
    # These settings control workflow execution limits without requiring server restart
    dynamic_config = optional(object({
      # History limits - control total workflow history size

      # Maximum size in bytes for workflow execution history
      # Default: 52428800 (50 MB). Increase for workflows with large payloads.
      history_size_limit_error = optional(number)

      # Maximum number of events in workflow execution history
      # Default: 51200 events. Increase for workflows with many activities/signals.
      history_count_limit_error = optional(number)

      # Warning threshold for history size in bytes
      # Default: 10485760 (10 MB)
      history_size_limit_warn = optional(number)

      # Warning threshold for history event count
      # Default: 10240 events
      history_count_limit_warn = optional(number)

      # Blob size limits - control individual payload sizes (markers, signals, activity I/O)
      # This is different from history limits which control total workflow history size

      # Maximum size in bytes for a single blob/payload (marker details, signal data, activity I/O)
      # Default: 2097152 (2 MB). Increase for workflows that send large payloads like IaC diffs.
      blob_size_limit_error = optional(number)

      # Warning threshold for blob/payload size in bytes
      # Default: 524288 (512 KB)
      blob_size_limit_warn = optional(number)
    }))

    # Number of history shards for the Temporal cluster
    # This is an IMMUTABLE setting that must be decided at cluster creation time.
    # Higher values enable better parallelism. Default: 512
    # WARNING: Cannot be changed after initial deployment without data migration.
    num_history_shards = optional(number)

    # Per-service replica and resource configuration for Temporal services
    services = optional(object({
      # Frontend service configuration (API gateway for gRPC/HTTP requests)
      frontend = optional(object({
        # Number of replicas
        replicas = optional(number)
        # Container resources (CPU and memory)
        resources = optional(object({
          limits = optional(object({
            cpu    = optional(string)
            memory = optional(string)
          }))
          requests = optional(object({
            cpu    = optional(string)
            memory = optional(string)
          }))
        }))
      }))

      # History service configuration (manages workflow state, most resource-intensive)
      history = optional(object({
        # Number of replicas
        replicas = optional(number)
        # Container resources (CPU and memory)
        resources = optional(object({
          limits = optional(object({
            cpu    = optional(string)
            memory = optional(string)
          }))
          requests = optional(object({
            cpu    = optional(string)
            memory = optional(string)
          }))
        }))
      }))

      # Matching service configuration (task queue management and dispatch)
      matching = optional(object({
        # Number of replicas
        replicas = optional(number)
        # Container resources (CPU and memory)
        resources = optional(object({
          limits = optional(object({
            cpu    = optional(string)
            memory = optional(string)
          }))
          requests = optional(object({
            cpu    = optional(string)
            memory = optional(string)
          }))
        }))
      }))

      # Worker service configuration (internal Temporal system workflows)
      worker = optional(object({
        # Number of replicas
        replicas = optional(number)
        # Container resources (CPU and memory)
        resources = optional(object({
          limits = optional(object({
            cpu    = optional(string)
            memory = optional(string)
          }))
          requests = optional(object({
            cpu    = optional(string)
            memory = optional(string)
          }))
        }))
      }))
    }))
  })
}

