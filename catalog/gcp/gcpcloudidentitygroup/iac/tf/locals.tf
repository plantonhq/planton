locals {
  # The console name defaults to metadata.name when the spec leaves
  # display_name empty -- the same naming basis every kind uses.
  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  # Every Google Group carries the discussion-forum label (the API requires
  # it and accepts only an empty value); a security group adds the security
  # label on top. The map is the module's, derived from spec.security --
  # never a user-facing map, because Google accepts nothing else here.
  group_labels = merge(
    { "cloudidentity.googleapis.com/groups.discussion_forum" = "" },
    var.spec.security ? { "cloudidentity.googleapis.com/groups.security" = "" } : {}
  )

  # initial_group_config carries a proto default (EMPTY) the manifest loader
  # applies before either engine runs; the coalesce is the same rule for a
  # tfvars file written by hand.
  initial_group_config = coalesce(var.spec.initial_group_config, "EMPTY")

  # Memberships keyed by the member's email (flattened from the reference),
  # so a member removed from the list is destroyed by key and the others are
  # left alone. Unset roles mean MEMBER only.
  memberships = {
    for m in var.spec.memberships : m.member => {
      member_namespace             = m.member_namespace != "" ? m.member_namespace : null
      roles                        = length(m.roles) > 0 ? m.roles : [{ name = "MEMBER", expire_time = "" }]
      create_ignore_already_exists = m.create_ignore_already_exists ? true : null
    }
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
