output "queue_id" {
  description = "OCID of the queue"
  value       = oci_queue_queue.this.id
}

output "messages_endpoint" {
  description = "Endpoint URL for consuming or publishing messages"
  value       = oci_queue_queue.this.messages_endpoint
}
