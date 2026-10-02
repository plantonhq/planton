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
  description = "GcpOrgPolicy specification"
  type = object({
    # Where the policy applies: a project, a folder, or the organization.
    # At most one arm; all empty means the provider's default project (the
    # project the deploying credentials are configured for). Everything
    # beneath the scope inherits the policy unless a lower policy overrides
    # it (`inherit_from_parent`) or resets it (`reset`). Immutable.
    scope = optional(object({
      # Project scope: a literal project ID (Google also accepts the project
      # number) or a reference to a GcpProject resource. The API stores the
      # policy under the project NUMBER and the provider treats the two forms
      # as equal, so a literal ID never shows a spurious change.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      project_id = optional(string, "")

      # Folder scope: the folder's numeric ID -- a literal, or a reference to
      # a GcpFolder resource (its folder_id output). Every project and folder
      # beneath it inherits the policy.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      folder_id = optional(string, "")

      # Organization scope: the numeric organization ID, without the
      # `organizations/` prefix. The root of inheritance for the whole estate.
      organization_id = optional(string, "")
    }))

    # A predefined or managed constraint by name, exactly as Google lists it
    # (`gcloud org-policies list-constraints`): `compute.disableSerialPortAccess`,
    # `gcp.resourceLocations`, `iam.managed.disableServiceAccountKeyCreation`,
    # `storage.publicAccessPrevention`. Immutable. Exactly one of this and
    # custom_constraint.
    constraint = optional(string, "")

    # The organization's own constraint to enforce: a reference to a
    # GcpOrgPolicyCustomConstraint resource (its `constraint` output, the
    # `custom.<name>` handle), or that handle as a literal. Immutable.
    # Exactly one of this and constraint. A custom constraint is always
    # organization-defined, but the policy that enforces it may sit at any
    # scope -- this is how one rule is written once and applied to three
    # folders by three small policies.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    custom_constraint = optional(string, "")

    # The rules Google ENFORCES at this scope. Omit to set no live rules --
    # meaningful only together with dry_run_policy, the audit-first rollout:
    # first a dry-run policy alone, read the violations it would have caused
    # in the audit log, then add the same rules here.
    policy = optional(object({
      # For LIST constraints only: whether the values allowed or denied by
      # policies higher in the hierarchy stay in effect here alongside this
      # policy's rules (true), or this policy becomes the new root of
      # evaluation and nothing above it applies (false, the default). Google
      # ignores it on boolean constraints. Cannot be true together with reset.
      inherit_from_parent = optional(bool, false)

      # Discard every rule inherited from above and restore the constraint's
      # own default behavior at this scope (the enforcement Google applies
      # when nobody has set a policy). Mutually exclusive with rules and with
      # inherit_from_parent: a reset policy carries nothing else. Works for
      # both list and boolean constraints.
      reset = optional(bool, false)

      # The rules. For a BOOLEAN constraint there must be exactly one rule
      # without a condition, and every conditional rule must set `enforce` to
      # the opposite of that unconditional rule (Google's rule; violations are
      # rejected at apply time). For a LIST constraint the rules combine:
      # unconditional rules set the baseline, conditional rules refine it for
      # resources whose tags match. Google compares rules as a set, so their
      # order never causes a spurious change.
      rules = optional(list(object({
        # LIST constraints: every value is allowed at this scope (`true`).
        # Setting it to false is meaningful only as the unconditional baseline
        # that a conditional rule overrides. Both engines send Google's string
        # form ("TRUE"/"FALSE").
        allow_all = optional(bool)

        # LIST constraints: every value is denied at this scope (`true`).
        # Both engines send Google's string form ("TRUE"/"FALSE").
        deny_all = optional(bool)

        # BOOLEAN constraints: the constraint is enforced here (`true`) or
        # explicitly NOT enforced (`false`) -- the way a child scope relaxes
        # a guardrail its parent enforces. Both engines send Google's string
        # form ("TRUE"/"FALSE").
        enforce = optional(bool)

        # LIST constraints: the explicit allow and deny lists.
        values = optional(object({
          # Values permitted at this scope.
          allowed_values = optional(list(string), [])

          # Values forbidden at this scope.
          denied_values = optional(list(string), [])
        }))

        # Apply this rule only to resources whose Resource Manager tags match.
        # Omit for an unconditional rule. A boolean constraint needs exactly one
        # unconditional rule; conditional rules refine it.
        condition = optional(object({
          # A Common Expression Language expression of 1 to 10 tag tests joined by
          # `||` or `&&`. Each test is either
          # `resource.matchTag('<org_id>/<tag_key_short_name>', '<tag_value_short_name>')`
          # (by short names) or `resource.matchTagId('tagKeys/<id>', 'tagValues/<id>')`
          # (by the ids GcpTagKey and GcpTagValue output). Example:
          # `resource.matchTag('123456789012/environment', 'prod')`.
          expression = string

          # A short label for the condition, shown in the console and audit logs.
          title = optional(string, "")

          # A longer explanation of what the condition selects and why.
          description = optional(string, "")

          # Where the expression came from, for error reporting (a file name and
          # position, or any free text).
          location = optional(string, "")
        }))

        # For MANAGED constraints that declare parameters (the `*.managed.*`
        # family): the parameter values as one JSON object, typed as the
        # constraint defines them -- e.g.
        # `{"allowedLocations": ["us-east1", "us-west1"], "allowAll": true}`.
        # Google validates the object against the constraint at apply time;
        # the provider validates only that it parses as JSON.
        parameters = optional(string, "")
      })), [])
    }))

    # The same rule shape, evaluated in AUDIT mode: violations are logged
    # (`cloudaudit.googleapis.com/policy`, `dryRunPolicyViolation`), nothing
    # is blocked. Set it beside `policy` to preview a tightening before it
    # bites, or alone to measure a new guardrail against live traffic first.
    dry_run_policy = optional(object({
      # For LIST constraints only: whether the values allowed or denied by
      # policies higher in the hierarchy stay in effect here alongside this
      # policy's rules (true), or this policy becomes the new root of
      # evaluation and nothing above it applies (false, the default). Google
      # ignores it on boolean constraints. Cannot be true together with reset.
      inherit_from_parent = optional(bool, false)

      # Discard every rule inherited from above and restore the constraint's
      # own default behavior at this scope (the enforcement Google applies
      # when nobody has set a policy). Mutually exclusive with rules and with
      # inherit_from_parent: a reset policy carries nothing else. Works for
      # both list and boolean constraints.
      reset = optional(bool, false)

      # The rules. For a BOOLEAN constraint there must be exactly one rule
      # without a condition, and every conditional rule must set `enforce` to
      # the opposite of that unconditional rule (Google's rule; violations are
      # rejected at apply time). For a LIST constraint the rules combine:
      # unconditional rules set the baseline, conditional rules refine it for
      # resources whose tags match. Google compares rules as a set, so their
      # order never causes a spurious change.
      rules = optional(list(object({
        # LIST constraints: every value is allowed at this scope (`true`).
        # Setting it to false is meaningful only as the unconditional baseline
        # that a conditional rule overrides. Both engines send Google's string
        # form ("TRUE"/"FALSE").
        allow_all = optional(bool)

        # LIST constraints: every value is denied at this scope (`true`).
        # Both engines send Google's string form ("TRUE"/"FALSE").
        deny_all = optional(bool)

        # BOOLEAN constraints: the constraint is enforced here (`true`) or
        # explicitly NOT enforced (`false`) -- the way a child scope relaxes
        # a guardrail its parent enforces. Both engines send Google's string
        # form ("TRUE"/"FALSE").
        enforce = optional(bool)

        # LIST constraints: the explicit allow and deny lists.
        values = optional(object({
          # Values permitted at this scope.
          allowed_values = optional(list(string), [])

          # Values forbidden at this scope.
          denied_values = optional(list(string), [])
        }))

        # Apply this rule only to resources whose Resource Manager tags match.
        # Omit for an unconditional rule. A boolean constraint needs exactly one
        # unconditional rule; conditional rules refine it.
        condition = optional(object({
          # A Common Expression Language expression of 1 to 10 tag tests joined by
          # `||` or `&&`. Each test is either
          # `resource.matchTag('<org_id>/<tag_key_short_name>', '<tag_value_short_name>')`
          # (by short names) or `resource.matchTagId('tagKeys/<id>', 'tagValues/<id>')`
          # (by the ids GcpTagKey and GcpTagValue output). Example:
          # `resource.matchTag('123456789012/environment', 'prod')`.
          expression = string

          # A short label for the condition, shown in the console and audit logs.
          title = optional(string, "")

          # A longer explanation of what the condition selects and why.
          description = optional(string, "")

          # Where the expression came from, for error reporting (a file name and
          # position, or any free text).
          location = optional(string, "")
        }))

        # For MANAGED constraints that declare parameters (the `*.managed.*`
        # family): the parameter values as one JSON object, typed as the
        # constraint defines them -- e.g.
        # `{"allowedLocations": ["us-east1", "us-west1"], "allowAll": true}`.
        # Google validates the object against the constraint at apply time;
        # the provider validates only that it parses as JSON.
        parameters = optional(string, "")
      })), [])
    }))

    # What destroying this resource does to the policy in GCP:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the policy is deleted and the scope falls back to what
    #                it inherits from above (or the constraint's default)
    #   "PREVENT" -- destroy FAILS; the guard for the guardrails a landing
    #                zone depends on
    #   "ABANDON" -- the policy is removed from management but keeps being
    #                enforced in GCP
    deletion_policy = optional(string, "")
  })
}
