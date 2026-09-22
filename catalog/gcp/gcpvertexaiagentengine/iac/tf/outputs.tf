# Google's `name` attribute is the numeric ID; the resource id is the full
# path the Vertex AI SDK and a :query call take, so both are exported.
output "name" {
  description = "Full resource name of the agent (projects/{project}/locations/{location}/reasoningEngines/{id})"
  value       = google_vertex_ai_reasoning_engine.this.id
}

output "reasoning_engine_id" {
  description = "The numeric ID Vertex AI assigned to the agent"
  value       = google_vertex_ai_reasoning_engine.this.name
}

output "location" {
  description = "The location the agent runs in"
  value       = google_vertex_ai_reasoning_engine.this.region
}

output "create_time" {
  description = "RFC 3339 timestamp of the agent's creation"
  value       = google_vertex_ai_reasoning_engine.this.create_time
}

output "update_time" {
  description = "RFC 3339 timestamp of the agent's last update"
  value       = google_vertex_ai_reasoning_engine.this.update_time
}
