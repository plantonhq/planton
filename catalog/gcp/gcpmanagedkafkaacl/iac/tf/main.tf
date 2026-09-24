# The ACL for one resource pattern. project, location, cluster, and acl_id
# (the pattern) are immutable; the entries update in place.
resource "google_managed_kafka_acl" "this" {
  project  = local.project_id
  location = var.spec.location
  cluster  = local.cluster
  acl_id   = var.spec.acl_id

  # permission_type and host are sent only when set: the provider defaults
  # them to ALLOW and "*", the only host Google accepts.
  dynamic "acl_entries" {
    for_each = var.spec.acl_entries
    content {
      principal       = acl_entries.value.principal
      operation       = acl_entries.value.operation
      permission_type = acl_entries.value.permission_type != "" ? acl_entries.value.permission_type : null
      host            = acl_entries.value.host != "" ? acl_entries.value.host : null
    }
  }

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy
}
