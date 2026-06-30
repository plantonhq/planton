variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name = string,
    id   = optional(string),
    org  = optional(string),
    env  = optional(string),
    labels = optional(map(string)),
    tags = optional(list(string)),
    version = optional(object({
      id      = string,
      message = string
    }))
  })
}

variable "spec" {
  description = "Specification for the DigitalOcean Certificate"
  type = object({
    certificate_name = string
    type             = string # "letsEncrypt" or "custom"
    
    # Let's Encrypt parameters (use when type = "letsEncrypt")
    lets_encrypt = optional(object({
      domains            = list(string)
      disable_auto_renew = optional(bool, false)
    }))
    
    # Custom certificate parameters (use when type = "custom")
    custom = optional(object({
      leaf_certificate  = string
      private_key       = string
      certificate_chain = optional(string)
    }))
    
    # Optional fields
    description = optional(string)
    tags        = optional(list(string))
  })
}
