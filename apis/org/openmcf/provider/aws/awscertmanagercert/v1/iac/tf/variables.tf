variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "AwsCertManagerCert specification"
  type = object({
    region = string
    primary_domain_name = string
    alternate_domain_names = optional(list(string), [])
    route53_hosted_zone_id = string
    validation_method = optional(string, "")
  })
}
