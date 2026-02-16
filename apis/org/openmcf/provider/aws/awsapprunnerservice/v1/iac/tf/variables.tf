variable "provider_config" {
  description = "AWS provider configuration"
  type = object({
    region            = string
    access_key_id     = optional(string)
    secret_access_key = optional(string)
    session_token     = optional(string)
  })
}

variable "metadata" {
  description = "Resource metadata"
  type = object({
    id   = string
    name = string
    org  = optional(string, "")
    env  = optional(string, "")
  })
}

variable "spec" {
  description = "AwsAppRunnerServiceSpec - desired state of the App Runner service"
  type = object({
    # Source (exactly one of image_source or code_source must be set)
    image_source = optional(object({
      image_identifier      = string
      image_repository_type = string
      access_role_arn       = optional(string, "")
    }), null)

    code_source = optional(object({
      repository_url       = string
      branch               = string
      source_directory     = optional(string, "")
      connection_arn       = string
      configuration_source = string
      runtime              = optional(string, "")
      build_command        = optional(string, "")
    }), null)

    # Runtime configuration (middleware-populated defaults)
    port                    = string
    start_command           = optional(string, "")
    environment_variables   = optional(map(string), {})
    environment_secrets     = optional(map(string), {})

    # Instance configuration (middleware-populated defaults)
    cpu               = string
    memory            = string
    instance_role_arn = optional(string, "")

    # Health check
    health_check = optional(object({
      protocol            = string
      path                = string
      interval_seconds    = number
      timeout_seconds     = number
      healthy_threshold   = number
      unhealthy_threshold = number
    }), null)

    # Auto scaling
    auto_scaling = optional(object({
      min_size        = number
      max_size        = number
      max_concurrency = number
    }), null)

    # VPC egress networking
    vpc_connector_arn  = optional(string, "")
    subnet_ids         = optional(list(string), [])
    security_group_ids = optional(list(string), [])

    # Network ingress (middleware-populated defaults)
    is_publicly_accessible = bool
    ip_address_type        = string

    # Encryption
    kms_key_arn = optional(string, "")

    # Observability
    observability_enabled           = optional(bool, false)
    observability_configuration_arn = optional(string, "")

    # Deployment behavior (middleware-populated default)
    auto_deployments_enabled = bool
  })
}
