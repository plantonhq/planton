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
  description = "NatsKubernetes specification"
  type = object({
    # Kubernetes namespace to install NATS
    namespace = string

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, true)

    # Server container configuration
    server_container = object({
      # Number of NATS replicas
      replicas = number

      # CPU and memory resources
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

      # PVC size for JetStream file store (e.g., "10Gi")
      disk_size = string
    })

    # Disable JetStream persistence
    disable_jet_stream = optional(bool, false)

    # Authentication configuration
    auth = optional(object({
      enabled = bool
      scheme  = string # "bearer_token" or "basic_auth"
      no_auth_user = optional(object({
        enabled          = bool
        publish_subjects = list(string)
      }))
    }))

    # TLS encryption
    tls_enabled = optional(bool, false)

    # Ingress configuration for external access
    ingress = optional(object({
      enabled  = bool
      hostname = string
    }))

    # Toggle to deploy nats-box utility pod
    disable_nats_box = optional(bool, false)

    # NACK JetStream controller configuration (opt-in)
    # When enabled, deploys the NACK controller alongside NATS for managing
    # streams and consumers via Kubernetes CRDs.
    nack_controller = optional(object({
      # Enable the NACK JetStream controller
      enabled = bool
      # Enable control-loop mode for KeyValue/ObjectStore support
      enable_control_loop = optional(bool, false)
      # NACK Helm chart version (default: "0.31.1")
      helm_chart_version = optional(string, "0.31.1")
      # NACK app version / GitHub release tag (default: "0.21.1")
      # Used for fetching CRDs - differs from chart version
      app_version = optional(string, "0.21.1")
    }))

    # JetStream streams to create (requires nack_controller.enabled = true)
    streams = optional(list(object({
      # Unique stream name (1-255 chars, alphanumeric with - _ .)
      name = string
      # List of subjects to consume (supports wildcards like "orders.*", "events.>")
      subjects = list(string)
      # Storage backend: "file" (persistent) or "memory" (ephemeral)
      storage = optional(string, "memory")
      # Number of replicas (1-5, odd recommended for quorum)
      replicas = optional(number, 1)
      # Retention policy: "limits", "interest", or "workqueue"
      retention = optional(string, "limits")
      # Maximum age of messages (e.g., "24h", "7d"), empty for unlimited
      max_age = optional(string, "")
      # Maximum size in bytes (-1 for unlimited)
      max_bytes = optional(number, -1)
      # Maximum number of messages (-1 for unlimited)
      max_msgs = optional(number, -1)
      # Maximum message size in bytes (-1 for unlimited)
      max_msg_size = optional(number, -1)
      # Maximum number of consumers (-1 for unlimited)
      max_consumers = optional(number, -1)
      # Discard policy when limits reached: "old" or "new_msgs"
      # Note: Use "new_msgs" instead of "new" because "new" is a reserved keyword in Java.
      # The module converts "new_msgs" to "new" when sending to NACK CRDs.
      discard = optional(string, "old")
      # Description of the stream
      description = optional(string, "")
      # Consumers for this stream
      consumers = optional(list(object({
        # Durable name of the consumer (unique within stream)
        durable_name = string
        # Delivery policy: "all", "last", or "new_msgs"
        # Note: Use "new_msgs" instead of "new" because "new" is a reserved keyword in Java.
        # The module converts "new_msgs" to "new" when sending to NACK CRDs.
        deliver_policy = optional(string, "all")
        # Acknowledgment policy: "none", "all", or "explicit"
        ack_policy = optional(string, "none")
        # Filter subject (supports wildcards)
        filter_subject = optional(string, "")
        # Deliver subject for push consumers (empty for pull)
        deliver_subject = optional(string, "")
        # Queue group name for load balancing
        deliver_group = optional(string, "")
        # Maximum unacknowledged messages
        max_ack_pending = optional(number, 0)
        # Maximum delivery attempts (-1 for unlimited)
        max_deliver = optional(number, -1)
        # Time to wait for acknowledgment (e.g., "30s", "1m")
        ack_wait = optional(string, "")
        # Replay policy: "original" or "instant"
        replay_policy = optional(string, "instant")
        # Description of the consumer
        description = optional(string, "")
      })), [])
    })), [])

    # NATS Helm chart version (default: "2.12.3")
    nats_helm_chart_version = optional(string, "2.12.3")
  })
}
