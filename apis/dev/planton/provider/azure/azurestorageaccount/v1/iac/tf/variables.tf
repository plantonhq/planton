variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Azure Storage Account specification"
  type = object({
    # The Azure region where the Storage Account will be deployed
    region = string

    # The Azure Resource Group name where the Storage Account will be created
    resource_group = string

    # The kind of storage account (StorageV2, BlobStorage, BlockBlobStorage, FileStorage, Storage)
    account_kind = optional(string, "StorageV2")

    # The performance tier (Standard or Premium)
    account_tier = optional(string, "Standard")

    # The replication strategy (LRS, ZRS, GRS, GZRS, RAGRS, RAGZRS)
    replication_type = optional(string, "LRS")

    # The default access tier for blob data (Hot or Cool)
    access_tier = optional(string, "Hot")

    # Enable HTTPS traffic only
    enable_https_traffic_only = optional(bool, true)

    # Minimum TLS version (TLS1_0, TLS1_1, TLS1_2)
    min_tls_version = optional(string, "TLS1_2")

    # Network access control configuration
    network_rules = optional(object({
      default_action             = optional(string, "Deny")
      bypass_azure_services      = optional(bool, true)
      ip_rules                   = optional(list(string), [])
      virtual_network_subnet_ids = optional(list(string), [])
    }))

    # Blob service properties configuration
    blob_properties = optional(object({
      enable_versioning                  = optional(bool, false)
      soft_delete_retention_days         = optional(number, 7)
      container_soft_delete_retention_days = optional(number, 7)
    }))

    # List of blob containers to create
    containers = optional(list(object({
      name        = string
      access_type = optional(string, "private")
    })), [])
  })
}
