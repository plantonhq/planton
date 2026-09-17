variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for the additive key-scoped IAM grant"
  type = object({
    # The crypto key whose IAM policy receives this grant, as its
    # fully-qualified resource path
    # (projects/<project>/locations/<location>/keyRings/<ring>/cryptoKeys/<key>).
    # The CLI's tfvars converter resolves StringValueOrRef fields to their
    # literal string before the module runs, so these arrive as plain strings.
    # There is no separate project or location input: both are embedded in
    # the key path.
    crypto_key_id = string

    # The role to grant on the key: a predefined role (typically
    # roles/cloudkms.cryptoKeyEncrypterDecrypter for CMEK consumers) or a
    # custom role's fully-qualified name ("projects/<project>/roles/<role_id>").
    role = string

    # The identity receiving the grant, in GCP IAM member format
    # (serviceAccount:<email> — most commonly a Google service agent for CMEK —
    # user:<email>, group:<email>, domain:<domain>, principal://...,
    # principalSet://...).
    member = string

    # Optional IAM Condition restricting when the grant applies. The condition
    # is part of the grant's identity — same role with and without a condition
    # are two independent grants.
    condition = optional(object({
      title       = string
      expression  = string
      description = optional(string)
    }))
  })

  validation {
    # The fully-qualified key path a GcpKmsKey reference resolves to, plus the
    # provider's two shorthand forms (<project>/<location>/<ring>/<key> and
    # <location>/<ring>/<key> riding the provider's default project).
    condition     = can(regex("^(projects/[^/]+/locations/[^/]+/keyRings/[^/]+/cryptoKeys/[^/]+|[^/]+/[^/]+/[^/]+/[^/]+|[^/]+/[^/]+/[^/]+)$", var.spec.crypto_key_id))
    error_message = "crypto_key_id must be a crypto key identifier (projects/<project>/locations/<location>/keyRings/<ring>/cryptoKeys/<key>)."
  }

  validation {
    condition     = var.spec.role != ""
    error_message = "role is required."
  }

  validation {
    # Mirrors the IAM member format: either <type>:<value> or one of the bare
    # public literals. Deleted principals cannot be granted to.
    condition = (
      var.spec.member != "" &&
      !startswith(var.spec.member, "deleted:") &&
      (can(regex("^.+:.+$", var.spec.member)) || contains(["allUsers", "allAuthenticatedUsers"], var.spec.member))
    )
    error_message = "member must be in IAM member format (e.g. serviceAccount:<email>, user:<email>, group:<email>) or one of allUsers / allAuthenticatedUsers; grants to deleted principals are not supported."
  }
}
