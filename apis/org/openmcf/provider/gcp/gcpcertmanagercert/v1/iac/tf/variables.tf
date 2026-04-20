variable "metadata" {
  description = "Metadata for the GCP Certificate Manager Cert resource"
  type = object({
    name = string
    id   = string
    org  = string
    env = object({
      id = string
    })
  })
}

variable "spec" {
  description = "Specification for the GCP Certificate Manager Cert"
  type = object({
    gcp_project_id         = string
    primary_domain_name    = string
    alternate_domain_names = list(string)
    cloud_dns_zone_id = optional(object({
      value = string
    }), null)
    certificate_type  = optional(number, 0) # 0 = MANAGED, 1 = LOAD_BALANCER
    validation_method = optional(string, "DNS")
  })
}

