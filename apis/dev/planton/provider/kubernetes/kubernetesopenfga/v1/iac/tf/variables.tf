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
    # Kubernetes namespace to install the component
    namespace = string

    # Flag to indicate if the namespace should be created
    create_namespace = bool

    # The container specifications for the OpenFGA deployment.
    container = object({

      # The number of OpenFGA replicas to deploy. This determines the level of concurrency and availability.
      replicas = number

      # The CPU and memory resources allocated to the OpenFGA container.
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
    })

    # The ingress configuration for the OpenFGA deployment.
    ingress = object({

      # A flag to enable or disable ingress.
      enabled = bool

      # The full hostname for external access (e.g., "openfga.example.com").
      hostname = string
    })

    # The data store configuration for OpenFGA.
    # This specifies the backend database engine and connection details.
    datastore = object({

      # Specifies the type of data store engine to use.
      # Allowed values are "mysql" for MySQL database and "postgres" for PostgreSQL database.
      engine = string

      # The hostname or endpoint of the database server.
      host = string

      # The port number of the database server.
      # Defaults to 5432 for PostgreSQL and 3306 for MySQL.
      port = optional(number)

      # The name of the database to connect to.
      database = string

      # The username for authenticating to the database.
      username = string

      # The password for authenticating to the database.
      # Can be provided either as a plain string value or as a reference to an existing Kubernetes Secret.
      password = object({
        string_value = optional(string)
        secret_ref = optional(object({
          namespace = optional(string)
          name = string
          key = string
        }))
      })

      # Whether to use SSL/TLS connection to the database.
      is_secure = optional(bool, false)
    })
  })
}