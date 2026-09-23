output "name" {
  description = "Full resource name of the request (projects/{project}/locations/{zone}/queuedResources/{queued_resource_id})"
  value       = google_tpu_v2_queued_resource.this.id
}

output "queued_resource_id" {
  description = "The request's id"
  value       = google_tpu_v2_queued_resource.this.name
}

output "zone" {
  description = "The zone the capacity is requested in"
  value       = google_tpu_v2_queued_resource.this.zone
}
