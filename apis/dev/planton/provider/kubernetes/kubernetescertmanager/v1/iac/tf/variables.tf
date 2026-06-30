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
  description = "Specification for Kubernetes Cert-Manager deployment"
  type = object({
    # Kubernetes namespace where cert-manager will be deployed
    namespace = optional(string, "cert-manager")

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, false)

    # cert-manager version (e.g., "v1.19.1") -- sets the image tag
    kubernetes_cert_manager_version = optional(string, "v1.19.1")

    # Helm chart version
    helm_chart_version = optional(string, "v1.19.1")

    # Skip installation of self-signed issuer
    skip_install_self_signed_issuer = optional(bool, false)

    # Workload identity configuration (optional)
    workload_identity = optional(object({
      # GKE Workload Identity (optional)
      gke = optional(object({
        service_account_email = string
      }))

      # EKS IRSA (optional)
      eks = optional(object({
        role_arn = string
      }))

      # AKS Workload Identity (optional)
      aks = optional(object({
        client_id = string
      }))
    }))
  })
}
