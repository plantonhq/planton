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
  description = "Specification for KubernetesIssuer"
  type = object({
    # Namespace where the Issuer will be created (must already exist)
    namespace = string

    # CA issuer -- signs using a CA keypair from a Kubernetes Secret
    ca = optional(object({
      ca_secret_name = string
    }))

    # Self-signed issuer -- no external CA needed
    self_signed = optional(object({}))
  })

  validation {
    condition     = (var.spec.ca != null) != (var.spec.self_signed != null)
    error_message = "Exactly one of 'ca' or 'self_signed' must be set."
  }
}
