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
  description = "Specification for the Zalando Postgres Operator deployment"
  type = object({
    # Kubernetes namespace to install the operator
    namespace = string

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, false)

    # The container specifications for the operator deployment
    container = object({

      # The CPU and memory resources allocated to the operator container
      resources = object({

        # The resource limits for the container
        # Specify the maximum amount of CPU and memory that the container can use
        limits = object({

          # The amount of CPU allocated (e.g., "1000m" for 1 CPU core)
          cpu = string

          # The amount of memory allocated (e.g., "1Gi" for 1 gibibyte)
          memory = string
        })

        # The resource requests for the container
        # Specify the minimum amount of CPU and memory that the container is guaranteed
        requests = object({

          # The amount of CPU allocated (e.g., "50m" for 0.05 CPU cores)
          cpu = string

          # The amount of memory allocated (e.g., "100Mi" for 100 mebibytes)
          memory = string
        })
      })
    })

    # Optional: Backup configuration for all databases managed by this operator. The
    # module composes the WAL-G target from bucket + object_prefix and appends the
    # per-cluster/per-version suffix.
    backup_config = optional(object({

      # Bucket that stores backups for every database on this cluster. Planton resolves
      # the value-or-ref to a plain bucket name before tfvars.
      bucket = string

      # Base path under the bucket; the module appends the per-cluster/per-version suffix.
      object_prefix = optional(string)

      # Cron schedule for base backups (e.g., "0 2 * * *" for 2 AM daily).
      schedule = string

      # Enable WAL-G for backups (default: true).
      enable_wal_g_backup = optional(bool, true)

      # Enable WAL-G for restores (default: true).
      enable_wal_g_restore = optional(bool, true)

      # Enable WAL-G for clone operations (default: true).
      enable_clone_wal_g_restore = optional(bool, true)

      # Credentials WAL-G uses to read and write backups.
      credentials = object({
        cloudflare_account_id = string
        access_key_id         = string
        secret_access_key     = string
      })
    }))
  })
}
