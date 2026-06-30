variable "metadata" {
  description = "Resource metadata"
  type = object({
    name = string
    org  = optional(string)
    env  = optional(string)
  })
}

variable "spec" {
  description = "AzureVirtualMachine specification"
  type = object({
    region         = string
    resource_group = string
    vm_size        = optional(string, "Standard_D2s_v3")
    subnet_id      = string
    image = object({
      publisher       = optional(string)
      offer           = optional(string)
      sku             = optional(string)
      version         = optional(string, "latest")
      custom_image_id = optional(string)
    })
    os_disk = optional(object({
      size_gb        = optional(number)
      storage_type   = optional(string, "premium_lrs")
      caching        = optional(string, "read_write")
      delete_with_vm = optional(bool, true)
    }))
    data_disks = optional(list(object({
      name           = string
      size_gb        = number
      lun            = number
      storage_type   = optional(string, "premium_lrs")
      caching        = optional(string, "read_only")
      delete_with_vm = optional(bool, true)
    })), [])
    admin_username            = optional(string, "azureuser")
    ssh_public_key            = optional(string)
    admin_password            = optional(string)
    network = optional(object({
      enable_public_ip             = optional(bool, false)
      public_ip_sku                = optional(string, "standard")
      public_ip_allocation         = optional(string, "static")
      network_security_group_id    = optional(string)
      enable_accelerated_networking = optional(bool, true)
      private_ip_allocation        = optional(string, "private_dynamic")
      private_ip_address           = optional(string)
    }))
    availability_zone              = optional(string)
    enable_boot_diagnostics        = optional(bool, true)
    enable_system_assigned_identity = optional(bool, false)
    user_assigned_identity_ids     = optional(list(string), [])
    custom_data                    = optional(string)
    tags                           = optional(map(string), {})
    is_spot_instance               = optional(bool, false)
    spot_max_price                 = optional(number, -1)
  })
}

variable "provider_config" {
  description = "Azure provider configuration"
  type = object({
    subscription_id = string
    tenant_id       = string
    client_id       = string
    client_secret   = string
  })
  sensitive = true
}
