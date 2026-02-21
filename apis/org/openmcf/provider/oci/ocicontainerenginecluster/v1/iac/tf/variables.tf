variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "OciContainerEngineCluster specification"
  type = object({
    compartment_id = object({
      value = string
    })

    vcn_id = object({
      value = string
    })

    name              = optional(string, "")
    kubernetes_version = string
    type              = optional(string, "")
    cni_type          = optional(string, "")

    endpoint_config = optional(object({
      subnet_id = object({
        value = string
      })
      is_public_ip_enabled = optional(bool, null)
      nsg_ids = optional(list(object({
        value = string
      })), [])
    }), null)

    kms_key_id = optional(object({
      value = string
    }), null)

    image_policy_config = optional(object({
      is_policy_enabled = optional(bool, false)
      key_details = optional(list(object({
        kms_key_id = object({
          value = string
        })
      })), [])
    }), null)

    options = optional(object({
      kubernetes_network_config = optional(object({
        pods_cidr     = optional(string, "")
        services_cidr = optional(string, "")
      }), null)

      service_lb_subnet_ids = optional(list(object({
        value = string
      })), [])

      ip_families = optional(list(string), [])

      service_lb_config = optional(object({
        backend_nsg_ids = optional(list(object({
          value = string
        })), [])
        freeform_tags = optional(map(string), {})
        defined_tags  = optional(map(string), {})
      }), null)

      persistent_volume_config = optional(object({
        freeform_tags = optional(map(string), {})
        defined_tags  = optional(map(string), {})
      }), null)

      open_id_connect_token_authentication_config = optional(object({
        is_open_id_connect_auth_enabled = optional(bool, false)
        configuration_file              = optional(string, "")
        issuer_url                      = optional(string, "")
        client_id                       = optional(string, "")
        ca_certificate                  = optional(string, "")
        username_claim                  = optional(string, "")
        username_prefix                 = optional(string, "")
        groups_claim                    = optional(string, "")
        groups_prefix                   = optional(string, "")
        signing_algorithms              = optional(list(string), [])
        required_claims = optional(list(object({
          key   = string
          value = string
        })), [])
      }), null)

      is_open_id_connect_discovery_enabled = optional(bool, false)
    }), null)
  })
}
