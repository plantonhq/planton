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

    # Description for cluster_name
    cluster_name = object({

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

    # Description for node_role_arn
    node_role_arn = object({

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

    # Description for subnet_ids
    subnet_ids = list(object({

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

    # Description for instance_type
    instance_type = string

    # Description for scaling
    scaling = object({

      # Description for min_size
      min_size = number

      # Description for max_size
      max_size = number

      # Description for desired_size
      desired_size = number
    })

    # Description for capacity_type
    capacity_type = string

    # Description for disk_size_gb
    disk_size_gb = number

    # Description for ssh_key_name
    ssh_key_name = string

    # Description for labels
    labels = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })
  })
}