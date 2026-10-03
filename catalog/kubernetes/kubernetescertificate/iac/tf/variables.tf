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
  description = "KubernetesCertificate specification"
  type = object({
    namespace   = string
    secret_name = string
    issuer_ref = object({
      cluster_issuer = optional(object({
        name = string
      }))
      issuer = optional(object({
        name = string
      }))
      external = optional(object({
        group = string
        kind  = string
        name  = string
      }))
    })
    dns_names       = optional(list(string), [])
    ip_addresses    = optional(list(string), [])
    uris            = optional(list(string), [])
    email_addresses = optional(list(string), [])
    common_name     = optional(string, "")
    subject = optional(object({
      organizations        = optional(list(string), [])
      organizational_units = optional(list(string), [])
      countries            = optional(list(string), [])
      provinces            = optional(list(string), [])
      localities           = optional(list(string), [])
      street_addresses     = optional(list(string), [])
      postal_codes         = optional(list(string), [])
      serial_number        = optional(string, "")
    }))
    literal_subject = optional(string, "")
    other_names = optional(list(object({
      oid        = string
      utf8_value = string
    })), [])
    duration                = optional(string)
    renew_before            = optional(string, "")
    renew_before_percentage = optional(number)
    private_key = optional(object({
      algorithm       = optional(string)
      size            = optional(number)
      encoding        = optional(string)
      rotation_policy = optional(string)
    }))
    usages                   = optional(list(string), [])
    encode_usages_in_request = optional(bool, false)
    is_ca                    = optional(bool, false)
    signature_algorithm      = optional(string)
    keystores = optional(object({
      jks = optional(object({
        create   = optional(bool, false)
        alias    = optional(string)
        password = string
      }))
      pkcs12 = optional(object({
        create   = optional(bool, false)
        password = string
        profile  = optional(string)
      }))
    }))
    additional_output_formats = optional(list(object({
      type = string
    })), [])
    name_constraints = optional(object({
      critical = optional(bool, false)
      permitted = optional(object({
        dns_domains     = optional(list(string), [])
        ip_ranges       = optional(list(string), [])
        email_addresses = optional(list(string), [])
        uri_domains     = optional(list(string), [])
      }))
      excluded = optional(object({
        dns_domains     = optional(list(string), [])
        ip_ranges       = optional(list(string), [])
        email_addresses = optional(list(string), [])
        uri_domains     = optional(list(string), [])
      }))
    }))
    secret_template = optional(object({
      labels      = optional(map(string), {})
      annotations = optional(map(string), {})
    }))
    revision_history_limit = optional(number)
  })
}