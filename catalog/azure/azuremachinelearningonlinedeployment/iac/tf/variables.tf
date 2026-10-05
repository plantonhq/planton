variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "AzureMachineLearningOnlineDeployment specification"
  type = object({
    endpoint_id      = string
    name             = string
    region           = string
    instance_type    = optional(string, "")
    instance_count   = optional(number)
    model            = optional(string, "")
    model_mount_path = optional(string, "")
    code_configuration = optional(object({
      code_id        = optional(string, "")
      scoring_script = string
    }))
    environment_id                       = optional(string, "")
    environment_variables                = optional(map(string), {})
    app_insights_enabled                 = optional(bool, false)
    egress_public_network_access_enabled = optional(bool)
    liveness_probe = optional(object({
      failure_threshold = optional(number)
      success_threshold = optional(number)
      initial_delay     = optional(string, "")
      period            = optional(string, "")
      timeout           = optional(string, "")
    }))
    readiness_probe = optional(object({
      failure_threshold = optional(number)
      success_threshold = optional(number)
      initial_delay     = optional(string, "")
      period            = optional(string, "")
      timeout           = optional(string, "")
    }))
    startup_probe = optional(object({
      failure_threshold = optional(number)
      success_threshold = optional(number)
      initial_delay     = optional(string, "")
      period            = optional(string, "")
      timeout           = optional(string, "")
    }))
    request_settings = optional(object({
      max_concurrent_requests_per_instance = optional(number)
      request_timeout                      = optional(string, "")
    }))
    data_collector = optional(object({
      collections = optional(map(object({
        enabled       = optional(bool, false)
        data_id       = optional(string, "")
        client_id     = optional(string, "")
        sampling_rate = optional(number)
      })), {})
      rolling_rate = optional(string, "")
      request_logging = optional(object({
        capture_headers = optional(list(string), [])
      }))
    }))
    properties  = optional(map(string), {})
    description = optional(string, "")
    tags        = optional(map(string), {})
  })
}
