# The grant, resolved (after reference resolution) -- echoed so audits see
# exactly what was applied without re-resolving references.
output "topic" {
  description = "The topic whose IAM policy received the grant (projects/<project>/topics/<topic>)"
  value       = google_pubsub_topic_iam_member.this.topic
}

output "role" {
  description = "The role that was granted"
  value       = google_pubsub_topic_iam_member.this.role
}

output "member" {
  description = "The member the role was granted to, in IAM member format"
  value       = google_pubsub_topic_iam_member.this.member
}

output "etag" {
  description = "The etag of the topic's IAM policy after the grant"
  value       = google_pubsub_topic_iam_member.this.etag
}
