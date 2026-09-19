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
  description = "GcpOrgPolicyCustomConstraint specification"
  type = object({
    # The organization that owns the constraint: the numeric organization
    # ID (from `gcloud organizations list`), without the `organizations/`
    # prefix. Immutable -- a constraint cannot move between organizations.
    organization_id = string

    # The constraint's name WITHOUT the `custom.` prefix -- the module adds
    # it, so Google knows the constraint as `custom.<constraint_name>` and
    # that full handle is the `constraint` output a GcpOrgPolicy references.
    # Defaults to metadata.name when empty. Google's rule: starts with a
    # letter, then letters and digits, at most 62 characters in all;
    # unique within the organization. Immutable: a rename is a delete and a
    # create, and every policy enforcing the old name would lapse.
    constraint_name = optional(string, "")

    # A human-friendly name for the constraint, shown in the console's
    # organization-policy pages. Mutable.
    display_name = optional(string, "")

    # What Google shows a user whose request the constraint blocks -- the
    # violation message. Write it as the instruction the blocked engineer
    # needs ("Node pools must enable auto-upgrade; set
    # management.autoUpgrade to true"). Mutable.
    description = optional(string, "")

    # The Google Cloud REST resource types the condition is evaluated
    # against, fully qualified: `container.googleapis.com/NodePool`,
    # `compute.googleapis.com/Instance`, `sqladmin.googleapis.com/Instance`,
    # `storage.googleapis.com/Bucket`. At least one; every type in one
    # constraint must belong to the same service. Immutable: changing the
    # list recreates the constraint.
    resource_types = list(string)

    # The operations the constraint is evaluated on. CREATE and UPDATE are
    # supported by every service that supports custom constraints; DELETE,
    # REMOVE_GRANT, and GOVERN_TAGS are supported by a few (the supported
    # services list says which). Google rejects a method a service does not
    # support at apply time. Mutable.
    method_types = list(string)

    # The Common Expression Language test over the resource, evaluated on
    # each method in method_types: `resource.management.autoUpgrade == false`,
    # `resource.settings.ipConfiguration.ipv4Enabled == true`,
    # `!has(resource.shieldedInstanceConfig) || resource.shieldedInstanceConfig.enableSecureBoot == false`.
    # The fields are the service's own REST resource fields; Google's per-
    # service pages list the ones the constraint engine can see. Together
    # with action_type it reads: "when this is true, ALLOW or DENY the
    # request". Mutable.
    condition = string

    # What happens when the condition is true for a request: DENY blocks it
    # (the usual guardrail -- the condition describes the forbidden shape),
    # ALLOW permits it and implicitly denies everything the condition does
    # not match (an allow-list -- the condition describes the only
    # acceptable shape). Mutable.
    action_type = string

    # What destroying this resource does to the constraint in GCP:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the constraint is deleted; every GcpOrgPolicy still
    #                enforcing it starts failing to apply, so destroy the
    #                policies first (a chart's dependency order does this
    #                when the policies reference this resource)
    #   "PREVENT" -- destroy FAILS; the guard for a rule many policies rely on
    #   "ABANDON" -- the constraint is removed from management but keeps
    #                existing in GCP, still enforceable by policies
    deletion_policy = optional(string, "")
  })
}
