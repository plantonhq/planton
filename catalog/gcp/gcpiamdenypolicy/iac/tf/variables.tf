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
  description = "GcpIamDenyPolicy specification"
  type = object({
    # Where the policy attaches. Omit entirely to attach to the
    # provider's default project.
    parent = optional(object({
      # Attach to a project — a literal project ID or a reference to a
      # GcpProject resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      project_id = optional(string, "")

      # Attach to a folder: the folder ID (numeric, with or without the
      # "folders/" prefix).
      folder_id = optional(string, "")

      # Attach to an organization: the numeric organization ID.
      organization_id = optional(string, "")
    }))

    # The policy's resource ID (the last segment of its name). Defaults
    # to metadata.name when left empty. Immutable: changing it destroys
    # and recreates the policy.
    policy_name = optional(string, "")

    # Human-readable name shown in consoles.
    display_name = optional(string, "")

    # The deny rules — each names the principals denied, the permissions
    # they are denied, any exceptions, and an optional condition.
    rules = list(object({
      # What this rule guards and why — for the operator auditing the
      # policy later.
      description = optional(string, "")

      # The rule body.
      deny_rule = object({
        # The identities denied, in the v2 principal formats:
        #   principalSet://goog/public:all                       -- everyone
        #   principal://goog/subject/{email}                     -- one user
        #   principalSet://goog/group/{group-email}              -- a group
        #   principal://iam.googleapis.com/projects/-/serviceAccounts/{email}
        #                                                        -- a service
        #                                                           account
        #   principalSet://cloudresourcemanager.googleapis.com/organizations/{org-id}
        #                                                        -- everyone in
        #                                                           the org
        denied_principals = optional(list(string), [])

        # Identities EXCLUDED from the rule even when denied_principals
        # covers them — the break-glass carve-out (e.g. deny a group but
        # exempt the on-call account).
        exception_principals = optional(list(string), [])

        # The permissions denied, as {service-fqdn}/{resource}.{verb} — e.g.
        # "secretmanager.googleapis.com/versions.access",
        # "iam.googleapis.com/roles.delete". Only permissions the deny API
        # supports may be listed (see Google's supported-permissions list).
        denied_permissions = optional(list(string), [])

        # Permissions EXCLUDED from denied_permissions — a permission
        # appearing in both lists is NOT denied.
        exception_permissions = optional(list(string), [])

        # Optional CEL condition scoping when the denial applies, evaluated
        # on resource tags (e.g.
        # !resource.matchTag('12345678/env', 'production') denies everywhere
        # EXCEPT tagged production resources).
        denial_condition = optional(object({
          # The CEL expression, e.g.
          # resource.matchTag('12345678/env', 'sandbox').
          expression = string

          # Short title identifying the condition's purpose.
          title = optional(string, "")

          # What the condition scopes and why.
          description = optional(string, "")

          # Where the expression came from, for error attribution in UIs (e.g.
          # a file/line marker). Rarely set by hand.
          location = optional(string, "")
        }))
      })
    }))

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the policy is deleted; the permissions it denied
    #                become usable again wherever roles allow them
    #   "PREVENT" -- destroy FAILS; protects a guardrail whose silent
    #                removal would re-open the surface it guards
    #   "ABANDON" -- the policy is removed from management but keeps
    #                denying in GCP
    deletion_policy = optional(string, "")
  })
}
