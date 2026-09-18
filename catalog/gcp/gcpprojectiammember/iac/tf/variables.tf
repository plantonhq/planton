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
  description = "GcpProjectIamMember specification"
  type = object({
    # The GCP project whose IAM policy receives this grant.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The role to grant. Either a predefined role ("roles/storage.objectViewer")
    # or a custom role's fully-qualified name ("projects/<project>/roles/<role_id>").
    # Reference a GcpIamCustomRole resource to grant a custom role — its `name`
    # output is exactly this value.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    role = string

    # The identity receiving the grant, in GCP IAM member format:
    #   serviceAccount:<email>       — a service account (the most common in IaC;
    #                                  reference a GcpServiceAccount resource —
    #                                  its `member` output is exactly this value)
    #   user:<email>                 — a Google account
    #   group:<email>                — a Google group
    #   domain:<domain>              — everyone in a Workspace/Cloud Identity domain
    #   principal:// / principalSet:// — workload identity federation principals
    #   allUsers / allAuthenticatedUsers — public access (grant with extreme care)
    # Grants to deleted principals ("deleted:...") are not supported.
    # Format validation happens at deploy time in the modules rather than here,
    # because the value usually arrives through a reference that is only
    # resolved at deploy time.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    member = string

    # Optional IAM Condition restricting when this grant applies (e.g. only on
    # resources with a given prefix, or before an expiry date). The condition is
    # part of the grant's identity: the same role with and without a condition
    # are two independent grants that do not interfere.
    condition = optional(object({
      # Short human-readable title identifying the condition's intent,
      # e.g. "expires-2026-12-31" or "prod-buckets-only".
      title = string

      # The CEL condition expression, e.g.
      # request.time < timestamp("2027-01-01T00:00:00Z") or
      # resource.name.startsWith("projects/_/buckets/prod-").
      expression = string

      # Optional longer explanation of what the condition does and why it exists.
      description = optional(string, "")
    }))
  })
}
