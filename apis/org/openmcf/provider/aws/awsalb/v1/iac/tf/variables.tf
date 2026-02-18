variable "metadata" {
  description = "metadata captures identifying information (name, org, version, etc.)\nand must pass standard validations for resource naming."
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
  description = "spec holds the core configuration data defining how the ECS service is deployed."
  type = object({

    # The AWS region where the resource will be created.
    region = string

    # subnets is a list of subnet IDs in which to create the ALB.
    # Often private subnets for internal or public subnets for internet-facing.
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

    # securityGroups is a list of security group IDs to attach to the ALB.
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

    # scheme indicates whether the ALB is internet-facing or internal.
    # Valid values: "internet-facing" or "internal".
    # If omitted, default to "internet-facing".
    internal = bool

    # enable_deletion_protection indicates whether the ALB should have deletion protection enabled.
    # This prevents accidental deletion.
    delete_protection_enabled = bool

    # idle_timeout_seconds sets the idle timeout in seconds for connections to the ALB.
    # If omitted, AWS default is 60 seconds.
    idle_timeout_seconds = number

    # dns configuration allows the resource to manage Route53 DNS if enabled.
    dns = object({

      # enabled, when set to true, indicates that the ALB resource
      # should create DNS records in Route53 .
      enabled = bool

      # route53_zone_id is the Route53 Hosted Zone ID where DNS records
      # will be created.
      route53_zone_id = object({

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

      # hostnames is a list of domain names (e.g., ["app.example.com"])
      # that will point to this ALB.
      hostnames = list(string)
    })

    # ssl configuration allows a single toggle for SSL, plus a certificate ARN if enabled.
    ssl = object({

      # Description for enabled
      enabled = bool

      # Description for certificate_arn
      certificate_arn = object({

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
    })
  })
}