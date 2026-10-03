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
  description = "AzureBastionHost specification"
  type = object({
    region         = string
    resource_group = string
    name           = string
    sku            = optional(string, "")
    ip_configuration = optional(object({
      name                 = string
      subnet_id            = string
      public_ip_address_id = optional(string, "")
    }))
    virtual_network_id        = optional(string, "")
    scale_units               = optional(number)
    copy_paste_enabled        = optional(bool)
    file_copy_enabled         = optional(bool, false)
    ip_connect_enabled        = optional(bool, false)
    kerberos_enabled          = optional(bool, false)
    shareable_link_enabled    = optional(bool, false)
    tunneling_enabled         = optional(bool, false)
    session_recording_enabled = optional(bool, false)
    zones                     = optional(list(string), [])
    tags                      = optional(map(string), {})
  })
}
