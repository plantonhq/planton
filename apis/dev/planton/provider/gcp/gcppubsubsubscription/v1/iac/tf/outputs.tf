output "subscription_id" {
  description = "The fully qualified subscription ID (projects/{project}/subscriptions/{name})"
  value       = google_pubsub_subscription.this.id
}

output "subscription_name" {
  description = "The short subscription name"
  value       = google_pubsub_subscription.this.name
}
