output "name" {
  description = "Full resource name of the singleton configuration (projects/{project}/locations/{location}/ragEngineConfig)"
  value       = google_vertex_ai_rag_engine_config.this.name
}

output "location" {
  description = "The location the configuration governs"
  value       = google_vertex_ai_rag_engine_config.this.region
}
