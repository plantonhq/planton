variable "metadata" {
  description = "Resource metadata from the manifest"
  type = object({
    name = string
    id   = string
    org  = string
    env  = string
    labels = object({
      key   = string
      value = string
    })
    annotations = object({
      key   = string
      value = string
    })
    tags = list(string)
  })
}

variable "spec" {
  description = "AwsBatchComputeEnvironmentSpec configuration"
  type = object({
    # Compute environment state: ENABLED or DISABLED
    state = string

    # Service role ARN (optional)
    service_role = object({
      value = string
    })

    # Compute resources configuration
    compute_resources = object({
      type                = string
      max_vcpus           = number
      min_vcpus           = number
      desired_vcpus       = number
      subnet_ids          = list(object({ value = string }))
      security_group_ids  = list(object({ value = string }))
      instance_types      = list(string)
      allocation_strategy = string
      instance_role       = object({ value = string })
      ec2_key_pair        = string
      bid_percentage      = number
      spot_iam_fleet_role = object({ value = string })
      launch_template = optional(object({
        launch_template_id   = string
        launch_template_name = string
        version              = string
      }))
      ec2_configurations = list(object({
        image_type        = string
        image_id_override = string
      }))
      resource_tags = map(string)
    })

    # Update policy
    update_policy = optional(object({
      terminate_jobs_on_update        = bool
      job_execution_timeout_minutes   = number
    }))

    # Job queues
    job_queues = list(object({
      name     = string
      state    = string
      priority = number
      job_state_time_limit_actions = list(object({
        action           = string
        max_time_seconds = number
        reason           = string
        state            = string
      }))
    }))

    # Scheduling policy
    scheduling_policy = optional(object({
      compute_reservation = number
      share_decay_seconds = number
      share_distributions = list(object({
        share_identifier = string
        weight_factor    = number
      }))
    }))
  })
}

variable "provider_config" {
  description = "AWS provider configuration"
  type = object({
    region            = string
    access_key_id     = string
    secret_access_key = string
    session_token     = string
  })
  default = {
    region            = ""
    access_key_id     = ""
    secret_access_key = ""
    session_token     = ""
  }
}
