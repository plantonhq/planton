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
  description = "AzureMachineLearningComputeInstance specification"
  type = object({
    workspace_id         = string
    name                 = string
    virtual_machine_size = string
    authorization_type   = optional(string, "")
    assign_to_user = optional(object({
      tenant_id = optional(string, "")
      object_id = optional(string, "")
    }))
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))
    local_auth_enabled = optional(bool)
    ssh = optional(object({
      public_key = string
    }))
    subnet_id              = optional(string, "")
    node_public_ip_enabled = optional(bool)
    description            = optional(string, "")
    tags                   = optional(map(string), {})
  })
}