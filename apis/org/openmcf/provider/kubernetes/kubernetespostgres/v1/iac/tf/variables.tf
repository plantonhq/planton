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
    # Kubernetes namespace to install PostgreSQL.
    namespace = string

    # flag to indicate if the namespace should be created
    create_namespace = optional(bool, true)

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
