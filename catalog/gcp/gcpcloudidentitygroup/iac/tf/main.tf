# The Google Group: the identity IAM bindings name. The group key, parent,
# namespace, initial configuration, and labels are immutable; the display
# name and description change in place.
resource "google_cloud_identity_group" "this" {
  parent               = var.spec.customer_id
  display_name         = local.display_name
  description          = var.spec.description != "" ? var.spec.description : null
  initial_group_config = local.initial_group_config
  labels               = local.group_labels

  group_key {
    id        = var.spec.group_email
    namespace = var.spec.group_namespace != "" ? var.spec.group_namespace : null
  }

  # What destroy does to the group: DELETE (default), PREVENT (refuse), or
  # ABANDON (drop from state, keep the group).
  deletion_policy = local.deletion_policy
}

# One membership per member, keyed by email. A different member is a
# different membership (the key is immutable); roles change in place.
resource "google_cloud_identity_group_membership" "this" {
  for_each = local.memberships

  group = google_cloud_identity_group.this.id

  preferred_member_key {
    id        = each.key
    namespace = each.value.member_namespace
  }

  dynamic "roles" {
    for_each = each.value.roles
    content {
      name = roles.value.name

      dynamic "expiry_detail" {
        for_each = roles.value.expire_time != "" ? [roles.value.expire_time] : []
        content {
          expire_time = expiry_detail.value
        }
      }
    }
  }

  # Adopt an existing membership for this member instead of failing; sent
  # only when true (the provider default is false).
  create_ignore_already_exists = each.value.create_ignore_already_exists

  # The memberships share the group's fate on destroy.
  deletion_policy = local.deletion_policy
}
