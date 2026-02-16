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
  description = "spec holds the core configuration data defining how the ECS cluster is deployed."
  type = object({

    # enable_container_insights determines whether to enable CloudWatch
    # Container Insights for this cluster. This is highly recommended
    # for production monitoring, though it incurs CloudWatch costs.
    # If omitted, it is recommended to be "true".
    enable_container_insights = bool

    # capacity_providers is a list of capacity providers attached
    # to this cluster. For a Fargate-only cluster, typically ["FARGATE"]
    # or ["FARGATE", "FARGATE_SPOT"] for cost-optimized Spot usage.
    capacity_providers = list(string)

    # default_capacity_provider_strategy defines the base/weight
    # distribution for tasks across capacity providers. This is the
    # primary cost-optimization lever for Fargate workloads.
    default_capacity_provider_strategy = list(object({
      capacity_provider = string
      base              = number
      weight            = number
    }))

    # execute_command_configuration defines cluster-level auditing
    # settings for ECS Exec. This controls logging and encryption
    # for exec sessions. If not specified, exec is disabled.
    execute_command_configuration = object({
      logging = string
      log_configuration = object({
        cloud_watch_log_group_name      = string
        cloud_watch_encryption_enabled  = bool
        s3_bucket_name                  = string
        s3_key_prefix                   = string
        s3_encryption_enabled           = bool
      })
      kms_key_id = optional(object({ value = string }))
    })
  })
}