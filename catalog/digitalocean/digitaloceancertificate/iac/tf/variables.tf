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
  description = "DigitalOceanCertificate specification"
  type = object({
    certificate_name = string
    lets_encrypt = optional(object({
      domains = list(string)
    }))
    custom = optional(object({
      leaf_certificate  = string
      private_key       = string
      certificate_chain = optional(string, "")
    }))
  })
}