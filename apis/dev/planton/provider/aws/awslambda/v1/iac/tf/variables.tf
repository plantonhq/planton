variable "metadata" {
  description = "metadata for all resource objects on planton"
  type = object({

    # name of the resource
    name = string

    # id of the resource
    id = string

    # id of the organization to which the api-resource belongs to
    org = string

    # environment to which the resource belongs to
    env = string

    # labels for the resource
    labels = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # annotations for the resource
    annotations = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # tags for the resource
    tags = list(string)
  })
}

variable "spec" {
  description = "Specification for Deployment Component"
  type = object({

    # The AWS region where the resource will be created.
    region = string

    # Description for function_name
    function_name = string

    # Description for description
    description = string

    # Description for role_arn
    role_arn = object({

      # Description for value
      value = string

      # Description for value_from
      value_from = object({

        # Description for kind
        kind = string

        # Description for env
        env = string

        # Description for name
        name = string

        # Description for field_path
        field_path = string
      })
    })

    # Description for runtime
    runtime = string

    # Description for handler
    handler = string

    # Description for memory_mb
    memory_mb = number

    # Description for timeout_seconds
    timeout_seconds = number

    # Description for reserved_concurrency
    reserved_concurrency = number

    # Description for environment
    environment = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # Description for subnets
    subnets = list(object({

      # Description for value
      value = string

      # Description for value_from
      value_from = object({

        # Description for kind
        kind = string

        # Description for env
        env = string

        # Description for name
        name = string

        # Description for field_path
        field_path = string
      })
    }))

    # Description for security_groups
    security_groups = list(object({

      # Description for value
      value = string

      # Description for value_from
      value_from = object({

        # Description for kind
        kind = string

        # Description for env
        env = string

        # Description for name
        name = string

        # Description for field_path
        field_path = string
      })
    }))

    # Description for architecture
    architecture = string

    # Description for layer_arns
    layer_arns = list(object({

      # Description for value
      value = string

      # Description for value_from
      value_from = object({

        # Description for kind
        kind = string

        # Description for env
        env = string

        # Description for name
        name = string

        # Description for field_path
        field_path = string
      })
    }))

    # Description for kms_key_arn
    kms_key_arn = object({

      # Description for value
      value = string

      # Description for value_from
      value_from = object({

        # Description for kind
        kind = string

        # Description for env
        env = string

        # Description for name
        name = string

        # Description for field_path
        field_path = string
      })
    })

    # Description for code_source_type
    code_source_type = string

    # Description for s3
    s3 = object({

      # Description for bucket
      bucket = string

      # Description for key
      key = string

      # Description for object_version
      object_version = string
    })

    # Description for image_uri
    image_uri = string
  })
}