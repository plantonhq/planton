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
  description = "AzureMachineLearningDatastore specification"
  type = object({
    workspace_id          = string
    name                  = string
    description           = optional(string, "")
    service_data_identity = optional(string, "")
    tags                  = optional(map(string), {})
    blob_storage = optional(object({
      storage_container_id    = string
      is_default              = optional(bool, false)
      account_key             = optional(string, "")
      shared_access_signature = optional(string, "")
    }))
    data_lake_gen2 = optional(object({
      storage_container_id = string
      tenant_id            = optional(string, "")
      client_id            = optional(string, "")
      client_secret        = optional(string, "")
      authority_url        = optional(string, "")
    }))
    file_share = optional(object({
      storage_fileshare_id    = string
      account_key             = optional(string, "")
      shared_access_signature = optional(string, "")
    }))
  })
}