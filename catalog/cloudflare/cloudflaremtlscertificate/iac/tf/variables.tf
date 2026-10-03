variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "CloudflareMtlsCertificate specification"
  type = object({
    account_id = string
    name = optional(string, "")
    ca = bool
    certificates = string
    private_key = optional(string, "")
  })
}
