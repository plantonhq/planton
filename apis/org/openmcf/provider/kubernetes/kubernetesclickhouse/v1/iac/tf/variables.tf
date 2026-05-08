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
  description = "ClickHouse cluster specification"
  type = object({
    # Kubernetes namespace to install ClickHouse
    namespace = string

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, true)

    # The name of the ClickHouse cluster (used for ClickHouseInstallation resource name)
    # Defaults to metadata.name if not specified
    cluster_name = optional(string)

    # The ClickHouse version to deploy (e.g., "24.8")
    # If not specified, defaults to a recent stable version
    version = optional(string)

    # The container specifications for the ClickHouse deployment.
    container = object({

      # The number of ClickHouse pods to deploy (for standalone mode).
      # Ignored if clustering is enabled (uses shard_count * replica_count instead)
      replicas = number

      # The CPU and memory resources allocated to each ClickHouse container.
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

      # A flag to enable or disable data persistence for ClickHouse.
      # When enabled, data is persisted to a storage volume.
      persistence_enabled = bool

      # The size of the persistent volume attached to each ClickHouse pod (e.g., "50Gi").
      disk_size = string
    })

    # The ingress configuration for the ClickHouse deployment.
    ingress = optional(object({
      is_enabled = bool
      dns_domain = string
    }))

    # The cluster configuration for ClickHouse sharding and replication.
    cluster = optional(object({
      is_enabled    = bool
      shard_count   = number
      replica_count = number
    }))

    # ZooKeeper configuration for cluster coordination (optional)
    # If not specified, the operator automatically manages ZooKeeper for clustered deployments
    zookeeper = optional(object({
      # Flag to use external ZooKeeper instead of operator-managed
      use_external = bool
      # List of external ZooKeeper nodes in format "host:port"
      nodes = optional(list(string))
    }))
  })
}
