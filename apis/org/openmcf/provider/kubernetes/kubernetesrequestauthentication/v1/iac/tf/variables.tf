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
  description = "Specification for KubernetesRequestAuthentication"
  type = object({
    # Namespace the RequestAuthentication is created in (resolved foreign key).
    namespace = string

    # Workload selector. When omitted (and target_refs is also omitted), the
    # policy applies to all workloads in the namespace. Matched by label at
    # runtime by istiod; not an OpenMCF foreign key. Mutually exclusive with
    # target_refs.
    selector = optional(object({
      match_labels = optional(map(string))
    }))

    # Attaches the policy to specific resources (Gateway, Service, ServiceEntry)
    # instead of selecting workloads by label. Mutually exclusive with selector.
    # Plain cross-resource references, not OpenMCF foreign keys.
    target_refs = optional(list(object({
      group     = optional(string)
      kind      = string
      name      = string
      namespace = optional(string)
    })))

    # JWT validation rules. When empty, the policy installs no JWT requirements.
    jwt_rules = optional(list(object({
      issuer    = string
      audiences = optional(list(string))
      jwks_uri  = optional(string)
      jwks      = optional(string)
      from_headers = optional(list(object({
        name   = string
        prefix = optional(string)
      })))
      from_params              = optional(list(string))
      from_cookies             = optional(list(string))
      output_payload_to_header = optional(string)
      forward_original_token   = optional(bool)
      output_claim_to_headers = optional(list(object({
        header = string
        claim  = string
      })))
      timeout = optional(string)
    })))
  })
}
