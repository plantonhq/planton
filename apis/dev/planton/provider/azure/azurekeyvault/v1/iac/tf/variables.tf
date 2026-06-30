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
  description = "Azure Key Vault specification"
  type = object({
    # The Azure region where the Key Vault will be deployed
    region = string

    # The Azure Resource Group name where the Key Vault will be created
    resource_group = string

    # The SKU tier for the Key Vault (standard or premium)
    sku = optional(string, "standard")

    # Enable Azure RBAC for authorization instead of vault access policies
    enable_rbac_authorization = optional(bool, true)

    # Enable purge protection to prevent permanent deletion
    enable_purge_protection = optional(bool, true)

    # Soft delete retention period in days (7-90)
    soft_delete_retention_days = optional(number, 90)

    # Network access control configuration
    network_acls = optional(object({
      default_action              = optional(string, "Deny")
      bypass_azure_services       = optional(bool, true)
      ip_rules                    = optional(list(string), [])
      virtual_network_subnet_ids  = optional(list(string), [])
    }))

    # List of secret names to create in the Key Vault
    secret_names = optional(list(string), [])
  })
}
