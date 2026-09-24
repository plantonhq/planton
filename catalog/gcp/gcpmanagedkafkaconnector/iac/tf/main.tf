# The connector. project, location, connect_cluster, and connector_id are
# immutable; configs and the restart policy update in place.
resource "google_managed_kafka_connector" "this" {
  project         = local.project_id
  location        = var.spec.location
  connect_cluster = local.connect_cluster
  connector_id    = local.connector_id
  configs         = local.configs

  # Omitted, failed tasks stay failed; declared, each backoff is sent only
  # when set so Google's default fills the other.
  dynamic "task_restart_policy" {
    for_each = var.spec.task_restart_policy != null ? [var.spec.task_restart_policy] : []
    content {
      minimum_backoff = task_restart_policy.value.minimum_backoff != "" ? task_restart_policy.value.minimum_backoff : null
      maximum_backoff = task_restart_policy.value.maximum_backoff != "" ? task_restart_policy.value.maximum_backoff : null
    }
  }

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy
}
