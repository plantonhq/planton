output "topic_id" {
  description = "The fully qualified topic ID (projects/{project}/topics/{name})"
  value       = google_pubsub_topic.this.id
}

output "topic_name" {
  description = "The short topic name"
  value       = google_pubsub_topic.this.name
}
