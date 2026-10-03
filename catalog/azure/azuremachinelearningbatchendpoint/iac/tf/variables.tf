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
  description = "AzureMachineLearningBatchEndpoint specification"
  type = object({
    workspace_id = string
    name         = string
    region       = string
    auth_mode    = optional(string, "")
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))
    default_deployment_name = optional(string, "")
    properties              = optional(map(string), {})
    description             = optional(string, "")
    tags                    = optional(map(string), {})
  })
}