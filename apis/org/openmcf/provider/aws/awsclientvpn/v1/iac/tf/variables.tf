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

    # Description for description
    description = string

    # Description for vpc_id
    vpc_id = object({

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

    # Description for client_cidr_block
    client_cidr_block = string

    # Description for authentication_type
    authentication_type = string

    # Description for server_certificate_arn
    server_certificate_arn = object({

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

    # Description for cidr_authorization_rules
    cidr_authorization_rules = list(string)

    # Description for disable_split_tunnel
    disable_split_tunnel = bool

    # Description for vpn_port
    vpn_port = number

    # Description for transport_protocol
    transport_protocol = string

    # Description for log_group_name
    log_group_name = string

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

    # Description for dns_servers
    dns_servers = list(string)
  })
}