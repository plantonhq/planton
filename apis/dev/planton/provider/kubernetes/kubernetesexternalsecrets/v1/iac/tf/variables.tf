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
  description = "Specification for Kubernetes External Secrets Operator"
  type = object({
    # Namespace where the operator will be installed
    namespace = optional(string)

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, false)

    # Helm chart version to deploy
    helm_chart_version = optional(string)

    # Polling interval for secrets in seconds
    poll_interval_seconds = optional(number)

    # Container resource specifications
    container = optional(object({
      resources = optional(object({
        requests = optional(object({
          cpu    = optional(string)
          memory = optional(string)
        }))
        limits = optional(object({
          cpu    = optional(string)
          memory = optional(string)
        }))
      }))
    }))

    # GKE-specific configuration
    gke = optional(object({
      gsa_email = optional(string)
    }))

    # EKS-specific configuration
    eks = optional(object({
      irsa_role_arn_override = optional(string)
    }))

    # AKS-specific configuration
    aks = optional(object({
      managed_identity_client_id = optional(string)
    }))
  })
}