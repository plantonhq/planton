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
  description = "AzureExpressRoutePort specification"
  type = object({
    region            = string
    resource_group    = string
    name              = string
    peering_location  = string
    bandwidth_in_gbps = number
    encapsulation     = optional(string, "")
    billing_type      = optional(string)
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))
    link1 = optional(object({
      admin_enabled                 = optional(bool, false)
      macsec_cipher                 = optional(string)
      macsec_ckn_keyvault_secret_id = optional(string, "")
      macsec_cak_keyvault_secret_id = optional(string, "")
      macsec_sci_enabled            = optional(bool, false)
    }))
    link2 = optional(object({
      admin_enabled                 = optional(bool, false)
      macsec_cipher                 = optional(string)
      macsec_ckn_keyvault_secret_id = optional(string, "")
      macsec_cak_keyvault_secret_id = optional(string, "")
      macsec_sci_enabled            = optional(bool, false)
    }))
    authorizations = optional(list(object({
      name = string
    })), [])
    tags = optional(map(string), {})
  })
}