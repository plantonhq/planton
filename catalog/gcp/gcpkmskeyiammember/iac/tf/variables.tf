variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "GcpKmsKeyIamMember specification"
  type = object({
    # The crypto key whose IAM policy receives this grant, as its
    # fully-qualified resource path:
    #   projects/<project>/locations/<location>/keyRings/<ring>/cryptoKeys/<key>
    # Reference a GcpKmsKey resource — its `key_id` output is exactly this
    # value. There is no separate project or location field: both are embedded
    # in the key path.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    crypto_key_id = string

    # The role to grant on the key. Either a predefined role (typically
    # "roles/cloudkms.cryptoKeyEncrypterDecrypter" for CMEK consumers) or a
    # custom role's fully-qualified name
    # ("projects/<project>/roles/<role_id>"). Reference a GcpIamCustomRole
    # resource to grant a custom role — its `name` output is exactly this value.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    role = string

    # The identity receiving the grant, in GCP IAM member format:
    #   serviceAccount:<email>       — a service account or Google service
    #                                  agent (the most common for CMEK — the
    #                                  consuming service's agent, e.g.
    #                                  service-<project_number>@gs-project-accounts.iam.gserviceaccount.com;
    #                                  for your own accounts reference a
    #                                  GcpServiceAccount resource — its `member`
    #                                  output is exactly this value)
    #   user:<email>                 — a Google account
    #   group:<email>                — a Google group
    #   domain:<domain>              — everyone in a Workspace/Cloud Identity domain
    #   principal:// / principalSet:// — workload identity federation principals
    # Grants to deleted principals ("deleted:...") are not supported.
    # Format validation happens at deploy time in the modules rather than here,
    # because the value usually arrives through a reference that is only
    # resolved at deploy time.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    member = string

    # Optional IAM Condition restricting when this grant applies (e.g. only
    # before an expiry date). The condition is part of the grant's identity:
    # the same role with and without a condition are two independent grants
    # that do not interfere.
    condition = optional(object({
      # Short human-readable title identifying the condition's intent,
      # e.g. "expires-2026-12-31".
      title = string

      # The CEL condition expression, e.g.
      # request.time < timestamp("2027-01-01T00:00:00Z").
      expression = string

      # Optional longer explanation of what the condition does and why it exists.
      description = optional(string, "")
    }))
  })
}
