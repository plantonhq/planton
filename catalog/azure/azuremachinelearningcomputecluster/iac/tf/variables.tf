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
  description = "AzureMachineLearningComputeCluster specification"
  type = object({
    workspace_id = string
    name         = string
    region       = string
    vm_size      = string
    vm_priority  = string
    scale_settings = object({
      max_node_count                       = optional(number, 0)
      min_node_count                       = optional(number, 0)
      scale_down_nodes_after_idle_duration = string
    })
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))
    ssh = optional(object({
      admin_username = string
      admin_password = optional(string, "")
      key_value      = optional(string, "")
    }))
    ssh_public_access_enabled = optional(bool, false)
    local_auth_enabled        = optional(bool)
    node_public_ip_enabled    = optional(bool)
    subnet_id                 = optional(string, "")
    description               = optional(string, "")
    tags                      = optional(map(string), {})
  })
}