output "endpoint_id" {
  description = "Fully qualified endpoint resource path (projects/{project}/locations/{location}/endpoints/{name})"
  value       = google_vertex_ai_endpoint.this.id
}

output "display_name" {
  description = "Display name of the endpoint"
  value       = google_vertex_ai_endpoint.this.display_name
}

output "dedicated_endpoint_dns" {
  description = "DNS of the dedicated endpoint (populated only when dedicated_endpoint_enabled is true)"
  value       = google_vertex_ai_endpoint.this.dedicated_endpoint_dns
}

output "create_time" {
  description = "RFC3339 timestamp of when the endpoint was created"
  value       = google_vertex_ai_endpoint.this.create_time
}
