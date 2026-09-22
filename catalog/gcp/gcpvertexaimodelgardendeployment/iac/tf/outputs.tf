# The provider's `endpoint` is the endpoint's numeric ID segment; the full
# resource path is rebuilt the way Google names it so the output reads the
# same as GcpVertexAiEndpoint's endpoint_id.
output "endpoint_id" {
  description = "Fully qualified endpoint resource path (projects/{project}/locations/{location}/endpoints/{endpoint_name})"
  value       = "projects/${google_vertex_ai_endpoint_with_model_garden_deployment.this.project}/locations/${google_vertex_ai_endpoint_with_model_garden_deployment.this.location}/endpoints/${google_vertex_ai_endpoint_with_model_garden_deployment.this.endpoint}"
}

output "endpoint_name" {
  description = "The numeric endpoint ID Vertex AI assigned"
  value       = google_vertex_ai_endpoint_with_model_garden_deployment.this.endpoint
}

output "deployed_model_id" {
  description = "The numeric ID Vertex AI assigned to the deployed model"
  value       = google_vertex_ai_endpoint_with_model_garden_deployment.this.deployed_model_id
}

output "deployed_model_display_name" {
  description = "The display name Vertex AI gave the deployed model"
  value       = google_vertex_ai_endpoint_with_model_garden_deployment.this.deployed_model_display_name
}

output "location" {
  description = "The location the model is deployed in"
  value       = google_vertex_ai_endpoint_with_model_garden_deployment.this.location
}
