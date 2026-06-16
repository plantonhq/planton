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
    # Kubernetes namespace to install PostgreSQL.
    namespace = string

    # flag to indicate if the namespace should be created.
    # Default MUST be false to match the proto3 default (spec.proto field 3) and the
    # Pulumi module: an unset create_namespace serializes away (proto omits false), so a
    # tofu default of true diverged -- the module created a namespace from spec.namespace
    # even when callers (e.g. an InfraChart with a dedicated KubernetesNamespace) never
    # asked it to, which on an unresolved/empty namespace produced an invalid (empty) name.
    create_namespace = optional(bool, false)

    # The container specifications for the PostgreSQL deployment.
    container = object({

      # The number of replicas of PostgreSQL pods.
      replicas = number

      # The CPU and memory resources allocated to the PostgreSQL container.
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

      # The storage size to allocate for each PostgreSQL instance (e.g., "1Gi").
      # A default value is set if the client does not provide a value.
      disk_size = string
    })

    # The ingress configuration for the PostgreSQL deployment.
    ingress = optional(object({

      # A flag to enable or disable ingress.
      enabled = bool

      # The full hostname for external access.
      hostname = string
    }))

    # Per-database backup configuration. When set, these settings override the
    # operator-level backup configuration. Mirrors KubernetesPostgresBackupConfig
    # in spec.proto and the Pulumi module's backup_config.go / restore_config.go.
    backup_config = optional(object({

      # Explicitly enable/disable backups for this database (USE_WALG_BACKUP).
      enable_backup = optional(bool)

      # Custom S3/R2 prefix path for this database's backups (WALG_S3_PREFIX).
      s3_prefix = optional(string)

      # Custom backup schedule in cron format (BACKUP_SCHEDULE).
      backup_schedule = optional(string)

      # Number of base backups to retain (BACKUP_NUM_TO_RETAIN).
      backup_retain_count = optional(number)

      # Dedicated R2 credentials + endpoint for this database's backups, independent
      # of any operator-level S3 config. When set, the module creates a Secret from
      # these credentials and references it via secretKeyRef in the pod env.
      r2_config = optional(object({
        cloudflare_account_id = string
        access_key_id         = string
        secret_access_key     = string
      }))

      # Disaster-recovery restore configuration (Zalando spec.standby + STANDBY_* env).
      restore_config = optional(object({

        # When true, the database bootstraps as a read-only standby from backups.
        enabled = optional(bool, false)

        # S3/R2 bucket holding the backup source.
        bucket_name = optional(string)

        # S3 path to the backup directory (without s3:// prefix or bucket name).
        s3_path = optional(string)

        # R2/S3 credentials used during standby bootstrap.
        r2_config = optional(object({
          cloudflare_account_id = string
          access_key_id         = string
          secret_access_key     = string
        }))
      }))
    }))

    # List of databases to create.
    # Each database has a name and an optional owner role.
    # The operator will create these databases during cluster initialization.
    # If not specified, only the default "postgres" database will be available.
    # Note: Owner roles must be declared in the 'users' field.
    databases = optional(list(object({
      name       = string
      owner_role = optional(string, "")
    })), [])

    # List of PostgreSQL users/roles to create.
    # Users must be declared here before being used as database owners.
    # Each user has a name and optional flags (e.g., ["createdb"], ["superuser"]).
    # Empty flags array means standard user with login privileges only.
    users = optional(list(object({
      name  = string
      flags = optional(list(string), [])
    })), [])
  })
}
