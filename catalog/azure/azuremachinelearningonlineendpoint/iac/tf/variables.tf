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
  description = "AzureMachineLearningOnlineEndpoint specification"
  type = object({
    workspace_id = string
    name         = string
    region       = string
    auth_mode    = string
    identity = object({
      type         = string
      identity_ids = optional(list(string), [])
    })
    traffic                       = optional(map(number), {})
    mirror_traffic                = optional(map(number), {})
    public_network_access_enabled = optional(bool)
    initial_auth_keys = optional(object({
      primary_key   = optional(string, "")
      secondary_key = optional(string, "")
    }))
    properties  = optional(map(string), {})
    description = optional(string, "")
    tags        = optional(map(string), {})
  })
}
