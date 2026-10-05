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
  description = "AwsPrivateCa specification"
  type = object({
    region = string
    type = optional(string, "")
    key_algorithm = optional(string, "")
    signing_algorithm = optional(string, "")
    subject = object({
      common_name = optional(string, "")
      organization = optional(string, "")
      organizational_unit = optional(string, "")
      country = optional(string, "")
      state = optional(string, "")
      locality = optional(string, "")
    })
    usage_mode = optional(string, "")
    key_storage_security_standard = optional(string, "")
    revocation = optional(object({
      crl = optional(object({
        enabled = optional(bool, false)
        expiration_in_days = optional(number, 0)
        s3_bucket_name = optional(string, "")
        s3_object_acl = optional(string, "")
        custom_cname = optional(string, "")
        custom_path = optional(string, "")
      }))
      ocsp = optional(object({
        enabled = optional(bool, false)
        custom_cname = optional(string, "")
      }))
    }))
    root_ca_validity = optional(object({
      type = optional(string, "")
      value = string
    }))
    subordinate_activation = optional(object({
      parent_ca_arn = string
      path_length = optional(number, 0)
      validity = object({
        type = optional(string, "")
        value = string
      })
    }))
    issued_certificates = optional(list(object({
      name = string
      csr = string
      signing_algorithm = optional(string, "")
      validity = object({
        type = optional(string, "")
        value = string
      })
      template_arn = optional(string, "")
      api_passthrough = optional(string, "")
    })), [])
    acm_renewal_permission = optional(bool, false)
    policy = optional(string, "")
    permanent_deletion_time_in_days = optional(number, 0)
    enabled = optional(bool)
  })
}