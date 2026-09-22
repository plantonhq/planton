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
  description = "GcpCloudIdentityGroup specification"
  type = object({
    # The group's email address, in a domain the customer owns
    # (e.g. platform-admins@example.com). This is the group's identity in
    # IAM (`group:platform-admins@example.com`). Required. Immutable.
    group_email = string

    # The Cloud Identity customer the group is created under, as
    # `customers/{id}` (the customer ID from the Google Admin console or
    # `gcloud organizations list`). Required. Immutable.
    customer_id = string

    # The namespace of an identity-mapped (non-Google) group, in the form
    # `identitysources/{id}`. Unset means a Google Group. Immutable.
    group_namespace = optional(string, "")

    # Name shown in the Admin console and Groups UI. Defaults to
    # metadata.name.
    display_name = optional(string, "")

    # Free-text description of the group's purpose (up to 4096 characters).
    description = optional(string, "")

    # Make this a security group: adds the
    # cloudidentity.googleapis.com/groups.security label on top of the
    # discussion-forum label every Google Group carries. Security groups are
    # for access control only (no mailing-list features) and the label
    # cannot be removed once added. Immutable in practice.
    security = optional(bool, false)

    # Who the group starts with: EMPTY (the default; the members below are
    # the whole membership), WITH_INITIAL_OWNER (the caller becomes an owner
    # so the group is never orphaned). Immutable.
    initial_group_config = optional(string)

    # The group's members. Each is one membership resource keyed by the
    # member's email; a member removed here is removed from the group, one
    # added is added. Roles change in place.
    memberships = optional(list(object({
      # The member's email: a Google user, a Google Group, or a service
      # account -- a GcpServiceAccount reference or a literal address.
      # Immutable: a different member is a different membership.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      member = string

      # The namespace of an identity-mapped (non-Google) member, in the form
      # `identitysources/{id}`. Unset means a Google-managed identity.
      # Immutable.
      member_namespace = optional(string, "")

      # The roles the member holds. Unset means MEMBER only; listed roles must
      # include MEMBER. Roles change in place.
      roles = optional(list(object({
        # The role: MEMBER (belongs to the group), MANAGER (manages members),
        # OWNER (full control, including deleting the group).
        name = string

        # When the MEMBER role expires and the membership is removed, as an RFC
        # 3339 timestamp (e.g. 2027-01-01T00:00:00Z). Unset never expires.
        expire_time = optional(string, "")
      })), [])

      # Skip creating the membership when one for this member already exists
      # (adopt it instead of failing). Unset means fail on a duplicate.
      create_ignore_already_exists = optional(bool, false)
    })), [])

    # What destroy does to the group:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the group and every membership are deleted; IAM
    #                bindings that named the group stop granting anything
    #   "PREVENT" -- destroy FAILS; protects a group IAM policies depend on
    #   "ABANDON" -- the group leaves management but keeps existing
    deletion_policy = optional(string, "")
  })
}
