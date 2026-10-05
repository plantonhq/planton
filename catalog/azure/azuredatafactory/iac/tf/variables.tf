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
  description = "AzureDataFactory specification"
  type = object({
    resource_group = string
    name           = string
    region         = string
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))
    github_configuration = optional(object({
      account_name       = string
      branch_name        = string
      git_url            = optional(string, "")
      repository_name    = string
      root_folder        = string
      publishing_enabled = optional(bool)
    }))
    vsts_configuration = optional(object({
      account_name       = string
      branch_name        = string
      project_name       = string
      repository_name    = string
      root_folder        = string
      tenant_id          = string
      publishing_enabled = optional(bool)
    }))
    global_parameters = optional(list(object({
      name  = string
      type  = string
      value = string
    })), [])
    managed_virtual_network_enabled = optional(bool)
    public_network_enabled          = optional(bool)
    purview_id                      = optional(string, "")
    customer_managed_key = optional(object({
      key_vault_key_id          = string
      user_assigned_identity_id = string
    }))
    user_managed_identity_credentials = optional(list(object({
      name        = string
      identity_id = string
      description = optional(string, "")
      annotations = optional(list(string), [])
    })), [])
    service_principal_credentials = optional(list(object({
      name                 = string
      tenant_id            = string
      service_principal_id = string
      service_principal_key = optional(object({
        linked_service_name = string
        secret_name         = string
        secret_version      = optional(string, "")
      }))
      description = optional(string, "")
      annotations = optional(list(string), [])
    })), [])
    managed_private_endpoints = optional(list(object({
      name               = string
      target_resource_id = string
      subresource_name   = optional(string, "")
      fqdns              = optional(list(string), [])
    })), [])
    tags = optional(map(string), {})
  })
}
