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
  description = "Specification for KubernetesClusterIssuer"
  type = object({
    # Namespace where cert-manager is installed
    cert_manager_namespace = string

    # DNS domain this issuer manages (becomes the ClusterIssuer k8s name)
    dns_domain = string

    # ACME configuration
    acme = object({
      # ACME account email
      email = string
      # ACME server URL (defaults to Let's Encrypt production)
      server = optional(string, "https://acme-v02.api.letsencrypt.org/directory")
    })

    # Cloudflare DNS solver (optional)
    cloudflare = optional(object({
      api_token = string
    }))

    # GCP Cloud DNS solver (optional)
    gcp_cloud_dns = optional(object({
      project_id = string
    }))

    # AWS Route53 solver (optional)
    aws_route53 = optional(object({
      region = string
    }))

    # Azure DNS solver (optional)
    azure_dns = optional(object({
      subscription_id = string
      resource_group  = string
    }))
  })
}
