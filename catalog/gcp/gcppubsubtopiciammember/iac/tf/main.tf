# A single ADDITIVE IAM grant ON a Pub/Sub topic: one role, to one member, on
# one topic. Additive means the provider merges this (role, member) pair into
# the topic's IAM policy without touching any other member's bindings on the
# same role, and destroy subtracts only this exact pair.
#
# The grant is its own resource, not a field on the topic, because the
# identities that most need publish rights (a logging sink's writer, a
# Security Command Center export's publisher) belong to resources that name
# the topic themselves; depending on both, it always lands last.
#
# Every argument is immutable (ForceNew): an IAM grant has no update -- any
# change replaces it, which is also how the API behaves.
resource "google_pubsub_topic_iam_member" "this" {
  # The topic arrives as its full name (projects/<project>/topics/<topic>);
  # the provider reads the project from it, so no project argument is set.
  topic  = var.spec.topic
  role   = var.spec.role
  member = var.spec.member

  # No condition block: Pub/Sub topics do not accept conditional role
  # bindings (the provider never requests the version-3 policy conditions
  # need).
}
