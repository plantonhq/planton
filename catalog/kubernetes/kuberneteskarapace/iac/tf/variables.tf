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
  description = "KubernetesKarapace specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    replicas         = optional(number)
    kafka = object({
      bootstrap_servers = string
      security_protocol = optional(string)
      tls = optional(object({
        ca_secret_name          = string
        ca_certificate          = optional(string)
        client_cert_secret_name = optional(string, "")
        client_certificate      = optional(string)
        client_key              = optional(string)
      }))
      sasl = optional(object({
        mechanism = string
        username  = string
        password_secret = optional(object({
          secret_name = string
          key         = optional(string)
        }))
        password = optional(string, "")
      }))
    })
    registry = optional(object({
      topic_name               = optional(string)
      replication_factor       = optional(number)
      compatibility            = optional(string)
      group_id                 = optional(string, "")
      master_election_strategy = optional(string)
    }))
    rest_proxy = optional(object({
      enabled  = optional(bool, false)
      replicas = optional(number)
      port     = optional(number)
      resources = optional(object({
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
    }))
    server_tls = optional(object({
      secret_name = string
      certificate = optional(string)
      key         = optional(string)
    }))
    http_authentication = optional(object({
      basic = optional(object({
        secret_name = string
        key         = optional(string)
      }))
      oidc = optional(object({
        jwks_endpoint_url = string
        expected_issuer   = optional(string, "")
        expected_audience = optional(string, "")
      }))
    }))
    port      = optional(number)
    log_level = optional(string)
    image     = optional(string, "")
    resources = optional(object({
      limits = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
      requests = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
    }))
    node_selector = optional(map(string), {})
    tolerations = optional(list(object({
      key                = optional(string, "")
      operator           = optional(string, "")
      value              = optional(string, "")
      effect             = optional(string, "")
      toleration_seconds = optional(number)
    })), [])
  })
}
