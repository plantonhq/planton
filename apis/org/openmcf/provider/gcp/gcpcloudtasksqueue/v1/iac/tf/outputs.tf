output "queue_id" {
  description = "The fully qualified queue ID (projects/{project}/locations/{location}/queues/{name})"
  value       = google_cloud_tasks_queue.this.id
}

output "queue_name" {
  description = "The short queue name"
  value       = google_cloud_tasks_queue.this.name
}

# Note: The 'state' attribute is not exported by the Terraform Google provider
# for google_cloud_tasks_queue. It is available via the Pulumi module.
# To check queue state in Terraform, use the GCP API or gcloud CLI.
