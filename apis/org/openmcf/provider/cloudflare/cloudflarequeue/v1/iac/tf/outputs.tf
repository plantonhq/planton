output "queue_id" {
  description = "The Cloudflare-assigned ID of the queue (referenced by a consumer and by event-notification producers)"
  value       = cloudflare_queue.main.queue_id
}

output "queue_name" {
  description = "The queue name (referenced by a Worker producer binding and by R2 event notifications)"
  value       = cloudflare_queue.main.queue_name
}

output "created_on" {
  description = "RFC3339 timestamp of when the queue was created"
  value       = cloudflare_queue.main.created_on
}

output "modified_on" {
  description = "RFC3339 timestamp of when the queue was last modified"
  value       = cloudflare_queue.main.modified_on
}
