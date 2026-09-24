# The topic. project, location, cluster, topic_id, and replication_factor
# are immutable (a change replaces the topic and deletes its messages);
# partition_count grows in place, and configs update in place. The
# provider waits five seconds after create for the topic to settle.
resource "google_managed_kafka_topic" "this" {
  project            = local.project_id
  location           = var.spec.location
  cluster            = local.cluster
  topic_id           = local.topic_id
  replication_factor = var.spec.replication_factor
  partition_count    = local.partition_count
  configs            = local.configs

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy
}
