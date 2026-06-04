variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for ExternalDNS on a Kubernetes cluster"
  type = object({
    # Kubernetes namespace where ExternalDNS is installed. StringValueOrRef:
    # the platform resolves value_from before apply, so only .value is consumed.
    namespace = optional(object({
      value      = optional(string)
      value_from = optional(any)
    }))

    # Whether this module should create the namespace (vs. read an existing one).
    create_namespace = optional(bool, false)

    # ExternalDNS image tag (e.g. "v0.19.0").
    external_dns_version = optional(string)

    # ExternalDNS Helm chart version (e.g. "1.19.0").
    helm_chart_version = optional(string)

    # GKE + Google Cloud DNS configuration (one provider_config is set).
    gke = optional(object({
      project_id = optional(object({
        value      = optional(string)
        value_from = optional(any)
      }))
      dns_zone_id = optional(object({
        value      = optional(string)
        value_from = optional(any)
      }))
    }))

    # EKS + AWS Route53 configuration.
    eks = optional(object({
      route53_zone_id = optional(object({
        value      = optional(string)
        value_from = optional(any)
      }))
      irsa_role_arn_override = optional(string)
    }))

    # AKS + Azure DNS configuration.
    aks = optional(object({
      dns_zone_id = optional(object({
        value      = optional(string)
        value_from = optional(any)
      }))
      managed_identity_client_id = optional(string)
    }))

    # Cloudflare DNS configuration.
    cloudflare = optional(object({
      api_token = optional(string)
      dns_zone_id = optional(object({
        value      = optional(string)
        value_from = optional(any)
      }))
      is_proxied = optional(bool, false)
    }))
  })
}
