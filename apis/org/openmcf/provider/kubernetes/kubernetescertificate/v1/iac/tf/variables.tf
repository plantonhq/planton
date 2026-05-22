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
  description = "Specification for KubernetesCertificate"
  type = object({
    # Namespace where the Certificate resource will be created
    namespace = string

    # DNS Subject Alternative Names (at least one required)
    dns_names = list(string)

    # Kubernetes Secret name for the signed certificate and private key
    secret_name = string

    # Issuer reference -- exactly one of cluster_issuer or issuer must be set.
    # The oneof maps to two optional objects; middleware validation ensures
    # exactly one is populated.
    issuer_ref = object({
      cluster_issuer = optional(object({
        name = string
      }))
      issuer = optional(object({
        name = string
      }))
    })

    # When true, the issued certificate is a CA certificate
    is_ca = optional(bool, false)

    # Certificate lifetime and renewal timing (optional)
    duration_config = optional(object({
      duration     = optional(string, "2160h")
      renew_before = optional(string, "360h")
    }))

    # Private key configuration (optional).
    # Values should already be in cert-manager CRD format (RSA, PKCS1, Always)
    # since middleware populates defaults from proto options.
    private_key = optional(object({
      algorithm       = optional(string, "RSA")
      size            = optional(number, 2048)
      encoding        = optional(string, "PKCS1")
      rotation_policy = optional(string, "Always")
    }))
  })
}
