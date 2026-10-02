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
  description = "GcpServiceAccountIamMember specification"
  type = object({
    # The service account whose IAM policy receives this grant, as its
    # fully-qualified resource name:
    #   projects/<project>/serviceAccounts/<email>
    # Reference a GcpServiceAccount resource — its `name` output is exactly
    # this value. There is no separate project field: the account's project is
    # embedded in the resource name.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    service_account_id = string

    # The role to grant on the service account. Either a predefined role
    # (typically one of the service-account usage roles listed above) or a
    # custom role's fully-qualified name
    # ("projects/<project>/roles/<role_id>"). Reference a GcpIamCustomRole
    # resource to grant a custom role — its `name` output is exactly this value.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    role = string

    # The identity receiving the grant, in GCP IAM member format:
    #   serviceAccount:<email>       — another service account (reference a
    #                                  GcpServiceAccount resource — its `member`
    #                                  output is exactly this value)
    #   principal:// / principalSet:// — workload identity federation principals;
    #                                  the workloadIdentityUser + principalSet
    #                                  pair is how an external identity (e.g. a
    #                                  GitHub repository) gains impersonation
    #   user:<email>                 — a Google account
    #   group:<email>                — a Google group
    #   domain:<domain>              — everyone in a Workspace/Cloud Identity domain
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
