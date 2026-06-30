variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "GCP Cloud Function specification"
  type = object({
    # GCP project ID where the Cloud Function will be created
    project_id = string

    # Region where the function is deployed (e.g., "us-central1")
    region = string

    # Optional: Custom function name. If not provided, metadata.name is used
    function_name = optional(string)

    # Build configuration
    build_config = object({
      runtime     = string # e.g., "python311", "nodejs20"
      entry_point = string # Function name in source code
      source = object({
        bucket     = string           # GCS bucket containing source code
        object     = string           # Object path in bucket (e.g., "function.zip")
        generation = optional(number) # Optional: specific object version
      })
      build_environment_variables = optional(map(string))
    })

    # Service configuration (optional, with defaults)
    service_config = optional(object({
      service_account_email            = optional(string)
      available_memory_mb              = optional(number) # 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768
      timeout_seconds                  = optional(number) # 1-3600
      max_instance_request_concurrency = optional(number) # 1-1000
      environment_variables            = optional(map(string))
      secret_environment_variables     = optional(map(string))
      vpc_connector                    = optional(string)
      vpc_connector_egress_settings    = optional(string) # "PRIVATE_RANGES_ONLY" or "ALL_TRAFFIC"
      ingress_settings                 = optional(string) # "ALLOW_ALL", "ALLOW_INTERNAL_ONLY", "ALLOW_INTERNAL_AND_GCLB"
      scaling = optional(object({
        min_instance_count = optional(number) # 0-100
        max_instance_count = optional(number) # 1-3000
      }))
      allow_unauthenticated = optional(bool)
    }))

    # Trigger configuration (optional, defaults to HTTP)
    trigger = optional(object({
      trigger_type = optional(number) # 0=HTTP, 1=EVENT_TRIGGER
      event_trigger = optional(object({
        event_type   = string
        pubsub_topic = optional(string)
        event_filters = optional(list(object({
          attribute = string
          value     = string
          operator  = optional(string)
        })))
        trigger_region        = optional(string)
        retry_policy          = optional(number) # 0=DO_NOT_RETRY, 1=RETRY
        service_account_email = optional(string)
      }))
    }))
  })
}
