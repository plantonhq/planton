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
  description = "KubernetesClusterIssuer specification"
  type = object({
    cert_manager_namespace = string
    config = object({
      acme = optional(object({
        email           = string
        server          = optional(string)
        profile         = optional(string, "")
        preferred_chain = optional(string, "")
        ca_bundle       = optional(string, "")
        skip_tls_verify = optional(bool, false)
        external_account_binding = optional(object({
          key_id   = string
          hmac_key = string
        }))
        disable_account_key_generation = optional(bool, false)
        enable_duration_feature        = optional(bool, false)
        solvers = list(object({
          selector = optional(object({
            dns_zones    = optional(list(string), [])
            dns_names    = optional(list(string), [])
            match_labels = optional(map(string), {})
          }))
          http01 = optional(object({
            ingress = optional(object({
              ingress_class_name = optional(string, "")
              name               = optional(string, "")
              service_type       = optional(string)
            }))
            gateway_http_route = optional(object({
              parent_refs = list(object({
                name         = string
                namespace    = optional(string, "")
                section_name = optional(string, "")
              }))
              labels       = optional(map(string), {})
              service_type = optional(string)
            }))
          }))
          dns01 = optional(object({
            cname_strategy = optional(string)
            cloudflare = optional(object({
              api_token = optional(object({
                token = string
              }))
              api_key = optional(object({
                email = string
                key   = string
              }))
            }))
            route53 = optional(object({
              region          = string
              hosted_zone_id  = optional(string, "")
              assume_role_arn = optional(string, "")
              static_credentials = optional(object({
                access_key_id     = string
                secret_access_key = string
              }))
              service_account = optional(object({
                service_account_name = string
                audiences            = optional(list(string), [])
              }))
            }))
            azure_dns = optional(object({
              subscription_id     = string
              resource_group_name = string
              hosted_zone_name    = optional(string, "")
              zone_type           = optional(string)
              environment         = optional(string)
              client_id           = optional(string, "")
              client_secret       = optional(string, "")
              tenant_id           = optional(string, "")
              managed_identity = optional(object({
                client_id   = optional(string, "")
                resource_id = optional(string, "")
              }))
            }))
            gcp_cloud_dns = optional(object({
              project_id               = string
              hosted_zone_name         = optional(string, "")
              service_account_key_json = optional(string, "")
            }))
            digitalocean = optional(object({
              token = string
            }))
            rfc2136 = optional(object({
              nameserver     = string
              tsig_key_name  = optional(string, "")
              tsig_algorithm = optional(string, "")
              tsig_secret    = optional(string, "")
            }))
            acme_dns = optional(object({
              host         = string
              account_json = string
            }))
            akamai = optional(object({
              service_consumer_domain = string
              client_token            = string
              client_secret           = string
              access_token            = string
            }))
            webhook = optional(object({
              group_name  = string
              solver_name = string
              config_yaml = optional(string, "")
            }))
          }))
        }))
      }))
      ca = optional(object({
        ca_secret_name           = string
        crl_distribution_points  = optional(list(string), [])
        ocsp_servers             = optional(list(string), [])
        issuing_certificate_urls = optional(list(string), [])
      }))
      self_signed = optional(object({
        crl_distribution_points = optional(list(string), [])
      }))
      vault = optional(object({
        server          = string
        path            = string
        vault_namespace = optional(string, "")
        ca_bundle       = optional(string, "")
        server_name     = optional(string, "")
        token_auth = optional(object({
          token = string
        }))
        app_role_auth = optional(object({
          path      = string
          role_id   = string
          secret_id = string
        }))
        kubernetes_auth = optional(object({
          role                 = string
          mount_path           = optional(string)
          service_account_name = string
          audiences            = optional(list(string), [])
        }))
      }))
    })
  })
}