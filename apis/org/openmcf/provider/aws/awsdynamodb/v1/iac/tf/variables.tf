variable "metadata" {
  description = "metadata"
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
  description = "spec"
  type = object({

    # The AWS region where the resource will be created.
    region = string

    # Controls how you are charged for read and write throughput and how you manage
    # capacity. The valid values are `PROVISIONED` and `PAY_PER_REQUEST`. Defaults
    # to `PROVISIONED`.
    billing_mode = string

    # Description for provisioned_throughput
    provisioned_throughput = object({

      # Description for read_capacity_units
      read_capacity_units = number

      # Description for write_capacity_units
      write_capacity_units = number
    })

    # Description for attribute_definitions
    attribute_definitions = list(object({

      # Description for name
      name = string

      # Description for type
      type = string
    }))

    # Description for key_schema
    key_schema = list(object({

      # Description for attribute_name
      attribute_name = string

      # Description for key_type
      key_type = string
    }))

    # Describe a GSI for the table; subject to the normal limits on the number of GSIs, projected attributes, etc.
    global_secondary_indexes = list(object({

      # Description for name
      name = string

      # Description for key_schema
      key_schema = list(object({

        # Description for attribute_name
        attribute_name = string

        # Description for key_type
        key_type = string
      }))

      # Description for projection
      projection = object({

        # Description for type
        type = string

        # Description for non_key_attributes
        non_key_attributes = list(string)
      })

      # Description for provisioned_throughput
      provisioned_throughput = object({

        # Description for read_capacity_units
        read_capacity_units = number

        # Description for write_capacity_units
        write_capacity_units = number
      })
    }))

    # Describe an LSI on the table; these can only be allocated _at creation_
    # so you cannot change this definition after you have created the resource.
    local_secondary_indexes = list(object({

      # Description for name
      name = string

      # Description for key_schema
      key_schema = list(object({

        # Description for attribute_name
        attribute_name = string

        # Description for key_type
        key_type = string
      }))

      # Description for projection
      projection = object({

        # Description for type
        type = string

        # Description for non_key_attributes
        non_key_attributes = list(string)
      })
    }))

    # Configuration block for TTL.
    ttl = object({

      # Description for enabled
      enabled = bool

      # Description for attribute_name
      attribute_name = string
    })

    # Description for stream_enabled
    stream_enabled = bool

    # When an item in the table is modified, StreamViewType determines what information
    # is written to the table's stream. Valid values are
    # `KEYS_ONLY`, `NEW_IMAGE`, `OLD_IMAGE`, `NEW_AND_OLD_IMAGES`.
    stream_view_type = string

    # Description for point_in_time_recovery_enabled
    point_in_time_recovery_enabled = bool

    # Encryption at rest options. AWS DynamoDB tables are automatically
    # encrypted at rest with an AWS-owned Customer Master Key if this argument
    # isn't specified.
    server_side_encryption = object({

      # Description for enabled
      enabled = bool

      # Description for kms_key_arn
      kms_key_arn = string
    })

    # Description for table_class
    table_class = string

    # Description for deletion_protection_enabled
    deletion_protection_enabled = bool

    # Description for contributor_insights_enabled
    contributor_insights_enabled = bool
  })
}