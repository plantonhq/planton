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

    # cluster_arn is the ARN of the ECS cluster where this service will run.
    # This must already exist (created by a separate EcsCluster resource or otherwise).
    # Example: "arn:aws:ecs:us-east-1:123456789012:cluster/my-mixed-cluster"
    cluster_arn = object({

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

    # AWS ECS Service container configuration.
    container = object({

      # container image
      image = object({

        # The repository of the image (e.g., "gcr.io/project/image").
        repo = string

        # The tag of the image (e.g., "latest" or "1.0.0").
        tag = string
      })

      # container environment variables and secrets
      env = object({

        # map of environment variables to be set in the container.
        # The key is the name of the variable, and the value is the value to be set.
        variables = object({

          # Description for key
          key = string

          # Description for value
          value = string
        })

        # map of environment secrets to be set in the container.
        # The key is the name of the variable, and the value is the value to be set.
        # The value can be a plaintext value or a reference to a secret in AWS Secrets Manager or SSM Parameter Store.
        secrets = object({

          # Description for key
          key = string

          # Description for value
          value = string
        })

        # Description for s3_files
        s3_files = list(string)
      })

      # container_port is the port inside the container that should be exposed to receive traffic.
      # This is optional if the service doesn't need inbound requests (e.g., a background worker).
      # Example: 80 for HTTP
      port = number

      # replicas is the number of task replicas to run for this service.
      # higher values improve availability at increased cost.
      replicas = number

      # cpu is the amount of vCPU (in CPU units) to allocate for the entire task.
      # Valid Fargate values include 256, 512, 1024, 2048, etc., subject to ECS constraints.
      # Example: 512
      cpu = number

      # memory is the total MiB of RAM for the task.
      # Valid values depend on CPU. For example, 512 CPU can pair with 1024–4096 MiB.
      # Example: 1024
      memory = number

      # Description for logging
      logging = object({

        # Description for enabled
        enabled = bool
      })
    })

    # ECS service network configuration.
    network = object({

      # subnets is a list of VPC subnet IDs where the Fargate tasks will run.
      # Typically private subnets for production, often at least two for high availability.
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

      # security_groups is a list of security group IDs to apply to each task's ENI.
      # If not provided, ECS may use the default VPC security group, which is not ideal for production.
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
    })

    # IAM configuration for the ECS service.
    iam = object({

      # task_execution_role_arn is the IAM role used by ECS to pull private images and write logs.
      # If omitted, a default "ecsTaskExecutionRole" may be assumed, but it must already exist.
      # Example: "arn:aws:iam::123456789012:role/ecsTaskExecutionRole"
      task_execution_role_arn = object({

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

      # task_role_arn is the IAM role your container uses if it needs AWS permissions.
      # Omit if your container does not call AWS APIs.
      # Example: "arn:aws:iam::123456789012:role/myAppTaskRole"
      task_role_arn = object({

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

    # alb defines how an ALB fronts traffic to this ECS service, supporting path- or hostname-based routing.
    alb = object({

      # enabled controls whether an ALB is used. If false, no ALB is attached.
      enabled = bool

      # arn is the ARN of the ALB. Required if enabled = true.
      arn = object({

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

      # routingType can be "PATH" or "HOSTNAME" if enabled.
      # If "PATH", specify a path (e.g. "/my-service").
      # If "HOSTNAME", specify a hostname (e.g. "api.example.com").
      routing_type = string

      # path is used if routingType = "path".
      path = string

      # hostname is used if routingType = "hostname".
      hostname = string

      # listener_port is the port on the ALB to forward traffic to.
      listener_port = number

      # Description for listener_priority
      listener_priority = number

      # Description for health_check
      health_check = object({

        # Description for protocol
        protocol = string

        # Description for path
        path = string

        # Description for port
        port = string

        # Description for interval
        interval = number

        # Description for timeout
        timeout = number

        # Description for healthy_threshold
        healthy_threshold = number

        # Description for unhealthy_threshold
        unhealthy_threshold = number
      })
    })
  })
}