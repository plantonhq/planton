variable "metadata" {
  description = "metadata for all resource objects on planton-cloud"
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

    # Description for instance_name
    instance_name = string

    # Description for ami_id
    ami_id = string

    # Description for instance_type
    instance_type = string

    # Description for subnet_id
    subnet_id = object({

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

    # Description for security_group_ids
    security_group_ids = list(object({

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

    # Description for connection_method
    connection_method = string

    # Description for iam_instance_profile_arn
    iam_instance_profile_arn = object({

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

    # Description for key_name
    key_name = string

    # Description for root_volume_size_gb
    root_volume_size_gb = number

    # Description for tags
    tags = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # Description for user_data
    user_data = string

    # Description for ebs_optimized
    ebs_optimized = bool

    # Description for disable_api_termination
    disable_api_termination = bool
  })
}