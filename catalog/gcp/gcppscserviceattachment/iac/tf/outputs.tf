# The composition handle: a consumer's PSC endpoint (a regional
# GcpGlobalForwardingRule with an empty scheme) names this as its target.
output "self_link" {
  description = "Self-link URI of the service attachment (the consumer forwarding rule's target)"
  value       = google_compute_service_attachment.this.self_link
}

output "attachment_name" {
  description = "Name of the service attachment as it exists in GCP"
  value       = google_compute_service_attachment.this.name
}

output "region" {
  description = "Region of the service attachment"
  value       = var.spec.region
}

output "fingerprint" {
  description = "Server-computed fingerprint of the attachment for concurrency control"
  value       = google_compute_service_attachment.this.fingerprint
}

output "connected_endpoints_count" {
  description = "Number of consumer endpoints connected to the attachment at provisioning time"
  value       = tostring(length(google_compute_service_attachment.this.connected_endpoints))
}
