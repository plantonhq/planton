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
  description = "AzureVpnServerConfiguration specification"
  type = object({
    region                   = string
    resource_group           = string
    name                     = string
    vpn_authentication_types = list(string)
    aad_authentication = optional(object({
      audience = string
      issuer   = string
      tenant   = string
    }))
    client_root_certificates = optional(list(object({
      name             = string
      public_cert_data = string
    })), [])
    client_revoked_certificates = optional(list(object({
      name       = string
      thumbprint = string
    })), [])
    ipsec_policy = optional(object({
      dh_group               = optional(string, "")
      ike_encryption         = optional(string, "")
      ike_integrity          = optional(string, "")
      ipsec_encryption       = optional(string, "")
      ipsec_integrity        = optional(string, "")
      pfs_group              = optional(string, "")
      sa_lifetime_seconds    = optional(number, 0)
      sa_data_size_kilobytes = optional(number, 0)
    }))
    radius = optional(object({
      servers = optional(list(object({
        address = string
        secret  = string
        score   = optional(number, 0)
      })), [])
      client_root_certificates = optional(list(object({
        name       = string
        thumbprint = string
      })), [])
      server_root_certificates = optional(list(object({
        name             = string
        public_cert_data = string
      })), [])
    }))
    vpn_protocols = optional(list(string), [])
    policy_groups = optional(list(object({
      name       = string
      is_default = optional(bool, false)
      priority   = optional(number, 0)
      policies = list(object({
        name  = string
        type  = optional(string, "")
        value = string
      }))
    })), [])
    tags = optional(map(string), {})
  })
}
