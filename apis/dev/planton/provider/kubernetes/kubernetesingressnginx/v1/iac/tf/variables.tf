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
  description = "Specification for the Kubernetes Ingress Nginx deployment"
  type = object({
    # The target cluster configuration (not directly used in Terraform, handled by provider)
    # Included for API compatibility but credentials should be configured via provider block

    # Namespace for deployment
    namespace = optional(string)

    # Whether to create the namespace
    create_namespace = optional(bool, false)

    # Upstream Helm chart version tag (e.g. "4.11.1")
    # If not specified, uses default stable version
    chart_version = optional(string)

    # Deploy the controller with an internal load balancer
    # Default (false) produces an external LB
    internal = optional(bool, false)

    # GKE-specific configuration
    gke = optional(object({
      # Name of an existing reserved static IP address (global or regional)
      static_ip_name = optional(string)

      # Sub-network self-link to use when internal = true
      subnetwork_self_link = optional(string)
    }))

    # EKS-specific configuration
    eks = optional(object({
      # Security group IDs to attach to the load balancer
      additional_security_group_ids = optional(list(string))

      # Subnet IDs where the ELB/NLB should be placed
      subnet_ids = optional(list(string))

      # Optional existing IAM role ARN for IRSA
      irsa_role_arn_override = optional(string)
    }))

    # AKS-specific configuration
    aks = optional(object({
      # Client ID of a user-assigned managed identity
      managed_identity_client_id = optional(string)

      # Name of a pre-existing public IP resource to reuse
      public_ip_name = optional(string)
    }))
  })
}